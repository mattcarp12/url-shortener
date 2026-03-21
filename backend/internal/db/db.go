package db

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/jackc/pgx/v5/pgxpool"
)

// DB is a global variable holding our connection pool
var Pool *pgxpool.Pool

// InitDB initializes the PostgreSQL connection pool
func InitDB() error {
	// In production, this comes from an environment variable.
	// For local dev, this matches our docker-compose.yml credentials.
	dsn := "postgres://devuser:devpassword@localhost:5432/shortener?sslmode=disable"

	// If we set an env var, override the local default
	if envDsn := os.Getenv("DATABASE_URL"); envDsn != "" {
		dsn = envDsn
	}

	config, err := pgxpool.ParseConfig(dsn)
	if err != nil {
		return fmt.Errorf("unable to parse database config: %w", err)
	}

	// System Design: Tuning the pool.
	// We limit connections to prevent overwhelming the DB under load.
	config.MaxConns = 20

	pool, err := pgxpool.NewWithConfig(context.Background(), config)
	if err != nil {
		return fmt.Errorf("unable to connect to database: %w", err)
	}

	// Ping to verify connection
	if err := pool.Ping(context.Background()); err != nil {
		return fmt.Errorf("database is unreachable: %w", err)
	}

	Pool = pool

	// Initialize the database schema
	if err := createTableIfNotExists(); err != nil {
		return fmt.Errorf("failed to initialize database schema: %w", err)
	}

	fmt.Println("✅ Connected to PostgreSQL pool")
	return nil
}

// Close gracefully shuts down the connection pool
func Close() {
	if Pool != nil {
		Pool.Close()
		fmt.Println("🔌 Database connection closed")
	}
}

func createTableIfNotExists() error {
	// Create the urls table if it doesn't exist
	createTableQuery := `
	CREATE TABLE IF NOT EXISTS urls (
		id BIGSERIAL PRIMARY KEY,
		original_url TEXT NOT NULL,
		short_code VARCHAR(50) UNIQUE NOT NULL,
		created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
	);`

	// Execute the schema creation
	_, err := Pool.Exec(context.Background(), createTableQuery)
	if err != nil {
		log.Fatalf("Failed to initialize database schema: %v", err)
	}

	log.Println("✅ Database schema verified/loaded")
	return nil
}
