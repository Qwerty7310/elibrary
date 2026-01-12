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
	GetByID(ctx context.Context, id uuid.UUID) (*readmodel.WorkDetailed, error)
	Delete(ctx context.Context, id uuid.UUID) error

	GetAll(ctx context.Context) ([]*readmodel.WorkShort, error)

	WithTx(ctx context.Context, fn func(tx WorkTx) error) error
}

type WorkTx interface {
	CreateWork(ctx context.Context, work domain.Work) error
	UpdateWork(ctx context.Context, work domain.Work) error
	GetDomainByID(ctx context.Context, id uuid.UUID) (*domain.Work, error)
	ReplaceWorkAuthors(ctx context.Context, workID uuid.UUID, authors []uuid.UUID) error
}
