package repository

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/lib/pq"

	"github.com/LimeOnTop/gamedev-workspace/backend/internal/apperr"
	"github.com/LimeOnTop/gamedev-workspace/backend/internal/entity"
	"github.com/LimeOnTop/gamedev-workspace/backend/internal/usecase"
)

type NodeRepository struct {
	db *sql.DB
}

func NewNodeRepository(db *sql.DB) *NodeRepository {
	return &NodeRepository{db: db}
}

var _ usecase.NodeRepository = (*NodeRepository)(nil)

const nodeColumns = `id, parent_id, kind, name, description, mechanics, characteristics, created_at, updated_at`

func (r *NodeRepository) Create(ctx context.Context, node entity.Node) (entity.Node, error) {
	node.ID = uuid.New().String()
	node.CreatedAt = time.Now()
	node.UpdatedAt = node.CreatedAt

	chars, err := marshalCharacteristics(node.Characteristics)
	if err != nil {
		return entity.Node{}, err
	}

	query := `
		INSERT INTO nodes (id, parent_id, kind, name, description, mechanics, characteristics, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
	`
	_, err = r.db.ExecContext(ctx, query,
		node.ID, node.ParentID, node.Kind, node.Name, node.Description,
		node.Mechanics, chars, node.CreatedAt, node.UpdatedAt,
	)
	if err != nil {
		return entity.Node{}, fmt.Errorf("create node: %w", err)
	}

	return node, nil
}

func (r *NodeRepository) GetByID(ctx context.Context, id string) (entity.Node, error) {
	if !isUUID(id) {
		return entity.Node{}, apperr.ErrNotFound
	}
	row := r.db.QueryRowContext(ctx, `SELECT `+nodeColumns+` FROM nodes WHERE id = $1`, id)
	node, err := scanNode(row)
	if errors.Is(err, sql.ErrNoRows) {
		return entity.Node{}, apperr.ErrNotFound
	}
	if err != nil {
		return entity.Node{}, fmt.Errorf("get node: %w", err)
	}
	return node, nil
}

func (r *NodeRepository) GetAllBrief(ctx context.Context) ([]entity.NodeBrief, error) {
	query := `
		SELECT n.id, n.parent_id, n.kind, n.name, left(n.description, 180),
		       (SELECT ref.stored_name FROM node_references ref
		         WHERE ref.node_id = n.id AND ref.content_type LIKE 'image/%'
		         ORDER BY ref.created_at, ref.id LIMIT 1)
		FROM nodes n
		ORDER BY CASE n.kind WHEN 'folder' THEN 0 ELSE 1 END, lower(n.name), n.created_at
	`
	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("get nodes: %w", err)
	}
	defer rows.Close()

	var nodes []entity.NodeBrief
	for rows.Next() {
		var n entity.NodeBrief
		if err := rows.Scan(&n.ID, &n.ParentID, &n.Kind, &n.Name, &n.Summary, &n.Preview); err != nil {
			return nil, fmt.Errorf("scan node: %w", err)
		}
		nodes = append(nodes, n)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate nodes: %w", err)
	}
	return nodes, nil
}

func (r *NodeRepository) GetPath(ctx context.Context, id string) ([]entity.PathItem, error) {
	query := `
		WITH RECURSIVE up AS (
			SELECT id, parent_id, name, 0 AS depth FROM nodes WHERE id = $1
			UNION ALL
			SELECT n.id, n.parent_id, n.name, up.depth + 1 FROM nodes n JOIN up ON n.id = up.parent_id
		)
		SELECT id, name FROM up WHERE id <> $1 ORDER BY depth DESC
	`
	rows, err := r.db.QueryContext(ctx, query, id)
	if err != nil {
		return nil, fmt.Errorf("get node path: %w", err)
	}
	defer rows.Close()

	path := []entity.PathItem{}
	for rows.Next() {
		var p entity.PathItem
		if err := rows.Scan(&p.ID, &p.Name); err != nil {
			return nil, fmt.Errorf("scan path item: %w", err)
		}
		path = append(path, p)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate path: %w", err)
	}
	return path, nil
}

func (r *NodeRepository) Update(ctx context.Context, node entity.Node) (entity.Node, error) {
	node.UpdatedAt = time.Now()

	chars, err := marshalCharacteristics(node.Characteristics)
	if err != nil {
		return entity.Node{}, err
	}

	query := `
		UPDATE nodes
		SET name = $1, description = $2, mechanics = $3, characteristics = $4, updated_at = $5
		WHERE id = $6
	`
	res, err := r.db.ExecContext(ctx, query,
		node.Name, node.Description, node.Mechanics, chars, node.UpdatedAt, node.ID,
	)
	if err != nil {
		return entity.Node{}, fmt.Errorf("update node: %w", err)
	}
	if n, _ := res.RowsAffected(); n == 0 {
		return entity.Node{}, apperr.ErrNotFound
	}
	return node, nil
}

func (r *NodeRepository) SetParent(ctx context.Context, id string, parentID *string) error {
	res, err := r.db.ExecContext(ctx,
		`UPDATE nodes SET parent_id = $1, updated_at = now() WHERE id = $2`, parentID, id)
	if err != nil {
		return fmt.Errorf("move node: %w", err)
	}
	if n, _ := res.RowsAffected(); n == 0 {
		return apperr.ErrNotFound
	}
	return nil
}

func (r *NodeRepository) IsDescendant(ctx context.Context, id, candidate string) (bool, error) {
	query := `
		WITH RECURSIVE up AS (
			SELECT id, parent_id FROM nodes WHERE id = $1
			UNION ALL
			SELECT n.id, n.parent_id FROM nodes n JOIN up ON n.id = up.parent_id
		)
		SELECT EXISTS (SELECT 1 FROM up WHERE id = $2)
	`
	var found bool
	if err := r.db.QueryRowContext(ctx, query, candidate, id).Scan(&found); err != nil {
		return false, fmt.Errorf("check node ancestry: %w", err)
	}
	return found, nil
}

func (r *NodeRepository) Delete(ctx context.Context, id string) ([]string, error) {
	if !isUUID(id) {
		return nil, apperr.ErrNotFound
	}
	// Data-modifying CTEs see the pre-delete snapshot, so the subtree's
	// references are collected in the same statement that removes them.
	query := `
		WITH RECURSIVE sub AS (
			SELECT id FROM nodes WHERE id = $1
			UNION ALL
			SELECT n.id FROM nodes n JOIN sub ON n.parent_id = sub.id
		),
		files AS (
			SELECT ref.stored_name FROM node_references ref JOIN sub ON ref.node_id = sub.id
		),
		deleted AS (
			DELETE FROM nodes WHERE id = $1 RETURNING id
		)
		SELECT (SELECT count(*) FROM deleted),
		       COALESCE((SELECT array_agg(stored_name) FROM files), '{}')
	`
	var deleted int
	var files []string
	if err := r.db.QueryRowContext(ctx, query, id).Scan(&deleted, pq.Array(&files)); err != nil {
		return nil, fmt.Errorf("delete node: %w", err)
	}
	if deleted == 0 {
		return nil, apperr.ErrNotFound
	}
	return files, nil
}

func (r *NodeRepository) Search(ctx context.Context, query string, limit int64) ([]entity.NodeBrief, error) {
	searchQuery := `
		SELECT id, parent_id, kind, name, left(description, 180)
		FROM nodes
		WHERE name ILIKE $1 OR description ILIKE $1 OR mechanics ILIKE $1 OR characteristics::text ILIKE $1
		ORDER BY (name ILIKE $1) DESC, lower(name)
		LIMIT $2
	`
	pattern := "%" + strings.NewReplacer(`\`, `\\`, "%", `\%`, "_", `\_`).Replace(query) + "%"
	rows, err := r.db.QueryContext(ctx, searchQuery, pattern, limit)
	if err != nil {
		return nil, fmt.Errorf("search nodes: %w", err)
	}
	defer rows.Close()

	var nodes []entity.NodeBrief
	for rows.Next() {
		var n entity.NodeBrief
		if err := rows.Scan(&n.ID, &n.ParentID, &n.Kind, &n.Name, &n.Summary); err != nil {
			return nil, fmt.Errorf("scan node: %w", err)
		}
		nodes = append(nodes, n)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate nodes: %w", err)
	}
	return nodes, nil
}

func scanNode(row *sql.Row) (entity.Node, error) {
	var node entity.Node
	var chars []byte
	if err := row.Scan(
		&node.ID, &node.ParentID, &node.Kind, &node.Name, &node.Description,
		&node.Mechanics, &chars, &node.CreatedAt, &node.UpdatedAt,
	); err != nil {
		return entity.Node{}, err
	}
	if err := json.Unmarshal(chars, &node.Characteristics); err != nil {
		return entity.Node{}, fmt.Errorf("decode characteristics: %w", err)
	}
	return node, nil
}

func marshalCharacteristics(chars []entity.Characteristic) ([]byte, error) {
	if chars == nil {
		chars = []entity.Characteristic{}
	}
	b, err := json.Marshal(chars)
	if err != nil {
		return nil, fmt.Errorf("encode characteristics: %w", err)
	}
	return b, nil
}

func isUUID(s string) bool {
	_, err := uuid.Parse(s)
	return err == nil
}
