package routes

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

func TestRouter(t *testing.T) {
	r := InitRouter()
	srv := httptest.NewServer(r)
	defer srv.Close()

	t.Run("Create link", func(t *testing.T) {
		response, err := http.Post(srv.URL, "text/plain", strings.NewReader("https://practicum.yandex.ru/"))

		defer response.Body.Close()

		require.NoError(t, err)
		assert.Equal(t, http.StatusCreated, response.StatusCode, fmt.Sprintf("Invalid status code. Expected 201, received %d", response.StatusCode))
		assert.Equal(t, "text/plain", response.Header.Get("content-type"), fmt.Sprintf("Invalid content type. Expected text/plain, received %s", response.Header.Get("content-type")))
	})

	t.Run("Create link (invalid body)", func(t *testing.T) {
		response, err := http.Post(srv.URL, "text/plain", nil)

		defer response.Body.Close()

		require.NoError(t, err)
		assert.Equal(t, http.StatusBadRequest, response.StatusCode, fmt.Sprintf("Invalid status code. Expected 400, received %d", response.StatusCode))
	})

	t.Run("Get invalid link", func(t *testing.T) {
		response, err := http.Get(srv.URL + "/invalid-link")

		defer response.Body.Close()

		require.NoError(t, err)
		assert.Equal(t, http.StatusBadRequest, response.StatusCode, fmt.Sprintf("Invalid status code. Expected 400, received %d", response.StatusCode))
	})

	t.Run("Get valid link", func(t *testing.T) {
		response, err := http.Post(srv.URL, "text/plain", strings.NewReader("https://practicum.yandex.ru/"))

		require.NoError(t, err)
		require.Equal(t, http.StatusCreated, response.StatusCode, fmt.Sprintf("Invalid status code. Expected 201, received %d", response.StatusCode))

		defer response.Body.Close()
		body, err := io.ReadAll(response.Body)

		require.NoError(t, err)

		shortURL := string(body)
		shortURLParts := strings.Split(shortURL, "/")

		client := &http.Client{
			CheckRedirect: func(req *http.Request, via []*http.Request) error {
				return http.ErrUseLastResponse
			},
		}

		response, err = client.Get(srv.URL + "/" + shortURLParts[len(shortURLParts)-1])

		defer response.Body.Close()

		require.NoError(t, err)
		assert.Equal(t, http.StatusTemporaryRedirect, response.StatusCode, fmt.Sprintf("Invalid status code. Expected 307, received %d", response.StatusCode))
	})
}
