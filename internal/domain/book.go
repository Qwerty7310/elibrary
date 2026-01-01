package domain

import (
	"time"

	"github.com/google/uuid"
)

type Book struct {
	ID             uuid.UUID      `json:"id"`
	Barcode        string         `json:"barcode"`
	FactoryBarcode string         `json:"factory_barcode,omitempty"`
	Title          string         `json:"title"`
	Publisher      *Publisher     `json:"publisher,omitempty"`
	Year           int            `json:"year"`
	Description    string         `json:"description,omitempty"`
	Content        []Work         `json:"content"`
	Location       *Location      `json:"location,omitempty"`
	Extra          map[string]any `json:"extra,omitempty"`
	CreatedAt      time.Time      `json:"created_at"`
	UpdatedAt      time.Time      `json:"updated_at"`
}

type Work struct {
	Title   string   `json:"title"`
	Authors []Author `json:"authors,omitempty"`
}
