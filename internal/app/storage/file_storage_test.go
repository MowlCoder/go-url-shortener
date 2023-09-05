package storage

import (
	"context"
	"testing"

	"github.com/MowlCoder/go-url-shortener/internal/app/domain"
	"github.com/MowlCoder/go-url-shortener/internal/app/storage/models"

	"github.com/stretchr/testify/assert"
)

func TestFileStorage_SaveURL(t *testing.T) {
	t.Run("Save url", func(t *testing.T) {
		urlToAdd := "https://test.com"
		storage, _ := NewFileStorage("")
		shortenedURL, err := storage.SaveURL(context.Background(), domain.SaveShortUrlDto{
			OriginalURL: urlToAdd,
			ShortURL:    "short-url",
		})

		if assert.NoError(t, err) {
			if assert.NotEmpty(t, shortenedURL) {
				assert.Equal(t, urlToAdd, shortenedURL.OriginalURL)
			}
		}
	})

	t.Run("Save url twice", func(t *testing.T) {
		urlToAdd := "https://test.com"
		storage, _ := NewFileStorage("")
		shortenedURL, err := storage.SaveURL(context.Background(), domain.SaveShortUrlDto{
			OriginalURL: urlToAdd,
			ShortURL:    "short-url-1",
		})

		if assert.NoError(t, err) {
			if assert.NotEmpty(t, shortenedURL) {
				assert.Equal(t, urlToAdd, shortenedURL.OriginalURL)
			}
		}

		secondShortenedURL, err := storage.SaveURL(context.Background(), domain.SaveShortUrlDto{
			OriginalURL: urlToAdd,
			ShortURL:    "short-url-2",
		})

		assert.ErrorIs(t, err, ErrRowConflict)
		assert.Equal(t, secondShortenedURL.ShortURL, shortenedURL.ShortURL)
	})
}

func TestFileStorage_GetOriginalURLByShortURL(t *testing.T) {
	t.Run("Get url", func(t *testing.T) {
		testID := "testid"
		testURL := "https://test.com"
		storage, _ := NewFileStorage("")
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
		storage, _ := NewFileStorage("")

		_, err := storage.GetOriginalURLByShortURL(context.Background(), testID)

		assert.Error(t, err)
	})
}
