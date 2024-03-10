package config

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"strings"
)

type (
	App struct {
		Dev    bool   `json:"dev"`
		Port   int    `json:"port"`
		CORS   CORS   `json:"cors"`
		Cookie Cookie `json:"cookie"`
		JWT    JWT    `json:"jwt"`
		DB     DB     `json:"db"`
		Redis  Redis  `json:"redis"`
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

	DB struct {
		Driver     string `json:"driver"`
		DataSource string `json:"data_source"`
	}

	Redis struct {
		Addr          string `json:"addr"`
		Password      string `json:"password"`
		DB            int    `json:"db"`
		TimeoutMillis int    `json:"timeout_millis"`
	}
)

func NewApp(confPath string) (App, error) {
	var app App
	if err := readConfig(confPath, &app); err != nil {
		return app, fmt.Errorf("read config: %w", err)
	}

	return app, validateApp(app)
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

func validateApp(app App) error {
	res := make([]string, 0, 10)
	if app.Port <= 0 || app.Port > 65535 {
		return fmt.Errorf("invalid port: %d", app.Port)
	}
	if app.Redis.Addr == "" {
		res = append(res, "empty redis addr")
	}
	if app.DB.Driver == "" {
		res = append(res, "empty DB driver")
	}
	if app.DB.DataSource == "" {
		res = append(res, "empty DB data source")
	}

	if len(res) != 0 {
		return fmt.Errorf("invalid app config: [%s]", strings.Join(res, "; "))
	}

	return nil
}
