package main

import (
	"context"
	"flag"
	stdLog "log"
	"time"

	"go.uber.org/zap"

	"github.com/Roma7-7-7/shared-clipboard/cmd/web/app"
	"github.com/Roma7-7-7/shared-clipboard/internal/config"
	"github.com/Roma7-7-7/shared-clipboard/tools/trace"
)

var configPath = flag.String("config", "web.json", "path to config file")

func main() {
	flag.Parse()

	var (
		ctx, cancel = context.WithTimeout(context.Background(), 1*time.Minute)
		conf        config.Web
		l           *zap.Logger
		a           *app.App
		err         error
	)
	defer cancel()
	ctx = trace.WithTraceID(ctx, "bootstrap")

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

	if a, err = app.New(ctx, conf, l.Sugar()); err != nil {
		stdLog.Fatalf("create app: %v", err)
	}

	runCtx, runCancel := context.WithCancel(ctx)
	defer runCancel()
	if err = a.Run(trace.WithTraceID(runCtx, "run")); err != nil {
		stdLog.Fatalf("start web: %s", err)
	}
}
