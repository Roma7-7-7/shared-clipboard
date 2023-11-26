package cmd

import (
	"database/sql"
	"fmt"

	"github.com/go-chi/chi/v5"
	bolt "go.etcd.io/bbolt"

	"github.com/Roma7-7-7/shared-clipboard/internal/config"
	"github.com/Roma7-7-7/shared-clipboard/internal/dal"
	"github.com/Roma7-7-7/shared-clipboard/internal/domain"
	"github.com/Roma7-7-7/shared-clipboard/internal/handler/api"
	"github.com/Roma7-7-7/shared-clipboard/internal/handler/api/cookie"
	"github.com/Roma7-7-7/shared-clipboard/internal/handler/api/jwt"
	"github.com/Roma7-7-7/shared-clipboard/internal/handler/web"
	"github.com/Roma7-7-7/shared-clipboard/tools/app"
	"github.com/Roma7-7-7/shared-clipboard/tools/log"
)

type (
	API struct {
		*app.App
	}
	Web struct {
		*app.App
	}
)

func NewAPI(conf config.API, traced log.TracedLogger) (*API, error) {
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
	userRpo, err := dal.NewUserRepository(sqlDB)
	if err != nil {
		return nil, fmt.Errorf("create user repository: %w", err)
	}
	sessionRepo, err := dal.NewSessionRepository(boltDB)
	if err != nil {
		return nil, fmt.Errorf("create session repository: %w", err)
	}
	clipboardRepo, err := dal.NewClipboardRepository(boltDB)
	if err != nil {
		return nil, fmt.Errorf("create clipboard repository: %w", err)
	}
	jwtRepo, err := dal.NewJWTRepository(boltDB)
	if err != nil {
		return nil, fmt.Errorf("create jwt repository: %w", err)
	}

	traced.Infow(domain.RuntimeTraceID, "Initializing services")
	userService := domain.NewUserService(userRpo, traced)

	traced.Infow(domain.RuntimeTraceID, "Initializing components")
	jwtProcessor := jwt.NewProcessor(conf.JWT)
	cookieProcessor := cookie.NewProcessor(jwtProcessor, conf.Cookie)

	traced.Infow(domain.RuntimeTraceID, "Creating router")
	h, err := api.NewRouter(api.Dependencies{
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

	return &API{
		app.New(conf.Port, h, traced),
	}, nil
}

func NewWeb(conf config.Web, l log.TracedLogger) (*Web, error) {
	var (
		h   *chi.Mux
		err error
	)

	l.Infow(domain.RuntimeTraceID, "Creating router")
	if h, err = web.NewRouter(conf, l); err != nil {
		return nil, fmt.Errorf("create router: %w", err)
	}

	return &Web{
		app.New(conf.Port, h, l),
	}, nil
}
