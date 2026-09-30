package filestorage

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/LimeOnTop/gamedev-workspace/backend/internal/apperr"
	"github.com/LimeOnTop/gamedev-workspace/backend/internal/usecase"
)

// Local stores files in a directory on disk (a Docker volume in production).
type Local struct {
	dir string
}

func NewLocal(dir string) (*Local, error) {
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return nil, fmt.Errorf("create upload dir: %w", err)
	}
	return &Local{dir: dir}, nil
}

var _ usecase.FileStorage = (*Local)(nil)

func (l *Local) path(name string) (string, error) {
	if name == "" || name != filepath.Base(name) || name[0] == '.' {
		return "", apperr.ErrNotFound
	}
	return filepath.Join(l.dir, name), nil
}

func (l *Local) Save(_ context.Context, name string, r io.Reader, limit int64) (int64, error) {
	path, err := l.path(name)
	if err != nil {
		return 0, err
	}
	f, err := os.OpenFile(path, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0o644)
	if err != nil {
		return 0, fmt.Errorf("create file: %w", err)
	}
	size, err := io.Copy(f, io.LimitReader(r, limit+1))
	if cerr := f.Close(); err == nil {
		err = cerr
	}
	if err == nil && size > limit {
		err = apperr.ErrTooLarge
	}
	if err != nil {
		os.Remove(path)
		if errors.Is(err, apperr.ErrTooLarge) {
			return 0, err
		}
		return 0, fmt.Errorf("write file: %w", err)
	}
	return size, nil
}

func (l *Local) Open(_ context.Context, name string) (io.ReadSeekCloser, error) {
	path, err := l.path(name)
	if err != nil {
		return nil, err
	}
	f, err := os.Open(path)
	if errors.Is(err, os.ErrNotExist) {
		return nil, apperr.ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("open file: %w", err)
	}
	return f, nil
}

func (l *Local) Remove(_ context.Context, name string) error {
	path, err := l.path(name)
	if err != nil {
		return err
	}
	if err := os.Remove(path); err != nil && !errors.Is(err, os.ErrNotExist) {
		return fmt.Errorf("remove file: %w", err)
	}
	return nil
}
