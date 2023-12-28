package interceptors

import (
	"context"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"

	contextUtil "github.com/MowlCoder/go-url-shortener/internal/context"
	"github.com/MowlCoder/go-url-shortener/internal/jwt"
)

type userService interface {
	GenerateUniqueID() string
}

func CreateAuthInterceptor(
	userService userService,
) func(ctx context.Context,
	req any,
	info *grpc.UnaryServerInfo,
	handler grpc.UnaryHandler) (any, error) {
	return func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (any, error) {
		var tokenString string
		var err error
		md, ok := metadata.FromIncomingContext(ctx)

		if !ok {
			return nil, status.Error(codes.Internal, "can not get token")
		}

		if len(md.Get("token")) == 0 {
			tokenString, err = jwt.GenerateToken(userService.GenerateUniqueID())
			if err != nil {
				return nil, status.Error(codes.Internal, "can not generate token")
			}
		} else {
			tokenString = md.Get("token")[0]
		}

		jwtClaim, err := jwt.ParseToken(tokenString)
		if err != nil {
			return nil, status.Error(codes.Internal, "can not parse token")
		}

		ctxWithUserID := contextUtil.SetUserIDToContext(ctx, jwtClaim.UserID)

		return handler(ctxWithUserID, req)
	}
}
