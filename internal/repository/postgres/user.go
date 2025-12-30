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

type UserRepository struct {
	db *pgxpool.Pool
}

func NewUserRepository(db *pgxpool.Pool) *UserRepository {
	return &UserRepository{
		db: db,
	}
}

func (r *UserRepository) GetByLogin(ctx context.Context, login string) (*domain.User, error) {
	var user domain.User

	err := r.db.QueryRow(ctx, `
		SELECT id, login, first_name, last_name, middle_name, email, password_hash, is_active, created_at, updated_at
		FROM users
		WHERE login = $1
	`, login).Scan(
		&user.ID,
		&user.Login,
		&user.FirstName,
		&user.LastName,
		&user.MiddleName,
		&user.Email,
		&user.PasswordHash,
		&user.IsActive,
		&user.CreatedAt,
		&user.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, repository.ErrNotFound
		}
		return nil, err
	}

	return &user, nil
}

func (r *UserRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.User, error) {
	var user domain.User

	err := r.db.QueryRow(ctx, `
		SELECT id, login, first_name, last_name, middle_name, email, password_hash, is_active, created_at, updated_at
		FROM users
		WHERE id = $1
	`, id).Scan(
		&user.ID,
		&user.Login,
		&user.FirstName,
		&user.LastName,
		&user.MiddleName,
		&user.Email,
		&user.PasswordHash,
		&user.IsActive,
		&user.CreatedAt,
		&user.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, repository.ErrNotFound
		}
		return nil, err
	}

	return &user, nil
}
