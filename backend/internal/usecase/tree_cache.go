package usecase

import "context"

// TreeCache keeps the rendered folder tree between requests.
// Implementations must treat failures as cache misses where possible.
type TreeCache interface {
	Get(ctx context.Context) (tree []TreeNodeDTO, found bool, err error)
	Set(ctx context.Context, tree []TreeNodeDTO) error
	Invalidate(ctx context.Context) error
}
