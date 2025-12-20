package repository

import "context"

type SequenceRepository interface {
	GetNext(ctx context.Context, prefix int) (int64, error)
	SetPrefix(ctx context.Context, prefix int, description string) error
}
