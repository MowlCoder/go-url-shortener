package storage

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestURLStorage_SaveURL(t *testing.T) {
	t.Run("Save url", func(t *testing.T) {
		urlToAdd := "https://test.com"
		storage := NewURLStorage()
		id, err := storage.SaveURL(urlToAdd)

		if assert.NoError(t, err) {
			if assert.NotEmpty(t, id) {
				assert.Equal(t, urlToAdd, storage.structure[id])
			}
		}
	})
}

func TestURLStorage_GetURLByID(t *testing.T) {
	t.Run("Get url", func(t *testing.T) {
		testID := "testid"
		testURL := "https://test.com"
		storage := NewURLStorage()
		storage.structure[testID] = testURL

		url, err := storage.GetURLByID(testID)

		if assert.NoError(t, err) {
			assert.Equal(t, testURL, url)
		}
	})

	t.Run("Get not existing url", func(t *testing.T) {
		testID := "testid"
		storage := NewURLStorage()

		_, err := storage.GetURLByID(testID)

		assert.Error(t, err)
	})
}
