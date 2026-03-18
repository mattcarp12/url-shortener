package db

import (
	"context"
	"fmt"
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