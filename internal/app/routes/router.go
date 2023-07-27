package routes

import (
	"github.com/MowlCoder/go-url-shortener/internal/app/handlers"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

func InitRouter() *chi.Mux {
	r := chi.NewRouter()

	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	r.Post("/", handlers.ShortURL)
	r.Get("/{id}", handlers.RedirectToURLByID)

	return r
}
