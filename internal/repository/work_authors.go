package repository

import (
	"context"

	"github.com/google/uuid"
)

type WorkAuthorsRepository interface {
	AddAuthorToWork(ctx context.Context, workID uuid.UUID, authorID uuid.UUID) error
	RemoveAuthorFromWork(ctx context.Context, workID uuid.UUID, authorID uuid.UUID) error
	ReplaceWorkAuthors(ctx context.Context, workID uuid.UUID, authorsIDs []uuid.UUID) error
}
