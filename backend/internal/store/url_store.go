package store

import (
	"context"
	"fmt"
	"time"

	"github.com/mattcarp12/url-shortener/internal/base62"
	"github.com/mattcarp12/url-shortener/internal/db"
	"github.com/redis/go-redis/v9"
)

type URLRecord struct {
	ID          uint64
	OriginalURL string
	ShortCode   string
	CreatedAt   time.Time
}

// CreateShortURL handles both auto-generated and custom aliases
func CreateShortURL(ctx context.Context, originalURL string, customAlias string) (*URLRecord, error) {
	var record URLRecord

	if customAlias != "" {
		// --- PATH A: Custom Alias ---
		// Attempt to insert the custom alias directly.
		// If someone else already took it, Postgres will throw a unique constraint violation error.
		query := `
			INSERT INTO urls (original_url, short_code) 
			VALUES ($1, $2) 
			RETURNING id, original_url, short_code, created_at`

		err := db.Pool.QueryRow(ctx, query, originalURL, customAlias).Scan(
			&record.ID, &record.OriginalURL, &record.ShortCode, &record.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to save custom alias (might be taken): %w", err)
		}
		return &record, nil
	}

	// --- PATH B: Auto-Generated Base-62 ---
	// Step 1: Get the next ID from the sequence without inserting a row yet.
	// This is highly concurrent and thread-safe in Postgres.
	var nextID uint64
	err := db.Pool.QueryRow(ctx, "SELECT nextval('urls_id_seq')").Scan(&nextID)
	if err != nil {
		return nil, fmt.Errorf("failed to get next sequence id: %w", err)
	}

	// Step 2: Base-62 encode the sequence ID
	shortCode := base62.Encode(nextID)

	// Step 3: Insert the final record
	query := `
		INSERT INTO urls (id, original_url, short_code) 
		VALUES ($1, $2, $3) 
		RETURNING id, original_url, short_code, created_at`

	err = db.Pool.QueryRow(ctx, query, nextID, originalURL, shortCode).Scan(
		&record.ID, &record.OriginalURL, &record.ShortCode, &record.CreatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to insert auto-generated url: %w", err)
	}

	return &record, nil
}

// GetOriginalURL performs a Cache-Aside lookup: Redis -> Postgres -> Update Redis
func GetOriginalURL(ctx context.Context, shortCode string) (string, error) {
	// Step 1: Check Redis
	// We prefix the key with "url:" to keep our Redis keyspace organized
	cacheKey := "url:" + shortCode
	cachedURL, err := db.RedisClient.Get(ctx, cacheKey).Result()

	if err == nil {
		// CACHE HIT! Return immediately. No DB query needed.
		fmt.Printf("🚀 Cache Hit for %s\n", shortCode)
		return cachedURL, nil
	} else if err != redis.Nil {
		// An actual error occurred with Redis (not just a missing key)
		// We log it, but we don't fail! We fall back to the DB to ensure high availability.
		fmt.Printf("⚠️ Redis error: %v\n", err)
	}

	// Step 2: CACHE MISS. Query Postgres via B-Tree Index
	fmt.Printf("🐢 Cache Miss for %s. Querying DB...\n", shortCode)
	var originalURL string
	query := `SELECT original_url FROM urls WHERE short_code = $1`

	err = db.Pool.QueryRow(ctx, query, shortCode).Scan(&originalURL)
	if err != nil {
		return "", fmt.Errorf("url not found: %w", err)
	}

	// Step 3: Populate the Cache for next time
	// System Design: We set a TTL (Time To Live) of 7 days.
	// If a link isn't clicked for 7 days, it drops out of RAM to save memory.
	// The next time it's clicked, it will just be pulled from Postgres again.
	err = db.RedisClient.Set(ctx, cacheKey, originalURL, 7*24*time.Hour).Err()
	if err != nil {
		fmt.Printf("⚠️ Failed to update cache for %s: %v\n", shortCode, err)
	}

	return originalURL, nil
}
