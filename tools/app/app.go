package app

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/pkg/errors"

	"github.com/Roma7-7-7/shared-clipboard/tools/log"
	"github.com/Roma7-7-7/shared-clipboard/tools/trace"
)

type App struct {
	port int
	mux  *chi.Mux
	log  log.TracedLogger
}

func New(port int, mux *chi.Mux, log log.TracedLogger) *App {
	return &App{
		port: port,
		mux:  mux,
		log:  log,
	}
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
			a.log.Infow(trace.RuntimeTraceID, "Shutting down server")
			if err := s.Shutdown(ctx); err != nil {
				a.log.Errorw(trace.RuntimeTraceID, "Shutdown server", err)
			}
			a.log.Infow(trace.RuntimeTraceID, "Server stopped")
			return
		}
	}()

	a.log.Infow(trace.RuntimeTraceID, "Starting server", "address", addr)
	if err := s.ListenAndServe(); !errors.Is(err, http.ErrServerClosed) {
		return fmt.Errorf("server listen: %w", err)
	}
	a.log.Infow(trace.RuntimeTraceID, "Server stopped")

	return nil
}
