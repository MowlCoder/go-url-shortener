package routes

import (
	"github.com/MowlCoder/go-url-shortener/internal/app/config"
	"github.com/MowlCoder/go-url-shortener/internal/app/handlers"
	"github.com/MowlCoder/go-url-shortener/internal/app/storage"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

func InitRouter(config *config.AppConfig) *chi.Mux {
	r := chi.NewRouter()

	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	urlStorage := storage.NewURLStorage()

	shortenerHandler := handlers.NewShortenerHandler(config, urlStorage)

	r.Post("/", shortenerHandler.ShortURL)
	r.Get("/{id}", shortenerHandler.RedirectToURLByID)

	return r
}
