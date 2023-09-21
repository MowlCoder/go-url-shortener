package context

import "context"

type contextKey string

const UserIDKey = contextKey("user_id")

func SetUserIDToContext(ctx context.Context, userID string) context.Context {
	return context.WithValue(ctx, UserIDKey, userID)
}

func GetUserIDFromContext(ctx context.Context) string {
	return ctx.Value(UserIDKey).(string)
}
