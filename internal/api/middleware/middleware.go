package middleware

import (
	"net/http"
	"time"

	"github.com/OleksandrOleniuk/twitchong/pkg/utils"
	"go.uber.org/zap"
)

// Logger logs incoming requests with their method, path, and duration
func Logger(next http.Handler) http.Handler {
	logger := utils.With(zap.String("component", "http"))
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		// Call the next handler
		next.ServeHTTP(w, r)

		// Log after request is processed
		duration := time.Since(start)
		logger.Info("request completed",
			zap.String("method", r.Method),
			zap.String("path", r.URL.Path),
			zap.Duration("duration", duration),
		)
	})
}
