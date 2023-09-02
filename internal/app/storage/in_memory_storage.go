package storage

import (
	"context"
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

func (storage *InMemoryStorage) GetOriginalURLByShortURL(ctx context.Context, shortURL string) (string, error) {
	if url, ok := storage.structure[shortURL]; ok {
		return url.OriginalURL, nil
	}

	return "", errorURLNotFound
}

func (storage *InMemoryStorage) SaveURL(ctx context.Context, url string) (models.ShortenedURL, error) {
	shortURL := util.Base62Encode(rand.Uint64())
	storage.structure[shortURL] = models.ShortenedURL{
		ID:          len(storage.structure) + 1,
		ShortURL:    shortURL,
		OriginalURL: url,
	}

	return storage.structure[shortURL], nil
}

func (storage *InMemoryStorage) SaveSeveralURL(ctx context.Context, urls []string) ([]models.ShortenedURL, error) {
	shortenedURLs := make([]models.ShortenedURL, 0, len(urls))

	for _, url := range urls {
		shortenedURL, err := storage.SaveURL(ctx, url)

		if err != nil {
			return nil, err
		}

		shortenedURLs = append(shortenedURLs, shortenedURL)
	}

	return shortenedURLs, nil
}

func (storage *InMemoryStorage) Ping(ctx context.Context) error {
	return nil
}
