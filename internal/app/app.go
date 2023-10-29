package app

import (
	"context"
	"fmt"

	"github.com/dgraph-io/badger/v4"
	"github.com/go-chi/chi/v5"
	"go.uber.org/zap"

	"github.com/Roma7-7-7/shared-clipboard/internal/config"
	"github.com/Roma7-7-7/shared-clipboard/internal/dal"
	"github.com/Roma7-7-7/shared-clipboard/internal/handler/api"
	"github.com/Roma7-7-7/shared-clipboard/internal/handler/web"
	"github.com/Roma7-7-7/shared-clipboard/tools/app"
	"github.com/Roma7-7-7/shared-clipboard/tools/trace"
)

type (
	API struct {
		*app.App
	}
	Web struct {
		*app.App
	}
)

func NewAPI(ctx context.Context, conf config.API, l *zap.SugaredLogger) (*API, error) {
	var (
		db  *badger.DB
		h   *chi.Mux
		err error
	)

	log := trace.NewSugaredLogger(l.With("service", "api"))

	log.Infow(ctx, "Initializing DB")
	badgerOpts := badger.DefaultOptions(conf.DB.Path)
	badgerOpts.Logger = trace.NewBadgerLogger(log)
	if db, err = badger.Open(badgerOpts); err != nil {
		return nil, fmt.Errorf("open db: %w", err)
	}

	log.Infow(ctx, "Creating router")
	if h, err = api.NewRouter(ctx, dal.NewSessionRepository(db), conf, log); err != nil {
		return nil, fmt.Errorf("create router: %w", err)
	}

	return &API{
		app.New(conf.Port, h, log),
	}, nil
}

func NewWeb(ctx context.Context, conf config.Web, l *zap.SugaredLogger) (*Web, error) {
	var (
		h   *chi.Mux
		err error
	)

	log := trace.NewSugaredLogger(l.With("service", "web"))

	log.Infow(ctx, "Creating router")
	if h, err = web.NewRouter(ctx, conf, log); err != nil {
		return nil, fmt.Errorf("create router: %w", err)
	}

	return &Web{
		app.New(conf.Port, h, log),
	}, nil
}
