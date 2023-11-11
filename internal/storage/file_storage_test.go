package storage

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/MowlCoder/go-url-shortener/internal/domain"
	"github.com/MowlCoder/go-url-shortener/internal/storage/models"
)

func TestFileStorage_SaveURL(t *testing.T) {
	t.Run("Save url", func(t *testing.T) {
		urlToAdd := "https://test.com"
		storage, _ := NewFileStorage("")
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
		storage, _ := NewFileStorage("")
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

		assert.ErrorIs(t, err, domain.ErrURLConflict)
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

		url, err := storage.GetByShortURL(context.Background(), testID)

		if assert.NoError(t, err) {
			assert.Equal(t, testURL, url.OriginalURL)
		}
	})

	t.Run("Get not existing url", func(t *testing.T) {
		testID := "testid"
		storage, _ := NewFileStorage("")

		_, err := storage.GetByShortURL(context.Background(), testID)

		assert.Error(t, err)
	})
}

func TestFileStorage_DeleteByShortURLs(t *testing.T) {
	t.Run("delete (valid)", func(t *testing.T) {
		testID := "testid"
		testURL := "https://test.com"
		userID := "32"

		storage, _ := NewInMemoryStorage()
		storage.structure[testID] = models.ShortenedURL{
			ShortURL:    testID,
			OriginalURL: testURL,
			IsDeleted:   false,
			UserID:      userID,
		}

		err := storage.DeleteByShortURLs(context.Background(), []string{testID}, userID)
		require.NoError(t, err)

		assert.Equal(t, storage.structure[testID].IsDeleted, true)
	})

	t.Run("delete many (valid)", func(t *testing.T) {
		testID1 := "testid1"
		testID2 := "testid2"
		testID3 := "testid3"
		testURL := "https://test.com"
		userID := "32"

		storage, _ := NewInMemoryStorage()
		storage.structure[testID1] = models.ShortenedURL{
			ShortURL:    testID1,
			OriginalURL: testURL,
			IsDeleted:   false,
			UserID:      userID,
		}

		storage.structure[testID2] = models.ShortenedURL{
			ShortURL:    testID2,
			OriginalURL: testURL,
			IsDeleted:   false,
			UserID:      userID,
		}

		storage.structure[testID3] = models.ShortenedURL{
			ShortURL:    testID3,
			OriginalURL: testURL,
			IsDeleted:   false,
			UserID:      userID,
		}

		err := storage.DeleteByShortURLs(context.Background(), []string{testID1, testID2}, userID)
		require.NoError(t, err)

		assert.Equal(t, storage.structure[testID1].IsDeleted, true)
		assert.Equal(t, storage.structure[testID2].IsDeleted, true)
		assert.Equal(t, storage.structure[testID3].IsDeleted, false)
	})

	t.Run("delete (invalid)", func(t *testing.T) {
		testID := "testid"
		testURL := "https://test.com"
		userIDMy := "32"
		userIDOther := "33"

		storage, _ := NewInMemoryStorage()
		storage.structure[testID] = models.ShortenedURL{
			ShortURL:    testID,
			OriginalURL: testURL,
			IsDeleted:   false,
			UserID:      userIDOther,
		}

		err := storage.DeleteByShortURLs(context.Background(), []string{testID}, userIDMy)
		require.NoError(t, err)

		assert.Equal(t, storage.structure[testID].IsDeleted, false)
	})
}
