package storage

import (
	"errors"
	"github.com/MowlCoder/go-url-shortener/internal/app/util"
	"math/rand"
)

type URLStorage struct {
	structure map[string]string
}

var (
	errorURLNotFound = errors.New("url not found")
)

func NewURLStorage() *URLStorage {
	return &URLStorage{
		structure: map[string]string{},
	}
}

func (storage *URLStorage) GetURLByID(id string) (string, error) {
	url, ok := storage.structure[id]

	if !ok {
		return "", errorURLNotFound
	}

	return url, nil
}

func (storage *URLStorage) SaveURL(url string) (string, error) {
	id := util.Base62Encode(rand.Uint64())
	storage.structure[id] = url

	return id, nil
}
