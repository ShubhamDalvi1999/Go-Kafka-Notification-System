package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"

	"kafka-notify/pkg/models"

	"github.com/IBM/sarama"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

const (
	ConsumerGroup = "notifications-group"
	ConsumerTopic = "notifications"
	ConsumerPort  = ":8081"
)

func getKafkaBroker() string {
	brokers := os.Getenv("KAFKA_BROKERS")
	if brokers == "" {
		return "kafka:9092"
	}
	if strings.Contains(brokers, ",") {
		parts := strings.SplitN(brokers, ",", 2)
		return strings.TrimSpace(parts[0])
	}
	return strings.TrimSpace(brokers)
}

// ============== HELPER FUNCTIONS ==============
var ErrNoMessagesFound = errors.New("no messages found")

func getUserIDFromRequest(ctx *gin.Context) (string, error) {
	userID := ctx.Param("userID")
	if userID == "" {
		return "", ErrNoMessagesFound
	}
	return userID, nil
}

// Real-time WebSocket functionality removed

// ====== NOTIFICATION STORAGE ======
type UserNotifications map[string][]models.Notification

type NotificationStore struct {
	data UserNotifications
	mu   sync.RWMutex
}

func (ns *NotificationStore) Add(userID string,
	notification models.Notification) {
	ns.mu.Lock()
	defer ns.mu.Unlock()
	ns.data[userID] = append(ns.data[userID], notification)
}

func (ns *NotificationStore) Get(userID string) []models.Notification {
	ns.mu.RLock()
	defer ns.mu.RUnlock()
	return ns.data[userID]
}

// ============== KAFKA RELATED FUNCTIONS ==============
type Consumer struct {
	store *NotificationStore
}

func (*Consumer) Setup(sarama.ConsumerGroupSession) error   { return nil }
func (*Consumer) Cleanup(sarama.ConsumerGroupSession) error { return nil }

func (consumer *Consumer) ConsumeClaim(
	sess sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {
	for msg := range claim.Messages() {
		userID := string(msg.Key)
		var notification models.Notification
		err := json.Unmarshal(msg.Value, &notification)
		if err != nil {
			log.Printf("failed to unmarshal notification: %v", err)
			continue
		}
		consumer.store.Add(userID, notification)
		sess.MarkMessage(msg, "")
	}
	return nil
}

func initializeConsumerGroup() (sarama.ConsumerGroup, error) {
	config := sarama.NewConfig()

	broker := getKafkaBroker()
	consumerGroup, err := sarama.NewConsumerGroup(
		[]string{broker}, ConsumerGroup, config)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize consumer group: %w", err)
	}

	return consumerGroup, nil
}

func setupConsumerGroup(ctx context.Context, store *NotificationStore) {
	backoff := 5 * time.Second
	for {
		cg, err := initializeConsumerGroup()
		if err != nil {
			log.Printf("initialization error: %v", err)
			select {
			case <-time.After(backoff):
				continue
			case <-ctx.Done():
				return
			}
		}

		consumer := &Consumer{
			store: store,
		}

		for {
			err = cg.Consume(ctx, []string{ConsumerTopic}, consumer)
			if err != nil {
				log.Printf("error from consumer: %v", err)
				break
			}
			if ctx.Err() != nil {
				_ = cg.Close()
				return
			}
		}
		_ = cg.Close()
		select {
		case <-time.After(backoff):
			// retry
		case <-ctx.Done():
			return
		}
	}
}

func handleNotifications(ctx *gin.Context, store *NotificationStore) {
	userID, err := getUserIDFromRequest(ctx)
	if err != nil {
		ctx.JSON(http.StatusNotFound, gin.H{"message": err.Error()})
		return
	}

	notes := store.Get(userID)
	if len(notes) == 0 {
		ctx.JSON(http.StatusOK,
			gin.H{
				"message":       "No notifications found for user",
				"notifications": []models.Notification{},
			})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"notifications": notes})
}

// WebSocket handler removed

func main() {
	store := &NotificationStore{
		data: make(UserNotifications),
	}

	ctx, cancel := context.WithCancel(context.Background())
	go setupConsumerGroup(ctx, store)
	defer cancel()

	gin.SetMode(gin.ReleaseMode)
	router := gin.Default()

	// Add CORS middleware for HTTP routes only
	corsMiddleware := cors.New(cors.Config{
		AllowOrigins:     []string{"http://localhost:3000", "http://127.0.0.1:3000"},
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Accept", "Authorization"},
		AllowCredentials: true,
	})

	// HTTP API routes with CORS
	router.GET("/notifications/:userID", corsMiddleware, func(ctx *gin.Context) {
		handleNotifications(ctx, store)
	})

	// WebSocket route removed

	// Health check endpoint
	router.GET("/health", func(ctx *gin.Context) {
		ctx.JSON(http.StatusOK, gin.H{
			"status":             "healthy",
			"service":            "kafka-consumer",
			"timestamp":          time.Now().Format(time.RFC3339),
			"active_connections": 0,
		})
	})

	// WebSocket test endpoint removed

	fmt.Printf("Kafka CONSUMER (Group: %s) 👥📥 "+
		"started at http://localhost%s\n", ConsumerGroup, ConsumerPort)
	// WebSocket endpoint removed

	if err := router.Run(ConsumerPort); err != nil {
		log.Printf("failed to run the server: %v", err)
	}
}
