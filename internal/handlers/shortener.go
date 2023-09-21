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
	contextUtil "github.com/MowlCoder/go-url-shortener/internal/context"
	"github.com/MowlCoder/go-url-shortener/internal/domain"
	"github.com/MowlCoder/go-url-shortener/internal/handlers/dtos"
	"github.com/MowlCoder/go-url-shortener/internal/storage"
	"github.com/MowlCoder/go-url-shortener/internal/storage/models"
)

type URLStorage interface {
	SaveSeveralURL(ctx context.Context, dtos []domain.SaveShortURLDto) ([]models.ShortenedURL, error)
	SaveURL(ctx context.Context, dto domain.SaveShortURLDto) (*models.ShortenedURL, error)
	GetByShortURL(ctx context.Context, shortURL string) (*models.ShortenedURL, error)
	GetURLsByUserID(ctx context.Context, userID string) ([]models.ShortenedURL, error)
	DeleteByShortURLs(ctx context.Context, shortURLs []string, userID string) error
	Ping(ctx context.Context) error
}

type StringGeneratorService interface {
	GenerateRandom() string
}

type DeleteURLQueue interface {
	Push(task *domain.DeleteURLsTask)
}

type ShortenerHandler struct {
	config          *config.AppConfig
	urlStorage      URLStorage
	stringGenerator StringGeneratorService
	deleteURLQueue  DeleteURLQueue
}

func NewShortenerHandler(
	config *config.AppConfig,
	urlStorage URLStorage,
	stringGenerator StringGeneratorService,
	deleteURLQueue DeleteURLQueue,
) *ShortenerHandler {
	return &ShortenerHandler{
		config:          config,
		urlStorage:      urlStorage,
		stringGenerator: stringGenerator,
		deleteURLQueue:  deleteURLQueue,
	}
}

func (h *ShortenerHandler) ShortURLJSON(w http.ResponseWriter, r *http.Request) {
	userID := contextUtil.GetUserIDFromContext(r.Context())
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
	userID := contextUtil.GetUserIDFromContext(r.Context())
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
	userID := contextUtil.GetUserIDFromContext(r.Context())
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
	userID := contextUtil.GetUserIDFromContext(r.Context())
	urls, err := h.urlStorage.GetURLsByUserID(r.Context(), userID)

	if err != nil {
		SendStatusCode(w, http.StatusInternalServerError)
		return
	}

	if len(urls) == 0 {
		SendStatusCode(w, http.StatusNoContent)
		return
	}

	responseURLs := make([]dtos.UserURLsResponse, 0, len(urls))

	for _, url := range urls {
		responseURLs = append(responseURLs, dtos.UserURLsResponse{
			ShortURL:    fmt.Sprintf("%s/%s", h.config.BaseShortURLAddr, url.ShortURL),
			OriginalURL: url.OriginalURL,
		})
	}

	SendJSONResponse(w, 200, responseURLs)
}

func (h *ShortenerHandler) DeleteURLs(w http.ResponseWriter, r *http.Request) {
	userID := contextUtil.GetUserIDFromContext(r.Context())
	var requestBody dtos.DeleteURLsRequest
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

	go h.deleteURLQueue.Push(&domain.DeleteURLsTask{
		ShortURLs: requestBody,
		UserID:    userID,
	})

	SendStatusCode(w, http.StatusAccepted)
}

func (h *ShortenerHandler) RedirectToURLByID(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	originalURL, err := h.urlStorage.GetByShortURL(r.Context(), id)

	if err != nil {
		SendStatusCode(w, http.StatusBadRequest)
		return
	}

	if originalURL.IsDeleted {
		SendStatusCode(w, http.StatusGone)
		return
	}

	SendRedirectResponse(w, originalURL.OriginalURL)
}

func (h *ShortenerHandler) Ping(w http.ResponseWriter, r *http.Request) {
	if err := h.urlStorage.Ping(r.Context()); err != nil {
		SendStatusCode(w, http.StatusInternalServerError)
		return
	}

	SendStatusCode(w, http.StatusOK)
}
