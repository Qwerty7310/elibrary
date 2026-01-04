package domain

import (
	"time"

	"github.com/google/uuid"
)

type Author struct {
	ID         uuid.UUID  `json:"id"`
	LastName   string     `json:"last_name"`
	FirstName  *string    `json:"first_name,omitempty"`
	MiddleName *string    `json:"middle_name,omitempty"`
	BirthDate  *time.Time `json:"birth_date,omitempty"`
	DeathDate  *time.Time `json:"death_date,omitempty"`
	Bio        *string    `json:"bio,omitempty"`
	PhotoURL   *string    `json:"photo_url,omitempty"`

	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}
