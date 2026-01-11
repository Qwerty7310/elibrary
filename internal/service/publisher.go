package service

import (
	"context"
	"elibrary/internal/domain"
	"elibrary/internal/readmodel"
	"elibrary/internal/repository"
	"errors"
	"log"
	"strings"

	"github.com/google/uuid"
)

type PublisherService struct {
	publisherRepo repository.PublisherRepository
}

func NewPublisherService(publisherRepo repository.PublisherRepository) *PublisherService {
	return &PublisherService{
		publisherRepo: publisherRepo,
	}
}

func (s *PublisherService) Create(ctx context.Context, publisher domain.Publisher) (*domain.Publisher, error) {
	publisher.ID = uuid.New()

	if strings.TrimSpace(publisher.Name) == "" {
		return nil, errors.New("name is required")
	}

	if err := s.publisherRepo.Create(ctx, publisher); err != nil {
		log.Printf("Error creating publisher: %v", err)
		return nil, err
	}

	return &publisher, nil
}

type UpdatePublisherRequest struct {
	Name    *string `json:"name,omitempty"`
	LogoURL *string `json:"logo_url,omitempty"`
	WebURL  *string `json:"web_url,omitempty"`
}

func (s *PublisherService) Update(ctx context.Context, id uuid.UUID, updates UpdatePublisherRequest) error {
	publisher, err := s.publisherRepo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return domain.ErrNotFound
		}
		return err
	}

	if updates.Name != nil {
		if strings.TrimSpace(*updates.Name) == "" {
			return errors.New("name is required")
		}
		publisher.Name = *updates.Name
	}

	if updates.LogoURL != nil {
		publisher.LogoURL = updates.LogoURL
	}

	if updates.WebURL != nil {
		publisher.WebURL = updates.WebURL
	}

	return s.publisherRepo.Update(ctx, *publisher)
}

func (s *PublisherService) GetByID(ctx context.Context, id uuid.UUID) (*domain.Publisher, error) {
	publisher, err := s.publisherRepo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return nil, domain.ErrNotFound
		}
		return nil, err
	}

	return publisher, nil
}

func (s *PublisherService) Delete(ctx context.Context, id uuid.UUID) error {
	err := s.publisherRepo.Delete(ctx, id)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return domain.ErrNotFound
		}
		log.Printf("Error deleting publisher: %v", err)
		return err
	}

	return nil
}

func (s *PublisherService) GetAll(ctx context.Context) ([]readmodel.Publisher, error) {
	return s.publisherRepo.GetAll(ctx)
}
