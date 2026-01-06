package repository

import (
	"context"
	"elibrary/internal/domain"
)

type SequenceRepository interface {
	GetNext(ctx context.Context, t domain.BarcodeType) (int64, int, error)
	SetType(ctx context.Context, t domain.BarcodeType, prefix int, description string) error
}
