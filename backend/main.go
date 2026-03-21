package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/mattcarp12/url-shortener/internal/db"
	"github.com/mattcarp12/url-shortener/internal/middleware"
	"github.com/mattcarp12/url-shortener/internal/store"

	"github.com/rs/cors"
)

// Define our JSON request and response payloads
type CreateURLRequest struct {
	OriginalURL string `json:"original_url"`
	CustomAlias string `json:"custom_alias,omitempty"`
}

type CreateURLResponse struct {
	ShortURL string `json:"short_url"`
}

func main() {
	// 1. Initialize the Database Connection Pool
	if err := db.InitDB(); err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}
	defer db.Close()

	// 1.5 Initialize Redis
	if err := db.InitRedis(); err != nil {
		log.Fatalf("Failed to initialize redis: %v", err)
	}
	defer db.CloseRedis()

	// 1.75 Initialize Rate Limiter
	middleware.InitRateLimiter()

	// 2. Set up the multiplexer (router)
	mux := http.NewServeMux()

	frontendOrigin := "http://localhost:4566"
	if envOrigin := os.Getenv("FRONTEND_URL"); envOrigin != "" {
		frontendOrigin = envOrigin
	}

	corsOptions := cors.New(cors.Options{
		AllowedOrigins:   []string{frontendOrigin},
		AllowedMethods:   []string{"GET", "POST"},
		AllowedHeaders:   []string{"Content-Type"},
		AllowCredentials: true,
	})

	// 3. Define the routes
	// Go 1.22+ allows us to specify the HTTP method directly in the route string
	mux.HandleFunc("GET /health", func(w http.ResponseWriter, r *http.Request) { w.Write([]byte("OK")) })
	mux.HandleFunc("POST /api/urls", middleware.RateLimitAPI(handleCreateURL))
	mux.HandleFunc("GET /{shortCode}", handleRedirect)

	// 4. Start the server
	port := ":8080"
	fmt.Printf("🚀 Server starting on http://localhost%s\n", port)
	if err := http.ListenAndServe(port, corsOptions.Handler(mux)); err != nil {
		log.Fatalf("Server failed to start: %v", err)
	}
}

// handleCreateURL accepts a long URL and returns the short version
func handleCreateURL(w http.ResponseWriter, r *http.Request) {
	var req CreateURLRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON body", http.StatusBadRequest)
		return
	}

	if req.OriginalURL == "" {
		http.Error(w, "original_url is required", http.StatusBadRequest)
		return
	}

	// Call our store to save/generate the URL
	record, err := store.CreateShortURL(r.Context(), req.OriginalURL, req.CustomAlias)
	if err != nil {
		// In a real app, we'd check if this is a unique constraint error for the custom alias
		log.Printf("Error creating URL: %v", err)
		http.Error(w, "Failed to create short URL", http.StatusInternalServerError)
		return
	}

	// We'll construct the full short URL. In production, the domain comes from an env var.
	baseURL := "http://localhost:8080/"
	resp := CreateURLResponse{
		ShortURL: baseURL + record.ShortCode,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(resp)
}

// handleRedirect looks up the short code and redirects the user
func handleRedirect(w http.ResponseWriter, r *http.Request) {
	// Extract the {shortCode} variable from the URL path
	shortCode := r.PathValue("shortCode")

	originalURL, err := store.GetOriginalURL(r.Context(), shortCode)
	if err != nil {
		http.Error(w, "URL not found", http.StatusNotFound)
		return
	}

	// Perform a 302 Temporary Redirect.
	// (We use 302 instead of 301 Permanent so analytics always hit our server)
	http.Redirect(w, r, originalURL, http.StatusFound)
}
