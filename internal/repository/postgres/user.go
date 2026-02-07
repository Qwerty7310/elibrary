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

func (r *UserRepository) Create(ctx context.Context, user domain.User) error {
	tx, err := r.db.BeginTx(ctx, pgx.TxOptions{
		IsoLevel:   pgx.ReadCommitted,
		AccessMode: pgx.ReadWrite,
	})
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	_, err = tx.Exec(ctx, `
		INSERT INTO users (id, login, first_name, last_name, middle_name, email, password_hash)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
	`,
		user.ID,
		user.Login,
		user.FirstName,
		user.LastName,
		user.MiddleName,
		user.Email,
		user.PasswordHash,
	)
	if err != nil {
		return err
	}

	roleCodes := make([]string, 0, len(user.Roles))
	for _, role := range user.Roles {
		if role.Code != "" {
			roleCodes = append(roleCodes, role.Code)
		}
	}

	_, err = tx.Exec(ctx, `
		INSERT INTO user_roles (user_id, role_id)
		SELECT $1, r.id
		FROM roles r
		WHERE r.code = ANY($2::text[])
	`, user.ID, roleCodes)
	if err != nil {
		return err
	}

	return tx.Commit(ctx)
}

func (r *UserRepository) Update(ctx context.Context, user domain.User) error {
	tx, err := r.db.BeginTx(ctx, pgx.TxOptions{
		IsoLevel:   pgx.ReadCommitted,
		AccessMode: pgx.ReadWrite,
	})
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	res, err := tx.Exec(ctx, `
		UPDATE users
		SET
		    login = $2,
		    first_name = $3,
		    last_name = $4,
		    middle_name = $5,
		    email = $6,
		    password_hash = $7,
		    is_active = $8,
		    updated_at = NOW()
		WHERE id = $1
	`,
		user.ID,
		user.Login,
		user.FirstName,
		user.LastName,
		user.MiddleName,
		user.Email,
		user.PasswordHash,
		user.IsActive,
	)
	if err != nil {
		return err
	}
	if res.RowsAffected() == 0 {
		return repository.ErrNotFound
	}

	_, err = tx.Exec(ctx, `
		DELETE FROM user_roles
		WHERE user_id = $1
	`, user.ID)
	if err != nil {
		return err
	}

	roleCodes := make([]string, 0, len(user.Roles))
	for _, role := range user.Roles {
		if role.Code != "" {
			roleCodes = append(roleCodes, role.Code)
		}
	}

	_, err = tx.Exec(ctx, `
		INSERT INTO user_roles (user_id, role_id)
		SELECT $1, r.id
		FROM roles r
		WHERE r.code = ANY($2::text[])
	`, user.ID, roleCodes)
	if err != nil {
		return err
	}

	return tx.Commit(ctx)
}

func (r *UserRepository) Delete(ctx context.Context, id uuid.UUID) error {
	tx, err := r.db.BeginTx(ctx, pgx.TxOptions{
		IsoLevel:   pgx.ReadCommitted,
		AccessMode: pgx.ReadWrite,
	})
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	_, err = tx.Exec(ctx, `
		DELETE FROM user_roles
		WHERE user_id = $1
	`, id)
	if err != nil {
		return err
	}

	res, err := tx.Exec(ctx, `
		DELETE FROM users
		WHERE id = $1
	`, id)
	if err != nil {
		return err
	}
	if res.RowsAffected() == 0 {
		return repository.ErrNotFound
	}

	return tx.Commit(ctx)
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

func (r *UserRepository) GetByIDWithRoles(ctx context.Context, id uuid.UUID) (*domain.User, error) {
	rows, err := r.db.Query(ctx, `
		SELECT
			u.id, u.login, u.first_name, u.last_name, u.middle_name, u.email,
			u.password_hash, u.is_active, u.created_at, u.updated_at,
			r.id, r.code, r.name
		FROM users u
		LEFT JOIN user_roles ur ON ur.user_id = u.id
		LEFT JOIN roles r ON r.id = ur.role_id
		WHERE u.id = $1
	`, id)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var user *domain.User

	for rows.Next() {
		var (
			roleID   *int
			roleCode *string
			roleName *string

			u domain.User
		)
		err := rows.Scan(
			&u.ID,
			&u.Login,
			&u.FirstName,
			&u.LastName,
			&u.MiddleName,
			&u.Email,
			&u.PasswordHash,
			&u.IsActive,
			&u.CreatedAt,
			&u.UpdatedAt,
			&roleID,
			&roleCode,
			&roleName,
		)
		if err != nil {
			return nil, err
		}

		if user == nil {
			user = &u
		}

		if roleID != nil {
			user.Roles = append(user.Roles, domain.Role{
				ID:   *roleID,
				Code: *roleCode,
				Name: *roleName,
			})
		}
	}

	if user == nil {
		return nil, repository.ErrNotFound
	}

	return user, nil
}

func (r *UserRepository) GetAllWithRoles(ctx context.Context) ([]*domain.User, error) {
	rows, err := r.db.Query(ctx, `
		SELECT
			u.id, u.login, u.first_name, u.last_name, u.middle_name, u.email,
			u.password_hash, u.is_active, u.created_at, u.updated_at,
			r.id, r.code, r.name
		FROM users u
		LEFT JOIN user_roles ur ON ur.user_id = u.id
		LEFT JOIN roles r ON r.id = ur.role_id
		ORDER BY u.login, r.code
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	usersByID := make(map[uuid.UUID]*domain.User)
	orderedIDs := make([]uuid.UUID, 0)

	for rows.Next() {
		var (
			roleID   *int
			roleCode *string
			roleName *string

			u domain.User
		)
		err := rows.Scan(
			&u.ID,
			&u.Login,
			&u.FirstName,
			&u.LastName,
			&u.MiddleName,
			&u.Email,
			&u.PasswordHash,
			&u.IsActive,
			&u.CreatedAt,
			&u.UpdatedAt,
			&roleID,
			&roleCode,
			&roleName,
		)
		if err != nil {
			return nil, err
		}

		existing, ok := usersByID[u.ID]
		if !ok {
			existing = &u
			usersByID[u.ID] = existing
			orderedIDs = append(orderedIDs, u.ID)
		}

		if roleID != nil && roleCode != nil && roleName != nil {
			existing.Roles = append(existing.Roles, domain.Role{
				ID:   *roleID,
				Code: *roleCode,
				Name: *roleName,
			})
		}
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	if len(orderedIDs) == 0 {
		return nil, repository.ErrNotFound
	}

	result := make([]*domain.User, 0, len(orderedIDs))
	for _, id := range orderedIDs {
		result = append(result, usersByID[id])
	}

	return result, nil
}
