package cmd

import (
	"fmt"

	"github.com/dgraph-io/badger/v4"
	"github.com/go-chi/chi/v5"

	"github.com/Roma7-7-7/shared-clipboard/internal/config"
	"github.com/Roma7-7-7/shared-clipboard/internal/dal"
	"github.com/Roma7-7-7/shared-clipboard/internal/handler/api"
	"github.com/Roma7-7-7/shared-clipboard/internal/handler/web"
	"github.com/Roma7-7-7/shared-clipboard/tools/app"
	"github.com/Roma7-7-7/shared-clipboard/tools/log"
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

func NewAPI(conf config.API, traced log.TracedLogger, badgerLog badger.Logger) (*API, error) {
	var (
		db  *badger.DB
		h   *chi.Mux
		err error
	)

	traced.Infow(trace.RuntimeTraceID, "Initializing DB")
	badgerOpts := badger.DefaultOptions(conf.DB.Path)
	badgerOpts.Logger = badgerLog
	if db, err = badger.Open(badgerOpts); err != nil {
		return nil, fmt.Errorf("open db: %w", err)
	}

	traced.Infow(trace.RuntimeTraceID, "Creating router")
	badgerRepo, err := dal.NewSessionRepository(db)
	if err != nil {
		return nil, fmt.Errorf("create repository: %w", err)
	}
	if h, err = api.NewRouter(badgerRepo, conf, traced); err != nil {
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

	l.Infow(trace.RuntimeTraceID, "Creating router")
	if h, err = web.NewRouter(conf, l); err != nil {
		return nil, fmt.Errorf("create router: %w", err)
	}

	return &Web{
		app.New(conf.Port, h, l),
	}, nil
}
