package handlers

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestHandleShortURL(t *testing.T) {
	t.Run("Create link", func(t *testing.T) {
		r := httptest.NewRequest(http.MethodPost, "/", strings.NewReader("https://practicum.yandex.ru/"))
		w := httptest.NewRecorder()

		HandleShortURL(w, r)

		assert.Equal(t, http.StatusCreated, w.Code, fmt.Sprintf("Invalid status code. Expected 201, received %d", w.Code))
		assert.Equal(t, "text/plain", w.Header().Get("content-type"), fmt.Sprintf("Invalid content type. Expected text/plain, received %s", w.Header().Get("content-type")))
	})

	t.Run("Get link", func(t *testing.T) {
		r := httptest.NewRequest(http.MethodPost, "/", strings.NewReader("https://practicum.yandex.ru/"))
		w := httptest.NewRecorder()

		HandleShortURL(w, r)

		require.Equal(t, http.StatusCreated, w.Code, fmt.Sprintf("Invalid status code. Expected 201, received %d", w.Code))

		res := w.Result()
		defer res.Body.Close()

		respBody, err := io.ReadAll(res.Body)
		require.NoError(t, err)

		shortURL := string(respBody)

		r = httptest.NewRequest(http.MethodGet, shortURL, nil)
		w = httptest.NewRecorder()

		HandleShortURL(w, r)

		assert.Equal(t, http.StatusTemporaryRedirect, w.Code, fmt.Sprintf("Invalid status code. Expected 307, received %d", w.Code))
	})

	t.Run("Get invalid link", func(t *testing.T) {
		r := httptest.NewRequest(http.MethodGet, "/invalid-link", nil)
		w := httptest.NewRecorder()

		HandleShortURL(w, r)

		assert.Equal(t, http.StatusBadRequest, w.Code, fmt.Sprintf("Invalid status code. Expected 400, received %d", w.Code))
	})

	t.Run("Invalid method type", func(t *testing.T) {
		r := httptest.NewRequest(http.MethodDelete, "/", nil)
		w := httptest.NewRecorder()

		HandleShortURL(w, r)

		require.Equal(t, http.StatusMethodNotAllowed, w.Code, fmt.Sprintf("Invalid status code. Expected 405, received %d", w.Code))
	})
}
