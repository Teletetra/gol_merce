package httpapi

import (
	"context"

	"ecommerce_go/internal/domain"
)

type contextKey string

const authUserKey contextKey = "auth_user"

func WithUser(ctx context.Context, user domain.User) context.Context {
	return context.WithValue(ctx, authUserKey, user)
}

func UserFromContext(ctx context.Context) (domain.User, bool) {
	user, ok := ctx.Value(authUserKey).(domain.User)
	return user, ok
}
