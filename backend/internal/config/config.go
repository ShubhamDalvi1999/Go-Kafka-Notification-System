package config

import (
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/joho/godotenv"
)

// Config holds all configuration for the application
type Config struct {
	Server   ServerConfig
	Database DatabaseConfig
	Kafka    KafkaConfig
	Logging  LoggingConfig
}

// ServerConfig holds HTTP server configuration
type ServerConfig struct {
	Port         string
	ReadTimeout  time.Duration
	WriteTimeout time.Duration
	IdleTimeout  time.Duration
}

// DatabaseConfig holds database connection configuration
type DatabaseConfig struct {
	Host            string
	Port            int
	User            string
	Password        string
	Database        string
	SSLMode         string
	MaxOpenConns    int
	MaxIdleConns    int
	ConnMaxLifetime time.Duration
	ConnMaxIdleTime time.Duration
}

// KafkaConfig holds Kafka configuration
type KafkaConfig struct {
	Brokers        []string
	Topic          string
	ConsumerGroup  string
	ProducerConfig ProducerConfig
	ConsumerConfig ConsumerConfig
}

// ProducerConfig holds Kafka producer configuration
type ProducerConfig struct {
	RequiredAcks int
	RetryMax     int
	Timeout      time.Duration
}

// ConsumerConfig holds Kafka consumer configuration
type ConsumerConfig struct {
	AutoOffsetReset   string
	SessionTimeout    time.Duration
	HeartbeatInterval time.Duration
}

// LoggingConfig holds logging configuration
type LoggingConfig struct {
	Level      string
	Format     string
	OutputPath string
}

// Load loads configuration from environment variables
func Load() (*Config, error) {
	// Load .env file if it exists
	if err := godotenv.Load(); err != nil {
		// Don't fail if .env doesn't exist
	}

	config := &Config{
		Server: ServerConfig{
			Port:         getEnv("SERVER_PORT", ":8082"),
			ReadTimeout:  getDurationEnv("SERVER_READ_TIMEOUT", 30*time.Second),
			WriteTimeout: getDurationEnv("SERVER_WRITE_TIMEOUT", 30*time.Second),
			IdleTimeout:  getDurationEnv("SERVER_IDLE_TIMEOUT", 60*time.Second),
		},
		Database: DatabaseConfig{
			Host:            getEnv("DB_HOST", "localhost"),
			Port:            getIntEnv("DB_PORT", 5432),
			User:            getEnv("DB_USER", "postgres"),
			Password:        getEnv("DB_PASSWORD", "postgres"),
			Database:        getEnv("DB_NAME", "postgres"),
			SSLMode:         getEnv("DB_SSLMODE", "disable"),
			MaxOpenConns:    getIntEnv("DB_MAX_OPEN_CONNS", 25),
			MaxIdleConns:    getIntEnv("DB_MAX_IDLE_CONNS", 5),
			ConnMaxLifetime: getDurationEnv("DB_CONN_MAX_LIFETIME", 5*time.Minute),
			ConnMaxIdleTime: getDurationEnv("DB_CONN_MAX_IDLE_TIME", 1*time.Minute),
		},
		Kafka: KafkaConfig{
			Brokers:       getStringSliceEnv("KAFKA_BROKERS", []string{"localhost:9092"}),
			Topic:         getEnv("KAFKA_TOPIC", "notifications"),
			ConsumerGroup: getEnv("KAFKA_CONSUMER_GROUP", "notifications-group"),
			ProducerConfig: ProducerConfig{
				RequiredAcks: getIntEnv("KAFKA_PRODUCER_REQUIRED_ACKS", -1),
				RetryMax:     getIntEnv("KAFKA_PRODUCER_RETRY_MAX", 3),
				Timeout:      getDurationEnv("KAFKA_PRODUCER_TIMEOUT", 10*time.Second),
			},
			ConsumerConfig: ConsumerConfig{
				AutoOffsetReset:   getEnv("KAFKA_CONSUMER_AUTO_OFFSET_RESET", "latest"),
				SessionTimeout:    getDurationEnv("KAFKA_CONSUMER_SESSION_TIMEOUT", 30*time.Second),
				HeartbeatInterval: getDurationEnv("KAFKA_CONSUMER_HEARTBEAT_INTERVAL", 3*time.Second),
			},
		},
		Logging: LoggingConfig{
			Level:      getEnv("LOG_LEVEL", "info"),
			Format:     getEnv("LOG_FORMAT", "json"),
			OutputPath: getEnv("LOG_OUTPUT_PATH", ""),
		},
	}

	return config, nil
}

// GetDatabaseDSN returns the database connection string
func (c *Config) GetDatabaseDSN() string {
	return fmt.Sprintf("postgres://%s:%s@%s:%d/%s?sslmode=%s",
		c.Database.User,
		c.Database.Password,
		c.Database.Host,
		c.Database.Port,
		c.Database.Database,
		c.Database.SSLMode,
	)
}

// Helper functions to get environment variables with defaults
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getIntEnv(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
	}
	return defaultValue
}

func getDurationEnv(key string, defaultValue time.Duration) time.Duration {
	if value := os.Getenv(key); value != "" {
		if duration, err := time.ParseDuration(value); err == nil {
			return duration
		}
	}
	return defaultValue
}

func getStringSliceEnv(key string, defaultValue []string) []string {
	if value := os.Getenv(key); value != "" {
		// Simple comma-separated values for now
		// Could be enhanced to support more complex formats
		return []string{value}
	}
	return defaultValue
}
