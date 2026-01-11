package service

import (
	"context"
	"elibrary/internal/domain"
	"elibrary/internal/readmodel"
	"elibrary/internal/repository"
	"errors"
	"strings"
	"time"

	"github.com/google/uuid"
)

type AuthorService struct {
	authorRepo repository.AuthorRepository
}

func NewAuthorService(authorRepo repository.AuthorRepository) *AuthorService {
	return &AuthorService{authorRepo: authorRepo}
}

func (s *AuthorService) Create(ctx context.Context, author domain.Author) (*domain.Author, error) {
	author.ID = uuid.New()

	if strings.TrimSpace(author.LastName) == "" {
		return nil, errors.New("author last name is required")
	}

	if err := s.authorRepo.Create(ctx, author); err != nil {
		return nil, err
	}

	return &author, nil
}

type UpdateAuthorRequest struct {
	LastName   *string    `json:"last_name,omitempty"`
	FirstName  *string    `json:"first_name,omitempty"`
	MiddleName *string    `json:"middle_name,omitempty"`
	BirthDate  *time.Time `json:"birth_date,omitempty"`
	DeathDate  *time.Time `json:"death_date,omitempty"`
	Bio        *string    `json:"bio,omitempty"`
	PhotoURL   *string    `json:"photo_url,omitempty"`
}

func (s *AuthorService) Update(ctx context.Context, id uuid.UUID, updates UpdateAuthorRequest) error {
	author, err := s.authorRepo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return domain.ErrNotFound
		}
		return err
	}

	if updates.LastName != nil {
		if strings.TrimSpace(*updates.LastName) == "" {
			return errors.New("last name is required")
		}
		author.LastName = *updates.LastName
	}
	if updates.FirstName != nil {
		author.FirstName = updates.FirstName
	}
	if updates.MiddleName != nil {
		author.MiddleName = updates.MiddleName
	}
	if updates.BirthDate != nil {
		author.BirthDate = updates.BirthDate
	}
	if updates.DeathDate != nil {
		author.DeathDate = updates.DeathDate
	}
	if updates.Bio != nil {
		author.Bio = updates.Bio
	}
	if updates.PhotoURL != nil {
		author.PhotoURL = updates.PhotoURL
	}

	return s.authorRepo.Update(ctx, *author)
}

func (s *AuthorService) GetByID(ctx context.Context, id uuid.UUID) (*domain.Author, error) {
	author, err := s.authorRepo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return nil, domain.ErrNotFound
		}
		return nil, err
	}

	return author, nil
}

func (s *AuthorService) Delete(ctx context.Context, id uuid.UUID) error {
	err := s.authorRepo.Delete(ctx, id)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return domain.ErrNotFound
		}
		return err
	}

	return nil
}

func (s *AuthorService) GetAll(ctx context.Context) ([]readmodel.Author, error) {
	return s.authorRepo.GetAll(ctx)
}
