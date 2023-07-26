package handlers

import (
	"fmt"
	"github.com/MowlCoder/go-url-shortener/internal/app/storage"
	"io"
	"net/http"
	"strings"
)

var urlStorage = storage.NewURLStorage()

func HandleShortURL(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		pathParts := strings.Split(r.URL.Path, "/")

		if len(pathParts) != 2 {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		id := pathParts[1]
		originalURL, err := urlStorage.GetURLById(id)

		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		w.Header().Set("Location", originalURL)
		w.WriteHeader(http.StatusTemporaryRedirect)
	case http.MethodPost:
		body, err := io.ReadAll(r.Body)

		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		if len(body) == 0 {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		id, err := urlStorage.SaveURL(string(body))

		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		w.Header().Set("content-type", "text/plain")
		w.WriteHeader(http.StatusCreated)
		io.WriteString(w, fmt.Sprintf("http://localhost:8080/%s", id))
	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}
