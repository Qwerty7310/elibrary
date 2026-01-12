package service

import (
	"context"
	"elibrary/internal/domain"
	"elibrary/internal/readmodel"
	"elibrary/internal/repository"
	"errors"
	"log"
	"strings"
	"time"

	"github.com/google/uuid"
)

type WorkService struct {
	workRepo repository.WorkRepository
}

func NewWorkService(workRepo repository.WorkRepository) *WorkService {
	return &WorkService{workRepo: workRepo}
}

func (s *WorkService) Create(ctx context.Context, work domain.Work, authors []uuid.UUID) (*domain.Work, error) {
	work.ID = uuid.New()

	if strings.TrimSpace(work.Title) == "" {
		return nil, errors.New("title is required")
	}

	err := s.workRepo.WithTx(ctx, func(tx repository.WorkTx) error {
		if err := tx.CreateWork(ctx, work); err != nil {
			return err
		}
		return tx.ReplaceWorkAuthors(ctx, work.ID, authors)
	})
	if err != nil {
		return nil, err
	}

	return &work, nil
}

type UpdateWorkRequest struct {
	Title       *string `json:"title,omitempty"`
	Description *string `json:"description,omitempty"`
	Year        *int    `json:"year,omitempty"`

	Authors *[]uuid.UUID `json:"authors,omitempty"`
}

func (s *WorkService) Update(ctx context.Context, id uuid.UUID, updates UpdateWorkRequest) error {
	return s.workRepo.WithTx(ctx, func(tx repository.WorkTx) error {

		work, err := tx.GetDomainByID(ctx, id)
		if err != nil {
			if errors.Is(err, repository.ErrNotFound) {
				return domain.ErrNotFound
			}
			return err
		}

		if updates.Title != nil {
			title := strings.TrimSpace(*updates.Title)
			if title == "" {
				return errors.New("title is required")
			}
			work.Title = title
		}
		if updates.Description != nil {
			work.Description = updates.Description
		}
		if updates.Year != nil {
			if *updates.Year < 0 || *updates.Year > time.Now().Year()+1 {
				return errors.New("invalid year")
			}
			work.Year = updates.Year
		}

		if err := tx.UpdateWork(ctx, *work); err != nil {
			return err
		}

		if updates.Authors != nil {
			if err := tx.ReplaceWorkAuthors(ctx, work.ID, *updates.Authors); err != nil {
				return err
			}
		}

		return nil

	})
}

func (s *WorkService) GetByID(ctx context.Context, id uuid.UUID) (*readmodel.WorkDetailed, error) {
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

func (s *WorkService) GetAll(ctx context.Context) ([]*readmodel.WorkShort, error) {
	return s.workRepo.GetAll(ctx)
}
