package domain

import (
	"time"

	"github.com/google/uuid"
)

type Role struct {
	ID   int    `json:"id"`
	Code string `json:"code"`
	Name string `json:"name"`
}

type User struct {
	ID         uuid.UUID `json:"id"`
	Login      string    `json:"login"`
	FirstName  string    `json:"first_name"`
	LastName   *string   `json:"last_name,omitempty"`
	MiddleName *string   `json:"middle_name,omitempty"`
	Email      *string   `json:"email,omitempty"`

	PasswordHash string `json:"-"`
	IsActive     bool   `json:"is_active"`

	Roles []Role `json:"roles"`

	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}
