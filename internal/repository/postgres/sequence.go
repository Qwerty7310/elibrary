package postgres

import (
	"context"
	"elibrary/internal/domain"

	"github.com/jackc/pgx/v5/pgxpool"
)

type SequenceRepository struct {
	db *pgxpool.Pool
}

func NewSequenceRepository(db *pgxpool.Pool) *SequenceRepository {
	return &SequenceRepository{db: db}
}

func (r *SequenceRepository) GetNext(ctx context.Context, t domain.BarcodeType) (seq int64, prefix int, err error) {
	err = r.db.QueryRow(ctx, `
		UPDATE barcode_sequences
		SET last_value = last_value + 1
		WHERE type = $1
		RETURNING last_value, prefix
	`, t).Scan(&seq, &prefix)

	return
}

func (r *SequenceRepository) SetType(ctx context.Context, t domain.BarcodeType, prefix int, description string) error {
	_, err := r.db.Exec(ctx, `
		INSERT INTO barcode_sequences (type, prefix, description)
		VALUES ($1, $2, $3)
		ON CONFLICT (type)
		DO UPDATE SET
		    prefix = EXCLUDED.prefix,
		    description = EXCLUDED.description,
		    updated_at = NOW()
	`, t, prefix, description)

	return err
}
