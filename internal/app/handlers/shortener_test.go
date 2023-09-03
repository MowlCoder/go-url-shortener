package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/MowlCoder/go-url-shortener/internal/app/handlers/dtos"

	"github.com/MowlCoder/go-url-shortener/internal/app/storage"

	"github.com/MowlCoder/go-url-shortener/internal/app/config"

	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestShortURL(t *testing.T) {
	appConfig := &config.AppConfig{}
	urlStorage := storage.NewInMemoryStorage()

	handler := NewShortenerHandler(appConfig, urlStorage)

	type want struct {
		code        int
		contentType string
	}

	tests := []struct {
		name string
		body io.Reader
		want want
	}{
		{
			name: "Create short link (valid)",
			body: strings.NewReader("https://practicum.yandex.ru"),
			want: want{
				code:        201,
				contentType: "text/plain",
			},
		},
		{
			name: "Create short link (invalid)",
			body: nil,
			want: want{
				code:        400,
				contentType: "",
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			request := httptest.NewRequest(http.MethodPost, "/", test.body)
			w := httptest.NewRecorder()

			handler.ShortURL(w, request)

			res := w.Result()

			assert.Equal(t, test.want.code, res.StatusCode)
			defer res.Body.Close()

			assert.Equal(t, test.want.contentType, res.Header.Get("Content-Type"))
		})
	}

	t.Run("Create short link twice", func(t *testing.T) {
		request := httptest.NewRequest(http.MethodPost, "/", strings.NewReader("https://practicum.yandex.ru/123"))
		w := httptest.NewRecorder()

		handler.ShortURL(w, request)

		res := w.Result()

		assert.Equal(t, 201, res.StatusCode)
		defer res.Body.Close()
		assert.Equal(t, "text/plain", res.Header.Get("Content-Type"))

		request = httptest.NewRequest(http.MethodPost, "/", strings.NewReader("https://practicum.yandex.ru/123"))
		w = httptest.NewRecorder()
		handler.ShortURL(w, request)

		res = w.Result()
		assert.Equal(t, 409, res.StatusCode)
		defer res.Body.Close()
		assert.Equal(t, "text/plain", res.Header.Get("Content-Type"))
	})
}

func TestShortURLJSON(t *testing.T) {
	appConfig := &config.AppConfig{}
	urlStorage := storage.NewInMemoryStorage()

	handler := NewShortenerHandler(appConfig, urlStorage)

	type want struct {
		code        int
		contentType string
	}

	tests := []struct {
		name string
		body dtos.ShortURLDto
		want want
	}{
		{
			name: "Create short link (valid)",
			body: dtos.ShortURLDto{
				URL: "https://vk.com",
			},
			want: want{
				code:        201,
				contentType: "application/json",
			},
		},
		{
			name: "Create short link (invalid)",
			body: dtos.ShortURLDto{
				URL: "",
			},
			want: want{
				code:        400,
				contentType: "",
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			jsonBody, err := json.Marshal(test.body)

			require.NoError(t, err)

			request := httptest.NewRequest(http.MethodPost, "/api/shorten", bytes.NewReader(jsonBody))
			request.Header.Set("content-type", "application/json")
			w := httptest.NewRecorder()

			handler.ShortURLJSON(w, request)

			res := w.Result()

			assert.Equal(t, test.want.code, res.StatusCode)
			defer res.Body.Close()

			assert.Equal(t, test.want.contentType, res.Header.Get("Content-Type"))
		})
	}

	t.Run("Create short link twice", func(t *testing.T) {
		jsonBody, err := json.Marshal(dtos.ShortURLDto{
			URL: "https://vk.com/123",
		})

		require.NoError(t, err)

		request := httptest.NewRequest(http.MethodPost, "/", bytes.NewReader(jsonBody))
		w := httptest.NewRecorder()

		handler.ShortURLJSON(w, request)

		res := w.Result()

		assert.Equal(t, 201, res.StatusCode)
		defer res.Body.Close()
		assert.Equal(t, "application/json", res.Header.Get("Content-Type"))

		request = httptest.NewRequest(http.MethodPost, "/", bytes.NewReader(jsonBody))
		w = httptest.NewRecorder()
		handler.ShortURLJSON(w, request)

		res = w.Result()
		assert.Equal(t, 409, res.StatusCode)
		defer res.Body.Close()
		assert.Equal(t, "application/json", res.Header.Get("Content-Type"))
	})
}

func TestShortBatchURL(t *testing.T) {
	appConfig := &config.AppConfig{}
	urlStorage := storage.NewInMemoryStorage()

	handler := NewShortenerHandler(appConfig, urlStorage)

	type want struct {
		code        int
		contentType string
	}

	tests := []struct {
		name string
		body []dtos.ShortBatchURLDto
		want want
	}{
		{
			name: "Create short links (valid)",
			body: []dtos.ShortBatchURLDto{
				{
					OriginalURL:   "https://youtube.com/1",
					CorrelationID: "123",
				},
				{
					OriginalURL:   "https://youtube.com/2",
					CorrelationID: "123456",
				},
				{
					OriginalURL:   "https://youtube.com/3",
					CorrelationID: "1236789",
				},
			},
			want: want{
				code:        201,
				contentType: "application/json",
			},
		},
		{
			name: "Create short link (invalid body)",
			body: nil,
			want: want{
				code:        400,
				contentType: "",
			},
		},
		{
			name: "Create short link (valid empty body)",
			body: []dtos.ShortBatchURLDto{},
			want: want{
				code:        400,
				contentType: "",
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			jsonBody, err := json.Marshal(test.body)

			require.NoError(t, err)

			request := httptest.NewRequest(http.MethodPost, "/api/shorten/batch", bytes.NewReader(jsonBody))
			request.Header.Set("content-type", "application/json")
			w := httptest.NewRecorder()

			handler.ShortBatchURL(w, request)

			res := w.Result()

			assert.Equal(t, test.want.code, res.StatusCode)
			defer res.Body.Close()

			assert.Equal(t, test.want.contentType, res.Header.Get("Content-Type"))

			if test.want.code == 201 {
				response, err := io.ReadAll(res.Body)
				require.NoError(t, err)

				var responseBody []dtos.ShortBatchURLResponse
				require.NoError(t, json.Unmarshal(response, &responseBody))

				assert.Equal(t, len(test.body), len(responseBody))

				allFound := true

				for _, dto := range test.body {
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

func TestRedirectToURLByID(t *testing.T) {
	appConfig := &config.AppConfig{}
	urlStorage := storage.NewInMemoryStorage()
	handler := NewShortenerHandler(appConfig, urlStorage)

	type want struct {
		code int
	}

	tests := []struct {
		name          string
		preCreateLink bool
		want          want
	}{
		{
			name:          "Get original url (valid)",
			preCreateLink: true,
			want: want{
				code: 307,
			},
		},
		{
			name:          "Get original url (invalid)",
			preCreateLink: false,
			want: want{
				code: 400,
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			id := "test-id"

			if test.preCreateLink {
				request := httptest.NewRequest(http.MethodPost, "/", strings.NewReader("https://practicum.yandex.ru"))
				w := httptest.NewRecorder()
				handler.ShortURL(w, request)

				res := w.Result()

				require.Equal(t, http.StatusCreated, res.StatusCode)
				defer res.Body.Close()

				resBody, err := io.ReadAll(res.Body)
				require.NoError(t, err)

				shortURL := string(resBody)
				shortURLParts := strings.Split(shortURL, "/")

				id = shortURLParts[1]
			}

			request := httptest.NewRequest(http.MethodGet, "/{id}", nil)
			w := httptest.NewRecorder()

			rctx := chi.NewRouteContext()
			rctx.URLParams.Add("id", id)

			request = request.WithContext(context.WithValue(request.Context(), chi.RouteCtxKey, rctx))

			handler.RedirectToURLByID(w, request)

			res := w.Result()

			assert.Equal(t, test.want.code, res.StatusCode)
			defer res.Body.Close()
		})
	}
}
