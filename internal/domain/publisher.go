package domain

import (
	"time"

	"github.com/google/uuid"
)

type Publisher struct {
	ID      uuid.UUID `json:"id"`
	Name    string    `json:"name"`
	LogoURL *string   `json:"logo_url,omitempty"`
	WebURL  *string   `json:"web_url,omitempty"`

	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}
