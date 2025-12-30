package postgres

import (
	"context"
	"elibrary/internal/domain"
	"elibrary/internal/repository"
	"errors"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

type LocationRepository struct {
	db *pgxpool.Pool
}

func NewLocationRepository(db *pgxpool.Pool) *LocationRepository {
	return &LocationRepository{db: db}
}

func (r *LocationRepository) Create(ctx context.Context, location domain.Location) error {
	_, err := r.db.Exec(ctx, `
		INSERT INTO locations (id, parent_id, type, name, barcode, address, description,)
		VALUES ($1, $2, $3, $4, $5, $6)
	`,
		location.ID,
		location.ParentID,
		location.Type,
		location.Name,
		location.Barcode,
		location.Address,
		location.Description,
	)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return errors.New("barcode already exists")
		}
		return err
	}

	return nil
}

func (r *LocationRepository) Update(ctx context.Context, location domain.Location) error {
	res, err := r.db.Exec(ctx, `
		UPDATE locations
		SET
		    parent_id = $2,
		    name = $3,
		    address = $4,
		    description = $5,
		    updated_at = NOW()
		WHERE id = $1
	`,
		location.ID,
		location.ParentID,
		location.Name,
		location.Address,
		location.Description,
	)
	if err != nil {
		return err
	}

	if res.RowsAffected() == 0 {
		return repository.ErrNotFound
	}

	return nil
}

func (r *LocationRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.Location, error) {
	var location domain.Location

	err := r.db.QueryRow(ctx, `
		SELECT id, parent_id, type, name, barcode, address, description, created_at, updated_at
		FROM locations
		WHERE id = $1
	`, id).Scan(
		&location.ID,
		&location.ParentID,
		&location.Type,
		&location.Name,
		&location.Barcode,
		&location.Address,
		&location.Description,
		&location.CreatedAt,
		&location.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, repository.ErrNotFound
		}
	}

	return &location, nil
}

func (r *LocationRepository) GetByTypeParentID(ctx context.Context, locType domain.LocationType, parentID uuid.UUID) ([]*domain.Location, error) {
	rows, err := r.db.Query(ctx, `
		SELECT id, parent_id, type, name, barcode, address, description, created_at, updated_at
		FROM locations
		WHERE type = $1
		  AND (
		    ($2::uuid IS NULL AND parent_id IS NULL)
		    OR parent_id = $2
		  )
		ORDER BY name
	`, locType, parentID)

	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var locations []*domain.Location
	for rows.Next() {
		var loc domain.Location

		if err := rows.Scan(
			&loc.ID,
			&loc.ParentID,
			&loc.Type,
			&loc.Name,
			&loc.Barcode,
			&loc.Address,
			&loc.Description,
			&loc.CreatedAt,
			&loc.UpdatedAt,
		); err != nil {
			return nil, err
		}

		locations = append(locations, &loc)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	if len(locations) == 0 {
		return nil, repository.ErrNotFound
	}

	return locations, nil
}

func (r *LocationRepository) GetByBarcode(ctx context.Context, barcode string) (*domain.Location, error) {
	var location domain.Location

	err := r.db.QueryRow(ctx, `
		SELECT id, parent_id, type, name, barcode, address, description, created_at, updated_at
		FROM locations
		WHERE barcode = $1
	`, barcode).Scan(
		&location.ID,
		&location.ParentID,
		&location.Type,
		&location.Name,
		&location.Barcode,
		&location.Address,
		&location.Description,
		&location.CreatedAt,
		&location.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, repository.ErrNotFound
		}
	}

	return &location, nil
}

func (r *LocationRepository) Delete(ctx context.Context, id uuid.UUID) error {
	res, err := r.db.Exec(ctx, `
		DELETE FROM locations
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
