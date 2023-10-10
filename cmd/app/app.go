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
	"github.com/Roma7-7-7/shared-clipboard/internal/handler"
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

	conf := newConfig()
	log.Info("Initializing DB")
	if db, err = badger.Open(badger.DefaultOptions(conf.DB.Path)); err != nil {
		log.Fatal("open db", zap.Error(err))
	}

	log.Info("Initializing handlers")
	if h, err = handler.New(dal.NewSessionRepository(db), log); err != nil {
		l.Fatal("create handler", zap.Error(err))
	}

	s := http.Server{
		Addr:        fmt.Sprintf(":%d", conf.Server.Port),
		Handler:     h,
		ReadTimeout: 30 * time.Second,
	}
	if err = s.ListenAndServe(); !errors.Is(err, http.ErrServerClosed) {
		log.Fatal(err)
	}
}

func newConfig() config.Config {
	return config.Config{
		Server: config.Server{
			Port: 8080,
		},
		DB: config.DB{
			Path: "data/badger",
		},
	}
}
