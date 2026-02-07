package repository

import (
	"context"
	"elibrary/internal/domain"

	"github.com/google/uuid"
)

type UserRepository interface {
	Create(ctx context.Context, user domain.User) error
	Update(ctx context.Context, user domain.User) error
	Delete(ctx context.Context, it uuid.UUID) error

	GetByLogin(ctx context.Context, login string) (*domain.User, error)
	GetByID(ctx context.Context, id uuid.UUID) (*domain.User, error)
	GetByIDWithRoles(ctx context.Context, id uuid.UUID) (*domain.User, error)
	GetAllWithRoles(ctx context.Context) ([]*domain.User, error)
}
