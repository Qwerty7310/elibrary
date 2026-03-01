package service

import (
	"bytes"
	"context"
	"errors"
	"io"
	"testing"

	"elibrary/internal/storage"

	"github.com/google/uuid"
)

type stubImageStorage struct {
	saveFunc   func(ctx context.Context, entity storage.EntityType, entityID uuid.UUID, file io.Reader) (string, error)
	deleteFunc func(ctx context.Context, url string) error
}

func (s stubImageStorage) Save(ctx context.Context, entity storage.EntityType, entityID uuid.UUID, file io.Reader) (string, error) {
	return s.saveFunc(ctx, entity, entityID, file)
}

func (s stubImageStorage) Delete(ctx context.Context, url string) error {
	if s.deleteFunc == nil {
		return nil
	}
	return s.deleteFunc(ctx, url)
}

func TestImageServiceUpload(t *testing.T) {
	t.Parallel()

	entityID := uuid.New()
	called := false
	svc := NewImageService(stubImageStorage{
		saveFunc: func(ctx context.Context, entity storage.EntityType, gotID uuid.UUID, file io.Reader) (string, error) {
			called = true
			if entity != storage.Book || gotID != entityID {
				t.Fatalf("Save() args = (%q, %v), want (%q, %v)", entity, gotID, storage.Book, entityID)
			}
			data, _ := io.ReadAll(file)
			if string(data) != "data" {
				t.Fatalf("file data = %q, want %q", string(data), "data")
			}
			return "/img", nil
		},
	})

	got, err := svc.Upload(context.Background(), storage.Book, entityID, bytes.NewBufferString("data"))
	if err != nil {
		t.Fatalf("Upload() error = %v", err)
	}
	if !called {
		t.Fatal("Save() was not called")
	}
	if got != "/img" {
		t.Fatalf("Upload() = %q, want %q", got, "/img")
	}
}

func TestImageServiceReplace(t *testing.T) {
	t.Parallel()

	entityID := uuid.New()
	oldURL := "/old"
	deleted := false
	svc := NewImageService(stubImageStorage{
		saveFunc: func(ctx context.Context, entity storage.EntityType, gotID uuid.UUID, file io.Reader) (string, error) {
			return "/new", nil
		},
		deleteFunc: func(ctx context.Context, url string) error {
			deleted = true
			if url != oldURL {
				t.Fatalf("Delete() url = %q, want %q", url, oldURL)
			}
			return errors.New("ignored")
		},
	})

	got, err := svc.Replace(context.Background(), storage.Author, entityID, &oldURL, bytes.NewBufferString("data"))
	if err != nil {
		t.Fatalf("Replace() error = %v", err)
	}
	if !deleted {
		t.Fatal("Delete() was not called")
	}
	if got != "/new" {
		t.Fatalf("Replace() = %q, want %q", got, "/new")
	}
}

func TestImageServiceReplaceWithoutOldURL(t *testing.T) {
	t.Parallel()

	deleteCalled := false
	svc := NewImageService(stubImageStorage{
		saveFunc: func(ctx context.Context, entity storage.EntityType, gotID uuid.UUID, file io.Reader) (string, error) {
			return "/new", nil
		},
		deleteFunc: func(ctx context.Context, url string) error {
			deleteCalled = true
			return nil
		},
	})

	empty := ""
	if _, err := svc.Replace(context.Background(), storage.Author, uuid.New(), nil, bytes.NewBuffer(nil)); err != nil {
		t.Fatalf("Replace() with nil oldURL error = %v", err)
	}
	if _, err := svc.Replace(context.Background(), storage.Author, uuid.New(), &empty, bytes.NewBuffer(nil)); err != nil {
		t.Fatalf("Replace() with empty oldURL error = %v", err)
	}
	if deleteCalled {
		t.Fatal("Delete() should not be called for nil or empty oldURL")
	}
}
