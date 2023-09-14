package storage

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/MowlCoder/go-url-shortener/internal/domain"
	"github.com/MowlCoder/go-url-shortener/internal/storage/models"
)

func TestInMemoryStorage_SaveURL(t *testing.T) {
	t.Run("Save url", func(t *testing.T) {
		urlToAdd := "https://test.com"
		storage, _ := NewInMemoryStorage()
		shortenedURL, err := storage.SaveURL(context.Background(), domain.SaveShortURLDto{
			OriginalURL: urlToAdd,
			ShortURL:    "short-url",
			UserID:      "1",
		})

		if assert.NoError(t, err) {
			if assert.NotEmpty(t, shortenedURL) {
				assert.Equal(t, urlToAdd, shortenedURL.OriginalURL)
			}
		}
	})

	t.Run("Save url twice", func(t *testing.T) {
		urlToAdd := "https://test.com"
		storage, _ := NewInMemoryStorage()
		shortenedURL, err := storage.SaveURL(context.Background(), domain.SaveShortURLDto{
			OriginalURL: urlToAdd,
			ShortURL:    "short-url-1",
			UserID:      "1",
		})

		if assert.NoError(t, err) {
			if assert.NotEmpty(t, shortenedURL) {
				assert.Equal(t, urlToAdd, shortenedURL.OriginalURL)
			}
		}

		secondShortenedURL, err := storage.SaveURL(context.Background(), domain.SaveShortURLDto{
			OriginalURL: urlToAdd,
			ShortURL:    "short-url-2",
			UserID:      "1",
		})

		assert.ErrorIs(t, err, ErrRowConflict)
		assert.Equal(t, secondShortenedURL.ShortURL, shortenedURL.ShortURL)
	})
}

func TestInMemoryStorage_GetOriginalURLByShortURL(t *testing.T) {
	t.Run("Get url", func(t *testing.T) {
		testID := "testid"
		testURL := "https://test.com"
		storage, _ := NewInMemoryStorage()
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
		storage, _ := NewInMemoryStorage()

		_, err := storage.GetOriginalURLByShortURL(context.Background(), testID)

		assert.Error(t, err)
	})
}
