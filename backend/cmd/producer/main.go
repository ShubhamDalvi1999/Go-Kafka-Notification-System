package main

import (
	"context"
	"log"
	"time"

	"kafka-notify/internal/config"
	"kafka-notify/internal/database"
	"kafka-notify/internal/kafka"
	"kafka-notify/internal/server"
	"kafka-notify/internal/services"
	"kafka-notify/pkg/handlers"
	"kafka-notify/pkg/repository"
)

func main() {
	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	// Initialize database connection
	dbManager, err := database.NewConnectionManager(&cfg.Database)
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}
	defer dbManager.Close()

	// Initialize Kafka client manager
	kafkaManager := kafka.NewClientManager(&cfg.Kafka)

	// Create Kafka producer
	producer, err := kafkaManager.NewProducer()
	if err != nil {
		log.Fatalf("Failed to create Kafka producer: %v", err)
	}
	defer kafkaManager.CloseProducer(producer)

	// Initialize repository
	notificationRepo := repository.NewPostgresNotificationRepository(dbManager.GetDB())

	// Initialize notification service
	notificationService := services.NewNotificationService(notificationRepo, producer, cfg.Kafka.Topic)

	// Initialize HTTP handlers
	notificationHandlers := handlers.NewNotificationHandlers(notificationService)

	// Initialize HTTP server
	httpServer := server.NewServer(&cfg.Server)

	// Setup routes
	setupRoutes(httpServer, notificationHandlers)

	// Start outbox processor in background
	go startOutboxProcessor(notificationService)

	// Start HTTP server
	log.Printf("Starting producer service on port %s", cfg.Server.Port)
	if err := httpServer.Start(); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}

// setupRoutes configures the HTTP routes
func setupRoutes(server *server.Server, handlers *handlers.NotificationHandlers) {
	// Health check is already set up in the server

	// API routes
	api := server.AddGroup("/api/v1")

	// Notification routes
	api.POST("/notifications", handlers.CreateNotification)
	api.GET("/notifications/:userID", handlers.GetUserNotifications)
	api.PUT("/notifications/:id/read", handlers.MarkAsRead)

	// Preference routes
	api.PUT("/preferences/:userID", handlers.UpdateUserPreferences)
	api.GET("/preferences/:userID", handlers.GetUserPreferences)

	// Reminder routes
	api.POST("/reminders/daily", handlers.CreateDailyReminder)
	api.POST("/reminders/streak", handlers.CreateStreakReminder)

	// Event routes (POC)
	api.POST("/events/practice-completed", handlers.PracticeCompleted)

	// Outbox processing
	api.POST("/outbox/process", handlers.ProcessOutbox)
}

// startOutboxProcessor starts the background outbox processor
func startOutboxProcessor(notificationService services.NotificationService) {
	ticker := time.NewTicker(30 * time.Second) // Process every 30 seconds
	defer ticker.Stop()

	log.Println("Starting outbox processor...")

	for range ticker.C {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		if err := notificationService.ProcessOutbox(ctx); err != nil {
			log.Printf("Outbox processing error: %v", err)
		}
		cancel()
	}
}
