package routes

import (
	"github.com/MowlCoder/go-url-shortener/internal/app/handlers"
	"github.com/go-chi/chi/v5"
)

func InitRouter() *chi.Mux {
	r := chi.NewRouter()

	r.Post("/", handlers.ShortURL)
	r.Get("/{id}", handlers.RedirectToURLByID)

	return r
}
