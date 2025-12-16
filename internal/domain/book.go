package domain

import (
	"github.com/google/uuid"
)

type Book struct {
	ID        uuid.UUID      `json:"id"`
	Barcode   string         `json:"barcode"`
	Title     string         `json:"title"`
	Author    string         `json:"author"`
	Publisher string         `json:"publisher"`
	Year      int            `json:"year"`
	Location  string         `json:"location"`
	Extra     map[string]any `json:"extra"`
}
