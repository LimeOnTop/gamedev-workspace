package usecase

import (
	"context"

	"github.com/LimeOnTop/gamedev-workspace/backend/internal/entity"
)

type NodeRepository interface {
	Create(ctx context.Context, node entity.Node) (entity.Node, error)
	GetByID(ctx context.Context, id string) (entity.Node, error)
	GetAllBrief(ctx context.Context) ([]entity.NodeBrief, error)
	GetPath(ctx context.Context, id string) ([]entity.PathItem, error)
	Update(ctx context.Context, node entity.Node) (entity.Node, error)
	SetParent(ctx context.Context, id string, parentID *string) error
	// IsDescendant reports whether candidate is id itself or lies in its subtree.
	IsDescendant(ctx context.Context, id, candidate string) (bool, error)
	// Delete removes the node with its subtree and returns the stored file
	// names of every reference and model that belonged to the removed nodes.
	Delete(ctx context.Context, id string) ([]string, error)
	Search(ctx context.Context, query string, limit int64) ([]entity.NodeBrief, error)
}
