package handle

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
	"github.com/Roma7-7-7/shared-clipboard/internal/domain"
	"github.com/Roma7-7-7/shared-clipboard/internal/log"
)

type Dependencies struct {
	Config config.App
	CookieProcessor
	UserService
	JWTRepository
	SessionService
	ClipboardRepository
}

func NewRouter(ctx context.Context, deps Dependencies, log log.TracedLogger) (*chi.Mux, error) {
	log.Infow(ctx, "Initializing router")

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

	resp := &responder{log: log}

	authHandler := NewAuthHandler(deps.UserService, deps.CookieProcessor, deps.JWTRepository, resp, log)
	r.Post("/signup", authHandler.SignUp)
	r.Post("/signin", authHandler.SignIn)
	r.Post("/signout", authHandler.SignOut)

	authorizedRouter := r.With(NewAuthorizedMiddleware(deps.CookieProcessor, deps.JWTRepository, resp, log).Handle)

	sessionHandler := NewSessionHandler(deps.SessionService, deps.ClipboardRepository, resp, log)
	authorizedRouter.Post("/v1/sessions", sessionHandler.Create)
	authorizedRouter.Get("/v1/sessions", sessionHandler.FilterBy)
	authorizedRouter.Get("/v1/sessions/{sessionID}", sessionHandler.GetByID)
	authorizedRouter.Put("/v1/sessions/{sessionID}", sessionHandler.Update)
	authorizedRouter.Delete("/v1/sessions/{sessionID}", sessionHandler.Delete)
	authorizedRouter.Get("/v1/sessions/{sessionID}/clipboard", sessionHandler.GetClipboard)
	authorizedRouter.Put("/v1/sessions/{sessionID}/clipboard", sessionHandler.SetClipboard)

	userHandler := NewUserHandler(resp, log)
	authorizedRouter.Get("/v1/user/info", userHandler.GetUserInfo)

	r.NotFound(handleNotFound(resp))
	r.MethodNotAllowed(handleMethodNotAllowed(resp))

	printRoutes(ctx, r, log)

	log.Infow(ctx, "Router initialized")
	return r, nil
}

func handleNotFound(resp *responder) func(rw http.ResponseWriter, r *http.Request) {
	return func(rw http.ResponseWriter, r *http.Request) {
		resp.SendNotFound(r.Context(), rw, "Not Found")
	}
}

func handleMethodNotAllowed(resp *responder) func(rw http.ResponseWriter, r *http.Request) {
	return func(rw http.ResponseWriter, r *http.Request) {
		resp.SendError(r.Context(), rw, http.StatusMethodNotAllowed, domain.ErrorCodeMethodNotAllowed.Value, "Method Not Allowed", nil)
	}
}

func printRoutes(ctx context.Context, r *chi.Mux, logger log.TracedLogger) {
	logger.Infow(ctx, "Routes:")
	err := chi.Walk(r, func(method string, route string, handler http.Handler, middlewares ...func(http.Handler) http.Handler) error {
		logger.Infow(ctx, fmt.Sprintf("[%s]: '%s' has %d middlewares", method, route, len(middlewares)))
		return nil
	})
	if err != nil {
		logger.Errorw(ctx, "Failed to walk routes", err)
	}
}
