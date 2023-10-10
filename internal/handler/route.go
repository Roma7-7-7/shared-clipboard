package handler

import (
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	log "go.uber.org/zap"

	"github.com/Roma7-7-7/shared-clipboard/internal/config"
)

func New(conf config.Config, log *log.SugaredLogger) (*echo.Echo, error) {
	log.Info("Initializing router")

	var (
		e *echo.Echo
	)
	e = echo.New()

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
	e.GET("/favicon.ico", HandleFavicon)
	e.GET("/", HandleIndex)

	e.Static("/*.html", "web")
	e.Static("/assets", "web/assets")
}

func printRoutes(e *echo.Echo, logger *log.SugaredLogger) {
	logger.Info("Routes:")
	for _, r := range e.Routes() {
		logger.Infof("%s %s", r.Method, r.Path)
	}
}
