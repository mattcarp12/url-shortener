package db

import (
	"context"
	"fmt"
	"os"

	"github.com/redis/go-redis/v9"
)

var RedisClient *redis.Client

// InitRedis initializes the connection to our Redis container
func InitRedis() error {
	redisAddr := "redis:6379"
	
	// Override for production/Docker environments
	if envAddr := os.Getenv("REDIS_URL"); envAddr != "" {
		redisAddr = envAddr
	}

	client := redis.NewClient(&redis.Options{
		Addr:     redisAddr,
		Password: "", // No password set in our devcontainer
		DB:       0,  // Use default DB
		PoolSize: 100, // Connection pooling! SOTA practice.
	})

	// Ping to verify connection
	if err := client.Ping(context.Background()).Err(); err != nil {
		return fmt.Errorf("redis is unreachable: %w", err)
	}

	RedisClient = client
	fmt.Println("✅ Connected to Redis")
	return nil
}

// CloseRedis gracefully shuts down the client
func CloseRedis() {
	if RedisClient != nil {
		RedisClient.Close()
		fmt.Println("🔌 Redis connection closed")
	}
}