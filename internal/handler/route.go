package handler

import (
	"fmt"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	log "go.uber.org/zap"

	"github.com/Roma7-7-7/shared-clipboard/internal/config"
)

func New(conf config.Config, log *log.SugaredLogger) (*echo.Echo, error) {
	log.Info("Initializing router")

	var (
		e   *echo.Echo
		err error
	)
	e = echo.New()

	if e.Renderer, err = NewTemplatesRenderer(conf.Web.TemplatesPath); err != nil {
		return nil, fmt.Errorf("create renderer: %w", err)
	}

	e.Use(middleware.RequestID())
	e.Use(middleware.Logger())
	e.Use(middleware.RemoveTrailingSlash())
	e.Use(middleware.Recover())
	e.Use(middleware.Gzip())

	setupWeb(e)

	printRoutes(e, log)

	e.HTTPErrorHandler = customHttpErrorHandler(log)

	log.Info("Router initialized")
	return e, nil
}

func setupWeb(e *echo.Echo) {
	e.GET("/", HandleIndex)
	e.GET("/index", HandleIndex)
	e.GET("/index.html", HandleIndex)

	e.Static("/static", "web/static")
}

func printRoutes(e *echo.Echo, logger *log.SugaredLogger) {
	logger.Info("Routes:")
	for _, r := range e.Routes() {
		logger.Infof("%s %s", r.Method, r.Path)
	}
}
