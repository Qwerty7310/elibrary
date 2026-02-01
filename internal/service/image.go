package service

import (
	"context"
	"elibrary/internal/storage"
	"io"

	"github.com/google/uuid"
)

type ImageService struct {
	storage storage.ImageStorage
}

func NewImageService(storage storage.ImageStorage) *ImageService {
	return &ImageService{storage: storage}
}

func (s *ImageService) Upload(ctx context.Context, entity storage.EntityType, entityID uuid.UUID, file io.Reader) (string, error) {
	return s.storage.Save(ctx, entity, entityID, file)
}

func (s *ImageService) Replace(ctx context.Context, entity storage.EntityType, entityID uuid.UUID, oldURL *string, file io.Reader) (string, error) {
	if oldURL != nil && *oldURL != "" {
		_ = s.storage.Delete(ctx, *oldURL)
	}

	return s.storage.Save(ctx, entity, entityID, file)
}
