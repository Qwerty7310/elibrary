package postgres

import (
	"context"
	"elibrary/internal/domain"
	"elibrary/internal/readmodel"
	"elibrary/internal/repository"
	"errors"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type WorkRepository struct {
	db *pgxpool.Pool
}

func NewWorkRepository(db *pgxpool.Pool) *WorkRepository {
	return &WorkRepository{db: db}
}

func (r *WorkRepository) Create(ctx context.Context, work domain.Work) error {
	_, err := r.db.Exec(ctx, `
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

func (r *WorkRepository) Update(ctx context.Context, work domain.Work) error {
	res, err := r.db.Exec(ctx, `
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

func (r *WorkRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.Work, error) {
	var work domain.Work

	err := r.db.QueryRow(ctx, `
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

func (r *WorkRepository) Delete(ctx context.Context, id uuid.UUID) error {
	res, err := r.db.Exec(ctx, `
		DELETE FROM works
		WHERE id = $1
	`, id)

	if err != nil {
		return err
	}

	if res.RowsAffected() == 0 {
		return repository.ErrNotFound
	}

	return nil
}

func (r *WorkRepository) GetAll(ctx context.Context) ([]*readmodel.Work, error) {
	rows, err := r.db.Query(ctx, `
		SELECT
			w.id,
			w.title,
			a.id,
			a.last_name,
			a.first_name,
			a.middle_name
		FROM works w
		LEFT JOIN work_authors wa ON wa.work_id = w.id
		LEFT JOIN authors a ON a.id = wa.author_id
		ORDER BY w.title, a.last_name, a.first_name
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	workMap := make(map[uuid.UUID]*readmodel.Work)
	res := make([]*readmodel.Work, 0, 64)

	for rows.Next() {
		var (
			workID uuid.UUID
			title  string

			authorID   *uuid.UUID
			lastName   *string
			firstName  *string
			middleName *string
		)

		if err := rows.Scan(
			&workID,
			&title,
			&authorID,
			&lastName,
			&firstName,
			&middleName,
		); err != nil {
			return nil, err
		}

		work, ok := workMap[workID]
		if !ok {
			work = &readmodel.Work{
				ID:    workID,
				Title: title,
			}
			workMap[workID] = work
			res = append(res, work)
		}

		if authorID != nil {
			work.Authors = append(work.Authors, readmodel.Author{
				ID:         *authorID,
				LastName:   derefStr(lastName),
				FirstName:  firstName,
				MiddleName: middleName,
			})
		}
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return res, nil
}
