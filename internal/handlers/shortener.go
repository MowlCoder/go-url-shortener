package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"

	"github.com/go-chi/chi/v5"

	"github.com/MowlCoder/go-url-shortener/internal/config"
	"github.com/MowlCoder/go-url-shortener/internal/domain"
	"github.com/MowlCoder/go-url-shortener/internal/handlers/dtos"
	"github.com/MowlCoder/go-url-shortener/internal/storage"
	"github.com/MowlCoder/go-url-shortener/internal/storage/models"
)

type URLStorage interface {
	SaveSeveralURL(ctx context.Context, dtos []domain.SaveShortURLDto) ([]models.ShortenedURL, error)
	SaveURL(ctx context.Context, dto domain.SaveShortURLDto) (*models.ShortenedURL, error)
	GetOriginalURLByShortURL(ctx context.Context, shortURL string) (string, error)
	GetURLsByUserID(ctx context.Context, userID string) ([]models.ShortenedURL, error)
	Ping(ctx context.Context) error
}

type StringGeneratorService interface {
	GenerateRandom() string
}

type ShortenerHandler struct {
	config          *config.AppConfig
	urlStorage      URLStorage
	stringGenerator StringGeneratorService
}

func NewShortenerHandler(
	config *config.AppConfig,
	urlStorage URLStorage,
	stringGenerator StringGeneratorService,
) *ShortenerHandler {
	return &ShortenerHandler{
		config:          config,
		urlStorage:      urlStorage,
		stringGenerator: stringGenerator,
	}
}

func (h *ShortenerHandler) ShortURLJSON(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value("user_id").(string)
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

	shortURL := h.stringGenerator.GenerateRandom()
	shortenedURL, err := h.urlStorage.SaveURL(r.Context(), domain.SaveShortURLDto{
		OriginalURL: requestBody.URL,
		ShortURL:    shortURL,
		UserID:      userID,
	})

	if errors.Is(err, storage.ErrRowConflict) {
		SendJSONResponse(w, http.StatusConflict, dtos.ShortURLResponse{
			Result: fmt.Sprintf("%s/%s", h.config.BaseShortURLAddr, shortenedURL.ShortURL),
		})
		return
	}

	if errors.Is(err, storage.ErrShortURLConflict) {
		shortURL = h.stringGenerator.GenerateRandom()
		shortenedURL, err = h.urlStorage.SaveURL(r.Context(), domain.SaveShortURLDto{
			OriginalURL: requestBody.URL,
			ShortURL:    shortURL,
			UserID:      "1",
		})
	}

	if err != nil {
		SendStatusCode(w, http.StatusInternalServerError)
		return
	}

	SendJSONResponse(w, http.StatusCreated, dtos.ShortURLResponse{
		Result: fmt.Sprintf("%s/%s", h.config.BaseShortURLAddr, shortenedURL.ShortURL),
	})
}

func (h *ShortenerHandler) ShortBatchURL(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value("user_id").(string)
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

	saveDtos := make([]domain.SaveShortURLDto, 0, len(requestBody))
	correlations := make(map[string]string)

	for _, dto := range requestBody {
		saveDtos = append(saveDtos, domain.SaveShortURLDto{
			OriginalURL: dto.OriginalURL,
			ShortURL:    h.stringGenerator.GenerateRandom(),
			UserID:      userID,
		})
		correlations[dto.OriginalURL] = dto.CorrelationID
	}

	shortenedURLs, err := h.urlStorage.SaveSeveralURL(r.Context(), saveDtos)

	if err != nil {
		log.Println(err)
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
	userID := r.Context().Value("user_id").(string)
	body, err := io.ReadAll(r.Body)

	if err != nil {
		SendStatusCode(w, http.StatusBadRequest)
		return
	}

	if len(body) == 0 {
		SendStatusCode(w, http.StatusBadRequest)
		return
	}

	shortURL := h.stringGenerator.GenerateRandom()
	shortenedURL, err := h.urlStorage.SaveURL(r.Context(), domain.SaveShortURLDto{
		OriginalURL: string(body),
		ShortURL:    shortURL,
		UserID:      userID,
	})

	if errors.Is(err, storage.ErrRowConflict) {
		SendTextResponse(w, http.StatusConflict, fmt.Sprintf("%s/%s", h.config.BaseShortURLAddr, shortenedURL.ShortURL))
		return
	}

	if errors.Is(err, storage.ErrShortURLConflict) {
		shortURL = h.stringGenerator.GenerateRandom()
		shortenedURL, err = h.urlStorage.SaveURL(r.Context(), domain.SaveShortURLDto{
			OriginalURL: string(body),
			ShortURL:    shortURL,
			UserID:      userID,
		})
	}

	if err != nil {
		SendStatusCode(w, http.StatusInternalServerError)
		return
	}

	SendTextResponse(w, http.StatusCreated, fmt.Sprintf("%s/%s", h.config.BaseShortURLAddr, shortenedURL.ShortURL))
}

func (h *ShortenerHandler) GetMyURLs(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value("user_id").(string)
	urls, err := h.urlStorage.GetURLsByUserID(r.Context(), userID)

	if err != nil {
		SendStatusCode(w, http.StatusInternalServerError)
		return
	}

	if len(urls) == 0 {
		SendStatusCode(w, http.StatusNoContent)
		return
	}

	SendJSONResponse(w, 200, urls)
}

func (h *ShortenerHandler) RedirectToURLByID(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	originalURL, err := h.urlStorage.GetOriginalURLByShortURL(r.Context(), id)

	if err != nil {
		SendStatusCode(w, http.StatusBadRequest)
		return
	}

	SendRedirectResponse(w, originalURL)
}

func (h *ShortenerHandler) Ping(w http.ResponseWriter, r *http.Request) {
	if err := h.urlStorage.Ping(r.Context()); err != nil {
		SendStatusCode(w, http.StatusInternalServerError)
		return
	}

	SendStatusCode(w, http.StatusOK)
}
