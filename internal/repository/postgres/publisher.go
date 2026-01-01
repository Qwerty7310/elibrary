package postgres

import (
	"context"
	"elibrary/internal/domain"
	"elibrary/internal/repository"
	"errors"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type PublisherRepository struct {
	db *pgxpool.Pool
}

func NewPublisherRepository(db *pgxpool.Pool) *PublisherRepository {
	return &PublisherRepository{db: db}
}

func (r *PublisherRepository) Create(ctx context.Context, publisher domain.Publisher) error {
	_, err := r.db.Exec(ctx, `
		INSERT INTO publishers (id, name, logo_url, web_url)
		VALUES ($1, $2, $3, $4)
	`,
		publisher.ID,
		publisher.Name,
		publisher.LogoURL,
		publisher.WebURL,
	)
	if err != nil {
		return err
	}

	return nil
}

func (r *PublisherRepository) Update(ctx context.Context, publisher domain.Publisher) error {
	res, err := r.db.Exec(ctx, `
		UPDATE publishers
		SET
		    name = $2,
		    logo_url = $3,
		    web_url = $4,
		    updated_at = NOW()
		WHERE id = $1
	`,
		publisher.ID,
		publisher.Name,
		publisher.LogoURL,
		publisher.WebURL,
	)
	if err != nil {
		return err
	}

	if res.RowsAffected() == 0 {
		return repository.ErrNotFound
	}

	return nil
}

func (r *PublisherRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.Publisher, error) {
	var publisher domain.Publisher

	err := r.db.QueryRow(ctx, `
		SELECT id, name, logo_url, web_url, created_at, updated_at
		FROM publishers
		WHERE id = $1
	`, id).Scan(
		&publisher.ID,
		&publisher.Name,
		&publisher.LogoURL,
		&publisher.WebURL,
		&publisher.CreatedAt,
		&publisher.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, repository.ErrNotFound
		}
		return nil, err
	}

	return &publisher, nil
}

func (r *PublisherRepository) Delete(ctx context.Context, id uuid.UUID) error {
	res, err := r.db.Exec(ctx, `
		DELETE FROM publishers
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
