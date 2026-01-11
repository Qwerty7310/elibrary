package repository

import (
	"context"
	"elibrary/internal/domain"
	"elibrary/internal/readmodel"

	"github.com/google/uuid"
)

type AuthorRepository interface {
	Create(ctx context.Context, author domain.Author) error
	Update(ctx context.Context, author domain.Author) error
	GetByID(ctx context.Context, id uuid.UUID) (*domain.Author, error)
	Delete(ctx context.Context, id uuid.UUID) error

	GetAll(ctx context.Context) ([]readmodel.Author, error)
}
