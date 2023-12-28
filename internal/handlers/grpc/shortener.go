package grpc

import (
	"context"
	"errors"
	"fmt"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/MowlCoder/go-url-shortener/internal/config"
	contextUtil "github.com/MowlCoder/go-url-shortener/internal/context"
	"github.com/MowlCoder/go-url-shortener/internal/domain"
	"github.com/MowlCoder/go-url-shortener/proto"
)

type shortenerService interface {
	ShortURL(ctx context.Context, url string, userID string) (*domain.ShortenedURL, error)
	ShortBatchURL(ctx context.Context, urls []domain.ShortBatchURL, userID string) ([]domain.ShortBatchURL, error)
	GetUserURLs(ctx context.Context, userID string) ([]domain.ShortenedURL, error)
	DeleteURLs(ctx context.Context, urls []string, userID string) error
	GetInternalStats(ctx context.Context) (*domain.InternalStats, error)
	Ping(ctx context.Context) error
}

type ShortenerHandler struct {
	proto.UnimplementedShortenerServer

	appConfig *config.AppConfig
	service   shortenerService
}

func NewShortenerHandler(
	appConfig *config.AppConfig,
	service shortenerService,
) *ShortenerHandler {
	return &ShortenerHandler{
		appConfig: appConfig,
		service:   service,
	}
}

func (h *ShortenerHandler) ShortURL(ctx context.Context, in *proto.ShortURLRequest) (*proto.ShortURLResponse, error) {
	userID, err := contextUtil.GetUserIDFromContext(ctx)
	if err != nil {
		return nil, status.Error(codes.Unauthenticated, "missing user id")
	}

	if len(in.Url) == 0 {
		return nil, status.Error(codes.InvalidArgument, "invalid url")
	}

	shortenedURL, err := h.service.ShortURL(ctx, in.Url, userID)
	if err != nil && !errors.Is(err, domain.ErrURLConflict) {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &proto.ShortURLResponse{
		Result: fmt.Sprintf("%s/%s", h.appConfig.BaseShortURLAddr, shortenedURL.ShortURL),
	}, nil
}

func (h *ShortenerHandler) ShortBatchURL(ctx context.Context, in *proto.ShortBatchURLRequest) (*proto.ShortBatchURLResponse, error) {
	userID, err := contextUtil.GetUserIDFromContext(ctx)
	if err != nil {
		return nil, status.Error(codes.Unauthenticated, "missing user id")
	}

	urls := make([]domain.ShortBatchURL, 0)
	for _, dto := range in.Dtos {
		urls = append(urls, domain.ShortBatchURL{
			CorrelationID: dto.CorrelationId,
			OriginalURL:   dto.OriginalUrl,
		})
	}

	shortenedURLs, err := h.service.ShortBatchURL(ctx, urls, userID)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	responseURLs := make([]*proto.ResponseBatchURLDto, 0, len(shortenedURLs))
	for _, url := range shortenedURLs {
		responseURLs = append(responseURLs, &proto.ResponseBatchURLDto{
			ShortUrl:      url.ShortURL,
			CorrelationId: url.CorrelationID,
		})
	}

	return &proto.ShortBatchURLResponse{
		Dtos: responseURLs,
	}, nil
}

func (h *ShortenerHandler) GetMyURLs(ctx context.Context, in *proto.GetMyURLsRequest) (*proto.GetMyURLsResponse, error) {
	userID, err := contextUtil.GetUserIDFromContext(ctx)
	if err != nil {
		return nil, status.Error(codes.Unauthenticated, "missing user id")
	}

	urls, err := h.service.GetUserURLs(ctx, userID)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	userShortenedURLs := make([]*proto.UserShortenedURL, 0, len(urls))

	for _, url := range urls {
		userShortenedURLs = append(userShortenedURLs, &proto.UserShortenedURL{
			OriginalUrl: url.OriginalURL,
			ShortUrl:    url.ShortURL,
		})
	}

	return &proto.GetMyURLsResponse{Result: userShortenedURLs}, nil
}

func (h *ShortenerHandler) DeleteURLs(ctx context.Context, in *proto.DeleteURLsRequest) (*proto.DeleteURLsResponse, error) {
	userID, err := contextUtil.GetUserIDFromContext(ctx)
	if err != nil {
		return nil, status.Error(codes.Unauthenticated, "missing user id")
	}

	if len(in.Urls) == 0 {
		return nil, status.Error(codes.InvalidArgument, "you have to send at least 1 url")
	}

	if err := h.service.DeleteURLs(ctx, in.Urls, userID); err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &proto.DeleteURLsResponse{}, nil
}

func (h *ShortenerHandler) GetStats(ctx context.Context, in *proto.GetStatsRequest) (*proto.GetStatsResponse, error) {
	stats, err := h.service.GetInternalStats(ctx)

	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &proto.GetStatsResponse{
		Users: int64(stats.Users),
		Urls:  int64(stats.URLs),
	}, nil
}

func (h *ShortenerHandler) Ping(ctx context.Context, in *proto.PingRequest) (*proto.PingResponse, error) {
	if err := h.service.Ping(ctx); err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &proto.PingResponse{
		Ok: true,
	}, nil
}
