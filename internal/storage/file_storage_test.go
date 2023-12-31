package storage

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/MowlCoder/go-url-shortener/internal/domain"
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

func TestFileStorage_GetURLsByUserID(t *testing.T) {
	type TestCase struct {
		Name           string
		UserID         string
		PrepareStorage func() *FileStorage
		IsError        bool
		ExpectedLen    int
	}

	testCases := []TestCase{
		{
			Name:   "Get urls (valid)",
			UserID: "123",
			PrepareStorage: func() *FileStorage {
				storage, _ := NewFileStorage("")
				storage.structure["1"] = domain.ShortenedURL{
					ID:          1,
					OriginalURL: "123",
					UserID:      "123",
					ShortURL:    "123",
				}

				return storage
			},
			IsError:     false,
			ExpectedLen: 1,
		},
		{
			Name:   "Get urls (zero)",
			UserID: "123",
			PrepareStorage: func() *FileStorage {
				storage, _ := NewFileStorage("")
				return storage
			},
			IsError:     false,
			ExpectedLen: 0,
		},
	}

	for _, testCase := range testCases {
		storage := testCase.PrepareStorage()
		urls, err := storage.GetURLsByUserID(context.Background(), testCase.UserID)

		if testCase.IsError {
			require.Error(t, err)
		} else {
			require.NoError(t, err)
			assert.Len(t, urls, testCase.ExpectedLen)
		}
	}
}

func TestFileStorage_SaveSeveralURL(t *testing.T) {
	type TestCase struct {
		Name    string
		DTOs    []domain.SaveShortURLDto
		IsError bool
	}

	testCases := []TestCase{
		{
			Name:    "Save URLs",
			IsError: false,
			DTOs: []domain.SaveShortURLDto{
				{
					OriginalURL: "123",
					ShortURL:    "123",
					UserID:      "123",
				},
			},
		},
	}

	for _, testCase := range testCases {
		storage, _ := NewFileStorage("")
		urls, err := storage.SaveSeveralURL(context.Background(), testCase.DTOs)

		if testCase.IsError {
			require.Error(t, err)
		} else {
			require.NoError(t, err)
			assert.Len(t, urls, len(testCase.DTOs))
		}
	}
}

func TestFileStorage_GetOriginalURLByShortURL(t *testing.T) {
	t.Run("Get url", func(t *testing.T) {
		testID := "testid"
		testURL := "https://test.com"
		storage, _ := NewFileStorage("")
		storage.structure[testID] = domain.ShortenedURL{
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

		storage, _ := NewFileStorage("")
		storage.structure[testID] = domain.ShortenedURL{
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

		storage, _ := NewFileStorage("")
		storage.structure[testID1] = domain.ShortenedURL{
			ShortURL:    testID1,
			OriginalURL: testURL,
			IsDeleted:   false,
			UserID:      userID,
		}

		storage.structure[testID2] = domain.ShortenedURL{
			ShortURL:    testID2,
			OriginalURL: testURL,
			IsDeleted:   false,
			UserID:      userID,
		}

		storage.structure[testID3] = domain.ShortenedURL{
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

		storage, _ := NewFileStorage("")
		storage.structure[testID] = domain.ShortenedURL{
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

func TestFileStorage_Ping(t *testing.T) {
	storage, _ := NewFileStorage("")

	t.Run("valid ping", func(t *testing.T) {
		err := storage.Ping(context.Background())
		assert.NoError(t, err)
	})
}
