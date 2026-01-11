package repository

import (
	"context"
	"elibrary/internal/domain"
	"elibrary/internal/readmodel"

	"github.com/google/uuid"
)

type PublisherRepository interface {
	Create(ctx context.Context, publisher domain.Publisher) error
	Update(ctx context.Context, publisher domain.Publisher) error
	GetByID(ctx context.Context, id uuid.UUID) (*domain.Publisher, error)
	Delete(ctx context.Context, id uuid.UUID) error

	GetAll(ctx context.Context) ([]readmodel.Publisher, error)
}
