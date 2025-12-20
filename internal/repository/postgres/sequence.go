package postgres

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
)

type SequenceRepository struct {
	db *pgxpool.Pool
}

func NewSequenceRepository(db *pgxpool.Pool) *SequenceRepository {
	return &SequenceRepository{db: db}
}

func (r *SequenceRepository) GetNext(ctx context.Context, prefix int) (int64, error) {
	var sequence int64
	err := r.db.QueryRow(ctx, `
		INSERT INTO barcode_sequences (prefix)
		VALUES ($1)
		ON CONFLICT (prefix)
		DO UPDATE SET last_value = barcode_sequences.last_value + 1
		RETURNING last_value
	`, prefix).Scan(&sequence)

	return sequence, err
}

func (r *SequenceRepository) SetPrefix(ctx context.Context, prefix int, description string) error {
	_, err := r.db.Exec(ctx, `
		INSERT INTO barcode_sequences (prefix, description)
		VALUES ($1, $2)
		ON CONFLICT (prefix)
		DO UPDATE SET description = EXCLUDED.description
	`, prefix, description)

	return err
}
