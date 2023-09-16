package middlewares

import (
	"net/http"

	"github.com/MowlCoder/go-url-shortener/internal/context"
	"github.com/MowlCoder/go-url-shortener/internal/handlers"
	"github.com/MowlCoder/go-url-shortener/internal/jwt"
)

type UserService interface {
	GenerateUniqueID() string
}

const CookieName = "token"

func AuthMiddleware(handler http.Handler, userService UserService) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authHandler(w, r, handler, userService)
	})
}

func authHandler(w http.ResponseWriter, r *http.Request, handler http.Handler, userService UserService) {
	var tokenString string

	cookie, err := r.Cookie(CookieName)

	if err != nil {
		tokenString, err = jwt.GenerateToken(userService.GenerateUniqueID())

		if err != nil {
			handlers.SendStatusCode(w, http.StatusInternalServerError)
			return
		}

		http.SetCookie(w, &http.Cookie{
			Name:  CookieName,
			Value: tokenString,
		})
	} else {
		tokenString = cookie.Value
	}

	jwtClaim, err := jwt.ParseToken(tokenString)

	if err != nil {
		handlers.SendStatusCode(w, http.StatusUnauthorized)
		return
	}

	ctx := context.SetUserIDToContext(r.Context(), jwtClaim.UserID)

	handler.ServeHTTP(w, r.WithContext(ctx))
}
