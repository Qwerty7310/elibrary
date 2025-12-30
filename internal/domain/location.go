package domain

import (
	"time"

	"github.com/google/uuid"
)

type LocationType string

const (
	LocationTypeBuilding LocationType = "building"
	LocationTypeRoom     LocationType = "room"
	LocationTypeCabinet  LocationType = "cabinet"
	LocationTypeShelf    LocationType = "shelf"
)

func (t LocationType) Level() int {
	switch t {
	case LocationTypeBuilding:
		return 1
	case LocationTypeRoom:
		return 2
	case LocationTypeCabinet:
		return 3
	case LocationTypeShelf:
		return 4
	default:
		return -1
	}
}

func (t LocationType) IsChildOf(parent LocationType) bool {
	return t.Level() == parent.Level()+1
}

type Location struct {
	ID       uuid.UUID    `json:"id"`
	ParentID *uuid.UUID   `json:"parent_id,omitempty"`
	Type     LocationType `json:"type"`
	Name     string       `json:"name"`

	Barcode     string  `json:"barcode"`
	Address     *string `json:"address,omitempty"`
	Description *string `json:"description,omitempty"`

	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}
