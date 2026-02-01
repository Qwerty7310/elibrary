package local

import (
	"context"
	"elibrary/internal/storage"
	"errors"
	"io"
	"net/url"
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

	publicURL, err := joinBaseURL(
		s.baseURL,
		string(entity),
		entityID.String(),
		filename,
	)
	if err != nil {
		return "", err
	}

	return publicURL, nil
}

func (s *ImageStorage) Delete(ctx context.Context, url string) error {
	rel, err := stripBaseURL(s.baseURL, url)
	if err != nil {
		return err
	}

	rel = filepath.Clean(rel)
	rel = strings.TrimPrefix(rel, string(filepath.Separator))
	if rel == "." || rel == "" || strings.HasPrefix(rel, ".."+string(filepath.Separator)) || rel == ".." {
		return errors.New("invalid image URL")
	}

	fullPath := filepath.Join(s.basePath, rel)

	return os.Remove(fullPath)
}

func joinBaseURL(base string, parts ...string) (string, error) {
	if base == "" {
		return path.Join(parts...), nil
	}
	if u, err := url.Parse(base); err == nil && u.IsAbs() {
		u.Path = path.Join(append([]string{u.Path}, parts...)...)
		return u.String(), nil
	}
	return path.Join(append([]string{base}, parts...)...), nil
}

func stripBaseURL(base, full string) (string, error) {
	if base == "" {
		return strings.TrimPrefix(full, "/"), nil
	}

	baseURL, err := url.Parse(base)
	if err == nil && baseURL.IsAbs() {
		fullURL, err := url.Parse(full)
		if err != nil || !fullURL.IsAbs() {
			return "", errors.New("invalid image URL")
		}
		if !sameURLBase(baseURL, fullURL) {
			return "", errors.New("invalid image URL")
		}
		basePath := path.Clean(baseURL.Path)
		fullPath := path.Clean(fullURL.Path)
		if !strings.HasPrefix(fullPath, basePath) {
			return "", errors.New("invalid image URL")
		}
		rel := strings.TrimPrefix(fullPath, basePath)
		return strings.TrimPrefix(rel, "/"), nil
	}

	if !strings.HasPrefix(full, base) {
		return "", errors.New("invalid image URL")
	}
	rel := strings.TrimPrefix(full, base)
	return strings.TrimPrefix(rel, "/"), nil
}

func sameURLBase(a, b *url.URL) bool {
	if !strings.EqualFold(a.Scheme, b.Scheme) {
		return false
	}
	if !strings.EqualFold(a.Host, b.Host) {
		return false
	}
	return true
}
