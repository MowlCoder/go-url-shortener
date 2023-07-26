package routes

import (
	"github.com/MowlCoder/go-url-shortener/internal/app/handlers"
	"net/http"
)

func InitRouter() *http.ServeMux {
	mux := http.NewServeMux()
	mux.HandleFunc("/", handlers.HandleShortUrl)

	return mux
}
