package config

import (
	"context"
	"fmt"
	"strings"

	"github.com/Roma7-7-7/shared-clipboard/tools/trace"
)

func NewWeb(ctx context.Context, dev bool, port int, staticFilesPath, apiHost string, log trace.Logger) (Web, error) {
	res := Web{
		Port:            port,
		StaticFilesPath: staticFilesPath,
		APIHost:         apiHost,
	}
	if dev {
		if res.StaticFilesPath == "" {
			res.StaticFilesPath = "web"
			log.Debugw(ctx, "dev mode is on, setting default static files path", "path", res.StaticFilesPath)
		}
		if res.APIHost == "" {
			log.Debugw(ctx, "dev mode is on, setting default api host", "host", "http://localhost:8080")
			res.APIHost = "http://localhost:8080"
		}
	}
	return res, validateWeb(res)
}

func NewAPI(ctx context.Context, dev bool, port int, dataPath string, log trace.Logger) (API, error) {
	api := API{
		Port: port,
		DB: DB{
			Path: dataPath,
		},
	}
	if dev {
		if api.Port == 0 {
			api.Port = 8080
			log.Debugw(ctx, "dev mode is on, setting default port", "port", api.Port)
		}
		if api.DB.Path == "" {
			api.DB.Path = "data"
			log.Debugw(ctx, "dev mode is on, setting default data path", "path", api.DB.Path)
		}
	}
	return api, validateAPI(api)
}

type Web struct {
	Port            int
	StaticFilesPath string
	APIHost         string
}

type API struct {
	Port int
	DB   DB
}

type DB struct {
	Path string
}

func validateWeb(conf Web) error {
	errors := make([]string, 0, 3)
	if conf.Port <= 0 || conf.Port > 65535 {
		return fmt.Errorf("invalid port: %d", conf.Port)
	}
	if conf.StaticFilesPath == "" {
		errors = append(errors, "empty static files path")
	}
	if conf.APIHost == "" {
		errors = append(errors, "empty api host")
	}

	if len(errors) != 0 {
		return fmt.Errorf("invalid web config: [%s]", strings.Join(errors, "; "))
	}

	return nil
}

func validateAPI(api API) error {
	errors := make([]string, 0, 2)
	if api.Port <= 0 || api.Port > 65535 {
		return fmt.Errorf("invalid port: %d", api.Port)
	}
	if api.DB.Path == "" {
		errors = append(errors, "empty data path")
	}

	if len(errors) != 0 {
		return fmt.Errorf("invalid api config: [%s]", strings.Join(errors, "; "))
	}

	return nil
}
