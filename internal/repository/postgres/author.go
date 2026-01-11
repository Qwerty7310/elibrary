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

type AuthorRepository struct {
	db *pgxpool.Pool
}

func NewAuthorRepository(db *pgxpool.Pool) *AuthorRepository {
	return &AuthorRepository{db: db}
}

func (r *AuthorRepository) Create(ctx context.Context, author domain.Author) error {
	_, err := r.db.Exec(ctx, `
		INSERT INTO authors (id, last_name, first_name, middle_name, birth_date, death_date, bio, photo_url)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
	`,
		author.ID,
		author.LastName,
		author.FirstName,
		author.MiddleName,
		author.BirthDate,
		author.DeathDate,
		author.Bio,
		author.PhotoURL,
	)
	if err != nil {
		return err
	}

	return nil
}

func (r *AuthorRepository) Update(ctx context.Context, author domain.Author) error {
	res, err := r.db.Exec(ctx, `
		UPDATE authors
		SET
		    last_name = $2,
		    first_name = $3,
		    middle_name = $4,
		    birth_date = $5,
		    death_date = $6,
		    bio = $7,
		    photo_url = $8,
			updated_at = NOW()
		WHERE id = $1
	`,
		author.ID,
		author.LastName,
		author.FirstName,
		author.MiddleName,
		author.BirthDate,
		author.DeathDate,
		author.Bio,
		author.PhotoURL,
	)

	if err != nil {
		return err
	}

	if res.RowsAffected() == 0 {
		return repository.ErrNotFound
	}

	return nil
}

func (r *AuthorRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.Author, error) {
	var author domain.Author

	err := r.db.QueryRow(ctx, `
		SELECT id, last_name, first_name, middle_name, birth_date, death_date, bio, photo_url, created_at, updated_at
		FROM authors
		WHERE id = $1
	`).Scan(
		&author.ID,
		&author.LastName,
		&author.FirstName,
		&author.MiddleName,
		&author.BirthDate,
		&author.DeathDate,
		&author.Bio,
		&author.PhotoURL,
		&author.CreatedAt,
		&author.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, repository.ErrNotFound
		}
		return nil, err
	}

	return &author, nil
}

func (r *AuthorRepository) Delete(ctx context.Context, id uuid.UUID) error {
	res, err := r.db.Exec(ctx, `
		DELETE FROM authors
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

func (r *AuthorRepository) GetAll(ctx context.Context) ([]readmodel.Author, error) {
	rows, err := r.db.Query(ctx, `
		SELECT id, last_name, first_name, middle_name
		FROM authors
		ORDER BY last_name, first_name, middle_name
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var authors []readmodel.Author
	for rows.Next() {
		var author readmodel.Author
		if err := rows.Scan(
			&author.ID,
			&author.LastName,
			&author.FirstName,
			&author.MiddleName,
		); err != nil {
			return nil, err
		}
		authors = append(authors, author)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return authors, nil
}
