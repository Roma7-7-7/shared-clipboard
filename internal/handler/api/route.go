package api

import (
	"fmt"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"github.com/go-chi/httprate"

	"github.com/Roma7-7-7/shared-clipboard/internal/config"
	"github.com/Roma7-7-7/shared-clipboard/internal/handler"
	"github.com/Roma7-7-7/shared-clipboard/tools/log"
	"github.com/Roma7-7-7/shared-clipboard/tools/trace"
)

func NewRouter(sessionRepo SessionRepository, clipboardRepo ClipboardRepository, conf config.API, log log.TracedLogger) (*chi.Mux, error) {
	log.Infow(trace.RuntimeTraceID, "Initializing router")

	r := chi.NewRouter()

	r.Use(handler.TraceID)
	r.Use(handler.Logger(log))
	r.Use(httprate.LimitByIP(10, 1*time.Second))
	r.Use(middleware.RedirectSlashes)
	r.Use(middleware.Recoverer)
	r.Use(middleware.Compress(5, "text/html", "text/css", "text/javascript"))
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   conf.CORS.AllowOrigins,
		AllowedMethods:   conf.CORS.AllowMethods,
		AllowedHeaders:   conf.CORS.AllowHeaders,
		ExposedHeaders:   conf.CORS.ExposeHeaders,
		MaxAge:           conf.CORS.MaxAge,
		AllowCredentials: conf.CORS.AllowCredentials,
	}))

	sessionHandler := NewSessionHandler(sessionRepo, clipboardRepo, log)

	r.Route("/sessions", sessionHandler.RegisterRoutes)

	r.NotFound(handleNotFound(log))
	r.MethodNotAllowed(handleMethodNotAllowed(log))

	printRoutes(r, log)

	log.Infow(trace.RuntimeTraceID, "Router initialized")
	return r, nil
}

func handleNotFound(log log.TracedLogger) func(rw http.ResponseWriter, r *http.Request) {
	return func(rw http.ResponseWriter, r *http.Request) {
		sendNotFound(r.Context(), rw, "Not Found", log)
	}
}

func handleMethodNotAllowed(log log.TracedLogger) func(rw http.ResponseWriter, r *http.Request) {
	return func(rw http.ResponseWriter, r *http.Request) {
		sendErrorMethodNotAllowed(r.Context(), r.Method, rw, log)
	}
}

func printRoutes(r *chi.Mux, logger log.TracedLogger) {
	logger.Infow(trace.RuntimeTraceID, "Routes:")
	err := chi.Walk(r, func(method string, route string, handler http.Handler, middlewares ...func(http.Handler) http.Handler) error {
		logger.Infow(trace.RuntimeTraceID, fmt.Sprintf("[%s]: '%s' has %d middlewares", method, route, len(middlewares)))
		return nil
	})
	if err != nil {
		logger.Errorw(trace.RuntimeTraceID, "Failed to walk routes", err)
	}
}
