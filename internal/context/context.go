package context

import (
	"context"
	"errors"
)

type contextKey string

const UserIDKey = contextKey("user_id")

var (
	ErrUserIDKeyNotFound = errors.New("user id key not found in context")
	ErrUserIDInvalidType = errors.New("user id found, but with invalid type")
)

func SetUserIDToContext(ctx context.Context, userID string) context.Context {
	return context.WithValue(ctx, UserIDKey, userID)
}

func GetUserIDFromContext(ctx context.Context) (string, error) {
	val := ctx.Value(UserIDKey)

	if val == nil {
		return "", ErrUserIDKeyNotFound
	}

	id, ok := val.(string)

	if !ok {
		return "", ErrUserIDInvalidType
	}

	return id, nil
}
