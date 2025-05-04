package server_echo

import (
	"errors"
	"html/template"
	"io"
	"net/http"
	"strconv"

	"github.com/OleksandrOleniuk/twitchong/internal/config"
	"github.com/OleksandrOleniuk/twitchong/pkg/utils"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"go.uber.org/zap"
)

var (
	logger = utils.With(zap.String("component", "main"))
)

type Templates struct {
	templates *template.Template
}

func (t *Templates) Render(w io.Writer, name string, data interface{}, c echo.Context) error {
	return t.templates.ExecuteTemplate(w, name, data)
}

func newTemplate() *Templates {
	return &Templates{
		templates: template.Must(template.ParseGlob("views/*.html")),
	}
}

type Count struct {
	Count int
}

func New(config *config.Config) *echo.Echo {
	e := echo.New()

	// Middleware
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())

	count := Count{Count: 0}
	e.Renderer = newTemplate()

	// Routes
	e.GET("/", func(c echo.Context) error {
		count.Count++
		return c.Render(200, "index", count)
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
