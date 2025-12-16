package domain

import (
	"github.com/google/uuid"
)

type Book struct {
	ID        uuid.UUID
	Title     string
	Author    string
	Publisher string
	Year      int
	Location  string
	Extra     map[string]any
}
