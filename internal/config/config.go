package config

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/caarlos0/env/v9"
)

// AppConfig store configuration for http server, storage (file or database)
// and configurable variables for application.
type AppConfig struct {
	BaseHTTPAddr     string `env:"SERVER_ADDRESS" json:"base_http_addr"`
	BaseShortURLAddr string `env:"BASE_URL" json:"base_url"`
	AppEnvironment   string `env:"APP_ENV" json:"app_env"`
	FileStoragePath  string `env:"FILE_STORAGE_PATH" json:"file_storage_path"`
	DatabaseDSN      string `env:"DATABASE_DSN" json:"database_dsn"`
	EnableHTTPS      bool   `env:"ENABLE_HTTPS" json:"enable_https"`
	SSLKeyPath       string `env:"SSL_KEY_PATH" json:"ssl_key_path"`
	SSLPemPath       string `env:"SSL_PEM_PATH" json:"ssl_pem_path"`
	TrustedSubnet    string `env:"TRUSTED_SUBNET" json:"trusted_subnet"`
}

// Available environments.
const (
	AppProductionEnv = "production"
	AppDevEnv        = "development"
)

// ParseFlags firstly parse flags and set defaults, after try to parse variables from environment variables.
func (appConfig *AppConfig) ParseFlags() {
	var configPath string
	flag.StringVar(&configPath, "c", "", "Path to config file")
	flag.StringVar(&appConfig.BaseHTTPAddr, "a", "localhost:8080", "Base http address that server running on")
	flag.StringVar(&appConfig.BaseShortURLAddr, "b", "http://localhost:8080", "Base short url address")
	flag.StringVar(&appConfig.FileStoragePath, "f", "/tmp/short-url-db.json", "Storage file path")
	flag.StringVar(&appConfig.DatabaseDSN, "d", "", "Database DSN")
	flag.BoolVar(&appConfig.EnableHTTPS, "s", false, "Enable HTTPS")
	flag.StringVar(&appConfig.SSLKeyPath, "sslk", "./certs/server.key", "Path to ssl key file")
	flag.StringVar(&appConfig.SSLPemPath, "sslp", "./certs/server.pem", "Path to ssl pem file")
	flag.StringVar(&appConfig.TrustedSubnet, "t", "", "Trusted subnet in CIDR format")
	flag.Parse()

	if configPathFromEnv, ok := os.LookupEnv("CONFIG"); ok {
		configPath = configPathFromEnv
	}

	if configPath != "" {
		rawContent, err := os.ReadFile(configPath)
		if err != nil {
			log.Fatal(err)
		}

		if err := json.Unmarshal(rawContent, appConfig); err != nil {
			log.Fatal(err)
		}
	}

	flag.Parse()

	if err := env.Parse(appConfig); err != nil {
		fmt.Printf("%+v\n", err)
	}
}
