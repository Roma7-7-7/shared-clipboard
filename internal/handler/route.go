package handler

import (
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	log "go.uber.org/zap"
)

func New(sessionRepo SessionRepository, log *log.SugaredLogger) (*echo.Echo, error) {
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

	setupAPI(NewAPI(sessionRepo, log), e)

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

func setupAPI(api *API, e *echo.Echo) {
	apiGroup := e.Group("/api")

	apiGroup.POST("/session", api.Create)
}

func printRoutes(e *echo.Echo, logger *log.SugaredLogger) {
	logger.Info("Routes:")
	for _, r := range e.Routes() {
		logger.Infof("%s %s", r.Method, r.Path)
	}
}
