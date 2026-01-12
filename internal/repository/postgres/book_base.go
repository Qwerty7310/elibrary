package postgres

import (
	"elibrary/internal/readmodel"
	"time"

	"github.com/google/uuid"
)

type bookBase struct {
	ID             uuid.UUID
	Barcode        string
	FactoryBarcode *string
	Title          string
	Publisher      *readmodel.Publisher
	Location       *readmodel.Location
	Year           *int
	Description    *string
	Extra          map[string]any
	Works          []*readmodel.WorkShort
	CreatedAt      time.Time
	UpdatedAt      time.Time
}
