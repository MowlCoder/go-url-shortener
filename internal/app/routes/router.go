package routes

import (
	"net/http"

	"github.com/MowlCoder/go-url-shortener/internal/app/config"
	"github.com/MowlCoder/go-url-shortener/internal/app/handlers"
	"github.com/MowlCoder/go-url-shortener/internal/app/middlewares"
	"github.com/MowlCoder/go-url-shortener/internal/app/storage"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

type Logger interface {
	Info(msg string)
}

func InitRouter(config *config.AppConfig, logger Logger) *chi.Mux {
	r := chi.NewRouter()

	r.Use(middleware.RealIP)
	r.Use(middleware.Recoverer)
	r.Use(func(handler http.Handler) http.Handler {
		return middlewares.WithLogging(handler, logger)
	})

	urlStorage := storage.NewURLStorage()

	shortenerHandler := handlers.NewShortenerHandler(config, urlStorage)

	r.Post("/", shortenerHandler.ShortURL)
	r.Get("/{id}", shortenerHandler.RedirectToURLByID)

	return r
}
