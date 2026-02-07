package repository

import (
	"context"
	"elibrary/internal/domain"

	"github.com/google/uuid"
)

type LocationRepository interface {
	Create(ctx context.Context, loc domain.Location) error
	Update(ctx context.Context, loc domain.Location) error
	GetByID(ctx context.Context, id uuid.UUID) (*domain.Location, error)
	GetByType(ctx context.Context, locType domain.LocationType) ([]*domain.Location, error)
	GetByTypeParentID(ctx context.Context, locType domain.LocationType, parentID uuid.UUID) ([]*domain.Location, error)
	GetByBarcode(ctx context.Context, barcode string) (*domain.Location, error)
	HasChildren(ctx context.Context, id uuid.UUID) (bool, error)
	Delete(ctx context.Context, id uuid.UUID) error
}
