package storage

import (
	"math/rand"

	"github.com/MowlCoder/go-url-shortener/internal/app/storage/models"

	"github.com/MowlCoder/go-url-shortener/internal/app/util"
)

type InMemoryStorage struct {
	structure map[string]models.ShortenedURL
}

func NewInMemoryStorage() *InMemoryStorage {
	storage := InMemoryStorage{
		structure: make(map[string]models.ShortenedURL),
	}

	return &storage
}

func (storage *InMemoryStorage) GetOriginalURLByShortURL(shortURL string) (string, error) {
	if url, ok := storage.structure[shortURL]; ok {
		return url.OriginalURL, nil
	}

	return "", errorURLNotFound
}

func (storage *InMemoryStorage) SaveURL(url string) (models.ShortenedURL, error) {
	shortURL := util.Base62Encode(rand.Uint64())
	storage.structure[shortURL] = models.ShortenedURL{
		ID:          len(storage.structure) + 1,
		ShortURL:    shortURL,
		OriginalURL: url,
	}

	return storage.structure[shortURL], nil
}

func (storage *InMemoryStorage) SaveSeveralURL(urls []string) ([]models.ShortenedURL, error) {
	shortenedURLs := make([]models.ShortenedURL, 0, len(urls))

	for _, url := range urls {
		shortenedURL, err := storage.SaveURL(url)

		if err != nil {
			return nil, err
		}

		shortenedURLs = append(shortenedURLs, shortenedURL)
	}

	return shortenedURLs, nil
}

func (storage *InMemoryStorage) Ping() error {
	return nil
}
