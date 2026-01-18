package repository

import (
	"context"

	"github.com/google/uuid"
)

type BookWorksRepository interface {
	AddWorkToBook(ctx context.Context, bookID uuid.UUID, workID uuid.UUID, position *int) error
	RemoveWorkFromBook(ctx context.Context, bookID uuid.UUID, workID uuid.UUID) error
	SetWorkPosition(ctx context.Context, bookID uuid.UUID, workID uuid.UUID, position int) error
	ReplaceBookWorks(ctx context.Context, bookID uuid.UUID, works []BookWorkInput) error
}

type BookWorkInput struct {
	WorkID   uuid.UUID `json:"work_id"`
	Position *int      `json:"position,omitempty"`
}
