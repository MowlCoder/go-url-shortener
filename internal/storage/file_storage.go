package storage

import (
	"context"
	"encoding/json"
	"os"

	"github.com/MowlCoder/go-url-shortener/internal/domain"
	"github.com/MowlCoder/go-url-shortener/internal/storage/models"
)

type FileStorage struct {
	structure     map[string]models.ShortenedURL
	file          *os.File
	savingChanges bool
}

func NewFileStorage(fileStoragePath string) (*FileStorage, error) {
	storage := FileStorage{
		structure:     make(map[string]models.ShortenedURL),
		savingChanges: false,
	}

	if fileStoragePath != "" {
		if file, err := os.OpenFile(fileStoragePath, os.O_RDWR|os.O_CREATE, 0644); err == nil {
			storage.file = file
			storage.savingChanges = true
		}
	}

	if storage.savingChanges {
		storage.parseFromFile()
	}

	return &storage, nil
}

func (storage *FileStorage) GetOriginalURLByShortURL(ctx context.Context, shortURL string) (string, error) {
	if url, ok := storage.structure[shortURL]; ok {
		return url.OriginalURL, nil
	}

	return "", errorURLNotFound
}

func (storage *FileStorage) FindByOriginalURL(ctx context.Context, originalURL string) (*models.ShortenedURL, error) {
	for _, value := range storage.structure {
		if value.OriginalURL == originalURL {
			return &value, nil
		}
	}

	return nil, ErrNotFound
}

func (storage *FileStorage) SaveURL(ctx context.Context, dto domain.SaveShortURLDto) (*models.ShortenedURL, error) {
	shortenedURL, err := storage.FindByOriginalURL(ctx, dto.OriginalURL)

	if err == nil {
		return shortenedURL, ErrRowConflict
	}

	shortenedURL = &models.ShortenedURL{
		ID:          len(storage.structure) + 1,
		ShortURL:    dto.ShortURL,
		OriginalURL: dto.OriginalURL,
	}
	storage.structure[dto.ShortURL] = *shortenedURL

	if storage.savingChanges {
		storage.saveToFile()
	}

	return shortenedURL, nil
}

func (storage *FileStorage) SaveSeveralURL(ctx context.Context, dtos []domain.SaveShortURLDto) ([]models.ShortenedURL, error) {
	shortenedURLs := make([]models.ShortenedURL, 0, len(dtos))

	for _, dto := range dtos {
		shortenedURL, err := storage.FindByOriginalURL(ctx, dto.OriginalURL)

		if err != nil {
			shortenedURL = &models.ShortenedURL{
				ID:          len(storage.structure) + 1,
				ShortURL:    dto.ShortURL,
				OriginalURL: dto.OriginalURL,
			}

			storage.structure[dto.ShortURL] = *shortenedURL
		}

		shortenedURLs = append(shortenedURLs, *shortenedURL)
	}

	if storage.savingChanges {
		storage.saveToFile()
	}

	return shortenedURLs, nil
}

func (storage *FileStorage) Ping(ctx context.Context) error {
	return nil
}

func (storage *FileStorage) parseFromFile() error {
	b, err := os.ReadFile(storage.file.Name())

	if err != nil {
		return err
	}

	if err := json.Unmarshal(b, &storage.structure); err != nil {
		return err
	}

	return nil
}

func (storage *FileStorage) saveToFile() error {
	b, err := json.Marshal(&storage.structure)

	if err != nil {
		return err
	}

	return os.WriteFile(storage.file.Name(), b, os.ModeAppend)
}