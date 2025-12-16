package repository

import (
	"context"
	"elibrary/internal/domain"

	"github.com/google/uuid"
)

type BookRepository interface {
	Create(ctx context.Context, book domain.Book) error
	GetByID(ctx context.Context, id uuid.UUID) (domain.Book, error)
}
