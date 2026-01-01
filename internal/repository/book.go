package repository

import (
	"context"
	"elibrary/internal/domain"
	"errors"

	"github.com/google/uuid"
)

var (
	ErrNotFound = errors.New("not found")
)

type BookRepository interface {
	Create(ctx context.Context, book domain.Book) error
	Update(ctx context.Context, book domain.Book) error
	GetByID(ctx context.Context, id uuid.UUID) (*domain.Book, error)
	GetByBarcode(ctx context.Context, barcode string) (*domain.Book, error)
	GetByFactoryBarcode(ctx context.Context, factoryBarcode string) ([]*domain.Book, error)
	Search(ctx context.Context, query string) ([]*domain.Book, error)
}
