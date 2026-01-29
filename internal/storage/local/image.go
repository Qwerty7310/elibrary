package local

import (
	"context"
	"elibrary/internal/storage"
	"errors"
	"io"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/google/uuid"
)

type ImageStorage struct {
	basePath string
	baseURL  string
}

func NewImageStorage(basePath string, baseURL string) *ImageStorage {
	return &ImageStorage{
		basePath: basePath,
		baseURL:  baseURL,
	}
}

func (s *ImageStorage) Save(ctx context.Context, entity storage.EntityType, entityID uuid.UUID, file io.Reader) (string, error) {
	id := uuid.New().String()

	dir := filepath.Join(s.basePath, string(entity), entityID.String())

	if err := os.MkdirAll(dir, 0755); err != nil {
		return "", err
	}

	filename := id + ".jpg"
	fullPath := filepath.Join(dir, filename)

	out, err := os.Create(fullPath)
	if err != nil {
		return "", err
	}
	defer out.Close()

	if _, err := io.Copy(out, file); err != nil {
		return "", err
	}

	publicURL := path.Join(
		s.baseURL,
		string(entity),
		entityID.String(),
		filename,
	)

	return publicURL, nil
}

func (s *ImageStorage) Delete(ctx context.Context, url string) error {
	if !strings.HasPrefix(url, s.baseURL) {
		return errors.New("invalid image URL")
	}

	rel := strings.TrimPrefix(url, s.baseURL)
	fullPath := filepath.Join(s.basePath, rel)

	return os.Remove(fullPath)
}
