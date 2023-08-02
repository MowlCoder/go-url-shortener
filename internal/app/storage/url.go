package storage

import (
	"errors"
	"math/rand"

	"github.com/MowlCoder/go-url-shortener/internal/app/util"
)

type URLStorage struct {
	structure map[string]string
}

var (
	errorURLNotFound = errors.New("url not found")
)

func NewURLStorage() *URLStorage {
	return &URLStorage{
		structure: make(map[string]string),
	}
}

func (storage *URLStorage) GetURLByID(id string) (string, error) {
	if url, ok := storage.structure[id]; ok {
		return url, nil
	}

	return "", errorURLNotFound
}

func (storage *URLStorage) SaveURL(url string) (string, error) {
	id := util.Base62Encode(rand.Uint64())
	storage.structure[id] = url

	return id, nil
}
