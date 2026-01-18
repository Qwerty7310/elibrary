package postgres

import (
	"context"
	"elibrary/internal/domain"
	"elibrary/internal/repository"
	"encoding/json"
	"errors"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

type bookTx struct {
	tx pgx.Tx
}

var _ repository.BookTx = (*bookTx)(nil)

func (t *bookTx) CreateBook(ctx context.Context, book domain.Book) error {
	extraJSON, err := json.Marshal(book.Extra)
	if err != nil {
		return err
	}

	_, err = t.tx.Exec(ctx, `
		INSERT INTO books (id, barcode, factory_barcode, title, publisher_id, year, description, location_id, extra)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
	`,
		book.ID,
		book.Barcode,
		book.FactoryBarcode,
		book.Title,
		book.PublisherID,
		book.Year,
		book.Description,
		book.LocationID,
		extraJSON,
	)

	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return domain.ErrBarcodeExists
		}
		return err
	}

	return nil
}

func (t *bookTx) UpdateBook(ctx context.Context, book domain.Book) error {
	extraJSON, err := json.Marshal(book.Extra)
	if err != nil {
		return err
	}

	res, err := t.tx.Exec(ctx, `
		UPDATE books
		SET
		    barcode = $2,
		    factory_barcode = $3,
		    title = $4,
		    publisher_id = $5,
		    year = $6,
		    description = $7,
		    location_id = $8,
		    extra = $9,
		    updated_at = NOW()
		WHERE id = $1
	`,
		book.ID,
		book.Barcode,
		book.FactoryBarcode,
		book.Title,
		book.PublisherID,
		book.Year,
		book.Description,
		book.LocationID,
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

func (t *bookTx) ReplaceBookWorks(ctx context.Context, bookID uuid.UUID, works []repository.BookWorkInput) error {
	_, err := t.tx.Exec(ctx, `
		DELETE FROM book_works
		WHERE book_id = $1
	`, bookID)
	if err != nil {
		return err
	}

	if len(works) == 0 {
		return nil
	}

	workIDs := make([]uuid.UUID, 0, len(works))
	positions := make([]*int, 0, len(works))

	for _, w := range works {
		workIDs = append(workIDs, w.WorkID)
		positions = append(positions, w.Position)
	}

	_, err = t.tx.Exec(ctx, `
		INSERT INTO book_works (book_id, work_id, position)
		SELECT $1, w, p
		FROM UNNEST($2::uuid[], $3::int[]) AS t(w, p)
	`, bookID, workIDs, positions)

	return err
}
func (t *bookTx) GetDomainByID(ctx context.Context, id uuid.UUID) (*domain.Book, error) {
	var book domain.Book
	var extraJSON []byte

	err := t.tx.QueryRow(ctx, `
		SELECT
		    id,
		    barcode,
		    factory_barcode,
		    title,
		    publisher_id,
		    year,
		    description,
		    location_id,
		    extra,
		    created_at,
		    updated_at
		FROM books
		WHERE id = $1
	`, id).Scan(
		&book.ID,
		&book.Barcode,
		&book.FactoryBarcode,
		&book.Title,
		&book.PublisherID,
		&book.Year,
		&book.Description,
		&book.LocationID,
		&extraJSON,
		&book.CreatedAt,
		&book.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, repository.ErrNotFound
		}
		return nil, err
	}

	if len(extraJSON) > 0 {
		if err := json.Unmarshal(extraJSON, &book.Extra); err != nil {
			return nil, err
		}
	} else {
		book.Extra = map[string]any{}
	}

	return &book, nil
}
