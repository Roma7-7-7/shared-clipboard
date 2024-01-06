package main

import (
	"context"
	"flag"
	stdLog "log"
	"os"

	"go.uber.org/zap"

	"github.com/Roma7-7-7/shared-clipboard/cmd"
	"github.com/Roma7-7-7/shared-clipboard/internal/config"
	"github.com/Roma7-7-7/shared-clipboard/internal/domain"
	"github.com/Roma7-7-7/shared-clipboard/tools/log"
)

var configPath = flag.String("config", "web.json", "path to config file")

func main() {
	flag.Parse()

	var (
		conf config.Web
		l    *zap.Logger
		a    *cmd.Web
		err  error
	)

	if conf, err = config.NewWeb(*configPath); err != nil {
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
	traced := log.NewZapTracedLogger(l.Sugar())

	if a, err = cmd.NewWeb(conf, traced); err != nil {
		traced.Errorw(domain.RuntimeTraceID, "Create app", err)
		os.Exit(1)
	}

	if err = a.Run(context.Background()); err != nil {
		traced.Errorw(domain.RuntimeTraceID, "Run", err)
		os.Exit(1)
	}
}
