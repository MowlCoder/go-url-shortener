package storage

import (
	"context"
	"testing"

	"github.com/MowlCoder/go-url-shortener/internal/app/storage/models"

	"github.com/stretchr/testify/assert"
)

func TestFileStorage_SaveURL(t *testing.T) {
	t.Run("Save url", func(t *testing.T) {
		urlToAdd := "https://test.com"
		storage := NewFileStorage("/tmp/go-url-shortener-test.json")
		shortenedURL, err := storage.SaveURL(context.Background(), urlToAdd)

		if assert.NoError(t, err) {
			if assert.NotEmpty(t, shortenedURL) {
				assert.Equal(t, urlToAdd, shortenedURL.OriginalURL)
			}
		}
	})
}

func TestFileStorage_GetOriginalURLByShortURL(t *testing.T) {
	t.Run("Get url", func(t *testing.T) {
		testID := "testid"
		testURL := "https://test.com"
		storage := NewFileStorage("/tmp/go-url-shortener-test.json")
		storage.structure[testID] = models.ShortenedURL{
			ShortURL:    testID,
			OriginalURL: testURL,
		}

		url, err := storage.GetOriginalURLByShortURL(context.Background(), testID)

		if assert.NoError(t, err) {
			assert.Equal(t, testURL, url)
		}
	})

	t.Run("Get not existing url", func(t *testing.T) {
		testID := "testid"
		storage := NewFileStorage("/tmp/go-url-shortener-test.json")

		_, err := storage.GetOriginalURLByShortURL(context.Background(), testID)

		assert.Error(t, err)
	})
}
