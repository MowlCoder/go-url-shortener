package handlers

import (
	"context"
	"github.com/MowlCoder/go-url-shortener/internal/app/storage"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/MowlCoder/go-url-shortener/internal/app/config"

	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestShortURL(t *testing.T) {
	appConfig := &config.AppConfig{}
	urlStorage := storage.NewURLStorage()

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
}

func TestRedirectToURLByID(t *testing.T) {
	appConfig := &config.AppConfig{}
	urlStorage := storage.NewURLStorage()
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
