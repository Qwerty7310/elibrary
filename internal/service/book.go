package service

import (
	"context"
	"elibrary/internal/domain"
	"elibrary/internal/repository"
	"errors"

	"github.com/google/uuid"
)

var ErrNotFound = errors.New("book not found")

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
