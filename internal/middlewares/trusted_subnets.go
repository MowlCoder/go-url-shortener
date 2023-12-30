package middlewares

import (
	"log"
	"net"
	"net/http"

	"github.com/MowlCoder/go-url-shortener/pkg/httputil"
)

// TrustedSubnetsMiddleware return middleware that not allow to request handlers from not trusted ip
func TrustedSubnetsMiddleware(trustedSubnet string) func(next http.Handler) http.Handler {
	var ipNet *net.IPNet
	var err error

	if trustedSubnet != "" {
		_, ipNet, err = net.ParseCIDR(trustedSubnet)
		if err != nil {
			log.Fatal("initialize trusted subnet middleware:", err)
		}
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if trustedSubnet == "" {
				httputil.SendStatusCode(w, http.StatusForbidden)
				return
			}

			userIP := net.ParseIP(r.Header.Get("X-Real-IP"))

			if !ipNet.Contains(userIP) {
				httputil.SendStatusCode(w, http.StatusForbidden)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}
