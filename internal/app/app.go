package app

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/redis/go-redis/v9"

	"github.com/Roma7-7-7/shared-clipboard/internal/config"
	"github.com/Roma7-7-7/shared-clipboard/internal/dal"
	"github.com/Roma7-7-7/shared-clipboard/internal/domain"
	"github.com/Roma7-7-7/shared-clipboard/internal/handle"
	"github.com/Roma7-7-7/shared-clipboard/internal/handle/cookie"
	"github.com/Roma7-7-7/shared-clipboard/internal/handle/jwt"
	"github.com/Roma7-7-7/shared-clipboard/internal/log"
)

type (
	App struct {
		port int
		mux  *chi.Mux
		log  log.TracedLogger
	}
)

func NewApp(ctx context.Context, conf config.App, traced log.TracedLogger) (*App, error) {
	dbURL := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		conf.DB.Host, conf.DB.Port, conf.DB.User, conf.DB.Password, conf.DB.Name, conf.DB.SSLMode,
	)
	traced.Infow(ctx, "Initializing SQL DB", "host", conf.DB.Host, "port", conf.DB.Port, "name", conf.DB.Name)
	sqlDB, err := sql.Open(conf.DB.Driver, dbURL)
	if err != nil {
		return nil, fmt.Errorf("open sql db: %w", err)
	}

	redis := redis.NewClient(&redis.Options{
		Addr:         conf.Redis.Addr,
		Password:     conf.Redis.Password,
		DB:           conf.Redis.DB,
		ReadTimeout:  time.Duration(conf.Redis.TimeoutMillis) * time.Millisecond,
		WriteTimeout: time.Duration(conf.Redis.TimeoutMillis) * time.Millisecond,
	})

	traced.Infow(ctx, "Initializing repositories")
	userRpo, err := dal.NewUserRepository(sqlDB)
	if err != nil {
		return nil, fmt.Errorf("create user repository: %w", err)
	}
	sessionRepo, err := dal.NewSessionRepository(sqlDB)
	if err != nil {
		return nil, fmt.Errorf("create session repository: %w", err)
	}
	traced.Infow(ctx, "Initializing services")
	userService := domain.NewUserService(userRpo, traced)

	traced.Infow(ctx, "Initializing components")
	jwtProcessor := jwt.NewProcessor(conf.JWT)
	cookieProcessor := cookie.NewProcessor(jwtProcessor, conf.Cookie)

	sessionService := domain.NewSessionService(sessionRepo, traced)

	traced.Infow(ctx, "Creating router")
	h, err := handle.NewRouter(ctx, handle.Dependencies{
		Config:           conf,
		CookieProcessor:  cookieProcessor,
		UserService:      userService,
		JTIService:       domain.NewJTIService(redis, traced),
		SessionService:   sessionService,
		ClipboardService: domain.NewClipboardService(redis, traced),
	}, traced)
	if err != nil {
		return nil, fmt.Errorf("create router: %w", err)
	}

	return &App{
		port: conf.Port,
		mux:  h,
		log:  traced,
	}, nil
}

func (a *App) Run(ctx context.Context) error {
	done := make(chan struct{})
	defer close(done)

	addr := fmt.Sprintf(":%d", a.port)
	s := http.Server{
		Addr:        addr,
		Handler:     a.mux,
		ReadTimeout: 30 * time.Second,
	}

	go func() {
		select {
		case <-done:
			return
		case <-ctx.Done():
			ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
			defer cancel()
			a.log.Infow(ctx, "Shutting down server")
			if err := s.Shutdown(ctx); err != nil {
				a.log.Errorw(ctx, "Shutdown server", err)
			}
			a.log.Infow(ctx, "Server stopped")
			return
		}
	}()

	a.log.Infow(ctx, "Starting server", "address", addr)
	if err := s.ListenAndServe(); !errors.Is(err, http.ErrServerClosed) {
		return fmt.Errorf("server listen: %w", err)
	}
	a.log.Infow(ctx, "Server stopped")

	return nil
}
