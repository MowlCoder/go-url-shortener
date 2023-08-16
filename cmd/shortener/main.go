package main

import (
	"fmt"
	"net/http"

	"github.com/MowlCoder/go-url-shortener/internal/app/logger"

	"github.com/MowlCoder/go-url-shortener/internal/app/config"
	"github.com/MowlCoder/go-url-shortener/internal/app/routes"
)

func main() {
	appConfig := &config.AppConfig{}
	appConfig.ParseFlags()

	customLogger, err := logger.NewLogger(logger.Options{
		Level:        logger.LogInfo,
		IsProduction: appConfig.AppEnvironment == config.AppProductionEnv,
	})

	if err != nil {
		panic(err)
	}

	mux := routes.InitRouter(appConfig, customLogger)

	fmt.Println("URL Shortener server is running on", appConfig.BaseHTTPAddr)

	if err := http.ListenAndServe(appConfig.BaseHTTPAddr, mux); err != nil {
		panic(err)
	}
}
