package server

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/OleksandrOleniuk/twitchong/internal/api/routes"
	"github.com/OleksandrOleniuk/twitchong/internal/config"
	"github.com/OleksandrOleniuk/twitchong/pkg/utils"
	"go.uber.org/zap"
)

// Server represents the HTTP server
type Server struct {
	httpServer *http.Server
	config     *config.Config
	logger     *zap.Logger
}

// New creates a new server instance
func New(cfg *config.Config) *Server {
	router := routes.SetupRoutes(cfg)

	// Create a child logger with server context
	logger := utils.With(
		zap.String("component", "server"),
		zap.Int("port", cfg.ServerPort),
	)

	return &Server{
		httpServer: &http.Server{
			Addr:         fmt.Sprintf(":%d", cfg.ServerPort),
			Handler:      router,
			ReadTimeout:  15 * time.Second,
			WriteTimeout: 15 * time.Second,
		},
		config: cfg,
		logger: logger,
	}
}

// Start runs the HTTP server and gracefully handles shutdown
func (s *Server) Start() error {
	// Server in a goroutine so shutdown can be handled gracefully
	go func() {
		s.logger.Info("starting server")
		if err := s.httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			s.logger.Error("server failed to start",
				zap.Error(err),
				zap.Int("port", s.config.ServerPort),
			)
		}
	}()

	// Wait for interrupt signal to gracefully shutdown the server
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	s.logger.Info("shutting down server")

	// Create a deadline for server shutdown
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := s.httpServer.Shutdown(ctx); err != nil {
		return fmt.Errorf("server forced to shutdown: %w", err)
	}

	s.logger.Info("server gracefully stopped")
	return nil
}
