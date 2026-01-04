package repository

import (
	"context"
	"elibrary/internal/domain"

	"github.com/google/uuid"
)

type WorkRepository interface {
	Create(ctx context.Context, work domain.Work) error
	Update(ctx context.Context, work domain.Work) error
	GetByID(ctx context.Context, id uuid.UUID) (*domain.Work, error)
	Delete(ctx context.Context, id uuid.UUID) error
}
