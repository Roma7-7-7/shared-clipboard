package app

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"go.uber.org/zap"

	"github.com/Roma7-7-7/shared-clipboard/internal/config"
	"github.com/Roma7-7-7/shared-clipboard/internal/handler/web"
	"github.com/Roma7-7-7/shared-clipboard/tools/trace"
)

type App struct {
	conf config.Web
	mux  *chi.Mux
	log  trace.Logger
}

func New(ctx context.Context, conf config.Web, l *zap.SugaredLogger) (*App, error) {
	var (
		h   *chi.Mux
		err error
	)

	log := trace.NewSugaredLogger(l.With("service", "web"))

	log.Infow(ctx, "Creating router")
	if h, err = web.NewRouter(ctx, conf, log); err != nil {
		return nil, fmt.Errorf("create router: %w", err)
	}

	return &App{
		conf: conf,
		mux:  h,
		log:  log,
	}, nil
}

func (a *App) Run(ctx context.Context) error {
	done := make(chan struct{})
	defer close(done)

	addr := fmt.Sprintf(":%d", a.conf.Port)
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
			return
		}
	}()

	a.log.Infow(ctx, "Starting server", "address", addr)
	if err := s.ListenAndServe(); !errors.Is(err, http.ErrServerClosed) {
		return fmt.Errorf("server listen: %w", err)
	}

	return nil
}
