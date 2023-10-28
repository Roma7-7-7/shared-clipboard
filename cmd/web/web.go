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

	"github.com/labstack/echo/v4"
	"go.uber.org/zap"

	"github.com/Roma7-7-7/shared-clipboard/internal/config"
	"github.com/Roma7-7-7/shared-clipboard/internal/handler/web"
	"github.com/Roma7-7-7/shared-clipboard/tools/trace"
)

var dev = flag.Bool("dev", false, "development mode")
var port = flag.Int("port", 80, "port to listen")
var staticFilesPath = flag.String("static-files-path", "", "path to static files")
var apiHost = flag.String("api-host", "", "api host")

func main() {
	flag.Parse()

	var (
		bootstrapCtx, cancel = context.WithTimeout(context.Background(), 1*time.Minute)
		l                    *zap.Logger
		log                  trace.Logger
		conf                 config.Web
		h                    *echo.Echo
		err                  error
	)
	defer cancel()
	bootstrapCtx = trace.WithTraceID(bootstrapCtx, "bootstrap")

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

	if conf, err = config.NewWeb(bootstrapCtx, *dev, *port, *staticFilesPath, *apiHost, log); err != nil {
		log.Errorw(bootstrapCtx, "create config", err)
		os.Exit(1)
	}

	log.Infow(bootstrapCtx, "Creating router")
	if h, err = web.NewRouter(bootstrapCtx, conf, log); err != nil {
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
