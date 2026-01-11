package repository

import (
	"context"
	"elibrary/internal/domain"
	"elibrary/internal/readmodel"

	"github.com/google/uuid"
)

type WorkRepository interface {
	Create(ctx context.Context, work domain.Work) error
	Update(ctx context.Context, work domain.Work) error
	GetByID(ctx context.Context, id uuid.UUID) (*domain.Work, error)
	Delete(ctx context.Context, id uuid.UUID) error

	GetAll(ctx context.Context) ([]*readmodel.Work, error)
}
