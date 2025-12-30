package domain

import "errors"

var (
	ErrNotFound       = errors.New("not found")
	ErrInvalidBarcode = errors.New("invalid barcode")
	ErrBarcodeExists  = errors.New("barcode already exists")
	ErrInvalidInput   = errors.New("invalid input")
	ErrForbidden      = errors.New("forbidden")
)
