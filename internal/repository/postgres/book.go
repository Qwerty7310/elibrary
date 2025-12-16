package postgres

import (
	"context"
	"elibrary/internal/domain"
	"encoding/json"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

type BookRepository struct {
	db *pgxpool.Pool
}

func NewBookRepository(db *pgxpool.Pool) *BookRepository {
	return &BookRepository{db: db}
}

func (r *BookRepository) Create(ctx context.Context, book domain.Book) error {
	extraJSON, err := json.Marshal(book.Extra)
	if err != nil {
		return err
	}

	_, err = r.db.Exec(ctx, `
		INSERT INTO books (id, title, author, publisher, year, location, extra)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
	`,
		book.ID,
		book.Title,
		book.Author,
		book.Publisher,
		book.Year,
		book.Location,
		extraJSON,
	)

	return err
}

func (r *BookRepository) GetByID(ctx context.Context, id uuid.UUID) (domain.Book, error) {
	var book domain.Book
	var extraJSON []byte

	err := r.db.QueryRow(ctx, `
		SELECT id, title, author, publisher, year, location, extra
		FROM books
		WHERE id = $1
	`, id).Scan(
		&book.ID,
		&book.Title,
		&book.Author,
		&book.Publisher,
		&book.Year,
		&book.Location,
		&extraJSON,
	)

	if err != nil {
		return domain.Book{}, err
	}

	if err := json.Unmarshal(extraJSON, &book.Extra); err != nil {
		return domain.Book{}, err
	}

	return book, nil
}
