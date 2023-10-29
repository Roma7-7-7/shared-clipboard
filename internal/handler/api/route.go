package api

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"github.com/go-chi/httprate"

	"github.com/Roma7-7-7/shared-clipboard/internal/config"
	"github.com/Roma7-7-7/shared-clipboard/internal/handler"
	"github.com/Roma7-7-7/shared-clipboard/internal/handler/api/session"
	"github.com/Roma7-7-7/shared-clipboard/tools/rest"
	"github.com/Roma7-7-7/shared-clipboard/tools/trace"
)

func NewRouter(ctx context.Context, sessionRepo session.Repository, conf config.API, log trace.Logger) (*chi.Mux, error) {
	log.Infow(ctx, "Initializing router")

	r := chi.NewRouter()

	r.Use(handler.TraceID)
	r.Use(handler.Logger(log))
	r.Use(httprate.LimitByIP(10, 1*time.Second))
	r.Use(middleware.RedirectSlashes)
	r.Use(middleware.Recoverer)
	r.Use(middleware.Compress(5, "text/javascript"))
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{fmt.Sprintf("http://%s", conf.CORS.AllowOrigin), fmt.Sprintf("https://%s", conf.CORS.AllowOrigin)},
		AllowedMethods:   conf.CORS.AllowMethods,
		AllowedHeaders:   conf.CORS.AllowHeaders,
		ExposedHeaders:   conf.CORS.ExposeHeaders,
		MaxAge:           conf.CORS.MaxAge,
		AllowCredentials: conf.CORS.AllowCredentials,
	}))

	sessionHandler := session.NewSessionHandler(sessionRepo, log)

	r.Route("/sessions", sessionHandler.RegisterRoutes)

	r.NotFound(handleNotFound(log))
	r.MethodNotAllowed(handleMethodNotAllowed(log))

	printRoutes(ctx, r, log)

	log.Infow(ctx, "Router initialized")
	return r, nil
}

func handleNotFound(log trace.Logger) func(rw http.ResponseWriter, r *http.Request) {
	return func(rw http.ResponseWriter, r *http.Request) {
		rest.SendNotFound(r.Context(), rw, log)
	}
}

func handleMethodNotAllowed(log trace.Logger) func(rw http.ResponseWriter, r *http.Request) {
	return func(rw http.ResponseWriter, r *http.Request) {
		rest.SendErrorMethodNotAllowed(r.Context(), r.Method, rw, log)
	}
}

func printRoutes(ctx context.Context, r *chi.Mux, logger trace.Logger) {
	logger.Infow(ctx, "Routes:")
	err := chi.Walk(r, func(method string, route string, handler http.Handler, middlewares ...func(http.Handler) http.Handler) error {
		logger.Infow(ctx, fmt.Sprintf("[%s]: '%s' has %d middlewares", method, route, len(middlewares)))
		return nil
	})
	if err != nil {
		logger.Errorw(ctx, "Failed to walk routes", err)
	}
}
