package usecase

import (
	"context"

	"github.com/LimeOnTop/gamedev-workspace/backend/internal/entity"
)

type ReferenceRepository interface {
	Create(ctx context.Context, reference entity.Reference) (entity.Reference, error)
	GetByNodeID(ctx context.Context, nodeID string) ([]entity.Reference, error)
	GetByStoredName(ctx context.Context, storedName string) (entity.Reference, error)
	// Delete removes the reference and returns its stored file name.
	Delete(ctx context.Context, id string) (string, error)
}
