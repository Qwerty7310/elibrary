package repository

import (
	"context"
	"elibrary/internal/domain"

	"github.com/google/uuid"
)

type PublisherRepository interface {
	Create(ctx context.Context, publisher domain.Publisher) error
	Update(ctx context.Context, publisher domain.Publisher) error
	GetByID(ctx context.Context, id uuid.UUID) (*domain.Publisher, error)
	Delete(ctx context.Context, id uuid.UUID) error
}
