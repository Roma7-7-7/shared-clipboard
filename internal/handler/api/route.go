package api

import (
	"context"
	"fmt"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"

	"github.com/Roma7-7-7/shared-clipboard/internal/handler"
	"github.com/Roma7-7-7/shared-clipboard/tools/trace"
)

func NewRouter(ctx context.Context, sessionRepo SessionRepository, log trace.Logger) (*echo.Echo, error) {
	log.Infow(ctx, "Initializing router")

	e := echo.New()

	e.Use(middleware.RequestID())
	e.Use(handler.Middleware)
	e.Use(middleware.Logger())
	e.Use(middleware.RateLimiter(middleware.NewRateLimiterMemoryStore(10)))
	e.Use(middleware.RemoveTrailingSlash())
	e.Use(middleware.Recover())
	e.Use(middleware.Gzip())

	apiService := NewAPIService(sessionRepo, log)
	apiGroup := e.Group("/")

	apiGroup.POST("/sessions", apiService.Create)

	printRoutes(ctx, e, log)

	e.HTTPErrorHandler = customHttpErrorHandler(log)

	log.Infow(ctx, "Router initialized")
	return e, nil
}

func printRoutes(ctx context.Context, e *echo.Echo, logger trace.Logger) {
	logger.Infow(ctx, "Routes:")
	for _, r := range e.Routes() {
		logger.Infow(ctx, fmt.Sprintf("%s %s", r.Method, r.Path))
	}
}
