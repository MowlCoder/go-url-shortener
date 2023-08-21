package storage

import (
	"testing"

	"github.com/MowlCoder/go-url-shortener/internal/app/storage/models"

	"github.com/stretchr/testify/assert"
)

func TestURLStorage_SaveURL(t *testing.T) {
	t.Run("Save url", func(t *testing.T) {
		urlToAdd := "https://test.com"
		storage := NewURLStorage("")
		id, err := storage.SaveURL(urlToAdd)

		if assert.NoError(t, err) {
			if assert.NotEmpty(t, id) {
				assert.Equal(t, urlToAdd, storage.structure[id].OriginalURL)
			}
		}
	})
}

func TestURLStorage_GetURLByID(t *testing.T) {
	t.Run("Get url", func(t *testing.T) {
		testID := "testid"
		testURL := "https://test.com"
		storage := NewURLStorage("")
		storage.structure[testID] = models.ShortenedURL{
			ShortURL:    testID,
			OriginalURL: testURL,
		}

		url, err := storage.GetOriginalURLByShortURL(testID)

		if assert.NoError(t, err) {
			assert.Equal(t, testURL, url)
		}
	})

	t.Run("Get not existing url", func(t *testing.T) {
		testID := "testid"
		storage := NewURLStorage("")

		_, err := storage.GetOriginalURLByShortURL(testID)

		assert.Error(t, err)
	})
}
