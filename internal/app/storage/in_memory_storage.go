package storage

import (
	"context"
	"errors"
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

func (storage *InMemoryStorage) FindByOriginalURL(ctx context.Context, originalURL string) (models.ShortenedURL, error) {
	for _, value := range storage.structure {
		if value.OriginalURL == originalURL {
			return value, nil
		}
	}

	return models.ShortenedURL{}, ErrNotFound
}

func (storage *InMemoryStorage) SaveURL(ctx context.Context, url string) (*models.ShortenedURL, error) {
	shortenedURL, err := storage.FindByOriginalURL(ctx, url)

	if err == nil {
		return &shortenedURL, ErrRowConflict
	}

	shortURL := storage.generateUniqueShortSlug(ctx)
	storage.structure[shortURL] = models.ShortenedURL{
		ID:          len(storage.structure) + 1,
		ShortURL:    shortURL,
		OriginalURL: url,
	}

	shortenedURL = storage.structure[shortURL]

	return &shortenedURL, nil
}

func (storage *InMemoryStorage) SaveSeveralURL(ctx context.Context, urls []string) ([]models.ShortenedURL, error) {
	shortenedURLs := make([]models.ShortenedURL, 0, len(urls))

	for _, url := range urls {
		shortenedURL, err := storage.SaveURL(ctx, url)

		if err != nil && !errors.Is(err, ErrRowConflict) {
			return nil, err
		}

		shortenedURLs = append(shortenedURLs, *shortenedURL)
	}

	return shortenedURLs, nil
}

func (storage *InMemoryStorage) Ping(ctx context.Context) error {
	return nil
}

func (storage *InMemoryStorage) generateUniqueShortSlug(ctx context.Context) string {
	shortURL := ""

	for original := "original"; original != ""; original, _ = storage.GetOriginalURLByShortURL(ctx, shortURL) {
		shortURL = util.Base62Encode(rand.Uint64())
	}

	return shortURL
}
