package storage

import (
	"context"

	"github.com/MowlCoder/go-url-shortener/internal/config"
	"github.com/MowlCoder/go-url-shortener/internal/domain"
)

// URLStorage is common interface for all storages.
type URLStorage interface {
	SaveSeveralURL(ctx context.Context, dtos []domain.SaveShortURLDto) ([]domain.ShortenedURL, error)
	SaveURL(ctx context.Context, dto domain.SaveShortURLDto) (*domain.ShortenedURL, error)
	GetByShortURL(ctx context.Context, shortURL string) (*domain.ShortenedURL, error)
	GetURLsByUserID(ctx context.Context, userID string) ([]domain.ShortenedURL, error)
	DeleteByShortURLs(ctx context.Context, shortURLs []string, userID string) error
	DoDeleteURLTasks(ctx context.Context, tasks []domain.DeleteURLsTask) error
	GetInternalStats(ctx context.Context) (*domain.InternalStats, error)
	Ping(ctx context.Context) error
}

// New create URLStorage base on given config.
func New(appConfig *config.AppConfig) (URLStorage, error) {
	switch {
	case appConfig.DatabaseDSN != "":
		return NewDatabaseStorage(appConfig.DatabaseDSN)
	case appConfig.FileStoragePath != "":
		return NewFileStorage(appConfig.FileStoragePath)
	default:
		return NewInMemoryStorage()
	}
}
