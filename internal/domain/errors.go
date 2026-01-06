package domain

import "errors"

var (
	ErrNotFound       = errors.New("not found")
	ErrInvalidInput   = errors.New("invalid input")
	ErrForbidden      = errors.New("forbidden")
	ErrInvalidBarcode = errors.New("invalid barcode")
	ErrBarcodeExists  = errors.New("barcode already exists")
)
