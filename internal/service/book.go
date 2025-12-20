package service

import (
	"context"
	"elibrary/internal/domain"
	"elibrary/internal/repository"
	"errors"
	"fmt"
	"strings"

	"github.com/google/uuid"
)

var (
	ErrNotFound       = errors.New("book not found")
	ErrInvalidBarcode = errors.New("invalid barcode")
	ErrBarcodeExists  = errors.New("barcode already exists")
)

type BookService struct {
	bookRepo   repository.BookRepository
	barcodeSvc *BarcodeService
}

func NewBookService(repo repository.BookRepository, barcodeSvc *BarcodeService) *BookService {
	return &BookService{
		bookRepo:   repo,
		barcodeSvc: barcodeSvc,
	}
}

func (s *BookService) CreateBook(ctx context.Context, book domain.Book) (*domain.Book, []byte, error) {
	ean13, err := s.barcodeSvc.GenerateEAN13(ctx)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to generate barcode: %w", err)
	}

	book.Barcode = ean13

	if book.ID == uuid.Nil {
		book.ID = uuid.New()
	}

	if book.Extra == nil {
		book.Extra = make(map[string]any)
	}

	if strings.TrimSpace(book.Title) == "" {
		return nil, nil, errors.New("title is required")
	}
	if strings.TrimSpace(book.Author) == "" {
		return nil, nil, errors.New("author is required")
	}

	if err := s.bookRepo.Create(ctx, book); err != nil {
		if strings.Contains(strings.ToLower(err.Error()), "duplicate") ||
			strings.Contains(strings.ToLower(err.Error()), "unique") {
			return nil, nil, ErrBarcodeExists
		}
		return nil, nil, fmt.Errorf("failed to save book: %w", err)
	}

	barcodeImage, err := s.barcodeSvc.GenerateBarcodeImage(ean13)
	if err != nil {
		fmt.Printf("Warning: failed to generate barcode image: %v\n", err)
	}

	return &book, barcodeImage, nil
}

func (s *BookService) GetByID(ctx context.Context, id uuid.UUID) (*domain.Book, error) {
	book, err := s.bookRepo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return nil, ErrNotFound
		}

		return nil, err
	}
	return &book, nil
}

func (s *BookService) GetByBarcode(ctx context.Context, barcode string) (*domain.Book, error) {
	if !s.barcodeSvc.ValidateEAN13(barcode) {
		return nil, ErrInvalidBarcode
	}

	book, err := s.bookRepo.GetByBarcode(ctx, barcode)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return &book, nil
}

func (s *BookService) GetByFactoryBarcode(ctx context.Context, factoryBarcode string) ([]domain.Book, error) {
	if factoryBarcode == "" {
		return nil, errors.New("factory barcode cannot be empty")
	}

	if !s.barcodeSvc.ValidateEAN13(factoryBarcode) {
		return nil, ErrInvalidBarcode
	}

	return s.bookRepo.GetByFactoryBarcode(ctx, factoryBarcode)
}

func (s *BookService) FindByScan(ctx context.Context, value string) ([]domain.Book, error) {
	value = strings.TrimSpace(value)
	if value == "" {
		return nil, ErrInvalidBarcode
	}

	// Поиск по нашему EAN-13
	if s.barcodeSvc.ValidateEAN13(value) {
		book, err := s.bookRepo.GetByBarcode(ctx, value)
		if err == nil {
			return []domain.Book{book}, nil
		}
	}

	// Поиск по UUID
	if isValidUUID(value) {
		id, _ := uuid.Parse(value)
		book, err := s.bookRepo.GetByID(ctx, id)
		if err == nil {
			return []domain.Book{book}, nil
		}
	}

	return s.bookRepo.GetByFactoryBarcode(ctx, value)
}

func (s *BookService) Search(ctx context.Context, query string) ([]domain.Book, error) {
	if query == "" {
		return nil, nil
	}
	return s.bookRepo.Search(ctx, query)
}

type UpdateBookRequest struct {
	Title          *string        `json:"title,omitempty"`
	Author         *string        `json:"author,omitempty"`
	Publisher      *string        `json:"publisher,omitempty"`
	Year           *int           `json:"year,omitempty"`
	Location       *string        `json:"location,omitempty"`
	FactoryBarcode *string        `json:"factory_barcode,omitempty"`
	Extra          map[string]any `json:"extra,omitempty"`
}

func (s *BookService) UpdateBook(ctx context.Context, id uuid.UUID, updates UpdateBookRequest) error {
	book, err := s.bookRepo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return ErrNotFound
		}
		return err
	}

	if updates.Title != nil {
		if title := strings.TrimSpace(*updates.Title); title != "" {
			book.Title = title
		}
	}
	if updates.Author != nil {
		if author := strings.TrimSpace(*updates.Author); author != "" {
			book.Author = author
		}
	}
	if updates.Publisher != nil {
		book.Publisher = *updates.Publisher
	}
	if updates.Year != nil {
		book.Year = *updates.Year
	}
	if updates.Location != nil {
		book.Location = *updates.Location
	}
	if updates.FactoryBarcode != nil {
		book.FactoryBarcode = *updates.FactoryBarcode
	}
	if updates.Extra != nil {
		book.Extra = updates.Extra
	}

	return s.bookRepo.Update(ctx, book)
}

func (s *BookService) GenerateBarcodeImage(barcode string) ([]byte, error) {
	return s.barcodeSvc.GenerateBarcodeImage(barcode)
}

func isValidUUID(s string) bool {
	_, err := uuid.Parse(s)
	return err == nil
}
