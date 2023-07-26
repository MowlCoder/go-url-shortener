package main

import (
	"fmt"
	"github.com/MowlCoder/go-url-shortener/internal/app/routes"
	"net/http"
)

func main() {
	mux := routes.InitRouter()

	fmt.Println("URL Shortener server is running...")
	if err := http.ListenAndServe(`:8080`, mux); err != nil {
		panic(err)
	}
}
