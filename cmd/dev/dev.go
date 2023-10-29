package main

import (
	"context"
	"flag"
	stdLog "log"
	"sync"
	"time"

	"go.uber.org/zap"

	apiApp "github.com/Roma7-7-7/shared-clipboard/cmd/api/app"
	webApp "github.com/Roma7-7-7/shared-clipboard/cmd/web/app"
	"github.com/Roma7-7-7/shared-clipboard/internal/config"
	"github.com/Roma7-7-7/shared-clipboard/tools/trace"
)

var apiConfigPath = flag.String("api-config", "configs/api.yaml", "path to api config")
var webConfigPath = flag.String("web-config", "configs/web.yaml", "path to web config")

func main() {
	flag.Parse()

	var (
		apiConf config.API
		api     *apiApp.App
		webConf config.Web
		web     *webApp.App
		l       *zap.Logger
		log     *zap.SugaredLogger
		err     error
	)

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()
	ctx = trace.WithTraceID(ctx, "bootstrap")

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
	log = l.Sugar()

	if api, err = apiApp.New(ctx, apiConf, log); err != nil {
		log.Fatalf("failed to create api app: %s", err)
	}
	if web, err = webApp.New(ctx, webConf, log); err != nil {
		log.Fatalf("failed to create web app: %s", err)
	}

	runCtx, cancelRun := context.WithCancel(context.Background())
	defer cancelRun()
	runCtx = trace.WithTraceID(runCtx, "run")

	var wg sync.WaitGroup
	wg.Add(2)
	go func() {
		defer wg.Done()
		if err := api.Run(runCtx); err != nil {
			log.Errorw("api app failed", err)
			cancelRun()
		}
	}()
	go func() {
		defer wg.Done()
		if err := web.Run(runCtx); err != nil {
			log.Errorw("web app failed", err)
			cancelRun()
		}
	}()

	wg.Wait()
}
