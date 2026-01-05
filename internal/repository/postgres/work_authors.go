package postgres

import (
	"context"
	"elibrary/internal/repository"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

type WorkAuthorsRepository struct {
	db *pgxpool.Pool
}

func NewWorkAuthorsRepository(db *pgxpool.Pool) *WorkAuthorsRepository {
	return &WorkAuthorsRepository{db: db}
}

func (r *WorkAuthorsRepository) AddAuthorToWork(ctx context.Context, workID uuid.UUID, authorID uuid.UUID) error {
	_, err := r.db.Exec(ctx, `
		INSERT INTO work_authors (work_id, author_id)
		VALUES ($1, $2)
	`, workID, authorID)

	return err
}

func (r *WorkAuthorsRepository) RemoveAuthorFromWork(ctx context.Context, workID uuid.UUID, authorID uuid.UUID) error {
	res, err := r.db.Exec(ctx, `
		DELETE FROM work_authors
		WHERE work_id = $1 AND author_id = $2
	`, workID, authorID)
	if err != nil {
		return err
	}
	if res.RowsAffected() == 0 {
		return repository.ErrNotFound
	}

	return nil
}

func (r *WorkAuthorsRepository) ReplaceAuthors(ctx context.Context, workID uuid.UUID, authorsIDs []uuid.UUID) error {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	_, err = tx.Exec(ctx, `
		DELETE FROM work_authors
		WHERE work_id = $1
	`, workID)
	if err != nil {
		return err
	}

	if len(authorsIDs) > 0 {
		_, err = tx.Exec(ctx, `
			INSERT INTO work_authors (work_id, author_id)
			SELECT $1, UNNEST($2::uuid[])
		`, workID, authorsIDs)
		if err != nil {
			return err
		}
	}

	return tx.Commit(ctx)
}
