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
	Author         string         `json:"author"`
	Publisher      string         `json:"publisher"`
	Year           int            `json:"year"`
	Location       string         `json:"location"`
	Extra          map[string]any `json:"extra"`
	CreatedAt      time.Time      `json:"created_at,omitempty"`
	UpdatedAt      time.Time      `json:"updated_at,omitempty"`
}

func NewBook() Book {
	return Book{
		ID:    uuid.New(),
		Extra: make(map[string]any),
	}
}
