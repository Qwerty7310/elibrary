package postgres

import (
	"context"
	"database/sql"
	"elibrary/internal/domain"
	"elibrary/internal/repository"
	"encoding/json"
	"errors"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
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
		INSERT INTO books (id, barcode, factory_barcode, title, author, publisher, year, location, extra)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
	`,
		book.ID,
		book.Barcode,
		sql.NullString{String: book.FactoryBarcode, Valid: book.FactoryBarcode != ""},
		book.Title,
		book.Author,
		sql.NullString{String: book.Publisher, Valid: book.Publisher != ""},
		sql.NullInt32{Int32: int32(book.Year), Valid: book.Year != 0},
		sql.NullString{String: book.Location, Valid: book.Location != ""},
		extraJSON,
	)

	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return errors.New("barcode already exists")
		}
		return err
	}

	return nil
}

func (r *BookRepository) Update(ctx context.Context, book domain.Book) error {
	extraJSON, err := json.Marshal(book.Extra)
	if err != nil {
		return err
	}

	res, err := r.db.Exec(ctx, `
		UPDATE books
		SET
		    title = $2,
		    author = $3,
		    publisher = $4,
		    year = $5,
		    location = $6,
		    factory_barcode = $7,
		    extra = $8,
		    updated_at = NOW()
		WHERE id = $1
	`,
		book.ID,
		book.Title,
		book.Author,
		sql.NullString{String: book.Publisher, Valid: book.Publisher != ""},
		sql.NullInt32{Int32: int32(book.Year), Valid: book.Year != 0},
		sql.NullString{String: book.Location, Valid: book.Location != ""},
		sql.NullString{String: book.FactoryBarcode, Valid: book.FactoryBarcode != ""},
		extraJSON,
	)

	if err != nil {
		return err
	}

	if res.RowsAffected() == 0 {
		return repository.ErrNotFound
	}

	return nil
}

func (r *BookRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.Book, error) {
	var book domain.Book
	var extraJSON []byte
	var factoryBarcode sql.NullString
	var publisher sql.NullString
	var year sql.NullInt32
	var location sql.NullString

	err := r.db.QueryRow(ctx, `
		SELECT id, barcode, factory_barcode, title, author, publisher, year, location, extra
		FROM books
		WHERE id = $1
	`, id).Scan(
		&book.ID,
		&book.Barcode,
		&factoryBarcode,
		&book.Title,
		&book.Author,
		&publisher,
		&year,
		&location,
		&extraJSON,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.Book{}, repository.ErrNotFound
		}

		return domain.Book{}, err
	}

	if factoryBarcode.Valid {
		book.FactoryBarcode = factoryBarcode.String
	}
	if publisher.Valid {
		book.Publisher = publisher.String
	}
	if year.Valid {
		book.Year = int(year.Int32)
	}
	if location.Valid {
		book.Location = location.String
	}

	if err := json.Unmarshal(extraJSON, &book.Extra); err != nil {
		return domain.Book{}, err
	}

	return book, nil
}

func (r *BookRepository) GetByBarcode(ctx context.Context, barcode string) (*domain.Book, error) {
	var book domain.Book
	var extraJSON []byte
	var factoryBarcode sql.NullString
	var publisher sql.NullString
	var year sql.NullInt32
	var location sql.NullString

	err := r.db.QueryRow(ctx, `
		SELECT id, barcode, factory_barcode, title, author, publisher, year, location, extra
		FROM books
		WHERE barcode = $1
	`, barcode).Scan(
		&book.ID,
		&book.Barcode,
		&factoryBarcode,
		&book.Title,
		&book.Author,
		&publisher,
		&year,
		&location,
		&extraJSON,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, repository.ErrNotFound
		}
		return nil, err
	}

	if factoryBarcode.Valid {
		book.FactoryBarcode = factoryBarcode.String
	}
	if publisher.Valid {
		book.Publisher = publisher.String
	}
	if year.Valid {
		book.Year = int(year.Int32)
	}
	if location.Valid {
		book.Location = location.String
	}

	if err := json.Unmarshal(extraJSON, &book.Extra); err != nil {
		return domain.Book{}, err
	}

	return book, nil
}

func (r *BookRepository) GetByFactoryBarcode(ctx context.Context, factoryBarcode string) ([]*domain.Book, error) {
	rows, err := r.db.Query(ctx, `
		SELECT id, barcode, factory_barcode, title, author, publisher, year, location, extra
		FROM books
		WHERE factory_barcode = $1
		ORDER BY created_at DESC
	`, factoryBarcode)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var books []*domain.Book
	for rows.Next() {
		var book domain.Book
		var extraJSON []byte
		var factoryBarcode sql.NullString
		var publisher sql.NullString
		var year sql.NullInt32
		var location sql.NullString

		if err := rows.Scan(
			&book.ID,
			&book.Barcode,
			&factoryBarcode,
			&book.Title,
			&book.Author,
			&publisher,
			&year,
			&location,
			&extraJSON,
		); err != nil {
			return nil, err
		}

		if factoryBarcode.Valid {
			book.FactoryBarcode = factoryBarcode.String
		}
		if publisher.Valid {
			book.Publisher = publisher.String
		}
		if year.Valid {
			book.Year = int(year.Int32)
		}
		if location.Valid {
			book.Location = location.String
		}

		if err := json.Unmarshal(extraJSON, &book.Extra); err != nil {
			return nil, err
		}

		books = append(books, &book)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	if len(books) == 0 {
		return nil, repository.ErrNotFound
	}

	return books, nil
}

func (r *BookRepository) Search(ctx context.Context, query string) ([]*domain.Book, error) {
	rows, err := r.db.Query(ctx, `
		SELECT id, barcode, factory_barcode, title, author, publisher, year, location, extra
		FROM books
		WHERE search_vector @@ plainto_tsquery('russian', $1)
		ORDER BY ts_rank(search_vector, plainto_tsquery('russian', $1)) DESC
	`, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var books []domain.Book

	for rows.Next() {
		var book domain.Book
		var extraJSON []byte
		var factoryBarcode sql.NullString
		var publisher sql.NullString
		var year sql.NullInt32
		var location sql.NullString

		if err := rows.Scan(
			&book.ID,
			&book.Barcode,
			&factoryBarcode,
			&book.Title,
			&book.Author,
			&publisher,
			&year,
			&location,
			&extraJSON,
		); err != nil {
			return nil, err
		}

		if factoryBarcode.Valid {
			book.FactoryBarcode = factoryBarcode.String
		}
		if publisher.Valid {
			book.Publisher = publisher.String
		}
		if year.Valid {
			book.Year = int(year.Int32)
		}
		if location.Valid {
			book.Location = location.String
		}

		if err := json.Unmarshal(extraJSON, &book.Extra); err != nil {
			return nil, err
		}

		books = append(books, book)
	}

	return books, nil
}
