package main

import (
	"fmt"
	"net/http"

	"github.com/MowlCoder/go-url-shortener/internal/app/config"
	"github.com/MowlCoder/go-url-shortener/internal/app/routes"
)

func main() {
	appConfig := &config.AppConfig{}
	appConfig.ParseFlags()

	mux := routes.InitRouter(appConfig)

	fmt.Println("URL Shortener server is running on", appConfig.BaseHTTPAddr)
	if err := http.ListenAndServe(appConfig.BaseHTTPAddr, mux); err != nil {
		panic(err)
	}
}
