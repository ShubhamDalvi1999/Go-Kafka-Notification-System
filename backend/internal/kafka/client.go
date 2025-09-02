package kafka

import (
	"fmt"
	"log"
	"time"

	"kafka-notify/internal/config"

	"github.com/IBM/sarama"
)

// ClientManager manages Kafka clients
type ClientManager struct {
	config *config.KafkaConfig
}

// NewClientManager creates a new Kafka client manager
func NewClientManager(cfg *config.KafkaConfig) *ClientManager {
	return &ClientManager{
		config: cfg,
	}
}

// NewProducer creates a new Kafka producer
func (cm *ClientManager) NewProducer() (sarama.SyncProducer, error) {
	config := sarama.NewConfig()

	// Producer configuration
	config.Producer.RequiredAcks = sarama.RequiredAcks(cm.config.ProducerConfig.RequiredAcks)
	config.Producer.Retry.Max = cm.config.ProducerConfig.RetryMax
	config.Producer.Timeout = cm.config.ProducerConfig.Timeout
	config.Producer.Return.Successes = true
	config.Producer.Return.Errors = true

	// Compression
	config.Producer.Compression = sarama.CompressionSnappy

	// Idempotent producer for exactly-once semantics
	config.Producer.Idempotent = true
	config.Net.MaxOpenRequests = 1

	producer, err := sarama.NewSyncProducer(cm.config.Brokers, config)
	if err != nil {
		return nil, fmt.Errorf("failed to create Kafka producer: %w", err)
	}

	log.Printf("Kafka producer created successfully, connected to brokers: %v", cm.config.Brokers)
	return producer, nil
}

// NewConsumerGroup creates a new Kafka consumer group
func (cm *ClientManager) NewConsumerGroup(groupID string) (sarama.ConsumerGroup, error) {
	config := sarama.NewConfig()

	// Consumer group configuration
	config.Consumer.Group.Session.Timeout = cm.config.ConsumerConfig.SessionTimeout
	config.Consumer.Group.Heartbeat.Interval = cm.config.ConsumerConfig.HeartbeatInterval
	config.Consumer.Group.Rebalance.Strategy = sarama.BalanceStrategyRoundRobin

	// Consumer configuration
	config.Consumer.Offsets.Initial = getOffsetReset(cm.config.ConsumerConfig.AutoOffsetReset)
	config.Consumer.Offsets.AutoCommit.Enable = true
	config.Consumer.Offsets.AutoCommit.Interval = 1 * time.Second

	// Network configuration
	config.Net.DialTimeout = 10 * time.Second
	config.Net.ReadTimeout = 30 * time.Second
	config.Net.WriteTimeout = 30 * time.Second

	consumerGroup, err := sarama.NewConsumerGroup(cm.config.Brokers, groupID, config)
	if err != nil {
		return nil, fmt.Errorf("failed to create Kafka consumer group: %w", err)
	}

	log.Printf("Kafka consumer group created successfully, group ID: %s, brokers: %v", groupID, cm.config.Brokers)
	return consumerGroup, nil
}

// CloseProducer closes a Kafka producer
func (cm *ClientManager) CloseProducer(producer sarama.SyncProducer) error {
	if producer != nil {
		log.Println("Closing Kafka producer...")
		return producer.Close()
	}
	return nil
}

// CloseConsumerGroup closes a Kafka consumer group
func (cm *ClientManager) CloseConsumerGroup(consumerGroup sarama.ConsumerGroup) error {
	if consumerGroup != nil {
		log.Println("Closing Kafka consumer group...")
		return consumerGroup.Close()
	}
	return nil
}

// getOffsetReset converts string offset reset to sarama constant
func getOffsetReset(offsetReset string) int64 {
	switch offsetReset {
	case "earliest":
		return sarama.OffsetOldest
	case "latest":
		return sarama.OffsetNewest
	default:
		return sarama.OffsetNewest
	}
}

// HealthCheck performs a health check on Kafka connectivity
func (cm *ClientManager) HealthCheck() error {
	// Try to create a temporary producer to test connectivity
	producer, err := cm.NewProducer()
	if err != nil {
		return fmt.Errorf("Kafka health check failed: %w", err)
	}
	defer cm.CloseProducer(producer)

	log.Println("Kafka health check passed")
	return nil
}
