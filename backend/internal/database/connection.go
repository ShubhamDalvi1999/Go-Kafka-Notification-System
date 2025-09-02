package database

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"os"
	"time"

	"kafka-notify/internal/config"

	_ "github.com/lib/pq"
)

// ConnectionManager manages database connections
type ConnectionManager struct {
	db     *sql.DB
	config *config.DatabaseConfig
}

// NewConnectionManager creates a new database connection manager
func NewConnectionManager(cfg *config.DatabaseConfig) (*ConnectionManager, error) {
	// Optional: allow forcing IPv4 by specifying DB_HOSTADDR (A record)
	hostaddr := os.Getenv("DB_HOSTADDR")
	var dsn string
	if hostaddr != "" {
		dsn = fmt.Sprintf(
			"host=%s hostaddr=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
			cfg.Host, hostaddr, cfg.Port, cfg.User, cfg.Password, cfg.Database, cfg.SSLMode,
		)
	} else {
		dsn = fmt.Sprintf(
			"host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
			cfg.Host, cfg.Port, cfg.User, cfg.Password, cfg.Database, cfg.SSLMode,
		)
	}

	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to open database connection: %w", err)
	}

	// Configure connection pool
	db.SetMaxOpenConns(cfg.MaxOpenConns)
	db.SetMaxIdleConns(cfg.MaxIdleConns)
	db.SetConnMaxLifetime(cfg.ConnMaxLifetime)
	db.SetConnMaxIdleTime(cfg.ConnMaxIdleTime)

	// Test connection
	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	manager := &ConnectionManager{
		db:     db,
		config: cfg,
	}

	// Start health check goroutine
	go manager.startHealthCheck()

	return manager, nil
}

// GetDB returns the underlying database connection
func (cm *ConnectionManager) GetDB() *sql.DB {
	return cm.db
}

// Close closes the database connection
func (cm *ConnectionManager) Close() error {
	log.Println("Closing database connection...")
	return cm.db.Close()
}

// HealthCheck performs a health check on the database
func (cm *ConnectionManager) HealthCheck(ctx context.Context) error {
	return cm.db.PingContext(ctx)
}

// startHealthCheck runs periodic health checks
func (cm *ConnectionManager) startHealthCheck() {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for range ticker.C {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		if err := cm.HealthCheck(ctx); err != nil {
			log.Printf("Database health check failed: %v", err)
		}
		cancel()
	}
}

// Stats returns database connection statistics
func (cm *ConnectionManager) Stats() sql.DBStats {
	return cm.db.Stats()
}
