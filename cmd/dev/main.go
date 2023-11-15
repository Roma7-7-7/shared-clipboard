package main

import (
	"context"
	"flag"
	stdLog "log"
	"os"
	"sync"

	"go.uber.org/zap"

	"github.com/Roma7-7-7/shared-clipboard/cmd"
	"github.com/Roma7-7-7/shared-clipboard/internal/config"
	"github.com/Roma7-7-7/shared-clipboard/tools/log"
	"github.com/Roma7-7-7/shared-clipboard/tools/trace"
)

var apiConfigPath = flag.String("api-config", "configs/api.yaml", "path to api config")
var webConfigPath = flag.String("web-config", "configs/web.yaml", "path to web config")

func main() {
	flag.Parse()

	var (
		apiConf config.API
		api     *cmd.API
		webConf config.Web
		web     *cmd.Web
		l       *zap.Logger
		sLog    *zap.SugaredLogger
		err     error
	)

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
	traced := log.NewZapTracedLogger(sLog)
	if api, err = cmd.NewAPI(apiConf, log.NewZapTracedLogger(sLog.With("service", "api"))); err != nil {
		traced.Errorw(trace.RuntimeTraceID, "failed to create api app: %s", err)
		os.Exit(1)
	}
	if web, err = cmd.NewWeb(webConf, log.NewZapTracedLogger(sLog.With("service", "web"))); err != nil {
		traced.Errorw(trace.RuntimeTraceID, "failed to create web app: %s", err)
		os.Exit(1)
	}

	runCtx, runCancel := context.WithCancel(context.Background())
	defer runCancel()

	var wg sync.WaitGroup
	wg.Add(2)
	go func() {
		defer wg.Done()
		defer runCancel()
		if err := api.Run(runCtx); err != nil {
			traced.Errorw(trace.RuntimeTraceID, "API run failed", err)
		}
	}()
	go func() {
		defer wg.Done()
		defer runCancel()
		if err := web.Run(runCtx); err != nil {
			traced.Errorw(trace.RuntimeTraceID, "Web run failed", err)
		}
	}()

	wg.Wait()
	traced.Infow(trace.RuntimeTraceID, "All apps stopped")
}
