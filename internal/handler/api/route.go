package api

import (
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	log "go.uber.org/zap"
)

func NewAPIRouter(sessionRepo SessionRepository, log *log.SugaredLogger) (*echo.Echo, error) {
	log.Info("Initializing router")

	var (
		e *echo.Echo
	)
	e = echo.New()

	e.Use(middleware.RequestID())
	e.Use(middleware.Logger())
	e.Use(middleware.RateLimiter(middleware.NewRateLimiterMemoryStore(10)))
	e.Use(middleware.RemoveTrailingSlash())
	e.Use(middleware.Recover())
	e.Use(middleware.Gzip())

	api := NewAPIService(sessionRepo, log)
	apiGroup := e.Group("/apis")
	apiGroup.Use(func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			ac := &APIContext{Context: c}
			if err := next(ac); err != nil {
				handleAPIError(ac, err, api.log)
			}

			return nil
		}
	})

	apiGroup.POST("/sessions", api.Create)

	printRoutes(e, log)

	e.HTTPErrorHandler = customHttpErrorHandler(log)

	log.Info("Router initialized")
	return e, nil
}

func printRoutes(e *echo.Echo, logger *log.SugaredLogger) {
	logger.Info("Routes:")
	for _, r := range e.Routes() {
		logger.Infof("%s %s", r.Method, r.Path)
	}
}
