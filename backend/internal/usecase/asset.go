package usecase

import "context"

// Asset is the catalog of files that describe objects with a 3D model.
type Asset interface {
	Categories(ctx context.Context) ([]AssetCategoryDTO, error)
	// List returns assets of one category, or of all categories when category is "".
	List(ctx context.Context, category string) ([]AssetDTO, error)
}
