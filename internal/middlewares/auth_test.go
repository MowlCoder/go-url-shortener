package middlewares

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/MowlCoder/go-url-shortener/internal/services"
)

func TestAuthMiddleware(t *testing.T) {
	t.Run("auth middleware", func(t *testing.T) {
		userService := services.NewUserService()

		request := httptest.NewRequest(http.MethodPost, "/", strings.NewReader("https://practicum.yandex.ru"))
		w := httptest.NewRecorder()
		authHandler(w, request, http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
			w.WriteString("OK")
		}), userService)

		res := w.Result()
		isFoundTokenCookie := false

		defer res.Body.Close()

		for _, cookie := range res.Cookies() {
			if cookie.Name == "token" {
				isFoundTokenCookie = true
			}
		}

		assert.True(t, isFoundTokenCookie)
	})
}
