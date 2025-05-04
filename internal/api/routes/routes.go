package routes

import (
	"net/http"

	"github.com/OleksandrOleniuk/twitchong/internal/api/handlers"
	"github.com/OleksandrOleniuk/twitchong/internal/api/middleware"
	"github.com/OleksandrOleniuk/twitchong/internal/config"
)

// SetupRoutes configures all application routes
func SetupRoutes(cfg *config.Config) http.Handler {
	mux := http.NewServeMux()

	// Setup middleware
	handler := middleware.Logger(mux)

	// API version group
	mux.HandleFunc("GET /api/v1/health", handlers.HealthCheck)

	// Twitch OAuth callback (handles both GET and POST)
	mux.HandleFunc("/twitch/callback", handlers.NewTwitchCallback(cfg))

	// Static file serving - more specific routes first
	staticHandler := handlers.NewStaticHandler(cfg)
	mux.HandleFunc("GET /index.html", staticHandler)
	mux.HandleFunc("GET /js/{file}", staticHandler)
	mux.HandleFunc("GET /css/{file}", staticHandler)
	mux.HandleFunc("GET /callback.html", staticHandler)

	return handler
}
