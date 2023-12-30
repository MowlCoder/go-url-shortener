package http

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"

	"github.com/go-chi/chi/v5"

	"github.com/MowlCoder/go-url-shortener/internal/handlers/http/dtos"

	"github.com/MowlCoder/go-url-shortener/internal/config"
	contextUtil "github.com/MowlCoder/go-url-shortener/internal/context"
	"github.com/MowlCoder/go-url-shortener/internal/domain"
	"github.com/MowlCoder/go-url-shortener/pkg/httputil"
)

type shortenerService interface {
	ShortURL(ctx context.Context, url string, userID string) (*domain.ShortenedURL, error)
	ShortBatchURL(ctx context.Context, urls []domain.ShortBatchURL, userID string) ([]domain.ShortBatchURL, error)
	GetUserURLs(ctx context.Context, userID string) ([]domain.ShortenedURL, error)
	DeleteURLs(ctx context.Context, urls []string, userID string) error
	GetByShortURL(ctx context.Context, url string) (*domain.ShortenedURL, error)
	GetInternalStats(ctx context.Context) (*domain.InternalStats, error)
	Ping(ctx context.Context) error
}

// ShortenerHandler contains handlers that responsible for handling http request and give proper http response.
type ShortenerHandler struct {
	config  *config.AppConfig
	service shortenerService
}

// NewShortenerHandler is contructor function for ShortenerHandler.
func NewShortenerHandler(
	config *config.AppConfig,
	service shortenerService,
) *ShortenerHandler {
	return &ShortenerHandler{
		config:  config,
		service: service,
	}
}

// ShortURLJSON godoc
// @Summary Short url (JSON)
// @Accept json
// @Produce json
// @Param dto body dtos.ShortURLDto true "Short url"
// @Success 201 {object} dtos.ShortURLResponse
// @Failure 400
// @Failure 401
// @Failure 409 {object} dtos.ShortURLResponse
// @Failure 500
// @Router /api/shorten [post]
func (h *ShortenerHandler) ShortURLJSON(w http.ResponseWriter, r *http.Request) {
	userID, err := contextUtil.GetUserIDFromContext(r.Context())
	if err != nil {
		httputil.SendStatusCode(w, http.StatusUnauthorized)
		return
	}

	requestBody := dtos.ShortURLDto{}
	rawBody, err := io.ReadAll(r.Body)

	if err != nil {
		httputil.SendStatusCode(w, http.StatusBadRequest)
		return
	}

	if jsonErr := json.Unmarshal(rawBody, &requestBody); jsonErr != nil {
		httputil.SendStatusCode(w, http.StatusBadRequest)
		return
	}

	if requestBody.URL == "" {
		httputil.SendStatusCode(w, http.StatusBadRequest)
		return
	}

	shortenedURL, err := h.service.ShortURL(r.Context(), requestBody.URL, userID)

	if errors.Is(err, domain.ErrURLConflict) {
		httputil.SendJSONResponse(w, http.StatusConflict, dtos.ShortURLResponse{
			Result: fmt.Sprintf("%s/%s", h.config.BaseShortURLAddr, shortenedURL.ShortURL),
		})
		return
	}

	if err != nil {
		httputil.SendStatusCode(w, http.StatusInternalServerError)
		return
	}

	httputil.SendJSONResponse(w, http.StatusCreated, dtos.ShortURLResponse{
		Result: fmt.Sprintf("%s/%s", h.config.BaseShortURLAddr, shortenedURL.ShortURL),
	})
}

// ShortBatchURL godoc
// @Summary Short batch urls
// @Accept json
// @Produce json
// @Param dto body []dtos.ShortBatchURLDto true "Short batch urls"
// @Success 201 {array} dtos.ShortBatchURLResponse
// @Failure 400
// @Failure 401
// @Failure 500
// @Router /api/shorten/batch [post]
func (h *ShortenerHandler) ShortBatchURL(w http.ResponseWriter, r *http.Request) {
	userID, err := contextUtil.GetUserIDFromContext(r.Context())
	if err != nil {
		httputil.SendStatusCode(w, http.StatusUnauthorized)
		return
	}

	requestBody := make([]dtos.ShortBatchURLDto, 0)
	rawBody, err := io.ReadAll(r.Body)

	if err != nil {
		httputil.SendStatusCode(w, http.StatusBadRequest)
		return
	}

	if jsonErr := json.Unmarshal(rawBody, &requestBody); jsonErr != nil {
		httputil.SendStatusCode(w, http.StatusBadRequest)
		return
	}

	if len(requestBody) == 0 {
		httputil.SendStatusCode(w, http.StatusBadRequest)
		return
	}

	urls := make([]domain.ShortBatchURL, 0, len(requestBody))

	for _, url := range requestBody {
		urls = append(urls, domain.ShortBatchURL{
			OriginalURL:   url.OriginalURL,
			CorrelationID: url.CorrelationID,
		})
	}

	shortenedURLs, err := h.service.ShortBatchURL(r.Context(), urls, userID)

	if err != nil {
		log.Println(err)
		httputil.SendStatusCode(w, http.StatusInternalServerError)
		return
	}

	responseBody := make([]dtos.ShortBatchURLResponse, 0, len(shortenedURLs))

	for _, shortenedURL := range shortenedURLs {
		responseBody = append(responseBody, dtos.ShortBatchURLResponse{
			ShortURL:      fmt.Sprintf("%s/%s", h.config.BaseShortURLAddr, shortenedURL.ShortURL),
			CorrelationID: shortenedURL.CorrelationID,
		})
	}

	httputil.SendJSONResponse(w, 201, responseBody)
}

// ShortURL godoc
// @Summary Short url (Text)
// @Accept plain
// @Produce plain
// @Param dto body string true "Short url"
// @Success 201 {string} string "Shortened url"
// @Failure 400
// @Failure 401
// @Failure 409 {string} string "Shortened url"
// @Failure 500
// @Router / [post]
func (h *ShortenerHandler) ShortURL(w http.ResponseWriter, r *http.Request) {
	userID, err := contextUtil.GetUserIDFromContext(r.Context())
	if err != nil {
		httputil.SendStatusCode(w, http.StatusUnauthorized)
		return
	}

	body, err := io.ReadAll(r.Body)

	if err != nil {
		httputil.SendStatusCode(w, http.StatusBadRequest)
		return
	}

	if len(body) == 0 {
		httputil.SendStatusCode(w, http.StatusBadRequest)
		return
	}

	shortenedURL, err := h.service.ShortURL(r.Context(), string(body), userID)

	if errors.Is(err, domain.ErrURLConflict) {
		httputil.SendTextResponse(w, http.StatusConflict, fmt.Sprintf("%s/%s", h.config.BaseShortURLAddr, shortenedURL.ShortURL))
		return
	}

	if err != nil {
		httputil.SendStatusCode(w, http.StatusInternalServerError)
		return
	}

	httputil.SendTextResponse(w, http.StatusCreated, fmt.Sprintf("%s/%s", h.config.BaseShortURLAddr, shortenedURL.ShortURL))
}

// GetMyURLs godoc
// @Summary Get user urls
// @Produce json
// @Success 200 {array} dtos.UserURLsResponse
// @Success 204
// @Failure 401
// @Failure 500
// @Router /api/user/urls [get]
func (h *ShortenerHandler) GetMyURLs(w http.ResponseWriter, r *http.Request) {
	userID, err := contextUtil.GetUserIDFromContext(r.Context())
	if err != nil {
		httputil.SendStatusCode(w, http.StatusUnauthorized)
		return
	}

	urls, err := h.service.GetUserURLs(r.Context(), userID)

	if err != nil {
		httputil.SendStatusCode(w, http.StatusInternalServerError)
		return
	}

	if len(urls) == 0 {
		httputil.SendStatusCode(w, http.StatusNoContent)
		return
	}

	responseURLs := make([]dtos.UserURLsResponse, 0, len(urls))

	for _, url := range urls {
		responseURLs = append(responseURLs, dtos.UserURLsResponse{
			ShortURL:    fmt.Sprintf("%s/%s", h.config.BaseShortURLAddr, url.ShortURL),
			OriginalURL: url.OriginalURL,
		})
	}

	httputil.SendJSONResponse(w, 200, responseURLs)
}

// DeleteURLs godoc
// @Summary Delete user urls
// @Accept json
// @Param dto body dtos.DeleteURLsRequest true "Delete user urls"
// @Success 202
// @Failure 400
// @Failure 401
// @Router /api/user/urls [delete]
func (h *ShortenerHandler) DeleteURLs(w http.ResponseWriter, r *http.Request) {
	userID, err := contextUtil.GetUserIDFromContext(r.Context())
	if err != nil {
		httputil.SendStatusCode(w, http.StatusUnauthorized)
		return
	}

	var requestBody dtos.DeleteURLsRequest
	rawBody, err := io.ReadAll(r.Body)

	if err != nil {
		httputil.SendStatusCode(w, http.StatusBadRequest)
		return
	}

	if err := json.Unmarshal(rawBody, &requestBody); err != nil {
		httputil.SendStatusCode(w, http.StatusBadRequest)
		return
	}

	if len(requestBody) == 0 {
		httputil.SendStatusCode(w, http.StatusBadRequest)
		return
	}

	h.service.DeleteURLs(r.Context(), requestBody, userID)

	httputil.SendStatusCode(w, http.StatusAccepted)
}

// RedirectToURLByID godoc
// @Summary Redirect from short url to original url
// @Param id path string true "Short URL ID"
// @Success 307
// @Failure 400
// @Failure 410
// @Router /{id} [get]
func (h *ShortenerHandler) RedirectToURLByID(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	originalURL, err := h.service.GetByShortURL(r.Context(), id)

	if err != nil {
		httputil.SendStatusCode(w, http.StatusBadRequest)
		return
	}

	if originalURL.IsDeleted {
		httputil.SendStatusCode(w, http.StatusGone)
		return
	}

	httputil.SendRedirectResponse(w, originalURL.OriginalURL)
}

// GetStats godoc
// @Summary Get internal statistics for metrics
// @Success 200 {object} dtos.GetStatsResponse
// @Failure 403
// @Failure 500
// @Router /api/internal/stats [get]
func (h *ShortenerHandler) GetStats(w http.ResponseWriter, r *http.Request) {
	stats, err := h.service.GetInternalStats(r.Context())
	if err != nil {
		httputil.SendStatusCode(w, http.StatusInternalServerError)
	}

	httputil.SendJSONResponse(w, 200, dtos.GetStatsResponse{
		URLs:  stats.URLs,
		Users: stats.Users,
	})
}

// Ping godoc
// @Summary Checking if server isn't down
// @Success 200
// @Failure 500
// @Router /ping [get]
func (h *ShortenerHandler) Ping(w http.ResponseWriter, r *http.Request) {
	if err := h.service.Ping(r.Context()); err != nil {
		httputil.SendStatusCode(w, http.StatusInternalServerError)
		return
	}

	httputil.SendStatusCode(w, http.StatusOK)
}
