package main

import (
	"errors"
	"flag"
	"fmt"
	stdlog "log"
	"net/http"
	"time"

	"github.com/labstack/echo/v4"
	"go.uber.org/zap"

	"github.com/Roma7-7-7/shared-clipboard/internal/config"
	"github.com/Roma7-7-7/shared-clipboard/internal/handler/web"
)

var dev = flag.Bool("dev", false, "development mode")
var port = flag.Int("port", 80, "port to listen")
var staticFilesPath = flag.String("static-files-path", "", "path to static files")
var apiHost = flag.String("api-host", "", "api host")

func main() {
	flag.Parse()

	var (
		l    *zap.Logger
		log  *zap.SugaredLogger
		conf config.Web
		h    *echo.Echo
		err  error
	)

	if *dev {
		if l, err = zap.NewDevelopment(); err != nil {
			stdlog.Fatalf("create logger: %v", err)
		}
	} else {
		if l, err = zap.NewProduction(); err != nil {
			stdlog.Fatalf("create logger: %v", err)
		}
	}
	log = l.Sugar()

	if conf, err = config.NewWeb(*dev, *port, *staticFilesPath, *apiHost, log); err != nil {
		log.Fatalw("create config", err)
	}

	log.Info("Creating router")
	if h, err = web.NewRouter(conf, log); err != nil {
		log.Fatalw("create router", err)
	}

	addr := fmt.Sprintf(":%d", conf.Port)
	s := http.Server{
		Addr:        addr,
		Handler:     h,
		ReadTimeout: 30 * time.Second,
	}
	log.Infow("Starting server", "address", addr)
	if err = s.ListenAndServe(); !errors.Is(err, http.ErrServerClosed) {
		log.Fatalf("server listen error: %s", err)
	}
}
