package http

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/go-chi/chi/v5"
	"go.uber.org/mock/gomock"

	handlersmock "github.com/MowlCoder/go-url-shortener/internal/handlers/http/mocks"

	"github.com/MowlCoder/go-url-shortener/internal/domain"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/MowlCoder/go-url-shortener/internal/handlers/http/dtos"

	"github.com/MowlCoder/go-url-shortener/internal/config"
	contextUtil "github.com/MowlCoder/go-url-shortener/internal/context"
)

func TestShortURL(t *testing.T) {
	ctrl := gomock.NewController(t)
	service := handlersmock.NewMockshortenerService(ctrl)

	handler := NewShortenerHandler(
		&config.AppConfig{},
		service,
	)

	type TestCase struct {
		PrepareServiceFunc func(
			ctx context.Context,
			body string,
		)
		Name               string
		Body               string
		NotAuth            bool
		ExpectedStatusCode int
	}

	testCases := []TestCase{
		{
			Name: "valid",
			Body: "https://url.com",
			PrepareServiceFunc: func(ctx context.Context, body string) {
				service.
					EXPECT().
					ShortURL(ctx, body, "1").
					Return(&domain.ShortenedURL{}, nil)
			},
			ExpectedStatusCode: http.StatusCreated,
		},
		{
			Name:               "invalid body",
			Body:               "",
			PrepareServiceFunc: nil,
			ExpectedStatusCode: http.StatusBadRequest,
		},
		{
			Name:               "not auth",
			NotAuth:            true,
			Body:               "",
			PrepareServiceFunc: nil,
			ExpectedStatusCode: http.StatusUnauthorized,
		},
		{
			Name: "err row conflict",
			Body: "https://url.com",
			PrepareServiceFunc: func(ctx context.Context, body string) {
				service.
					EXPECT().
					ShortURL(ctx, body, "1").
					Return(&domain.ShortenedURL{}, domain.ErrURLConflict)
			},
			ExpectedStatusCode: http.StatusConflict,
		},
		{
			Name: "internal server error",
			Body: "https://url.com",
			PrepareServiceFunc: func(ctx context.Context, body string) {
				service.
					EXPECT().
					ShortURL(ctx, body, "1").
					Return(nil, errors.New("undefined behaviour"))
			},
			ExpectedStatusCode: http.StatusInternalServerError,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			r := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(testCase.Body))
			r.Header.Set("Content-Type", "text/plain")

			if !testCase.NotAuth {
				ctx := contextUtil.SetUserIDToContext(r.Context(), "1")
				r = r.WithContext(ctx)
			}

			w := httptest.NewRecorder()

			if testCase.PrepareServiceFunc != nil {
				testCase.PrepareServiceFunc(r.Context(), testCase.Body)
			}

			handler.ShortURL(w, r)

			res := w.Result()
			defer res.Body.Close()

			assert.Equal(t, testCase.ExpectedStatusCode, res.StatusCode)
		})
	}
}

func TestShortURLJSON(t *testing.T) {
	ctrl := gomock.NewController(t)
	service := handlersmock.NewMockshortenerService(ctrl)

	handler := NewShortenerHandler(
		&config.AppConfig{},
		service,
	)

	type TestCase struct {
		PrepareServiceFunc func(
			ctx context.Context,
			body *dtos.ShortURLDto,
		)
		Body               *dtos.ShortURLDto
		Name               string
		NotAuth            bool
		ExpectedStatusCode int
	}

	testCases := []TestCase{
		{
			Name: "valid",
			Body: &dtos.ShortURLDto{
				URL: "https://url.com",
			},
			PrepareServiceFunc: func(ctx context.Context, body *dtos.ShortURLDto) {
				service.
					EXPECT().
					ShortURL(ctx, body.URL, "1").
					Return(&domain.ShortenedURL{}, nil)
			},
			ExpectedStatusCode: http.StatusCreated,
		},
		{
			Name:               "nil body",
			Body:               nil,
			PrepareServiceFunc: nil,
			ExpectedStatusCode: http.StatusBadRequest,
		},
		{
			Name:               "not auth",
			Body:               nil,
			NotAuth:            true,
			PrepareServiceFunc: nil,
			ExpectedStatusCode: http.StatusUnauthorized,
		},
		{
			Name:               "invalid body",
			Body:               &dtos.ShortURLDto{},
			PrepareServiceFunc: nil,
			ExpectedStatusCode: http.StatusBadRequest,
		},
		{
			Name: "err row conflict",
			Body: &dtos.ShortURLDto{
				URL: "https://url.com",
			},
			PrepareServiceFunc: func(ctx context.Context, body *dtos.ShortURLDto) {
				service.
					EXPECT().
					ShortURL(ctx, body.URL, "1").
					Return(&domain.ShortenedURL{}, domain.ErrURLConflict)
			},
			ExpectedStatusCode: http.StatusConflict,
		},
		{
			Name: "internal server error",
			Body: &dtos.ShortURLDto{
				URL: "https://url.com",
			},
			PrepareServiceFunc: func(ctx context.Context, body *dtos.ShortURLDto) {
				service.
					EXPECT().
					ShortURL(ctx, body.URL, "1").
					Return(nil, errors.New("undefined behaviour"))
			},
			ExpectedStatusCode: http.StatusInternalServerError,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			var rawBody []byte
			var err error

			if testCase.Body != nil {
				rawBody, err = json.Marshal(*testCase.Body)
				require.NoError(t, err)
			}

			r := httptest.NewRequest(http.MethodPost, "/", bytes.NewReader(rawBody))
			r.Header.Set("Content-Type", "application/json")

			if !testCase.NotAuth {
				ctx := contextUtil.SetUserIDToContext(r.Context(), "1")
				r = r.WithContext(ctx)
			}

			w := httptest.NewRecorder()

			if testCase.PrepareServiceFunc != nil {
				testCase.PrepareServiceFunc(r.Context(), testCase.Body)
			}

			handler.ShortURLJSON(w, r)

			res := w.Result()
			defer res.Body.Close()

			assert.Equal(t, testCase.ExpectedStatusCode, res.StatusCode)
		})
	}
}

func TestShortBatchURL(t *testing.T) {
	ctrl := gomock.NewController(t)
	service := handlersmock.NewMockshortenerService(ctrl)

	handler := NewShortenerHandler(
		&config.AppConfig{},
		service,
	)

	type TestCase struct {
		PrepareServiceFunc func(
			ctx context.Context,
			body []dtos.ShortBatchURLDto,
		)
		Name               string
		Body               []dtos.ShortBatchURLDto
		NotAuth            bool
		ExpectedStatusCode int
	}

	testCases := []TestCase{
		{
			Name: "valid",
			Body: []dtos.ShortBatchURLDto{
				{
					OriginalURL:   "https://url.com",
					CorrelationID: "1",
				},
				{
					OriginalURL:   "https://url.com/1",
					CorrelationID: "2",
				},
			},
			PrepareServiceFunc: func(ctx context.Context, body []dtos.ShortBatchURLDto) {
				shortenedUrls := make([]domain.ShortBatchURL, 0)

				for _, dto := range body {
					shortenedUrls = append(shortenedUrls, domain.ShortBatchURL{
						ShortURL:      dto.CorrelationID + "1234",
						OriginalURL:   dto.OriginalURL,
						CorrelationID: dto.CorrelationID,
					})
				}

				service.
					EXPECT().
					ShortBatchURL(ctx, gomock.Any(), "1").
					Return(shortenedUrls, nil)
			},
			ExpectedStatusCode: http.StatusCreated,
		},
		{
			Name:               "nil body",
			Body:               nil,
			PrepareServiceFunc: nil,
			ExpectedStatusCode: http.StatusBadRequest,
		},
		{
			Name:               "not auth",
			NotAuth:            true,
			Body:               nil,
			PrepareServiceFunc: nil,
			ExpectedStatusCode: http.StatusUnauthorized,
		},
		{
			Name:               "invalid body",
			Body:               []dtos.ShortBatchURLDto{},
			PrepareServiceFunc: nil,
			ExpectedStatusCode: http.StatusBadRequest,
		},
		{
			Name: "internal server error",
			Body: []dtos.ShortBatchURLDto{
				{
					OriginalURL:   "https://url.com",
					CorrelationID: "1",
				},
				{
					OriginalURL:   "https://url.com/1",
					CorrelationID: "2",
				},
			},
			PrepareServiceFunc: func(ctx context.Context, body []dtos.ShortBatchURLDto) {
				service.
					EXPECT().
					ShortBatchURL(ctx, gomock.Any(), "1").
					Return(nil, errors.New("undefined behavior"))
			},
			ExpectedStatusCode: http.StatusInternalServerError,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			var rawBody []byte
			var err error

			if testCase.Body != nil {
				rawBody, err = json.Marshal(testCase.Body)
				require.NoError(t, err)
			}

			r := httptest.NewRequest(http.MethodPost, "/", bytes.NewReader(rawBody))
			r.Header.Set("Content-Type", "application/json")

			if !testCase.NotAuth {
				ctx := contextUtil.SetUserIDToContext(r.Context(), "1")
				r = r.WithContext(ctx)
			}

			w := httptest.NewRecorder()

			if testCase.PrepareServiceFunc != nil {
				testCase.PrepareServiceFunc(r.Context(), testCase.Body)
			}

			handler.ShortBatchURL(w, r)

			res := w.Result()
			defer res.Body.Close()

			assert.Equal(t, testCase.ExpectedStatusCode, res.StatusCode)

			if res.StatusCode == http.StatusCreated {
				response, err := io.ReadAll(res.Body)
				require.NoError(t, err)

				var responseBody []dtos.ShortBatchURLResponse
				require.NoError(t, json.Unmarshal(response, &responseBody))

				assert.Equal(t, len(testCase.Body), len(responseBody))

				allFound := true

				for _, dto := range testCase.Body {
					isFound := false

					for _, resDto := range responseBody {
						if dto.CorrelationID == resDto.CorrelationID {
							isFound = true
							break
						}
					}

					if !isFound {
						allFound = false
						break
					}
				}

				assert.Equal(t, true, allFound)
			}
		})
	}
}

func TestGetMyURLs(t *testing.T) {
	ctrl := gomock.NewController(t)
	service := handlersmock.NewMockshortenerService(ctrl)
	userID := "1"

	handler := NewShortenerHandler(
		&config.AppConfig{},
		service,
	)

	type TestCase struct {
		PrepareServiceFunc func(
			ctx context.Context,
		)
		Name               string
		NotAuth            bool
		ExpectedStatusCode int
	}

	testCases := []TestCase{
		{
			Name: "valid",
			PrepareServiceFunc: func(ctx context.Context) {
				service.
					EXPECT().
					GetUserURLs(ctx, userID).
					Return([]domain.ShortenedURL{{}, {}}, nil)
			},
			ExpectedStatusCode: http.StatusOK,
		},
		{
			Name: "valid (no content)",
			PrepareServiceFunc: func(ctx context.Context) {
				service.
					EXPECT().
					GetUserURLs(ctx, userID).
					Return([]domain.ShortenedURL{}, nil)
			},
			ExpectedStatusCode: http.StatusNoContent,
		},
		{
			Name:               "not auth",
			NotAuth:            true,
			PrepareServiceFunc: nil,
			ExpectedStatusCode: http.StatusUnauthorized,
		},
		{
			Name: "internal server error",
			PrepareServiceFunc: func(ctx context.Context) {
				service.
					EXPECT().
					GetUserURLs(ctx, userID).
					Return(nil, errors.New("undefined behavior"))
			},
			ExpectedStatusCode: http.StatusInternalServerError,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			r := httptest.NewRequest(http.MethodGet, "/", nil)

			if !testCase.NotAuth {
				ctx := contextUtil.SetUserIDToContext(r.Context(), userID)
				r = r.WithContext(ctx)
			}

			w := httptest.NewRecorder()

			if testCase.PrepareServiceFunc != nil {
				testCase.PrepareServiceFunc(r.Context())
			}

			handler.GetMyURLs(w, r)

			res := w.Result()
			defer res.Body.Close()

			assert.Equal(t, testCase.ExpectedStatusCode, res.StatusCode)
		})
	}
}

func TestDeleteURLs(t *testing.T) {
	ctrl := gomock.NewController(t)
	service := handlersmock.NewMockshortenerService(ctrl)
	userID := "1"

	handler := NewShortenerHandler(
		&config.AppConfig{},
		service,
	)

	type TestCase struct {
		PrepareServiceFunc func(
			ctx context.Context,
		)
		Name               string
		Body               dtos.DeleteURLsRequest
		ExpectedStatusCode int
		NotAuth            bool
	}

	testCases := []TestCase{
		{
			Name: "valid",
			Body: dtos.DeleteURLsRequest{"123", "1234"},
			PrepareServiceFunc: func(ctx context.Context) {
				service.
					EXPECT().
					DeleteURLs(ctx, gomock.Any(), "1")
			},
			ExpectedStatusCode: http.StatusAccepted,
		},
		{
			Name:               "nil body",
			Body:               nil,
			PrepareServiceFunc: nil,
			ExpectedStatusCode: http.StatusBadRequest,
		},
		{
			Name:               "not auth",
			NotAuth:            true,
			Body:               nil,
			PrepareServiceFunc: nil,
			ExpectedStatusCode: http.StatusUnauthorized,
		},
		{
			Name:               "invalid body",
			Body:               dtos.DeleteURLsRequest{},
			PrepareServiceFunc: nil,
			ExpectedStatusCode: http.StatusBadRequest,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			var rawBody []byte
			var err error

			if testCase.Body != nil {
				rawBody, err = json.Marshal(testCase.Body)
				require.NoError(t, err)
			}

			r := httptest.NewRequest(http.MethodDelete, "/", bytes.NewReader(rawBody))

			if !testCase.NotAuth {
				ctx := contextUtil.SetUserIDToContext(r.Context(), userID)
				r = r.WithContext(ctx)
			}

			w := httptest.NewRecorder()

			if testCase.PrepareServiceFunc != nil {
				testCase.PrepareServiceFunc(r.Context())
			}

			handler.DeleteURLs(w, r)

			res := w.Result()
			defer res.Body.Close()

			assert.Equal(t, testCase.ExpectedStatusCode, res.StatusCode)
		})
	}
}

func TestRedirectToURLByID(t *testing.T) {
	ctrl := gomock.NewController(t)
	service := handlersmock.NewMockshortenerService(ctrl)

	handler := NewShortenerHandler(
		&config.AppConfig{},
		service,
	)

	type TestCase struct {
		PrepareServiceFunc func(
			ctx context.Context,
			body string,
		)
		Name               string
		Body               string
		ExpectedStatusCode int
	}

	testCases := []TestCase{
		{
			Name: "valid",
			Body: "1234",
			PrepareServiceFunc: func(ctx context.Context, body string) {
				service.
					EXPECT().
					GetByShortURL(ctx, body).
					Return(&domain.ShortenedURL{}, nil)
			},
			ExpectedStatusCode: http.StatusTemporaryRedirect,
		},
		{
			Name: "invalid",
			Body: "1234",
			PrepareServiceFunc: func(ctx context.Context, body string) {
				service.
					EXPECT().
					GetByShortURL(ctx, body).
					Return(nil, errors.New("undefined behavior"))
			},
			ExpectedStatusCode: http.StatusBadRequest,
		},
		{
			Name: "delete url",
			Body: "1234",
			PrepareServiceFunc: func(ctx context.Context, body string) {
				service.
					EXPECT().
					GetByShortURL(ctx, body).
					Return(&domain.ShortenedURL{IsDeleted: true}, nil)
			},
			ExpectedStatusCode: http.StatusGone,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			r := httptest.NewRequest(http.MethodGet, "/", strings.NewReader(testCase.Body))
			r.Header.Set("Content-Type", "text/plain")
			rctx := chi.NewRouteContext()
			rctx.URLParams.Add("id", testCase.Body)
			r = r.WithContext(context.WithValue(r.Context(), chi.RouteCtxKey, rctx))
			w := httptest.NewRecorder()

			if testCase.PrepareServiceFunc != nil {
				testCase.PrepareServiceFunc(r.Context(), testCase.Body)
			}

			handler.RedirectToURLByID(w, r)

			res := w.Result()
			defer res.Body.Close()

			assert.Equal(t, testCase.ExpectedStatusCode, res.StatusCode)
		})
	}
}

func TestPing(t *testing.T) {
	ctrl := gomock.NewController(t)
	service := handlersmock.NewMockshortenerService(ctrl)

	handler := NewShortenerHandler(
		&config.AppConfig{},
		service,
	)

	type TestCase struct {
		PrepareServiceFunc func(
			ctx context.Context,
		)
		Name               string
		ExpectedStatusCode int
	}

	testCases := []TestCase{
		{
			Name: "valid",
			PrepareServiceFunc: func(ctx context.Context) {
				service.
					EXPECT().
					Ping(ctx).
					Return(nil)
			},
			ExpectedStatusCode: http.StatusOK,
		},
		{
			Name: "not valid",
			PrepareServiceFunc: func(ctx context.Context) {
				service.
					EXPECT().
					Ping(ctx).
					Return(errors.New("undefined behavior"))
			},
			ExpectedStatusCode: http.StatusInternalServerError,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			r := httptest.NewRequest(http.MethodGet, "/", nil)
			w := httptest.NewRecorder()

			if testCase.PrepareServiceFunc != nil {
				testCase.PrepareServiceFunc(r.Context())
			}

			handler.Ping(w, r)

			res := w.Result()
			defer res.Body.Close()

			assert.Equal(t, testCase.ExpectedStatusCode, res.StatusCode)
		})
	}
}
