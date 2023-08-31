package storage

import (
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

func (storage *FileStorage) GetOriginalURLByShortURL(shortURL string) (string, error) {
	if url, ok := storage.structure[shortURL]; ok {
		return url.OriginalURL, nil
	}

	return "", errorURLNotFound
}

func (storage *FileStorage) SaveURL(url string) (string, error) {
	shortURL := util.Base62Encode(rand.Uint64())
	storage.structure[shortURL] = models.ShortenedURL{
		ID:          len(storage.structure) + 1,
		ShortURL:    shortURL,
		OriginalURL: url,
	}

	if storage.savingChanges {
		storage.saveToFile()
	}

	return shortURL, nil
}

func (storage *FileStorage) Ping() error {
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
