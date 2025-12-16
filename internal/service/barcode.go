package service

import (
	"bytes"
	"image/png"

	"github.com/boombuler/barcode/code128"
	"github.com/google/uuid"
)

type BarcodeService struct{}

func NewBarcodeService() *BarcodeService {
	return &BarcodeService{}
}

func (s *BarcodeService) Generate(id uuid.UUID) ([]byte, error) {
	code, err := code128.Encode(id.String())
	if err != nil {
		return nil, err
	}

	var buf bytes.Buffer
	if err := png.Encode(&buf, code); err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}
