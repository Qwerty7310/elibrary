package service

import (
	"bytes"
	"context"
	"elibrary/internal/domain"
	"elibrary/internal/repository"
	"fmt"
	"image/png"
	"strconv"

	"github.com/boombuler/barcode"
	"github.com/boombuler/barcode/ean"
)

type BarcodeService struct {
	seqRepo repository.SequenceRepository
}

func NewBarcodeService(seqRepo repository.SequenceRepository) *BarcodeService {
	return &BarcodeService{
		seqRepo: seqRepo,
	}
}

func (s *BarcodeService) GenerateEAN13(ctx context.Context, t domain.BarcodeType) (string, error) {
	sequence, prefix, err := s.seqRepo.GetNext(ctx, t)
	if err != nil {
		return "", fmt.Errorf("failed to get barcode sequence: %w", err)
	}

	if sequence > 999999999 {
		return "", fmt.Errorf("barcode sequence overflow for prefix %d", prefix)
	}

	base := fmt.Sprintf("%03d%09d", prefix, sequence)

	checkSum := s.calculateEAN13Checksum(base)

	ean13 := base + strconv.Itoa(checkSum)

	return ean13, nil
}

func (s *BarcodeService) calculateEAN13Checksum(first12 string) int {
	sum := 0
	for i, ch := range first12 {
		digit := int(ch - '0')
		// В EAN-13: позиции считаются с 1
		// Нечетные позиции (1,3,5...) умножаются на 1, четные (2,4,6...) на 3
		if i%2 == 0 { // i=0 соответствует позиции 1 (нечетная)
			sum += digit * 1
		} else { // i=1 соответствует позиции 2 (четная)
			sum += digit * 3
		}
	}
	return (10 - (sum % 10)) % 10
}

func (s *BarcodeService) GenerateBarcodeImage(ean13 string) ([]byte, error) {
	if !s.ValidateEAN13(ean13) {
		return nil, fmt.Errorf("invalid EAN-13: %s", ean13)
	}

	code, err := ean.Encode(ean13)
	if err != nil {
		return nil, fmt.Errorf("failed to encode EAN-13: %w", err)
	}

	var barcodeImg barcode.Barcode = code

	scaled, err := barcode.Scale(barcodeImg, 300, 150)
	if err != nil {
		return nil, fmt.Errorf("failed to scale barcode: %w", err)
	}

	var buf bytes.Buffer
	if err := png.Encode(&buf, scaled); err != nil {
		return nil, fmt.Errorf("failed to encode PNG: %w", err)
	}

	return buf.Bytes(), nil
}

func (s *BarcodeService) ValidateEAN13(ean13 string) bool {
	if len(ean13) != 13 {
		return false
	}

	for _, ch := range ean13 {
		if ch < '0' || ch > '9' {
			return false
		}
	}

	first12 := ean13[:12]
	expectedChecksum := s.calculateEAN13Checksum(first12)
	actualChecksum, _ := strconv.Atoi(ean13[12:])

	return expectedChecksum == actualChecksum
}
