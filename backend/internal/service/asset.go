package service

import (
	"context"
	"fmt"
	"strings"

	"github.com/LimeOnTop/gamedev-workspace/backend/internal/apperr"
	"github.com/LimeOnTop/gamedev-workspace/backend/internal/usecase"
)

type AssetService struct {
	assets usecase.AssetRepository
}

func NewAssetService(assets usecase.AssetRepository) *AssetService {
	return &AssetService{assets: assets}
}

var _ usecase.Asset = (*AssetService)(nil)

func (s *AssetService) Categories(ctx context.Context) ([]usecase.AssetCategoryDTO, error) {
	categories, err := s.assets.GetCategories(ctx)
	if err != nil {
		return nil, fmt.Errorf("get asset categories: %w", err)
	}
	result := make([]usecase.AssetCategoryDTO, 0, len(categories))
	for _, c := range categories {
		result = append(result, usecase.AssetCategoryDTO{
			ID: c.ID, Name: c.Name, Description: c.Description, AssetCount: c.AssetCount,
		})
	}
	return result, nil
}

func (s *AssetService) List(ctx context.Context, category string) ([]usecase.AssetDTO, error) {
	category = strings.TrimSpace(category)
	if category != "" {
		exists, err := s.assets.CategoryExists(ctx, category)
		if err != nil {
			return nil, err
		}
		if !exists {
			return nil, apperr.ErrNotFound
		}
	}

	assets, err := s.assets.GetAssets(ctx, category)
	if err != nil {
		return nil, fmt.Errorf("get assets: %w", err)
	}
	result := make([]usecase.AssetDTO, 0, len(assets))
	for _, a := range assets {
		result = append(result, usecase.AssetDTO{
			ID: a.ID, ParentID: a.ParentID, Category: a.Category, Name: a.Name,
			Summary: a.Summary, Preview: a.Preview, HasModel: a.HasModel, Path: a.Path,
		})
	}
	return result, nil
}
