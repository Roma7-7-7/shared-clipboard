package config

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/Roma7-7-7/shared-clipboard/tools/trace"
)

func NewWeb(ctx context.Context, dev bool, confPath string, log trace.Logger) (Web, error) {
	var res Web
	if err := readConfig(confPath, &res); err != nil {
		return res, fmt.Errorf("read config: %w", err)
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

func NewAPI(ctx context.Context, dev bool, confPath string, log trace.Logger) (API, error) {
	var api API
	if err := readConfig(confPath, &api); err != nil {
		return api, fmt.Errorf("read config: %w", err)
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
	Port            int    `json:"port"`
	StaticFilesPath string `json:"static_files_path"`
	APIHost         string `json:"api_host"`
}

type API struct {
	Port int  `json:"port"`
	CORS CORS `json:"cors"`
	DB   DB   `json:"db"`
}

type CORS struct {
	AllowOrigin      string   `json:"allow_origin"`
	AllowMethods     []string `json:"allow_methods"`
	AllowHeaders     []string `json:"allow_headers"`
	ExposeHeaders    []string `json:"expose_headers"`
	MaxAge           int      `json:"max_age"`
	AllowCredentials bool     `json:"allow_credentials"`
}

type DB struct {
	Path string `yaml:"path"`
}

func readConfig(path string, target any) error {
	if path == "" {
		return errors.New("empty path")
	}

	open, err := os.Open(path)
	if err != nil {
		return err
	}
	defer func(open *os.File) {
		_ = open.Close()
	}(open)

	if err = json.NewDecoder(open).Decode(target); err != nil {
		return fmt.Errorf("decode config file with path=\"%s\": %w", path, err)
	}

	return nil
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
