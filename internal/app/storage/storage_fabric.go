package storage

import (
	"context"

	"github.com/MowlCoder/go-url-shortener/internal/app/config"
	"github.com/MowlCoder/go-url-shortener/internal/app/domain"
	"github.com/MowlCoder/go-url-shortener/internal/app/storage/models"
)

type URLStorage interface {
	SaveSeveralURL(ctx context.Context, dtos []domain.SaveShortUrlDto) ([]models.ShortenedURL, error)
	SaveURL(ctx context.Context, dto domain.SaveShortUrlDto) (*models.ShortenedURL, error)
	GetOriginalURLByShortURL(ctx context.Context, shortURL string) (string, error)
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
