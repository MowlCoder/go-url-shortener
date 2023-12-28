package services

import (
	"context"

	"github.com/MowlCoder/go-url-shortener/internal/domain"
)

type urlStorageForService interface {
	SaveSeveralURL(ctx context.Context, dtos []domain.SaveShortURLDto) ([]domain.ShortenedURL, error)
	SaveURL(ctx context.Context, dto domain.SaveShortURLDto) (*domain.ShortenedURL, error)
	GetByShortURL(ctx context.Context, shortURL string) (*domain.ShortenedURL, error)
	GetURLsByUserID(ctx context.Context, userID string) ([]domain.ShortenedURL, error)
	DeleteByShortURLs(ctx context.Context, shortURLs []string, userID string) error
	GetInternalStats(ctx context.Context) (*domain.InternalStats, error)
	Ping(ctx context.Context) error
}

type stringGeneratorService interface {
	GenerateRandom() string
}

type deleteURLQueue interface {
	Push(task *domain.DeleteURLsTask)
}

type ShortenerService struct {
	urlStorage      urlStorageForService
	stringGenerator stringGeneratorService
	deleteURLQueue  deleteURLQueue
}

func NewShortenerService(
	urlStorage urlStorageForService,
	stringGenerator stringGeneratorService,
	deleteURLQueue deleteURLQueue,
) *ShortenerService {
	return &ShortenerService{
		urlStorage:      urlStorage,
		stringGenerator: stringGenerator,
		deleteURLQueue:  deleteURLQueue,
	}
}

func (s *ShortenerService) ShortURL(ctx context.Context, url string, userID string) (*domain.ShortenedURL, error) {
	shortURL := s.stringGenerator.GenerateRandom()

	return s.urlStorage.SaveURL(ctx, domain.SaveShortURLDto{
		OriginalURL: url,
		ShortURL:    shortURL,
		UserID:      userID,
	})
}

func (s *ShortenerService) ShortBatchURL(ctx context.Context, urls []domain.ShortBatchURL, userID string) ([]domain.ShortBatchURL, error) {
	correlations := make(map[string]string)
	saveDtos := make([]domain.SaveShortURLDto, 0, len(urls))

	for _, url := range urls {
		saveDtos = append(saveDtos, domain.SaveShortURLDto{
			OriginalURL: url.OriginalURL,
			ShortURL:    s.stringGenerator.GenerateRandom(),
			UserID:      userID,
		})
		correlations[url.OriginalURL] = url.CorrelationID
	}

	shortenedURLs, err := s.urlStorage.SaveSeveralURL(ctx, saveDtos)

	if err != nil {
		return nil, err
	}

	result := make([]domain.ShortBatchURL, 0)

	for _, url := range shortenedURLs {
		result = append(result, domain.ShortBatchURL{
			ShortURL:      url.ShortURL,
			OriginalURL:   url.OriginalURL,
			CorrelationID: correlations[url.OriginalURL],
		})
	}

	return result, nil
}

func (s *ShortenerService) GetByShortURL(ctx context.Context, url string) (*domain.ShortenedURL, error) {
	return s.urlStorage.GetByShortURL(ctx, url)
}

func (s *ShortenerService) GetUserURLs(ctx context.Context, userID string) ([]domain.ShortenedURL, error) {
	return s.urlStorage.GetURLsByUserID(ctx, userID)
}

func (s *ShortenerService) DeleteURLs(ctx context.Context, urls []string, userID string) error {
	go s.deleteURLQueue.Push(&domain.DeleteURLsTask{
		ShortURLs: urls,
		UserID:    userID,
	})

	return nil
}

func (s *ShortenerService) GetInternalStats(ctx context.Context) (*domain.InternalStats, error) {
	return s.urlStorage.GetInternalStats(ctx)
}

func (s *ShortenerService) Ping(ctx context.Context) error {
	return s.urlStorage.Ping(ctx)
}
