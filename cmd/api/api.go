package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	stdlog "log"
	"net/http"
	"os"
	"time"

	"github.com/dgraph-io/badger/v4"
	"github.com/labstack/echo/v4"
	"go.uber.org/zap"

	"github.com/Roma7-7-7/shared-clipboard/internal/config"
	"github.com/Roma7-7-7/shared-clipboard/internal/dal"
	"github.com/Roma7-7-7/shared-clipboard/internal/handler/api"
	"github.com/Roma7-7-7/shared-clipboard/tools/trace"
)

var dev = flag.Bool("dev", false, "development mode")
var port = flag.Int("port", 8080, "port to listen")
var dataPath = flag.String("data", "", "path to data directory")

func main() {
	flag.Parse()

	var (
		bootstrapCtx, cancel = context.WithTimeout(context.Background(), 1*time.Minute)
		l                    *zap.Logger
		log                  trace.Logger
		conf                 config.API
		db                   *badger.DB
		h                    *echo.Echo
		err                  error
	)
	defer cancel()
	bootstrapCtx = trace.WithTraceID(context.Background(), "bootstrap")

	if *dev {
		if l, err = zap.NewDevelopment(); err != nil {
			stdlog.Fatalf("create logger: %v", err)
		}
	} else {
		if l, err = zap.NewProduction(); err != nil {
			stdlog.Fatalf("create logger: %v", err)
		}
	}
	log = trace.NewSugaredLogger(l.Sugar())

	if conf, err = config.NewAPI(bootstrapCtx, *dev, *port, *dataPath, log); err != nil {
		log.Errorw(bootstrapCtx, "create config", err)
		os.Exit(1)
	}

	log.Infow(bootstrapCtx, "Initializing DB")
	if db, err = badger.Open(badger.DefaultOptions(conf.DB.Path)); err != nil {
		log.Errorw(bootstrapCtx, "open db", zap.Error(err))
		os.Exit(1)
	}

	log.Infow(bootstrapCtx, "Creating router")
	if h, err = api.NewRouter(bootstrapCtx, dal.NewSessionRepository(db), log); err != nil {
		log.Errorw(bootstrapCtx, "create router", err)
		os.Exit(1)
	}

	addr := fmt.Sprintf(":%d", conf.Port)
	s := http.Server{
		Addr:        addr,
		Handler:     h,
		ReadTimeout: 30 * time.Second,
	}
	log.Infow(bootstrapCtx, "Starting server", "address", addr)
	if err = s.ListenAndServe(); !errors.Is(err, http.ErrServerClosed) {
		log.Errorw(trace.WithTraceID(context.Background(), "termination"), "server listen error", err)
	}
}
