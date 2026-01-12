package readmodel

import (
	"time"

	"github.com/google/uuid"
)

type BookPublic struct {
	ID             uuid.UUID `json:"id"`
	Title          string    `json:"title"`
	Barcode        string    `json:"barcode"`
	FactoryBarcode *string   `json:"factory_barcode,omitempty"`

	Publisher   *Publisher   `json:"publisher,omitempty"`
	Works       []*WorkShort `json:"works,omitempty"`
	Year        *int         `json:"year,omitempty"`
	Description *string      `json:"description,omitempty"`

	Extra map[string]any `json:"extra,omitempty"`

	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type BookInternal struct {
	ID             uuid.UUID `json:"id"`
	Title          string    `json:"title"`
	Barcode        string    `json:"barcode"`
	FactoryBarcode *string   `json:"factory_barcode,omitempty"`

	Publisher   *Publisher   `json:"publisher,omitempty"`
	Location    *Location    `json:"location,omitempty"`
	Works       []*WorkShort `json:"works,omitempty"`
	Year        *int         `json:"year,omitempty"`
	Description *string      `json:"description,omitempty"`

	Extra map[string]any `json:"extra,omitempty"`

	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type Publisher struct {
	ID   uuid.UUID `json:"id"`
	Name string    `json:"name"`
}

type Location struct {
	ShelfID      uuid.UUID `json:"shelf_id"`
	ShelfName    string    `json:"shelf_name"`
	CabinetID    uuid.UUID `json:"cabinet_id"`
	CabinetName  string    `json:"cabinet_name"`
	RoomID       uuid.UUID `json:"room_id"`
	RoomName     string    `json:"room_name"`
	BuildingID   uuid.UUID `json:"building_id"`
	BuildingName string    `json:"building_name"`
	Address      string    `json:"address"`
}

type Author struct {
	ID         uuid.UUID `json:"id"`
	LastName   string    `json:"last_name"`
	FirstName  *string   `json:"first_name,omitempty"`
	MiddleName *string   `json:"middle_name,omitempty"`
}
