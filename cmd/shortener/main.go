package main

import (
	"fmt"
	"net/http"

	"github.com/MowlCoder/go-url-shortener/internal/app/config"
	"github.com/MowlCoder/go-url-shortener/internal/app/routes"
)

func main() {
	config.BaseConfig.ParseFlags()

	mux := routes.InitRouter()

	fmt.Println("URL Shortener server is running on", config.BaseConfig.BaseHTTPAddr)
	if err := http.ListenAndServe(config.BaseConfig.BaseHTTPAddr, mux); err != nil {
		panic(err)
	}
}
