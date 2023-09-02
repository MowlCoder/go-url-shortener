package storage

import (
	"context"
	"encoding/json"
	"math/rand"
	"os"

	"github.com/MowlCoder/go-url-shortener/internal/app/storage/models"

	"github.com/MowlCoder/go-url-shortener/internal/app/util"
)

type FileStorage struct {
	structure     map[string]models.ShortenedURL
	file          *os.File
	savingChanges bool
}

func NewFileStorage(fileStoragePath string) *FileStorage {
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

	return &storage
}

func (storage *FileStorage) GetOriginalURLByShortURL(ctx context.Context, shortURL string) (string, error) {
	if url, ok := storage.structure[shortURL]; ok {
		return url.OriginalURL, nil
	}

	return "", errorURLNotFound
}

func (storage *FileStorage) SaveURL(ctx context.Context, url string) (models.ShortenedURL, error) {
	shortURL := storage.generateUniqueShortSlug(ctx)
	storage.structure[shortURL] = models.ShortenedURL{
		ID:          len(storage.structure) + 1,
		ShortURL:    shortURL,
		OriginalURL: url,
	}

	if storage.savingChanges {
		storage.saveToFile()
	}

	return storage.structure[shortURL], nil
}

func (storage *FileStorage) SaveSeveralURL(ctx context.Context, urls []string) ([]models.ShortenedURL, error) {
	shortenedURLs := make([]models.ShortenedURL, 0, len(urls))

	for _, url := range urls {
		shortURL := storage.generateUniqueShortSlug(ctx)
		storage.structure[shortURL] = models.ShortenedURL{
			ID:          len(storage.structure) + 1,
			ShortURL:    shortURL,
			OriginalURL: url,
		}

		shortenedURLs = append(shortenedURLs, storage.structure[shortURL])
	}

	if storage.savingChanges {
		storage.saveToFile()
	}

	return shortenedURLs, nil
}

func (storage *FileStorage) Ping(ctx context.Context) error {
	return nil
}

func (storage *FileStorage) generateUniqueShortSlug(ctx context.Context) string {
	shortURL := ""

	for original := "original"; original != ""; original, _ = storage.GetOriginalURLByShortURL(ctx, shortURL) {
		shortURL = util.Base62Encode(rand.Uint64())
	}

	return shortURL
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
