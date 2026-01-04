package service

import (
	"context"
	"elibrary/internal/domain"
	"elibrary/internal/repository"
	"errors"
	"log"
	"strings"

	"github.com/google/uuid"
)

type WorkService struct {
	workRepo repository.WorkRepository
}

func NewWorkService(workRepo repository.WorkRepository) *WorkService {
	return &WorkService{workRepo: workRepo}
}

func (s *WorkService) Create(ctx context.Context, work domain.Work) (*domain.Work, error) {
	work.ID = uuid.New()

	if strings.TrimSpace(work.Title) == "" {
		return nil, errors.New("title is required")
	}

	err := s.workRepo.Create(ctx, work)
	if err != nil {
		log.Printf("Error creating work: %s", err)
		return nil, err
	}

	return &work, nil
}

type UpdateWorkRequest struct {
	Title       *string `json:"title"`
	Description *string `json:"description"`
	Year        *int    `json:"year"`
}

func (s *WorkService) Update(ctx context.Context, id uuid.UUID, updates UpdateWorkRequest) error {
	work, err := s.workRepo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return domain.ErrNotFound
		}
		return err
	}

	if updates.Title != nil {
		if strings.TrimSpace(*updates.Title) == "" {
			return errors.New("title is required")
		}
		work.Title = *updates.Title
	}

	if updates.Description != nil {
		work.Description = updates.Description
	}

	if updates.Year != nil {
		work.Year = updates.Year
	}

	return s.workRepo.Update(ctx, *work)
}

func (s *WorkService) GetByID(ctx context.Context, id uuid.UUID) (*domain.Work, error) {
	work, err := s.workRepo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return nil, domain.ErrNotFound
		}
		return nil, err
	}

	return work, nil
}

func (s *WorkService) Delete(ctx context.Context, id uuid.UUID) error {
	err := s.workRepo.Delete(ctx, id)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return domain.ErrNotFound
		}
		log.Printf("Error deleting work: %s", err)
		return err
	}

	return nil
}
