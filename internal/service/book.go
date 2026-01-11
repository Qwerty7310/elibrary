package service

import (
	"context"
	"elibrary/internal/domain"
	"elibrary/internal/readmodel"
	"elibrary/internal/repository"
	"errors"
	"fmt"
	"strings"

	"github.com/google/uuid"
)

type BookService struct {
	bookRepo        repository.BookRepository
	bookWorksRepo   repository.BookWorksRepository
	workRepo        repository.WorkRepository
	workAuthorsRepo repository.WorkAuthorsRepository
	barcodeSvc      *BarcodeService
}

func NewBookService(
	bookRepo repository.BookRepository,
	bookWorksRepo repository.BookWorksRepository,
	workRepo repository.WorkRepository,
	workAuthorsRepo repository.WorkAuthorsRepository,
	barcodeSvc *BarcodeService,
) *BookService {
	return &BookService{
		bookRepo:        bookRepo,
		bookWorksRepo:   bookWorksRepo,
		workRepo:        workRepo,
		workAuthorsRepo: workAuthorsRepo,
		barcodeSvc:      barcodeSvc,
	}
}

func (s *BookService) Create(ctx context.Context, book domain.Book, works []repository.BookWorkInput) (*domain.Book, error) {
	if strings.TrimSpace(book.Title) == "" {
		return nil, errors.New("title is required")
	}

	ean13, err := s.barcodeSvc.GenerateEAN13(ctx, domain.BarcodeTypeBook)
	if err != nil {
		return nil, fmt.Errorf("failed to generate barcode: %w", err)
	}

	book.ID = uuid.New()
	book.Barcode = ean13

	if book.Extra == nil {
		book.Extra = make(map[string]any)
	}

	for _, w := range works {
		if _, err := s.workRepo.GetByID(ctx, w.WorkID); err != nil {
			return nil, err
		}
	}

	err = s.bookRepo.WithTx(ctx, func(tx repository.BookTx) error {
		if err := tx.CreateBook(ctx, book); err != nil {
			return err
		}
		return tx.ReplaceBookWorks(ctx, book.ID, works)
	})

	return &book, nil
}

type UpdateBookRequest struct {
	FactoryBarcode *string        `json:"factory_barcode,omitempty"`
	Title          *string        `json:"title,omitempty"`
	PublisherID    *uuid.UUID     `json:"publisher_id,omitempty"`
	Year           *int           `json:"year,omitempty"`
	Description    *string        `json:"description,omitempty"`
	LocationID     *uuid.UUID     `json:"location_id,omitempty"`
	Extra          map[string]any `json:"extra,omitempty"`

	Works *[]repository.BookWorkInput `json:"works,omitempty"`
}

func (s *BookService) Update(ctx context.Context, id uuid.UUID, updates UpdateBookRequest) error {
	return s.bookRepo.WithTx(ctx, func(tx repository.BookTx) error {

		book, err := tx.GetDomainByID(ctx, id)
		if err != nil {
			if errors.Is(err, repository.ErrNotFound) {
				return domain.ErrNotFound
			}
			return err
		}

		if updates.FactoryBarcode != nil {
			book.FactoryBarcode = updates.FactoryBarcode
		}
		if updates.Title != nil {
			title := strings.TrimSpace(*updates.Title)
			if title == "" {
				return errors.New("title cannot be empty")
			}
			book.Title = title
		}
		if updates.PublisherID != nil {
			book.PublisherID = updates.PublisherID
		}
		if updates.Year != nil {
			book.Year = updates.Year
		}
		if updates.Description != nil {
			book.Description = updates.Description
		}
		if updates.LocationID != nil {
			book.LocationID = updates.LocationID
		}
		if updates.Extra != nil {
			book.Extra = updates.Extra
		}

		if err := tx.UpdateBook(ctx, *book); err != nil {
			return err
		}

		if updates.Works != nil {
			if err := tx.ReplaceBookWorks(ctx, book.ID, *updates.Works); err != nil {
				return err
			}
		}

		return nil
	})
}

func (s *BookService) GetPublicByID(ctx context.Context, id uuid.UUID) (*readmodel.BookPublic, error) {
	book, err := s.bookRepo.GetPublicByID(ctx, id)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return nil, domain.ErrNotFound
		}

		return nil, err
	}
	return book, nil
}

func (s *BookService) GetInternalByID(ctx context.Context, id uuid.UUID) (*readmodel.BookInternal, error) {
	if !auth.HasRole(ctx, auth.RoleAdmin) {
		return nil, domain.ErrForbidden
	}

	book, err := s.bookRepo.GetInternalByID(ctx, id)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return nil, domain.ErrNotFound
		}

		return nil, err
	}
	return book, nil
}

func (s *BookService) GetPublic(ctx context.Context, filter repository.BookFilter) ([]*readmodel.BookPublic, error) {
	books, err := s.bookRepo.GetPublic(ctx, filter)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return nil, domain.ErrNotFound
		}
		return nil, err
	}

	return books, nil
}

func (s *BookService) GetInternal(ctx context.Context, filter repository.BookFilter) ([]*readmodel.BookInternal, error) {
	if !auth.HasRole(ctx, auth.RoleAdmin) {
		return nil, domain.ErrForbidden
	}

	books, err := s.bookRepo.GetInternal(ctx, filter)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return nil, domain.ErrNotFound
		}
		return nil, err
	}

	return books, nil
}
