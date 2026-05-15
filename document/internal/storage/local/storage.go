package local

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"strings"

	"github.com/ossirdian/magistral/document/internal/storage"
)

type Storage struct {
	root string
}

func New(root string) *Storage {
	root = strings.TrimSpace(root)
	if root == "" {
		return nil
	}
	return &Storage{root: root}
}

func (s *Storage) Save(_ context.Context, objectKey string, content []byte) error {
	path, err := s.path(objectKey)
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	return os.WriteFile(path, content, 0o644)
}

func (s *Storage) Load(_ context.Context, objectKey string) ([]byte, error) {
	path, err := s.path(objectKey)
	if err != nil {
		return nil, err
	}
	data, err := os.ReadFile(path)
	if errors.Is(err, os.ErrNotExist) {
		return nil, storage.ErrNotFound
	}
	return data, err
}

func (s *Storage) Delete(_ context.Context, objectKey string) error {
	path, err := s.path(objectKey)
	if err != nil {
		return err
	}
	if err := os.Remove(path); err != nil && !errors.Is(err, os.ErrNotExist) {
		return err
	}
	return nil
}

func (s *Storage) path(objectKey string) (string, error) {
	cleanKey := filepath.Clean(filepath.FromSlash(strings.TrimSpace(objectKey)))
	path := filepath.Join(s.root, cleanKey)
	rootClean := filepath.Clean(s.root)
	if path != rootClean && !strings.HasPrefix(path, rootClean+string(filepath.Separator)) {
		return "", errors.New("object key escapes storage root")
	}
	return path, nil
}
