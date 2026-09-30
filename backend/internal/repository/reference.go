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

type ReferenceRepository struct {
	db *sql.DB
}

func NewReferenceRepository(db *sql.DB) *ReferenceRepository {
	return &ReferenceRepository{db: db}
}

var _ usecase.ReferenceRepository = (*ReferenceRepository)(nil)

const referenceColumns = `id, node_id, original_name, stored_name, content_type, size_bytes, created_at`

func (r *ReferenceRepository) Create(ctx context.Context, reference entity.Reference) (entity.Reference, error) {
	reference.ID = uuid.New().String()
	reference.CreatedAt = time.Now()

	query := `
		INSERT INTO node_references (id, node_id, original_name, stored_name, content_type, size_bytes, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
	`
	_, err := r.db.ExecContext(ctx, query,
		reference.ID, reference.NodeID, reference.OriginalName, reference.StoredName,
		reference.ContentType, reference.Size, reference.CreatedAt,
	)
	if err != nil {
		return entity.Reference{}, fmt.Errorf("create reference: %w", err)
	}
	return reference, nil
}

func (r *ReferenceRepository) GetByNodeID(ctx context.Context, nodeID string) ([]entity.Reference, error) {
	rows, err := r.db.QueryContext(ctx,
		`SELECT `+referenceColumns+` FROM node_references WHERE node_id = $1 ORDER BY created_at, id`, nodeID)
	if err != nil {
		return nil, fmt.Errorf("get references: %w", err)
	}
	defer rows.Close()

	references := []entity.Reference{}
	for rows.Next() {
		var ref entity.Reference
		if err := rows.Scan(
			&ref.ID, &ref.NodeID, &ref.OriginalName, &ref.StoredName,
			&ref.ContentType, &ref.Size, &ref.CreatedAt,
		); err != nil {
			return nil, fmt.Errorf("scan reference: %w", err)
		}
		references = append(references, ref)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate references: %w", err)
	}
	return references, nil
}

func (r *ReferenceRepository) GetByStoredName(ctx context.Context, storedName string) (entity.Reference, error) {
	var ref entity.Reference
	err := r.db.QueryRowContext(ctx,
		`SELECT `+referenceColumns+` FROM node_references WHERE stored_name = $1`, storedName).
		Scan(&ref.ID, &ref.NodeID, &ref.OriginalName, &ref.StoredName, &ref.ContentType, &ref.Size, &ref.CreatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return entity.Reference{}, apperr.ErrNotFound
	}
	if err != nil {
		return entity.Reference{}, fmt.Errorf("get reference: %w", err)
	}
	return ref, nil
}

func (r *ReferenceRepository) Delete(ctx context.Context, id string) (string, error) {
	if !isUUID(id) {
		return "", apperr.ErrNotFound
	}
	var storedName string
	err := r.db.QueryRowContext(ctx,
		`DELETE FROM node_references WHERE id = $1 RETURNING stored_name`, id).Scan(&storedName)
	if errors.Is(err, sql.ErrNoRows) {
		return "", apperr.ErrNotFound
	}
	if err != nil {
		return "", fmt.Errorf("delete reference: %w", err)
	}
	return storedName, nil
}
