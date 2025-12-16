package service

import (
	"context"
	"database/sql"
	"elibrary/internal/domain"
	"elibrary/internal/repository"
	"errors"
	"strings"

	"github.com/google/uuid"
)

var (
	ErrNotFound       = errors.New("book not found")
	ErrInvalidBarcode = errors.New("invalid barcode")
)

type BookService struct {
	repo repository.BookRepository
}

func NewBookService(repo repository.BookRepository) *BookService {
	return &BookService{repo: repo}
}

func (s *BookService) Create(ctx context.Context, book domain.Book) (domain.Book, error) {
	book.ID = uuid.New()

	if book.Extra == nil {
		book.Extra = map[string]any{}
	}

	if book.Barcode != "" {
		if !isValidEAN13(book.Barcode) {
			return domain.Book{}, errors.New(`invalid ean13 format`)
		}
	} else {
		book.Barcode = book.ID.String()
	}

	err := s.repo.Create(ctx, book)
	return book, err
}

func (s *BookService) GetByID(ctx context.Context, id uuid.UUID) (domain.Book, error) {
	book, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return domain.Book{}, ErrNotFound
	}
	return book, nil
}

func (s *BookService) Search(ctx context.Context, query string) ([]domain.Book, error) {
	if query == "" {
		return nil, nil
	}
	return s.repo.Search(ctx, query)
}

func (s *BookService) FindByScan(ctx context.Context, value string) (domain.Book, error) {
	value = strings.TrimSpace(value)

	if !isValidBarcode(value) {
		return domain.Book{}, ErrInvalidBarcode
	}

	book, err := s.repo.GetByBarcode(ctx, value)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return domain.Book{}, ErrNotFound
		}
		return domain.Book{}, err
	}

	return book, nil
}

func isValidEAN13(code string) bool {
	if len(code) != 13 {
		return false
	}

	sum := 0
	for i := 0; i < 12; i++ {
		c := code[i]
		if c < '0' || c > '9' {
			return false
		}

		d := int(c - '0')

		if i%2 == 0 {
			sum += d
		} else {
			sum += 3 * d
		}
	}

	last := code[12]
	if last < '0' || last > '9' {
		return false
	}

	checkDigit := (10 - (sum % 10)) % 10
	return int(last-'0') == checkDigit
}

func isValidUUID(s string) bool {
	_, err := uuid.Parse(s)
	return err == nil
}

func isValidBarcode(s string) bool {
	if isValidEAN13(s) {
		return true
	}
	if isValidUUID(s) {
		return true
	}
	return false
}
