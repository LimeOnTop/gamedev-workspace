package repository

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/LimeOnTop/gamedev-workspace/backend/internal/entity"
	"github.com/LimeOnTop/gamedev-workspace/backend/internal/usecase"
)

type AssetRepository struct {
	db *sql.DB
}

func NewAssetRepository(db *sql.DB) *AssetRepository {
	return &AssetRepository{db: db}
}

var _ usecase.AssetRepository = (*AssetRepository)(nil)

func (r *AssetRepository) GetCategories(ctx context.Context) ([]entity.AssetCategory, error) {
	query := `
		SELECT c.id, c.name, c.description, c.sort_order,
		       (SELECT count(*) FROM nodes n WHERE n.asset_category = c.id AND n.kind = 'file')
		FROM asset_categories c
		ORDER BY c.sort_order, c.name
	`
	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("get asset categories: %w", err)
	}
	defer rows.Close()

	var categories []entity.AssetCategory
	for rows.Next() {
		var c entity.AssetCategory
		if err := rows.Scan(&c.ID, &c.Name, &c.Description, &c.SortOrder, &c.AssetCount); err != nil {
			return nil, fmt.Errorf("scan asset category: %w", err)
		}
		categories = append(categories, c)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate asset categories: %w", err)
	}
	return categories, nil
}

func (r *AssetRepository) CategoryExists(ctx context.Context, id string) (bool, error) {
	var exists bool
	if err := r.db.QueryRowContext(ctx, `SELECT EXISTS (SELECT 1 FROM asset_categories WHERE id = $1)`, id).Scan(&exists); err != nil {
		return false, fmt.Errorf("check asset category: %w", err)
	}
	return exists, nil
}

func (r *AssetRepository) GetAssets(ctx context.Context, category string) ([]entity.Asset, error) {
	// The path is aggregated per asset by walking up the tree, so the whole
	// catalog page is served by a single query.
	query := `
		WITH RECURSIVE assets AS (
			SELECT id, parent_id, asset_category, name, left(description, 180) AS summary
			FROM nodes
			WHERE kind = 'file' AND asset_category IS NOT NULL
			  AND ($1 = '' OR asset_category = $1)
		),
		up AS (
			SELECT a.id AS asset_id, n.id, n.parent_id, n.name, 1 AS depth
			FROM assets a JOIN nodes n ON n.id = a.parent_id
			UNION ALL
			SELECT up.asset_id, n.id, n.parent_id, n.name, up.depth + 1
			FROM up JOIN nodes n ON n.id = up.parent_id
		)
		SELECT a.id, a.parent_id, a.asset_category, a.name, a.summary,
		       (SELECT ref.stored_name FROM node_references ref
		         WHERE ref.node_id = a.id AND ref.content_type LIKE 'image/%'
		         ORDER BY ref.created_at, ref.id LIMIT 1),
		       COALESCE((SELECT json_agg(json_build_object('id', up.id, 'name', up.name) ORDER BY up.depth DESC)
		                   FROM up WHERE up.asset_id = a.id), '[]')
		FROM assets a
		ORDER BY lower(a.name)
	`
	rows, err := r.db.QueryContext(ctx, query, category)
	if err != nil {
		return nil, fmt.Errorf("get assets: %w", err)
	}
	defer rows.Close()

	var assets []entity.Asset
	for rows.Next() {
		var a entity.Asset
		var path []byte
		if err := rows.Scan(&a.ID, &a.ParentID, &a.Category, &a.Name, &a.Summary, &a.Preview, &path); err != nil {
			return nil, fmt.Errorf("scan asset: %w", err)
		}
		if a.Path, err = decodePath(path); err != nil {
			return nil, err
		}
		assets = append(assets, a)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate assets: %w", err)
	}
	return assets, nil
}
