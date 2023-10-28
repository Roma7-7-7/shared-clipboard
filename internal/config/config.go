package config

import (
	"fmt"
	"strings"

	"go.uber.org/zap"
)

func NewWeb(dev bool, port int, staticFilesPath, apiHost string, log *zap.SugaredLogger) (Web, error) {
	res := Web{
		Port:            port,
		StaticFilesPath: staticFilesPath,
		APIHost:         apiHost,
	}
	if dev {
		if res.StaticFilesPath == "" {
			res.StaticFilesPath = "web"
			log.Debugw("dev mode is on, setting default static files path", "path", res.StaticFilesPath)
		}
		if res.APIHost == "" {
			log.Debugw("dev mode is on, setting default api host", "host", "http://localhost:8080")
			res.APIHost = "http://localhost:8080"
		}
	}
	return res, validateWeb(res)
}

func New() Config {
	return Config{
		Web: Web{
			Port:    80,
			APIHost: "",
		},
		API: API{
			Port: 8080,
			DB: DB{
				Path: "data/badger",
			},
		},
	}
}

type Config struct {
	Dev bool
	Web
	API
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
