package app

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	bolt "go.etcd.io/bbolt"

	"github.com/Roma7-7-7/shared-clipboard/internal/config"
	"github.com/Roma7-7-7/shared-clipboard/internal/dal/local"
	"github.com/Roma7-7-7/shared-clipboard/internal/dal/postgre"
	"github.com/Roma7-7-7/shared-clipboard/internal/domain"
	"github.com/Roma7-7-7/shared-clipboard/internal/handle"
	"github.com/Roma7-7-7/shared-clipboard/internal/handle/cookie"
	"github.com/Roma7-7-7/shared-clipboard/internal/handle/jwt"
	"github.com/Roma7-7-7/shared-clipboard/tools/log"
)

type (
	App struct {
		port int
		mux  *chi.Mux
		log  log.TracedLogger
	}
)

func NewApp(conf config.App, traced log.TracedLogger) (*App, error) {
	traced.Infow(domain.RuntimeTraceID, "Initializing SQL DB")
	sqlDB, err := sql.Open(conf.DB.SQL.Driver, conf.DB.SQL.DataSource)
	if err != nil {
		return nil, fmt.Errorf("open sql db: %w", err)
	}

	traced.Infow(domain.RuntimeTraceID, "Initializing Bolt DB")
	boltDB, err := bolt.Open(conf.DB.Bolt.Path, 0600, nil)
	if err != nil {
		return nil, fmt.Errorf("open bolt db: %w", err)
	}

	traced.Infow(domain.RuntimeTraceID, "Initializing repositories")
	userRpo, err := postgre.NewUserRepository(sqlDB)
	if err != nil {
		return nil, fmt.Errorf("create user repository: %w", err)
	}
	sessionRepo, err := postgre.NewSessionRepository(sqlDB)
	if err != nil {
		return nil, fmt.Errorf("create session repository: %w", err)
	}
	clipboardRepo, err := local.NewClipboardRepository(boltDB)
	if err != nil {
		return nil, fmt.Errorf("create clipboard repository: %w", err)
	}
	jwtRepo, err := local.NewJWTRepository(boltDB)
	if err != nil {
		return nil, fmt.Errorf("create jwt repository: %w", err)
	}

	traced.Infow(domain.RuntimeTraceID, "Initializing services")
	userService := domain.NewUserService(userRpo, traced)

	traced.Infow(domain.RuntimeTraceID, "Initializing components")
	jwtProcessor := jwt.NewProcessor(conf.JWT)
	cookieProcessor := cookie.NewProcessor(jwtProcessor, conf.Cookie)

	traced.Infow(domain.RuntimeTraceID, "Creating router")
	h, err := handle.NewRouter(handle.Dependencies{
		Config:              conf,
		CookieProcessor:     cookieProcessor,
		UserService:         userService,
		JWTRepository:       jwtRepo,
		SessionRepository:   sessionRepo,
		ClipboardRepository: clipboardRepo,
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
			a.log.Infow(domain.RuntimeTraceID, "Shutting down server")
			if err := s.Shutdown(ctx); err != nil {
				a.log.Errorw(domain.RuntimeTraceID, "Shutdown server", err)
			}
			a.log.Infow(domain.RuntimeTraceID, "Server stopped")
			return
		}
	}()

	a.log.Infow(domain.RuntimeTraceID, "Starting server", "address", addr)
	if err := s.ListenAndServe(); !errors.Is(err, http.ErrServerClosed) {
		return fmt.Errorf("server listen: %w", err)
	}
	a.log.Infow(domain.RuntimeTraceID, "Server stopped")

	return nil
}
