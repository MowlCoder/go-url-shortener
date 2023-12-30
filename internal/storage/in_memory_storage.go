package storage

import (
	"context"
	"errors"

	"github.com/MowlCoder/go-url-shortener/internal/domain"
)

// InMemoryStorage is storage that store all information in memory.
type InMemoryStorage struct {
	structure map[string]domain.ShortenedURL
}

// NewInMemoryStorage create in memory storage.
func NewInMemoryStorage() (*InMemoryStorage, error) {
	storage := InMemoryStorage{
		structure: make(map[string]domain.ShortenedURL),
	}

	return &storage, nil
}

// GetByShortURL return model where short url equal given short url.
func (storage *InMemoryStorage) GetByShortURL(ctx context.Context, shortURL string) (*domain.ShortenedURL, error) {
	if url, ok := storage.structure[shortURL]; ok {
		return &url, nil
	}

	return nil, domain.ErrURLNotFound
}

// GetURLsByUserID return list of models where user id equal given user id.
func (storage *InMemoryStorage) GetURLsByUserID(ctx context.Context, userID string) ([]domain.ShortenedURL, error) {
	urls := make([]domain.ShortenedURL, 0)

	for _, value := range storage.structure {
		if value.UserID == userID {
			urls = append(urls, value)
		}
	}

	return urls, nil
}

// FindByOriginalURL return model where original url equal given original url.
func (storage *InMemoryStorage) FindByOriginalURL(ctx context.Context, originalURL string) (domain.ShortenedURL, error) {
	for _, value := range storage.structure {
		if value.OriginalURL == originalURL {
			return value, nil
		}
	}

	return domain.ShortenedURL{}, domain.ErrURLNotFound
}

// SaveURL save short url to the memory.
func (storage *InMemoryStorage) SaveURL(ctx context.Context, dto domain.SaveShortURLDto) (*domain.ShortenedURL, error) {
	shortenedURL, err := storage.FindByOriginalURL(ctx, dto.OriginalURL)

	if err == nil {
		return &shortenedURL, domain.ErrURLConflict
	}

	storage.structure[dto.ShortURL] = domain.ShortenedURL{
		ID:          len(storage.structure) + 1,
		ShortURL:    dto.ShortURL,
		OriginalURL: dto.OriginalURL,
	}

	shortenedURL = storage.structure[dto.ShortURL]

	return &shortenedURL, nil
}

// SaveSeveralURL save several short url to the memory.
func (storage *InMemoryStorage) SaveSeveralURL(ctx context.Context, dtos []domain.SaveShortURLDto) ([]domain.ShortenedURL, error) {
	shortenedURLs := make([]domain.ShortenedURL, 0, len(dtos))

	for _, dto := range dtos {
		shortenedURL, err := storage.SaveURL(ctx, dto)

		if err != nil && !errors.Is(err, domain.ErrURLConflict) {
			return nil, err
		}

		shortenedURLs = append(shortenedURLs, *shortenedURL)
	}

	return shortenedURLs, nil
}

// DeleteByShortURLs delete short urls from the memory.
func (storage *InMemoryStorage) DeleteByShortURLs(ctx context.Context, shortURLs []string, userID string) error {
	for _, shortURL := range shortURLs {
		shortenedURL := storage.structure[shortURL]

		if shortenedURL.UserID != userID {
			continue
		}

		shortenedURL.IsDeleted = true
		storage.structure[shortURL] = shortenedURL
	}

	return nil
}

// DoDeleteURLTasks execute delete tasks and save result in the memory.
func (storage *InMemoryStorage) DoDeleteURLTasks(ctx context.Context, tasks []domain.DeleteURLsTask) error {
	for _, task := range tasks {
		for _, shortURL := range task.ShortURLs {
			shortenedURL := storage.structure[shortURL]

			if shortenedURL.UserID != task.UserID {
				continue
			}

			shortenedURL.IsDeleted = true
			storage.structure[shortURL] = shortenedURL
		}
	}

	return nil
}

// GetInternalStats get internal stats for metrics.
func (storage *InMemoryStorage) GetInternalStats(ctx context.Context) (*domain.InternalStats, error) {
	stats := domain.InternalStats{}
	uniqueUsers := make(map[string]struct{})

	for _, val := range storage.structure {
		uniqueUsers[val.UserID] = struct{}{}
	}

	stats.URLs = len(storage.structure)
	stats.Users = len(uniqueUsers)

	return &stats, nil
}

// Ping check if storage is available.
func (storage *InMemoryStorage) Ping(ctx context.Context) error {
	return nil
}
