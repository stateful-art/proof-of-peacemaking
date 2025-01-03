package storage

import (
	"context"
	"io"
)

// Storage defines the interface for storage operations
type Storage interface {
	UploadFile(ctx context.Context, key string, reader io.Reader, opts ...UploadOptions) error
	GetFile(ctx context.Context, key string) (io.ReadCloser, error)
	DeleteFile(ctx context.Context, key string) error
}
