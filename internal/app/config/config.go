package config

import (
	"flag"
	"fmt"

	"github.com/caarlos0/env/v9"
)

type AppConfig struct {
	BaseHTTPAddr     string `env:"SERVER_ADDRESS"`
	BaseShortURLAddr string `env:"BASE_URL"`
	AppEnvironment   string `env:"APP_ENV"`
}

const (
	AppProductionEnv = "production"
	AppDevEnv        = "development"
)

func (appConfig *AppConfig) ParseFlags() {
	flag.StringVar(&appConfig.BaseHTTPAddr, "a", "localhost:8080", "Base http address that server running on")
	flag.StringVar(&appConfig.BaseShortURLAddr, "b", "http://localhost:8080", "Base short url address")
	flag.Parse()

	if err := env.Parse(appConfig); err != nil {
		fmt.Printf("%+v\n", err)
	}
}
