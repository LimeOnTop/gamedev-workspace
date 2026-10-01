package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"

	"github.com/LimeOnTop/gamedev-workspace/backend/internal/apperr"
	"github.com/LimeOnTop/gamedev-workspace/backend/internal/entity"
	"github.com/LimeOnTop/gamedev-workspace/backend/internal/usecase"
)

type ModelRepository struct {
	db *sql.DB
}

func NewModelRepository(db *sql.DB) *ModelRepository {
	return &ModelRepository{db: db}
}

var _ usecase.ModelRepository = (*ModelRepository)(nil)

const modelColumns = `id, node_id, original_name, stored_name, content_type, size_bytes, created_at`

func (r *ModelRepository) Set(ctx context.Context, model entity.Model) (entity.Model, string, error) {
	model.ID = uuid.New().String()
	model.CreatedAt = time.Now()

	// The old row is read from the pre-statement snapshot, so the replaced
	// file name comes back from the same statement that overwrites it.
	query := `
		WITH old AS (
			SELECT stored_name FROM node_models WHERE node_id = $2
		),
		upsert AS (
			INSERT INTO node_models (id, node_id, original_name, stored_name, content_type, size_bytes, created_at)
			VALUES ($1, $2, $3, $4, $5, $6, $7)
			ON CONFLICT (node_id) DO UPDATE
			SET id = EXCLUDED.id, original_name = EXCLUDED.original_name, stored_name = EXCLUDED.stored_name,
			    content_type = EXCLUDED.content_type, size_bytes = EXCLUDED.size_bytes, created_at = EXCLUDED.created_at
			RETURNING id
		)
		SELECT COALESCE((SELECT stored_name FROM old), '') FROM upsert
	`
	var replaced string
	err := r.db.QueryRowContext(ctx, query,
		model.ID, model.NodeID, model.OriginalName, model.StoredName,
		model.ContentType, model.Size, model.CreatedAt,
	).Scan(&replaced)
	if err != nil {
		return entity.Model{}, "", fmt.Errorf("set model: %w", err)
	}
	return model, replaced, nil
}

func (r *ModelRepository) GetByNodeID(ctx context.Context, nodeID string) (entity.Model, error) {
	if !isUUID(nodeID) {
		return entity.Model{}, apperr.ErrNotFound
	}
	return r.getOne(ctx, `SELECT `+modelColumns+` FROM node_models WHERE node_id = $1`, nodeID)
}

func (r *ModelRepository) GetByStoredName(ctx context.Context, storedName string) (entity.Model, error) {
	return r.getOne(ctx, `SELECT `+modelColumns+` FROM node_models WHERE stored_name = $1`, storedName)
}

func (r *ModelRepository) Delete(ctx context.Context, nodeID string) (string, error) {
	if !isUUID(nodeID) {
		return "", apperr.ErrNotFound
	}
	var storedName string
	err := r.db.QueryRowContext(ctx,
		`DELETE FROM node_models WHERE node_id = $1 RETURNING stored_name`, nodeID).Scan(&storedName)
	if errors.Is(err, sql.ErrNoRows) {
		return "", apperr.ErrNotFound
	}
	if err != nil {
		return "", fmt.Errorf("delete model: %w", err)
	}
	return storedName, nil
}

func (r *ModelRepository) getOne(ctx context.Context, query string, arg string) (entity.Model, error) {
	var m entity.Model
	err := r.db.QueryRowContext(ctx, query, arg).
		Scan(&m.ID, &m.NodeID, &m.OriginalName, &m.StoredName, &m.ContentType, &m.Size, &m.CreatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return entity.Model{}, apperr.ErrNotFound
	}
	if err != nil {
		return entity.Model{}, fmt.Errorf("get model: %w", err)
	}
	return m, nil
}
