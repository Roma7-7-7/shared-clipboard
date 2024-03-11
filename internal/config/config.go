package config

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/kelseyhightower/envconfig"
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
		Domain string `json:"domain" envconfig:"APP_COOKIE_DOMAIN"`
	}

	JWT struct {
		Issuer          string   `json:"issuer"`
		Audience        []string `json:"audience" envconfig:"APP_JWT_AUDIENCE"`
		ExpireInMinutes uint64   `json:"expire_in_minutes"`
		Secret          string   `json:"secret"`
	}

	CORS struct {
		AllowOrigins     []string `json:"allow_origins" envconfig:"APP_CORS_ALLOW_ORIGINS"`
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
		Driver   string `json:"driver"`
		Host     string `json:"host" envconfig:"APP_DB_HOST"`
		Port     int    `json:"port"`
		Name     string `json:"name"`
		User     string `json:"user"`
		Password string `json:"password"`
		SSLMode  string `json:"ssl_mode"`
	}

	Redis struct {
		Addr          string `json:"addr" envconfig:"APP_REDIS_ADDR"`
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

	if err := envconfig.Process("app", &app); err != nil {
		return app, fmt.Errorf("process env: %w", err)
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
	if app.DB.Host == "" {
		res = append(res, "empty DB host")
	}
	if app.DB.Port <= 0 || app.DB.Port > 65535 {
		res = append(res, "invalid DB port")
	}
	if app.DB.Name == "" {
		res = append(res, "empty DB name")
	}
	if app.DB.User == "" {
		res = append(res, "empty DB user")
	}
	if app.DB.Password == "" {
		res = append(res, "empty DB password")
	}
	if app.DB.SSLMode == "" {
		res = append(res, "empty DB ssl mode")
	}

	if len(res) != 0 {
		return fmt.Errorf("invalid app config: [%s]", strings.Join(res, "; "))
	}

	return nil
}
