package middlewares

import (
	"net/http"

	"github.com/MowlCoder/go-url-shortener/internal/context"
	"github.com/MowlCoder/go-url-shortener/internal/jwt"
	"github.com/MowlCoder/go-url-shortener/pkg/httputil"
)

type userService interface {
	GenerateUniqueID() string
}

// CookieName is cookie name where store token.
const CookieName = "token"

// AuthMiddleware handle authorization. If user not middleware create token and save in cookie.
// If user provide valid token, parse token and save user id in request context.
func AuthMiddleware(handler http.Handler, userService userService) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authHandler(w, r, handler, userService)
	})
}

func authHandler(w http.ResponseWriter, r *http.Request, handler http.Handler, userService userService) {
	var tokenString string

	cookie, err := r.Cookie(CookieName)

	if err != nil {
		tokenString, err = jwt.GenerateToken(userService.GenerateUniqueID())

		if err != nil {
			httputil.SendStatusCode(w, http.StatusInternalServerError)
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
		httputil.SendStatusCode(w, http.StatusUnauthorized)
		return
	}

	ctx := context.SetUserIDToContext(r.Context(), jwtClaim.UserID)

	handler.ServeHTTP(w, r.WithContext(ctx))
}
