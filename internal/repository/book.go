package repository

import (
	"context"
	"elibrary/internal/domain"
	"elibrary/internal/readmodel"
	"errors"

	"github.com/google/uuid"
)

var ErrNotFound = errors.New("not found")

type BookRepository interface {
	Create(ctx context.Context, book domain.Book) error
	Update(ctx context.Context, book domain.Book) error

	GetPublicByID(ctx context.Context, id uuid.UUID) (*readmodel.BookPublic, error)
	GetInternalByID(ctx context.Context, id uuid.UUID) (*readmodel.BookInternal, error)

	GetPublic(ctx context.Context, filter BookFilter) ([]*readmodel.BookPublic, error)
	GetInternal(ctx context.Context, filter BookFilter) ([]*readmodel.BookInternal, error)

	WithTx(ctx context.Context, fn func(tx BookTx) error) error
}

type BookTx interface {
	GetDomainByID(ctx context.Context, id uuid.UUID) (*domain.Book, error)
	CreateBook(ctx context.Context, book domain.Book) error
	UpdateBook(ctx context.Context, book domain.Book) error
	ReplaceBookWorks(ctx context.Context, bookID uuid.UUID, works []BookWorkInput) error
}
