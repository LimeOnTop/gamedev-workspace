package usecase

import (
	"context"
	"io"
)

// FileStorage stores reference binaries.
type FileStorage interface {
	// Save writes at most limit bytes of r under name and returns the written size.
	// It returns apperr.ErrTooLarge (and removes the partial file) when r is longer.
	Save(ctx context.Context, name string, r io.Reader, limit int64) (int64, error)
	Open(ctx context.Context, name string) (io.ReadSeekCloser, error)
	Remove(ctx context.Context, name string) error
}
