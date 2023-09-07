package storage

import (
	"context"
	"errors"

	"github.com/MowlCoder/go-url-shortener/internal/app/domain"
	"github.com/MowlCoder/go-url-shortener/internal/app/storage/models"
)

type InMemoryStorage struct {
	structure map[string]models.ShortenedURL
}

func NewInMemoryStorage() (*InMemoryStorage, error) {
	storage := InMemoryStorage{
		structure: make(map[string]models.ShortenedURL),
	}

	return &storage, nil
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

func (storage *InMemoryStorage) SaveURL(ctx context.Context, dto domain.SaveShortURLDto) (*models.ShortenedURL, error) {
	shortenedURL, err := storage.FindByOriginalURL(ctx, dto.OriginalURL)

	if err == nil {
		return &shortenedURL, ErrRowConflict
	}

	storage.structure[dto.ShortURL] = models.ShortenedURL{
		ID:          len(storage.structure) + 1,
		ShortURL:    dto.ShortURL,
		OriginalURL: dto.OriginalURL,
	}

	shortenedURL = storage.structure[dto.ShortURL]

	return &shortenedURL, nil
}

func (storage *InMemoryStorage) SaveSeveralURL(ctx context.Context, dtos []domain.SaveShortURLDto) ([]models.ShortenedURL, error) {
	shortenedURLs := make([]models.ShortenedURL, 0, len(dtos))

	for _, dto := range dtos {
		shortenedURL, err := storage.SaveURL(ctx, dto)

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
