package middleware

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"github.com/go-redis/redis_rate/v10"
	"github.com/mattcarp12/url-shortener/internal/db"
)

var limiter *redis_rate.Limiter

// InitRateLimiter sets up the global limiter using our existing Redis connection
func InitRateLimiter() {
	limiter = redis_rate.NewLimiter(db.RedisClient)
	fmt.Println("🛡️  Rate Limiter initialized")
}

// RateLimitAPI is a middleware that enforces 5 requests per minute per IP
func RateLimitAPI(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// 1. Identify the user by IP address
		// In AWS behind a load balancer, you would check the "X-Forwarded-For" header instead.
		ip := strings.Split(r.RemoteAddr, ":")[0]

		// We prefix the key so we know what it is in Redis
		key := "rate_limit:api:" + ip

		// 2. Check the Token Bucket
		// Limit: 5 requests per minute
		res, err := limiter.Allow(context.Background(), key, redis_rate.PerMinute(5))
		if err != nil {
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}

		// 3. If no tokens are left (Allowed == 0), reject the request
		if res.Allowed == 0 {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusTooManyRequests) // HTTP 429
			fmt.Fprintf(w, `{"error": "Rate limit exceeded. Try again in %v"}`, res.RetryAfter)
			return
		}

		// 4. If they have tokens, pass the request to the actual handler
		next.ServeHTTP(w, r)
	}
}
