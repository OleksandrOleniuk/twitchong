package middleware

import (
	"time"

	"github.com/labstack/echo/v4"
	"go.uber.org/zap"
)

// Logger returns a middleware that logs HTTP requests with method, path, and duration
func Logger() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			// Create logger with component field
			logger := zap.L().With(zap.String("component", "http"))

			// Record start time
			start := time.Now()

			// Process request
			err := next(c)

			// Calculate duration after request is processed
			duration := time.Since(start)

			// Log request details
			logger.Info("request completed",
				zap.String("method", c.Request().Method),
				zap.String("path", c.Request().URL.Path),
				zap.Duration("duration", duration),
			)

			return err
		}
	}
}
