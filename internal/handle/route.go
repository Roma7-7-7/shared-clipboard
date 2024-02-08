package handle

import (
	"fmt"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"github.com/go-chi/httprate"

	"github.com/Roma7-7-7/shared-clipboard/internal/config"
	"github.com/Roma7-7-7/shared-clipboard/internal/domain"
	"github.com/Roma7-7-7/shared-clipboard/tools/log"
)

type Dependencies struct {
	Config config.App
	CookieProcessor
	UserService
	JWTRepository
	SessionService
	ClipboardRepository
}

func NewRouter(deps Dependencies, log log.TracedLogger) (*chi.Mux, error) {
	log.Infow(domain.RuntimeTraceID, "Initializing router")

	r := chi.NewRouter()
	conf := deps.Config

	r.Use(TraceID)
	r.Use(Logger(log))
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

	r.Route("/", NewAuthHandler(deps.UserService, deps.CookieProcessor, deps.JWTRepository, log).RegisterRoutes)

	r.With(NewAuthorizedMiddleware(deps.CookieProcessor, deps.JWTRepository, log).Handle).
		Route("/v1", func(r chi.Router) {
			r.Route("/sessions", NewSessionHandler(deps.SessionService, deps.ClipboardRepository, log).RegisterRoutes)
			r.Route("/user", NewUserHandler(log).RegisterRoutes)
		})

	r.NotFound(handleNotFound(log))
	r.MethodNotAllowed(handleMethodNotAllowed(log))

	printRoutes(r, log)

	log.Infow(domain.RuntimeTraceID, "Router initialized")
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
	logger.Infow(domain.RuntimeTraceID, "Routes:")
	err := chi.Walk(r, func(method string, route string, handler http.Handler, middlewares ...func(http.Handler) http.Handler) error {
		logger.Infow(domain.RuntimeTraceID, fmt.Sprintf("[%s]: '%s' has %d middlewares", method, route, len(middlewares)))
		return nil
	})
	if err != nil {
		logger.Errorw(domain.RuntimeTraceID, "Failed to walk routes", err)
	}
}
