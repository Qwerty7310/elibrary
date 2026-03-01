package service

import (
	"bytes"
	"context"
	"errors"
	"image/png"
	"strings"
	"testing"

	"elibrary/internal/domain"
)

type stubSequenceRepo struct {
	getNext func(ctx context.Context, t domain.BarcodeType) (int64, int, error)
}

func (s stubSequenceRepo) GetNext(ctx context.Context, t domain.BarcodeType) (int64, int, error) {
	return s.getNext(ctx, t)
}

func (s stubSequenceRepo) SetType(ctx context.Context, t domain.BarcodeType, prefix int, description string) error {
	return nil
}

func TestBarcodeServiceGenerateEAN13(t *testing.T) {
	t.Parallel()

	service := NewBarcodeService(stubSequenceRepo{
		getNext: func(ctx context.Context, barcodeType domain.BarcodeType) (int64, int, error) {
			if barcodeType != domain.BarcodeTypeBook {
				t.Fatalf("GetNext() type = %q, want %q", barcodeType, domain.BarcodeTypeBook)
			}
			return 123, 978, nil
		},
	})

	got, err := service.GenerateEAN13(context.Background(), domain.BarcodeTypeBook)
	if err != nil {
		t.Fatalf("GenerateEAN13() error = %v", err)
	}

	want := "9780000001238"
	if got != want {
		t.Fatalf("GenerateEAN13() = %q, want %q", got, want)
	}
	if !service.ValidateEAN13(got) {
		t.Fatalf("GenerateEAN13() returned invalid code %q", got)
	}
}

func TestBarcodeServiceGenerateEAN13PropagatesRepositoryError(t *testing.T) {
	t.Parallel()

	service := NewBarcodeService(stubSequenceRepo{
		getNext: func(ctx context.Context, t domain.BarcodeType) (int64, int, error) {
			return 0, 0, errors.New("db down")
		},
	})

	_, err := service.GenerateEAN13(context.Background(), domain.BarcodeTypeBook)
	if err == nil {
		t.Fatal("GenerateEAN13() error = nil, want repository error")
	}
	if !strings.Contains(err.Error(), "failed to get barcode sequence") {
		t.Fatalf("GenerateEAN13() error = %q, want wrapped repository error", err.Error())
	}
}

func TestBarcodeServiceGenerateEAN13RejectsOverflow(t *testing.T) {
	t.Parallel()

	service := NewBarcodeService(stubSequenceRepo{
		getNext: func(ctx context.Context, t domain.BarcodeType) (int64, int, error) {
			return 1000000000, 978, nil
		},
	})

	_, err := service.GenerateEAN13(context.Background(), domain.BarcodeTypeBook)
	if err == nil {
		t.Fatal("GenerateEAN13() error = nil, want overflow error")
	}
	if !strings.Contains(err.Error(), "overflow") {
		t.Fatalf("GenerateEAN13() error = %q, want overflow error", err.Error())
	}
}

func TestBarcodeServiceValidateEAN13(t *testing.T) {
	t.Parallel()

	service := NewBarcodeService(nil)

	tests := []struct {
		name string
		in   string
		want bool
	}{
		{name: "valid", in: "9780000001238", want: true},
		{name: "wrong checksum", in: "9780000001235", want: false},
		{name: "non digit", in: "97800000012A8", want: false},
		{name: "wrong length", in: "978000000123", want: false},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			if got := service.ValidateEAN13(tt.in); got != tt.want {
				t.Fatalf("ValidateEAN13(%q) = %v, want %v", tt.in, got, tt.want)
			}
		})
	}
}

func TestBarcodeServiceGenerateBarcodeImage(t *testing.T) {
	t.Parallel()

	service := NewBarcodeService(nil)

	data, err := service.GenerateBarcodeImage("9780000001238")
	if err != nil {
		t.Fatalf("GenerateBarcodeImage() error = %v", err)
	}
	if len(data) == 0 {
		t.Fatal("GenerateBarcodeImage() returned empty data")
	}

	cfg, err := png.DecodeConfig(bytes.NewReader(data))
	if err != nil {
		t.Fatalf("DecodeConfig() error = %v", err)
	}
	if cfg.Width != 300 || cfg.Height != 150 {
		t.Fatalf("barcode image size = %dx%d, want %dx%d", cfg.Width, cfg.Height, 300, 150)
	}
}

func TestBarcodeServiceGenerateBarcodeImageRejectsInvalidCode(t *testing.T) {
	t.Parallel()

	service := NewBarcodeService(nil)

	_, err := service.GenerateBarcodeImage("123")
	if err == nil {
		t.Fatal("GenerateBarcodeImage() error = nil, want invalid EAN-13 error")
	}
	if !strings.Contains(err.Error(), "invalid EAN-13") {
		t.Fatalf("GenerateBarcodeImage() error = %q, want invalid EAN-13 error", err.Error())
	}
}
