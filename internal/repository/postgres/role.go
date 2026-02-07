package postgres

import (
	"context"
	"elibrary/internal/domain"
	"elibrary/internal/readmodel"
	"elibrary/internal/repository"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

type RoleRepository struct {
	db *pgxpool.Pool
}

func NewRoleRepository(db *pgxpool.Pool) *RoleRepository {
	return &RoleRepository{db: db}
}

func (r *RoleRepository) GetAllWithPermissions(ctx context.Context) ([]*readmodel.RoleWithPermissions, error) {
	rows, err := r.db.Query(ctx, `
		SELECT
			rl.id, rl.code, rl.name,
			p.id, p.code, p.name
		FROM roles rl
		LEFT JOIN role_permissions rp ON rp.role_id = rl.id
		LEFT JOIN permissions p ON p.id = rp.permission_id
		ORDER BY rl.code, p.code
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	rolesByID := make(map[int]*readmodel.RoleWithPermissions)
	order := make([]int, 0)

	for rows.Next() {
		var (
			roleID   int
			roleCode string
			roleName string

			permID   *int
			permCode *string
			permName *string
		)
		if err := rows.Scan(
			&roleID,
			&roleCode,
			&roleName,
			&permID,
			&permCode,
			&permName,
		); err != nil {
			return nil, err
		}

		role, ok := rolesByID[roleID]
		if !ok {
			role = &readmodel.RoleWithPermissions{
				ID:   roleID,
				Code: roleCode,
				Name: roleName,
			}
			rolesByID[roleID] = role
			order = append(order, roleID)
		}

		if permID != nil && permCode != nil && permName != nil {
			role.Permissions = append(role.Permissions, readmodel.Permission{
				ID:   *permID,
				Code: *permCode,
				Name: *permName,
			})
		}
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}
	if len(order) == 0 {
		return nil, repository.ErrNotFound
	}

	result := make([]*readmodel.RoleWithPermissions, 0, len(order))
	for _, roleID := range order {
		result = append(result, rolesByID[roleID])
	}
	return result, nil
}

func (r *RoleRepository) GetAllPermissions(ctx context.Context) ([]*readmodel.Permission, error) {
	rows, err := r.db.Query(ctx, `
		SELECT id, code, name
		FROM permissions
		ORDER BY code
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	result := make([]*readmodel.Permission, 0)
	for rows.Next() {
		var perm readmodel.Permission
		if err := rows.Scan(&perm.ID, &perm.Code, &perm.Name); err != nil {
			return nil, err
		}
		result = append(result, &perm)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	if len(result) == 0 {
		return nil, repository.ErrNotFound
	}
	return result, nil
}

func (r *RoleRepository) Create(ctx context.Context, code, name string, permissionCodes []string) error {
	tx, err := r.db.BeginTx(ctx, pgx.TxOptions{
		IsoLevel:   pgx.ReadCommitted,
		AccessMode: pgx.ReadWrite,
	})
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	var roleID int
	err = tx.QueryRow(ctx, `
		INSERT INTO roles (code, name)
		VALUES ($1, $2)
		RETURNING id
	`, code, name).Scan(&roleID)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return domain.ErrRoleExists
		}
		return err
	}

	if len(permissionCodes) > 0 {
		_, err = tx.Exec(ctx, `
			INSERT INTO role_permissions (role_id, permission_id)
			SELECT $1, p.id
			FROM permissions p
			WHERE p.code = ANY($2::text[])
		`, roleID, permissionCodes)
		if err != nil {
			return err
		}
	}

	return tx.Commit(ctx)
}

