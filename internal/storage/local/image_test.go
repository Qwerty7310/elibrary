package local

import (
	"bytes"
	"context"
	"errors"
	"os"
	"path/filepath"
	"testing"

	"elibrary/internal/storage"

	"github.com/google/uuid"
)

func TestImageStorageSaveAndDeleteWithAbsoluteBaseURL(t *testing.T) {
	t.Parallel()

	basePath := t.TempDir()
	entityID := uuid.New()
	s := NewImageStorage(basePath, "https://cdn.example.com/static")

	gotURL, err := s.Save(context.Background(), storage.Book, entityID, bytes.NewBufferString("cover"))
	if err != nil {
		t.Fatalf("Save() error = %v", err)
	}

	wantURL := "https://cdn.example.com/static/book/" + entityID.String() + "/photo.jpg"
	if gotURL != wantURL {
		t.Fatalf("Save() URL = %q, want %q", gotURL, wantURL)
	}

	fullPath := filepath.Join(basePath, "book", entityID.String(), "photo.jpg")
	data, err := os.ReadFile(fullPath)
	if err != nil {
		t.Fatalf("ReadFile() error = %v", err)
	}
	if string(data) != "cover" {
		t.Fatalf("saved file content = %q, want %q", string(data), "cover")
	}

	if err := s.Delete(context.Background(), gotURL); err != nil {
		t.Fatalf("Delete() error = %v", err)
	}

	if _, err := os.Stat(fullPath); !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("expected file to be removed, stat error = %v", err)
	}
}

func TestImageStorageSaveWithRelativeBaseURL(t *testing.T) {
	t.Parallel()

	basePath := t.TempDir()
	entityID := uuid.New()
	s := NewImageStorage(basePath, "/images")

	gotURL, err := s.Save(context.Background(), storage.Author, entityID, bytes.NewBufferString("avatar"))
	if err != nil {
		t.Fatalf("Save() error = %v", err)
	}

	wantURL := "/images/author/" + entityID.String() + "/photo.jpg"
	if gotURL != wantURL {
		t.Fatalf("Save() URL = %q, want %q", gotURL, wantURL)
	}
}

func TestImageStorageDeleteRejectsTraversal(t *testing.T) {
	t.Parallel()

	s := NewImageStorage(t.TempDir(), "/images")

	err := s.Delete(context.Background(), "/images/../../etc/passwd")
	if err == nil {
		t.Fatal("Delete() error = nil, want invalid image URL")
	}
	if err.Error() != "invalid image URL" {
		t.Fatalf("Delete() error = %q, want %q", err.Error(), "invalid image URL")
	}
}

func TestImageStorageDeleteRejectsDifferentHost(t *testing.T) {
	t.Parallel()

	s := NewImageStorage(t.TempDir(), "https://cdn.example.com/static")

	err := s.Delete(context.Background(), "https://evil.example.com/static/book/id/photo.jpg")
	if err == nil {
		t.Fatal("Delete() error = nil, want invalid image URL")
	}
	if err.Error() != "invalid image URL" {
		t.Fatalf("Delete() error = %q, want %q", err.Error(), "invalid image URL")
	}
}

func TestJoinBaseURL(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		base    string
		parts   []string
		want    string
		wantErr bool
	}{
		{
			name:  "empty base",
			base:  "",
			parts: []string{"book", "123", "photo.jpg"},
			want:  "book/123/photo.jpg",
		},
		{
			name:  "relative base",
			base:  "/images",
			parts: []string{"book", "123", "photo.jpg"},
			want:  "/images/book/123/photo.jpg",
		},
		{
			name:  "absolute base",
			base:  "https://cdn.example.com/static",
			parts: []string{"book", "123", "photo.jpg"},
			want:  "https://cdn.example.com/static/book/123/photo.jpg",
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got, err := joinBaseURL(tt.base, tt.parts...)
			if (err != nil) != tt.wantErr {
				t.Fatalf("joinBaseURL() error = %v, wantErr %v", err, tt.wantErr)
			}
			if got != tt.want {
				t.Fatalf("joinBaseURL() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestStripBaseURL(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		base    string
		full    string
		want    string
		wantErr bool
	}{
		{
			name: "empty base",
			base: "",
			full: "/book/123/photo.jpg",
			want: "book/123/photo.jpg",
		},
		{
			name: "relative base",
			base: "/images",
			full: "/images/book/123/photo.jpg",
			want: "book/123/photo.jpg",
		},
		{
			name: "absolute base",
			base: "https://cdn.example.com/static",
			full: "https://cdn.example.com/static/book/123/photo.jpg",
			want: "book/123/photo.jpg",
		},
		{
			name:    "absolute invalid host",
			base:    "https://cdn.example.com/static",
			full:    "https://other.example.com/static/book/123/photo.jpg",
			wantErr: true,
		},
		{
			name:    "relative invalid prefix",
			base:    "/images",
			full:    "/files/book/123/photo.jpg",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got, err := stripBaseURL(tt.base, tt.full)
			if (err != nil) != tt.wantErr {
				t.Fatalf("stripBaseURL() error = %v, wantErr %v", err, tt.wantErr)
			}
			if got != tt.want {
				t.Fatalf("stripBaseURL() = %q, want %q", got, tt.want)
			}
		})
	}
}
