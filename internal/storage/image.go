package storage

import (
	"context"
	"io"

	"github.com/google/uuid"
)

type EntityType string

const (
	Author    EntityType = "author"
	Book      EntityType = "book"
	Publisher EntityType = "publisher"
)

type ImageStorage interface {
	Save(ctx context.Context, entity EntityType, entityID uuid.UUID, file io.Reader) (string, error)
	Delete(ctx context.Context, url string) error
}
