package middleware

import (
	"context"
	"elibrary/internal/service"
	"net/http"
	"strings"

	"github.com/google/uuid"
)

type ctxKey string

const userIDKey ctxKey = "user_id"

func Auth(jwt *service.JWTManager) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			h := r.Header.Get("Authorization")
			if !strings.HasPrefix(h, "Bearer ") {
				http.Error(w, "Unauthorized", http.StatusUnauthorized)
				return
			}

			id, err := jwt.Parse(strings.TrimPrefix(h, "Bearer "))
			if err != nil {
				http.Error(w, "Unauthorized", http.StatusUnauthorized)
			}

			ctx := context.WithValue(r.Context(), userIDKey, id)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func UserID(ctx context.Context) (uuid.UUID, bool) {
	id, ok := ctx.Value(userIDKey).(uuid.UUID)
	return id, ok
}
