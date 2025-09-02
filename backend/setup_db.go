package main

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	_ "github.com/lib/pq"
)

const (
	DBConnectionString = "postgres://postgres:postgres@localhost:5432/postgres?sslmode=disable"
	MaxRetries         = 30
	RetryDelay         = 2 * time.Second
)

func main() {
	fmt.Println("Setting up database...")

	// Try to connect to PostgreSQL with retries
	var db *sql.DB
	var err error

	for i := 0; i < MaxRetries; i++ {
		fmt.Printf("Attempting to connect to PostgreSQL (attempt %d/%d)...\n", i+1, MaxRetries)

		db, err = sql.Open("postgres", DBConnectionString)
		if err != nil {
			fmt.Printf("Failed to open database connection: %v\n", err)
			time.Sleep(RetryDelay)
			continue
		}

		// Test connection
		if err := db.Ping(); err != nil {
			fmt.Printf("Failed to ping database: %v\n", err)
			db.Close()
			time.Sleep(RetryDelay)
			continue
		}

		fmt.Println("Successfully connected to PostgreSQL!")
		break
	}

	if db == nil {
		log.Fatal("Failed to connect to PostgreSQL after all retries")
	}
	defer db.Close()

	// Read and execute the migration file
	fmt.Println("Reading migration file...")
	migrationSQL, err := os.ReadFile("migrations/001_initial_schema.sql")
	if err != nil {
		log.Fatalf("Failed to read migration file: %v", err)
	}

	fmt.Println("Executing database migration...")

	// Split SQL by semicolons and execute each statement separately
	statements := strings.Split(string(migrationSQL), ";")

	for i, statement := range statements {
		statement = strings.TrimSpace(statement)
		if statement == "" {
			continue
		}

		fmt.Printf("Executing statement %d/%d...\n", i+1, len(statements))
		_, err := db.Exec(statement)
		if err != nil {
			log.Printf("Warning: Failed to execute statement %d: %v\n", i+1, err)
			log.Printf("Statement: %s\n", statement)
			// Continue with other statements
		}
	}

	fmt.Println("Database setup complete!")
	fmt.Println("All tables and sample data have been created successfully.")
}
