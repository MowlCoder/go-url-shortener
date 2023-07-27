package config

import "flag"

type AppConfig struct {
	BaseHTTPAddr     string
	BaseShortURLAddr string
}

func (appConfig *AppConfig) ParseFlags() {
	flag.StringVar(&appConfig.BaseHTTPAddr, "a", "localhost:8080", "Base http address that server running on")
	flag.StringVar(&appConfig.BaseShortURLAddr, "b", "localhost:8080", "Base short url address")
	flag.Parse()
}

var BaseConfig = new(AppConfig)
