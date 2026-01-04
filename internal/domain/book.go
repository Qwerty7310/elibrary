package domain

import (
	"time"

	"github.com/google/uuid"
)

type Book struct {
	ID             uuid.UUID `json:"id"`
	Title          string    `json:"title"`
	Barcode        string    `json:"barcode"`
	FactoryBarcode *string   `json:"factory_barcode,omitempty"`

	PublisherID *uuid.UUID `json:"publisher_id,omitempty"`
	LocationID  *uuid.UUID `json:"location_id,omitempty"`
	Year        *int       `json:"year,omitempty"`
	Description *string    `json:"description,omitempty"`

	Extra map[string]any `json:"extra,omitempty"`

	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}
