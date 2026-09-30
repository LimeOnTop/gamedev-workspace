package usecase

import (
	"context"

	"github.com/LimeOnTop/gamedev-workspace/backend/internal/entity"
)

type AssetRepository interface {
	// GetCategories returns all categories ordered for display, with asset counts.
	GetCategories(ctx context.Context) ([]entity.AssetCategory, error)
	CategoryExists(ctx context.Context, id string) (bool, error)
	GetAssets(ctx context.Context, category string) ([]entity.Asset, error)
}
