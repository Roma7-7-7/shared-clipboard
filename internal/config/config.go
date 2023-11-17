package config

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"strings"
)

type (
	API struct {
		Dev    bool   `json:"dev"`
		Port   int    `json:"port"`
		CORS   CORS   `json:"cors"`
		Cookie Cookie `json:"cookie"`
		JWT    JWT    `json:"jwt"`
		DB     DB     `json:"db"`
	}

	Web struct {
		Dev             bool   `json:"dev"`
		Port            int    `json:"port"`
		StaticFilesPath string `json:"static_files_path"`
		APIHost         string `json:"api_host"`
	}

	Cookie struct {
		Path   string `json:"path"`
		Domain string `json:"domain"`
	}

	JWT struct {
		Issuer          string   `json:"issuer"`
		Audience        []string `json:"audience"`
		ExpireInMinutes uint64   `json:"expire_in_minutes"`
		Secret          string   `json:"secret"`
	}

	CORS struct {
		AllowOrigins     []string `json:"allow_origins"`
		AllowMethods     []string `json:"allow_methods"`
		AllowHeaders     []string `json:"allow_headers"`
		ExposeHeaders    []string `json:"expose_headers"`
		MaxAge           int      `json:"max_age"`
		AllowCredentials bool     `json:"allow_credentials"`
	}

	Bolt struct {
		Path string `json:"path"`
	}

	SQL struct {
		Driver     string `json:"driver"`
		DataSource string `json:"data_source"`
	}

	DB struct {
		Bolt Bolt `json:"bolt"`
		SQL  SQL  `json:"sql"`
	}
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
	res := make([]string, 0, 3)
	if conf.Port <= 0 || conf.Port > 65535 {
		return fmt.Errorf("invalid port: %d", conf.Port)
	}
	if conf.StaticFilesPath == "" {
		res = append(res, "empty static files path")
	}
	if conf.APIHost == "" {
		res = append(res, "empty api host")
	}

	if len(res) != 0 {
		return fmt.Errorf("invalid web config: [%s]", strings.Join(res, "; "))
	}

	return nil
}

func validateAPI(api API) error {
	res := make([]string, 0, 10)
	if api.Port <= 0 || api.Port > 65535 {
		return fmt.Errorf("invalid port: %d", api.Port)
	}
	if api.DB.Bolt.Path == "" {
		res = append(res, "empty data path")
	}
	if api.DB.SQL.Driver == "" {
		res = append(res, "empty sql driver")
	}
	if api.DB.SQL.DataSource == "" {
		res = append(res, "empty sql data source")
	}

	if len(res) != 0 {
		return fmt.Errorf("invalid api config: [%s]", strings.Join(res, "; "))
	}

	return nil
}
