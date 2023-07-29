package handlers

import (
	"fmt"
	"github.com/go-chi/chi/v5"
	"io"
	"net/http"

	"github.com/MowlCoder/go-url-shortener/internal/app/config"
	"github.com/MowlCoder/go-url-shortener/internal/app/storage"
)

type ShortenerHandler struct {
	config     *config.AppConfig
	urlStorage *storage.URLStorage
}

func NewShortenerHandler(config *config.AppConfig) *ShortenerHandler {
	return &ShortenerHandler{
		config:     config,
		urlStorage: storage.NewURLStorage(),
	}
}

func (h *ShortenerHandler) ShortURL(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(r.Body)

	if err != nil {
		SendStatusCode(w, http.StatusBadRequest)
		return
	}

	if len(body) == 0 {
		SendStatusCode(w, http.StatusBadRequest)
		return
	}

	id, err := h.urlStorage.SaveURL(string(body))

	if err != nil {
		SendStatusCode(w, http.StatusBadRequest)
		return
	}

	SendTextResponse(w, http.StatusCreated, fmt.Sprintf("%s/%s", h.config.BaseShortURLAddr, id))
}

func (h *ShortenerHandler) RedirectToURLByID(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	originalURL, err := h.urlStorage.GetURLByID(id)

	if err != nil {
		SendStatusCode(w, http.StatusBadRequest)
		return
	}

	w.Header().Set("Location", originalURL)
	SendStatusCode(w, http.StatusTemporaryRedirect)
}
