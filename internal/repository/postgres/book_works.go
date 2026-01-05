package postgres

import (
	"context"
	"elibrary/internal/repository"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

type BookWorksRepository struct {
	db *pgxpool.Pool
}

func NewBookWorksRepository(db *pgxpool.Pool) *BookWorksRepository {
	return &BookWorksRepository{db: db}
}

func (r *BookWorksRepository) AddWorkToBook(ctx context.Context, bookID uuid.UUID, workID uuid.UUID, position *int) error {
	_, err := r.db.Exec(ctx, `
		INSERT INTO book_works (book_id, work_id, position)
		VALUES ($1, $2, $3)
	`, bookID, workID, position)

	return err
}

func (r *BookWorksRepository) RemoveWorkFromBook(ctx context.Context, bookID uuid.UUID, workID uuid.UUID) error {
	res, err := r.db.Exec(ctx, `
		DELETE FROM book_works
		WHERE book_id = $1 AND work_id = $2
	`, bookID, workID)
	if err != nil {
		return err
	}
	if res.RowsAffected() == 0 {
		return repository.ErrNotFound
	}

	return nil
}

func (r *BookWorksRepository) SetWorkPosition(ctx context.Context, bookID uuid.UUID, workID uuid.UUID, position int) error {
	res, err := r.db.Exec(ctx, `
		UPDATE book_works
		SET position = $3
		WHERE book_id = $1 AND work_id = $2
	`, bookID, workID, position)
	if err != nil {
		return err
	}
	if res.RowsAffected() == 0 {
		return repository.ErrNotFound
	}

	return nil
}

func (r *BookWorksRepository) ReplaceBookWorks(ctx context.Context, bookID uuid.UUID, works []repository.BookWorkInput) error {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	_, err = tx.Exec(ctx, `
		DELETE FROM book_works
		WHERE book_id = $1
	`, bookID)
	if err != nil {
		return err
	}

	for _, w := range works {
		_, err := tx.Exec(ctx, `
			INSERT INTO book_works (book_id, work_id, position)
			VALUES ($1, $2, $3)
		`, bookID, w.WorkID, w.Position)
		if err != nil {
			return err
		}
	}

	return tx.Commit(ctx)
}
