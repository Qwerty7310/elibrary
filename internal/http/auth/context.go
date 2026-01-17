package auth

import (
	"context"
	"elibrary/internal/domain"
)

type ctxKey string

const userKey ctxKey = "user"

func UserFromContext(ctx context.Context) (*domain.User, bool) {
	user, ok := ctx.Value(userKey).(*domain.User)
	return user, ok
}

func HasRole(ctx context.Context, role string) bool {
	user, ok := UserFromContext(ctx)
	if !ok {
		return false
	}

	for _, r := range user.Roles {
		if r.Code == role {
			return true
		}
	}
	return false
}
