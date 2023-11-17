package storage

import (
	"context"
	"encoding/json"
	"os"

	"github.com/MowlCoder/go-url-shortener/internal/domain"
	"github.com/MowlCoder/go-url-shortener/internal/storage/models"
)

// FileStorage is storage that store all information in file on disk.
type FileStorage struct {
	structure     map[string]models.ShortenedURL
	file          *os.File
	savingChanges bool
}

// NewFileStorage create file storage with file at given path.
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

// GetByShortURL return model where short url equal given short url.
func (storage *FileStorage) GetByShortURL(ctx context.Context, shortURL string) (*models.ShortenedURL, error) {
	if url, ok := storage.structure[shortURL]; ok {
		return &url, nil
	}

	return nil, domain.ErrURLNotFound
}

// GetURLsByUserID return list of models where user id equal given user id.
func (storage *FileStorage) GetURLsByUserID(ctx context.Context, userID string) ([]models.ShortenedURL, error) {
	urls := make([]models.ShortenedURL, 0)

	for _, value := range storage.structure {
		if value.UserID == userID {
			urls = append(urls, value)
		}
	}

	return urls, nil
}

// FindByOriginalURL return model where original url equal given original url.
func (storage *FileStorage) FindByOriginalURL(ctx context.Context, originalURL string) (*models.ShortenedURL, error) {
	for _, value := range storage.structure {
		if value.OriginalURL == originalURL {
			return &value, nil
		}
	}

	return nil, domain.ErrURLNotFound
}

// SaveURL save short url to the file on disk.
func (storage *FileStorage) SaveURL(ctx context.Context, dto domain.SaveShortURLDto) (*models.ShortenedURL, error) {
	shortenedURL, err := storage.FindByOriginalURL(ctx, dto.OriginalURL)

	if err == nil {
		return shortenedURL, domain.ErrURLConflict
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

// SaveSeveralURL save several short url to the file on disk.
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

// DeleteByShortURLs delete short urls from the file on disk.
func (storage *FileStorage) DeleteByShortURLs(ctx context.Context, shortURLs []string, userID string) error {
	for _, shortURL := range shortURLs {
		shortenedURL := storage.structure[shortURL]

		if shortenedURL.UserID != userID {
			continue
		}

		shortenedURL.IsDeleted = true
		storage.structure[shortURL] = shortenedURL
	}

	if storage.savingChanges {
		storage.saveToFile()
	}

	return nil
}

// DoDeleteURLTasks execute delete tasks and save result to file.
func (storage *FileStorage) DoDeleteURLTasks(ctx context.Context, tasks []domain.DeleteURLsTask) error {
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

	if storage.savingChanges {
		storage.saveToFile()
	}

	return nil
}

// Ping check if storage is available.
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
