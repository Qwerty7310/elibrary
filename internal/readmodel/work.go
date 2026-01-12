package readmodel

import "github.com/google/uuid"

type WorkShort struct {
	ID      uuid.UUID `json:"id"`
	Title   string    `json:"title"`
	Authors []Author  `json:"authors,omitempty"`
}

type WorkDetailed struct {
	ID          uuid.UUID `json:"id"`
	Title       string    `json:"title"`
	Description *string   `json:"description,omitempty"`
	Year        *int      `json:"year,omitempty"`
	Authors     []Author  `json:"authors,omitempty"`
}
