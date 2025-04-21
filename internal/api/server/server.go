package server

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/OleksandrOleniuk/twitchong/internal/api/routes"
	"github.com/OleksandrOleniuk/twitchong/internal/config"
)

// Server represents the HTTP server
type Server struct {
	httpServer *http.Server
	config     *config.Provider
}

// New creates a new server instance
func New(cfg *config.Provider) *Server {
	router := routes.SetupRoutes(cfg)

	return &Server{
		httpServer: &http.Server{
			Addr:         fmt.Sprintf(":%d", cfg.Get().ServerPort),
			Handler:      router,
			ReadTimeout:  15 * time.Second,
			WriteTimeout: 15 * time.Second,
		},
		config: cfg,
	}
}

// Start runs the HTTP server and gracefully handles shutdown
func (s *Server) Start() error {
	// Server in a goroutine so shutdown can be handled gracefully
	go func() {
		log.Printf("Starting server on port %d", s.config.Get().ServerPort)
		if err := s.httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Could not listen on port %d: %v", s.config.Get().ServerPort, err)
		}
	}()

	// Wait for interrupt signal to gracefully shutdown the server
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down server...")

	// Create a deadline for server shutdown
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := s.httpServer.Shutdown(ctx); err != nil {
		return fmt.Errorf("server forced to shutdown: %w", err)
	}

	log.Println("Server gracefully stopped")
	return nil
}
