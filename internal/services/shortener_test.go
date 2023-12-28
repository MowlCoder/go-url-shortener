package services

import (
	"context"
	"errors"
	"testing"

	"go.uber.org/mock/gomock"

	servicesmocks "github.com/MowlCoder/go-url-shortener/internal/services/mocks"

	"github.com/MowlCoder/go-url-shortener/internal/domain"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/MowlCoder/go-url-shortener/internal/handlers/http/dtos"
)

func TestShortenerService_ShortURL(t *testing.T) {
	ctrl := gomock.NewController(t)
	storage := servicesmocks.NewMockurlStorageForService(ctrl)
	stringsGenerator := servicesmocks.NewMockstringGeneratorService(ctrl)
	deleteQueue := servicesmocks.NewMockdeleteURLQueue(ctrl)

	service := NewShortenerService(
		storage,
		stringsGenerator,
		deleteQueue,
	)

	type TestCase struct {
		PrepareServiceFunc func(
			ctx context.Context,
			body string,
		)
		Name    string
		URL     string
		IsError bool
	}

	testCases := []TestCase{
		{
			Name: "valid",
			URL:  "https://url.com",
			PrepareServiceFunc: func(ctx context.Context, body string) {
				stringsGenerator.
					EXPECT().
					GenerateRandom().
					Return("1234")
				storage.
					EXPECT().
					SaveURL(ctx, domain.SaveShortURLDto{
						OriginalURL: body,
						ShortURL:    "1234",
						UserID:      "1",
					}).
					Return(&domain.ShortenedURL{
						OriginalURL: body,
					}, nil)
			},
			IsError: false,
		},
		{
			Name: "err row conflict",
			URL:  "https://url.com",
			PrepareServiceFunc: func(ctx context.Context, body string) {
				stringsGenerator.
					EXPECT().
					GenerateRandom().
					Return("1234")
				storage.
					EXPECT().
					SaveURL(ctx, domain.SaveShortURLDto{
						OriginalURL: body,
						ShortURL:    "1234",
						UserID:      "1",
					}).
					Return(&domain.ShortenedURL{
						OriginalURL: body,
					}, domain.ErrURLConflict)
			},
			IsError: true,
		},
		{
			Name: "err",
			URL:  "https://url.com",
			PrepareServiceFunc: func(ctx context.Context, body string) {
				stringsGenerator.
					EXPECT().
					GenerateRandom().
					Return("1234")
				storage.
					EXPECT().
					SaveURL(ctx, domain.SaveShortURLDto{
						OriginalURL: body,
						ShortURL:    "1234",
						UserID:      "1",
					}).
					Return(nil, domain.ErrURLConflict)
			},
			IsError: true,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			ctx := context.Background()

			if testCase.PrepareServiceFunc != nil {
				testCase.PrepareServiceFunc(ctx, testCase.URL)
			}

			url, err := service.ShortURL(ctx, testCase.URL, "1")

			if testCase.IsError {
				require.Error(t, err)
			} else {
				assert.Equal(t, url.OriginalURL, testCase.URL)
			}
		})
	}
}

func TestShortenerService_ShortBatchURL(t *testing.T) {
	ctrl := gomock.NewController(t)
	storage := servicesmocks.NewMockurlStorageForService(ctrl)
	stringsGenerator := servicesmocks.NewMockstringGeneratorService(ctrl)
	deleteQueue := servicesmocks.NewMockdeleteURLQueue(ctrl)

	service := NewShortenerService(
		storage,
		stringsGenerator,
		deleteQueue,
	)

	type TestCase struct {
		PrepareServiceFunc func(
			ctx context.Context,
			body []domain.ShortBatchURL,
		)
		Name    string
		Body    []domain.ShortBatchURL
		IsError bool
	}

	testCases := []TestCase{
		{
			Name: "valid",
			Body: []domain.ShortBatchURL{
				{
					OriginalURL:   "https://url.com",
					CorrelationID: "1",
				},
				{
					OriginalURL:   "https://url.com/1",
					CorrelationID: "2",
				},
			},
			PrepareServiceFunc: func(ctx context.Context, body []domain.ShortBatchURL) {
				shortenedUrls := make([]domain.ShortenedURL, 0)

				for _, dto := range body {
					stringsGenerator.
						EXPECT().
						GenerateRandom().
						Return(dto.CorrelationID + "1234")

					shortenedUrls = append(shortenedUrls, domain.ShortenedURL{
						ShortURL:    dto.CorrelationID + "1234",
						OriginalURL: dto.OriginalURL,
					})
				}

				storage.
					EXPECT().
					SaveSeveralURL(ctx, gomock.Any()).
					Return(shortenedUrls, nil)
			},
			IsError: false,
		},
		{
			Name: "error",
			Body: []domain.ShortBatchURL{
				{
					OriginalURL:   "https://url.com",
					CorrelationID: "1",
				},
				{
					OriginalURL:   "https://url.com/1",
					CorrelationID: "2",
				},
			},
			PrepareServiceFunc: func(ctx context.Context, body []domain.ShortBatchURL) {
				for _, dto := range body {
					stringsGenerator.
						EXPECT().
						GenerateRandom().
						Return(dto.CorrelationID + "1234")
				}

				storage.
					EXPECT().
					SaveSeveralURL(ctx, gomock.Any()).
					Return(nil, errors.New("undefined behavior"))
			},
			IsError: true,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			ctx := context.Background()

			if testCase.PrepareServiceFunc != nil {
				testCase.PrepareServiceFunc(ctx, testCase.Body)
			}

			result, err := service.ShortBatchURL(ctx, testCase.Body, "1")

			if testCase.IsError {
				require.Error(t, err)
			} else {
				assert.Equal(t, len(testCase.Body), len(result))

				allFound := true

				for _, dto := range testCase.Body {
					isFound := false

					for _, resDto := range result {
						if dto.CorrelationID == resDto.CorrelationID {
							isFound = true
							break
						}
					}

					if !isFound {
						allFound = false
						break
					}
				}

				assert.Equal(t, true, allFound)
			}
		})
	}
}

func TestShortenerService_GetUserURLs(t *testing.T) {
	ctrl := gomock.NewController(t)
	storage := servicesmocks.NewMockurlStorageForService(ctrl)
	stringsGenerator := servicesmocks.NewMockstringGeneratorService(ctrl)
	deleteQueue := servicesmocks.NewMockdeleteURLQueue(ctrl)

	service := NewShortenerService(
		storage,
		stringsGenerator,
		deleteQueue,
	)

	type TestCase struct {
		PrepareServiceFunc func(
			ctx context.Context,
		)
		Name    string
		IsError bool
	}

	userID := "1"

	testCases := []TestCase{
		{
			Name: "valid",
			PrepareServiceFunc: func(ctx context.Context) {
				storage.
					EXPECT().
					GetURLsByUserID(ctx, userID).
					Return([]domain.ShortenedURL{{}, {}}, nil)
			},
			IsError: false,
		},
		{
			Name: "valid (no content)",
			PrepareServiceFunc: func(ctx context.Context) {
				storage.
					EXPECT().
					GetURLsByUserID(ctx, userID).
					Return([]domain.ShortenedURL{}, nil)
			},
			IsError: false,
		},
		{
			Name: "error",
			PrepareServiceFunc: func(ctx context.Context) {
				storage.
					EXPECT().
					GetURLsByUserID(ctx, userID).
					Return(nil, errors.New("undefined behavior"))
			},
			IsError: true,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			ctx := context.Background()

			if testCase.PrepareServiceFunc != nil {
				testCase.PrepareServiceFunc(ctx)
			}

			_, err := service.GetUserURLs(ctx, userID)

			if testCase.IsError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestShortenerService_DeleteURLs(t *testing.T) {
	ctrl := gomock.NewController(t)
	storage := servicesmocks.NewMockurlStorageForService(ctrl)
	stringsGenerator := servicesmocks.NewMockstringGeneratorService(ctrl)
	deleteQueue := servicesmocks.NewMockdeleteURLQueue(ctrl)

	service := NewShortenerService(
		storage,
		stringsGenerator,
		deleteQueue,
	)

	type TestCase struct {
		PrepareServiceFunc func(
			ctx context.Context,
		)
		Name    string
		Body    dtos.DeleteURLsRequest
		IsError bool
	}

	testCases := []TestCase{
		{
			Name: "valid",
			Body: dtos.DeleteURLsRequest{"123", "1234"},
			PrepareServiceFunc: func(ctx context.Context) {
				deleteQueue.
					EXPECT().
					Push(gomock.Any()).AnyTimes()
			},
			IsError: false,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			ctx := context.Background()

			if testCase.PrepareServiceFunc != nil {
				testCase.PrepareServiceFunc(ctx)
			}

			err := service.DeleteURLs(ctx, testCase.Body, "1")

			if testCase.IsError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestShortenerService_Ping(t *testing.T) {
	ctrl := gomock.NewController(t)
	storage := servicesmocks.NewMockurlStorageForService(ctrl)
	stringsGenerator := servicesmocks.NewMockstringGeneratorService(ctrl)
	deleteQueue := servicesmocks.NewMockdeleteURLQueue(ctrl)

	service := NewShortenerService(
		storage,
		stringsGenerator,
		deleteQueue,
	)

	type TestCase struct {
		PrepareServiceFunc func(
			ctx context.Context,
		)
		Name    string
		IsError bool
	}

	testCases := []TestCase{
		{
			Name: "valid",
			PrepareServiceFunc: func(ctx context.Context) {
				storage.
					EXPECT().
					Ping(ctx).
					Return(nil)
			},
			IsError: false,
		},
		{
			Name: "not valid",
			PrepareServiceFunc: func(ctx context.Context) {
				storage.
					EXPECT().
					Ping(ctx).
					Return(errors.New("undefined behavior"))
			},
			IsError: true,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			ctx := context.Background()

			if testCase.PrepareServiceFunc != nil {
				testCase.PrepareServiceFunc(ctx)
			}

			err := service.Ping(ctx)

			if testCase.IsError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
