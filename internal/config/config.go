package config

import (
	"flag"
	"fmt"

	"github.com/caarlos0/env/v9"
)

// AppConfig store configuration for http server, storage (file or database)
// and configurable variables for application.
type AppConfig struct {
	BaseHTTPAddr     string `env:"SERVER_ADDRESS"`
	BaseShortURLAddr string `env:"BASE_URL"`
	AppEnvironment   string `env:"APP_ENV"`
	FileStoragePath  string `env:"FILE_STORAGE_PATH"`
	DatabaseDSN      string `env:"DATABASE_DSN"`
}

// Available environments.
const (
	AppProductionEnv = "production"
	AppDevEnv        = "development"
)

// ParseFlags firstly parse flags and set defaults, after try to parse variables from environment variables.
func (appConfig *AppConfig) ParseFlags() {
	flag.StringVar(&appConfig.BaseHTTPAddr, "a", "localhost:8080", "Base http address that server running on")
	flag.StringVar(&appConfig.BaseShortURLAddr, "b", "http://localhost:8080", "Base short url address")
	flag.StringVar(&appConfig.FileStoragePath, "f", "/tmp/short-url-db.json", "Storage file path")
	flag.StringVar(&appConfig.DatabaseDSN, "d", "", "Database DSN")
	flag.Parse()

	if err := env.Parse(appConfig); err != nil {
		fmt.Printf("%+v\n", err)
	}
}
