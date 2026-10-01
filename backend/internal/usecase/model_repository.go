package usecase

import (
	"context"

	"github.com/LimeOnTop/gamedev-workspace/backend/internal/entity"
)

type ModelRepository interface {
	// Set stores the node's model and returns the stored file name of the
	// model it replaced, or "" when the node had none.
	Set(ctx context.Context, model entity.Model) (entity.Model, string, error)
	// GetByNodeID returns apperr.ErrNotFound when the node has no model.
	GetByNodeID(ctx context.Context, nodeID string) (entity.Model, error)
	GetByStoredName(ctx context.Context, storedName string) (entity.Model, error)
	// Delete removes the node's model and returns its stored file name.
	Delete(ctx context.Context, nodeID string) (string, error)
}
