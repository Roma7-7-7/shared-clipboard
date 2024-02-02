package main

import (
	"context"
	"flag"
	stdLog "log"
	"os"

	"go.uber.org/zap"

	"github.com/Roma7-7-7/shared-clipboard/internal/app"
	"github.com/Roma7-7-7/shared-clipboard/internal/config"
	"github.com/Roma7-7-7/shared-clipboard/internal/domain"
	"github.com/Roma7-7-7/shared-clipboard/tools/log"
)

var configPath = flag.String("config", "", "path to config file")

func main() {
	flag.Parse()

	var (
		conf config.App
		l    *zap.Logger
		a    *app.App
		err  error
	)

	if conf, err = config.NewApp(*configPath); err != nil {
		stdLog.Fatalf("create config: %v", err)
	}

	if conf.Dev {
		if l, err = zap.NewDevelopment(); err != nil {
			stdLog.Fatalf("create logger: %s", err)
		}
	} else {
		if l, err = zap.NewProduction(); err != nil {
			stdLog.Fatalf("create logger: %s", err)
		}
	}
	sLog := l.Sugar()
	traced := log.NewZapTracedLogger(sLog)

	if a, err = app.NewApp(conf, traced); err != nil {
		traced.Errorw(domain.RuntimeTraceID, "Create app", err)
		os.Exit(1)
	}

	if err = a.Run(context.Background()); err != nil {
		traced.Errorw(domain.RuntimeTraceID, "Run", err)
		os.Exit(1)
	}
}
