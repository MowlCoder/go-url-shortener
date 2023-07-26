package handlers

import (
	"fmt"
	"github.com/MowlCoder/go-url-shortener/internal/app/util"
	"io"
	"math/rand"
	"net/http"
	"strings"
)

var urlStorage = map[string]string{}

func HandleShortUrl(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		pathParts := strings.Split(r.URL.Path, "/")

		if len(pathParts) != 2 {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		id := pathParts[1]
		originalUrl, ok := urlStorage[id]

		if !ok {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		w.Header().Set("Location", originalUrl)
		w.WriteHeader(http.StatusTemporaryRedirect)

		return
	}

	body, err := io.ReadAll(r.Body)

	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	if len(body) == 0 {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	shortUrlId := util.Base62Encode(rand.Uint64())

	urlStorage[shortUrlId] = string(body)

	w.Header().Set("content-type", "text/plain")
	io.WriteString(w, fmt.Sprintf("http://localhost:8080/%s", shortUrlId))
}
