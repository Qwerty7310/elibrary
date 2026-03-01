package auth

import (
	"context"
	"testing"

	"elibrary/internal/domain"
)

func TestUserFromContext(t *testing.T) {
	t.Parallel()

	user := &domain.User{Login: "reader"}
	ctx := context.WithValue(context.Background(), userKey, user)

	got, ok := UserFromContext(ctx)
	if !ok {
		t.Fatal("UserFromContext() ok = false, want true")
	}
	if got != user {
		t.Fatalf("UserFromContext() = %#v, want %#v", got, user)
	}

	if _, ok := UserFromContext(context.Background()); ok {
		t.Fatal("UserFromContext() ok = true for empty context, want false")
	}
}

func TestHasRole(t *testing.T) {
	t.Parallel()

	user := &domain.User{
		Roles: []domain.Role{
			{Code: RoleAdmin},
			{Code: RoleUser},
		},
	}
	ctx := context.WithValue(context.Background(), userKey, user)

	if !HasRole(ctx, RoleAdmin) {
		t.Fatal("HasRole() = false, want true for existing role")
	}
	if HasRole(ctx, "missing") {
		t.Fatal("HasRole() = true, want false for missing role")
	}
	if HasRole(context.Background(), RoleAdmin) {
		t.Fatal("HasRole() = true, want false for missing user")
	}
}
