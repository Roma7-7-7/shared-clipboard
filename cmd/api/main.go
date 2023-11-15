package main

import (
	"context"
	"flag"
	stdLog "log"
	"os"

	"go.uber.org/zap"

	"github.com/Roma7-7-7/shared-clipboard/cmd"
	"github.com/Roma7-7-7/shared-clipboard/internal/config"
	"github.com/Roma7-7-7/shared-clipboard/tools/log"
	"github.com/Roma7-7-7/shared-clipboard/tools/trace"
)

var configPath = flag.String("config", "", "path to config file")

func main() {
	flag.Parse()

	var (
		conf config.API
		l    *zap.Logger
		a    *cmd.API
		err  error
	)

	if conf, err = config.NewAPI(*configPath); err != nil {
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

	if a, err = cmd.NewAPI(conf, traced); err != nil {
		traced.Errorw(trace.RuntimeTraceID, "Create app", err)
		os.Exit(1)
	}

	if err = a.Run(context.Background()); err != nil {
		traced.Errorw(trace.RuntimeTraceID, "Run", err)
		os.Exit(1)
	}
}
