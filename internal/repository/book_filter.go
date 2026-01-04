package repository

import (
	"github.com/google/uuid"
)

type BookFilter struct {
	ID             *uuid.UUID
	Barcode        *string
	FactoryBarcode *string
	Query          *string

	PublisherID *uuid.UUID
	YearFrom    *int
	YearTo      *int

	Limit  *int
	Offset *int
}

func (f BookFilter) LimitOr(def int) int {
	if f.Limit != nil {
		return *f.Limit
	}
	return def
}

func (f BookFilter) OffsetOr(def int) int {
	if f.Offset != nil {
		return *f.Offset
	}
	return def
}
