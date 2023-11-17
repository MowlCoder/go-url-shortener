package context

import (
	"context"
	"errors"
)

type contextKey string

// UserIDKey represent key in context to store user id. Need for avoiding magic string.
const UserIDKey = contextKey("user_id")

// Possible errors when working with package.
var (
	ErrUserIDKeyNotFound = errors.New("user id key not found in context")
	ErrUserIDInvalidType = errors.New("user id found, but with invalid type")
)

// SetUserIDToContext save user id in given context.
func SetUserIDToContext(ctx context.Context, userID string) context.Context {
	return context.WithValue(ctx, UserIDKey, userID)
}

// GetUserIDFromContext try to get user id from given context. If context not found or wrong value found return error.
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
