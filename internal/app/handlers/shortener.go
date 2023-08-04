package handlers

import (
	"fmt"
	"io"
	"net/http"

	"github.com/go-chi/chi/v5"

	"github.com/MowlCoder/go-url-shortener/internal/app/config"
)

type URLStorage interface {
	SaveURL(url string) (string, error)
	GetURLByID(id string) (string, error)
}

type ShortenerHandler struct {
	config     *config.AppConfig
	urlStorage URLStorage
}

func NewShortenerHandler(
	config *config.AppConfig,
	urlStorage URLStorage,
) *ShortenerHandler {
	return &ShortenerHandler{
		config:     config,
		urlStorage: urlStorage,
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

	SendRedirectResponse(w, originalURL)
}
