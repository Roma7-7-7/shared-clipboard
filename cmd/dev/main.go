package main

import (
	"context"
	"flag"
	stdLog "log"
	"sync"
	"time"

	"go.uber.org/zap"

	"github.com/Roma7-7-7/shared-clipboard/internal/app"
	"github.com/Roma7-7-7/shared-clipboard/internal/config"
	"github.com/Roma7-7-7/shared-clipboard/tools/trace"
)

var apiConfigPath = flag.String("api-config", "configs/api.yaml", "path to api config")
var webConfigPath = flag.String("web-config", "configs/web.yaml", "path to web config")

func main() {
	flag.Parse()

	var (
		apiConf config.API
		api     *app.API
		webConf config.Web
		web     *app.Web
		l       *zap.Logger
		sLog    *zap.SugaredLogger
		err     error
	)

	bootstrapCtx, bootstrapCancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer bootstrapCancel()
	bootstrapCtx = trace.WithTraceID(bootstrapCtx, "bootstrap")

	if apiConf, err = config.NewAPI(*apiConfigPath); err != nil {
		stdLog.Fatalf("create api config: %v", err)
	}

	if webConf, err = config.NewWeb(*webConfigPath); err != nil {
		stdLog.Fatalf("create web config: %v", err)
	}

	stdLog.Println("Overriding configs for dev mode")
	apiConf.Dev = true
	webConf.Dev = true

	if l, err = zap.NewDevelopment(); err != nil {
		stdLog.Fatalf("create logger: %s", err)
	}
	sLog = l.Sugar()
	apiLog := trace.NewSugaredLogger(sLog.With("service", "api"))
	webLog := trace.NewSugaredLogger(sLog.With("service", "web"))

	if api, err = app.NewAPI(bootstrapCtx, apiConf, apiLog); err != nil {
		sLog.Fatalf("failed to create api app: %s", err)
	}
	if web, err = app.NewWeb(bootstrapCtx, webConf, webLog); err != nil {
		sLog.Fatalf("failed to create web app: %s", err)
	}

	runCtx, runCancel := context.WithCancel(context.Background())
	defer runCancel()
	runCtx = trace.WithTraceID(runCtx, "run")

	var wg sync.WaitGroup
	wg.Add(2)
	go func() {
		defer wg.Done()
		defer runCancel()
		if err := api.Run(runCtx); err != nil {
			sLog.Errorw("API run failed", err)
		}
	}()
	go func() {
		defer wg.Done()
		defer runCancel()
		if err := web.Run(runCtx); err != nil {
			sLog.Errorw("Web run failed", err)
		}
	}()

	wg.Wait()
	sLog.Infow("All apps stopped")
}
