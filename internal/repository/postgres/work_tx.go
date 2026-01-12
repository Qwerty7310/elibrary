package postgres

import (
	"context"
	"elibrary/internal/domain"
	"elibrary/internal/repository"
	"errors"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

type workTx struct {
	tx pgx.Tx
}

var _ repository.WorkTx = (*workTx)(nil)

func (t *workTx) CreateWork(ctx context.Context, work domain.Work) error {
	_, err := t.tx.Exec(ctx, `
		INSERT INTO works (id, title, description, year)
		VALUES ($1, $2, $3, $4)
	`,
		work.ID,
		work.Title,
		work.Description,
		work.Year,
	)
	if err != nil {
		return err
	}

	return nil
}

func (t *workTx) UpdateWork(ctx context.Context, work domain.Work) error {
	res, err := t.tx.Exec(ctx, `
		UPDATE works
		SET
		    title = $2,
		    description = $3,
		    year = $4,
		    updated_at = NOW()
		WHERE id = $1
	`,
		work.ID,
		work.Title,
		work.Description,
		work.Year,
	)
	if err != nil {
		return err
	}

	if res.RowsAffected() == 0 {
		return repository.ErrNotFound
	}

	return nil
}

func (t *workTx) GetDomainByID(ctx context.Context, id uuid.UUID) (*domain.Work, error) {
	var work domain.Work

	err := t.tx.QueryRow(ctx, `
		SELECT id, title, description, year, created_at, updated_at
		FROM works
		WHERE id = $1
	`, id).Scan(
		&work.ID,
		&work.Title,
		&work.Description,
		&work.Year,
		&work.CreatedAt,
		&work.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, repository.ErrNotFound
		}
		return nil, err
	}
	return &work, nil
}

func (t *workTx) ReplaceWorkAuthors(ctx context.Context, workID uuid.UUID, authors []uuid.UUID) error {
	_, err := t.tx.Exec(ctx, `
		DELETE FROM work_authors
		WHERE work_id = $1
	`, workID)
	if err != nil {
		return err
	}

	if len(authors) == 0 {
		return nil
	}

	_, err = t.tx.Exec(ctx, `
		INSERT INTO work_authors (work_id, author_id)
		SELECT $1, a
		FROM UNNEST($2::uuid[]) AS a
	`, workID, authors)

	return err
}
