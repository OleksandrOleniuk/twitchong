package server

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/OleksandrOleniuk/twitchong/internal/api/handlers"
	"github.com/OleksandrOleniuk/twitchong/internal/api/middleware"
	"github.com/OleksandrOleniuk/twitchong/internal/config"
	"github.com/OleksandrOleniuk/twitchong/pkg/utils"
	"github.com/OleksandrOleniuk/twitchong/views"
	"github.com/labstack/echo/v4"
	"go.uber.org/zap"
)

var (
	logger = utils.With(zap.String("component", "main"))
)

func New(config *config.Config) *echo.Echo {
	e := echo.New()

	// Middleware
	e.Use(middleware.Logger())

	e.File("/index.js", "assets/js/index.js")
	e.File("/index.min.css", "assets/js/index.min.css")

	e.GET("/twitch/callback", handlers.HandleTwitchCallback)
	e.POST("/process-tokens", handlers.ProcessTokens)

	e.GET("/", func(c echo.Context) error {
		handlers.SetState(config.TwitchSecretState, true)
		return utils.TemplRender(c, http.StatusOK, views.IndexPage(config.ClientId, config.TwitchSecretState))
	})

	return e
}

// Start runs the HTTP server and gracefully handles shutdown
func Start(e *echo.Echo, cfg *config.Config) error {
	// Server in a goroutine so shutdown can be handled gracefully
	go func() {
		// Start server
		if err := e.Start(":" + strconv.Itoa(cfg.ServerPort)); err != nil && !errors.Is(err, http.ErrServerClosed) {
			logger.Error("failed to start server", zap.Error(err))
		}
	}()
	return nil
}
