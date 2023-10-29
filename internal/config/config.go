package config

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"strings"
)

func NewWeb(confPath string) (Web, error) {
	var res Web
	if err := readConfig(confPath, &res); err != nil {
		return res, fmt.Errorf("read config: %w", err)
	}
	return res, validateWeb(res)
}

func NewAPI(confPath string) (API, error) {
	var api API
	if err := readConfig(confPath, &api); err != nil {
		return api, fmt.Errorf("read config: %w", err)
	}

	return api, validateAPI(api)
}

type Web struct {
	Dev             bool   `json:"dev"`
	Port            int    `json:"port"`
	StaticFilesPath string `json:"static_files_path"`
	APIHost         string `json:"api_host"`
}

type API struct {
	Dev  bool `json:"dev"`
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
