package handlers

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/MowlCoder/go-url-shortener/internal/app/handlers/dtos"
	"github.com/MowlCoder/go-url-shortener/internal/app/storage/models"

	"github.com/go-chi/chi/v5"

	"github.com/MowlCoder/go-url-shortener/internal/app/config"
)

type URLStorage interface {
	SaveSeveralURL(urls []string) ([]models.ShortenedURL, error)
	SaveURL(url string) (models.ShortenedURL, error)
	GetOriginalURLByShortURL(shortURL string) (string, error)
	Ping() error
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

func (h *ShortenerHandler) ShortURLJSON(w http.ResponseWriter, r *http.Request) {
	requestBody := dtos.ShortURLDto{}
	rawBody, err := io.ReadAll(r.Body)

	if err != nil {
		SendStatusCode(w, http.StatusBadRequest)
		return
	}

	if err := json.Unmarshal(rawBody, &requestBody); err != nil {
		SendStatusCode(w, http.StatusBadRequest)
		return
	}

	if requestBody.URL == "" {
		SendStatusCode(w, http.StatusBadRequest)
		return
	}

	shortenedURL, err := h.urlStorage.SaveURL(requestBody.URL)

	if err != nil {
		SendStatusCode(w, http.StatusBadRequest)
		return
	}

	SendJSONResponse(w, http.StatusCreated, dtos.ShortURLResponse{
		Result: fmt.Sprintf("%s/%s", h.config.BaseShortURLAddr, shortenedURL.ShortURL),
	})
}

func (h *ShortenerHandler) ShortBatchURL(w http.ResponseWriter, r *http.Request) {
	requestBody := make([]dtos.ShortBatchURLDto, 0)
	rawBody, err := io.ReadAll(r.Body)

	if err != nil {
		SendStatusCode(w, http.StatusBadRequest)
		return
	}

	if err := json.Unmarshal(rawBody, &requestBody); err != nil {
		SendStatusCode(w, http.StatusBadRequest)
		return
	}

	if len(requestBody) == 0 {
		SendStatusCode(w, http.StatusBadRequest)
		return
	}

	urls := make([]string, 0, len(requestBody))
	correlations := make(map[string]string)

	for _, dto := range requestBody {
		urls = append(urls, dto.OriginalURL)
		correlations[dto.OriginalURL] = dto.CorrelationID
	}

	shortenedURLs, err := h.urlStorage.SaveSeveralURL(urls)

	if err != nil {
		fmt.Println(err)
		SendStatusCode(w, http.StatusInternalServerError)
		return
	}

	responseBody := make([]dtos.ShortBatchURLResponse, 0, len(shortenedURLs))

	for _, shortenedURL := range shortenedURLs {
		correlationID := correlations[shortenedURL.OriginalURL]

		responseBody = append(responseBody, dtos.ShortBatchURLResponse{
			ShortURL:      fmt.Sprintf("%s/%s", h.config.BaseShortURLAddr, shortenedURL.ShortURL),
			CorrelationID: correlationID,
		})
	}

	SendJSONResponse(w, 201, responseBody)
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

	shortenedURL, err := h.urlStorage.SaveURL(string(body))

	if err != nil {
		SendStatusCode(w, http.StatusInternalServerError)
		return
	}

	SendTextResponse(w, http.StatusCreated, fmt.Sprintf("%s/%s", h.config.BaseShortURLAddr, shortenedURL.ShortURL))
}

func (h *ShortenerHandler) RedirectToURLByID(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	originalURL, err := h.urlStorage.GetOriginalURLByShortURL(id)

	if err != nil {
		SendStatusCode(w, http.StatusBadRequest)
		return
	}

	SendRedirectResponse(w, originalURL)
}

func (h *ShortenerHandler) Ping(w http.ResponseWriter, r *http.Request) {
	if err := h.urlStorage.Ping(); err != nil {
		SendStatusCode(w, http.StatusInternalServerError)
		return
	}

	SendStatusCode(w, http.StatusOK)
}
