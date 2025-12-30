package repository

import (
	"context"
	"elibrary/internal/domain"

	"github.com/google/uuid"
)

type UserRepository interface {
	GetByLogin(ctx context.Context, login string) (*domain.User, error)
	GetByID(ctx context.Context, id uuid.UUID) (*domain.User, error)
}
