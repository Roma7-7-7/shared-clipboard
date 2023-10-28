package main

import (
	"errors"
	"fmt"
	stdlog "log"
	"net/http"
	"time"

	"github.com/dgraph-io/badger/v4"
	"github.com/labstack/echo/v4"
	"go.uber.org/zap"

	"github.com/Roma7-7-7/shared-clipboard/internal/config"
	"github.com/Roma7-7-7/shared-clipboard/internal/dal"
	"github.com/Roma7-7-7/shared-clipboard/internal/handler/api"
)

func main() {
	var (
		l   *zap.Logger
		log *zap.SugaredLogger
		db  *badger.DB
		h   *echo.Echo
		err error
	)

	if l, err = zap.NewDevelopment(); err != nil {
		stdlog.Fatalf("create logger: %v", err)
	}
	log = l.Sugar()

	conf := config.New().API
	log.Info("Initializing DB")
	if db, err = badger.Open(badger.DefaultOptions(conf.DB.Path)); err != nil {
		log.Fatal("open db", zap.Error(err))
	}

	log.Info("Creating router")
	if h, err = api.NewAPIRouter(dal.NewSessionRepository(db), log); err != nil {
		log.Fatalw("create router", err)
	}

	addr := fmt.Sprintf(":%d", conf.Port)
	s := http.Server{
		Addr:        addr,
		Handler:     h,
		ReadTimeout: 30 * time.Second,
	}
	log.Infof("Starting server on address=%s", addr)
	if err = s.ListenAndServe(); !errors.Is(err, http.ErrServerClosed) {
		log.Fatalf("server listen error: %s", err)
	}
}
