package storage

import (
	"context"

	"github.com/MowlCoder/go-url-shortener/internal/config"
	"github.com/MowlCoder/go-url-shortener/internal/domain"
	"github.com/MowlCoder/go-url-shortener/internal/storage/models"
)

type URLStorage interface {
	SaveSeveralURL(ctx context.Context, dtos []domain.SaveShortURLDto) ([]models.ShortenedURL, error)
	SaveURL(ctx context.Context, dto domain.SaveShortURLDto) (*models.ShortenedURL, error)
	GetByShortURL(ctx context.Context, shortURL string) (*models.ShortenedURL, error)
	GetURLsByUserID(ctx context.Context, userID string) ([]models.ShortenedURL, error)
	DeleteByShortURLs(ctx context.Context, shortURLs []string, userID string) error
	DoDeleteURLTasks(ctx context.Context, tasks []domain.DeleteURLsTask) error
	Ping(ctx context.Context) error
}

func New(appConfig *config.AppConfig) (URLStorage, error) {
	if appConfig.DatabaseDSN != "" {
		return NewDatabaseStorage(appConfig.DatabaseDSN)
	} else if appConfig.FileStoragePath != "" {
		return NewFileStorage(appConfig.FileStoragePath)
	} else {
		return NewInMemoryStorage()
	}
}
