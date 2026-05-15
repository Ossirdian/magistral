package storage

import (
	"context"
	"errors"
)

var ErrNotFound = errors.New("storage object not found")

type Storage interface {
	Save(ctx context.Context, objectKey string, content []byte) error
	Load(ctx context.Context, objectKey string) ([]byte, error)
	Delete(ctx context.Context, objectKey string) error
}
