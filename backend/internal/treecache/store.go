package treecache

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"

	"github.com/LimeOnTop/gamedev-workspace/backend/internal/usecase"
)

const treeKey = "workspace:tree:v1"

// Store caches the folder tree in Redis.
type Store struct {
	client *redis.Client
	ttl    time.Duration
}

func NewStore(client *redis.Client, ttl time.Duration) *Store {
	if ttl <= 0 {
		ttl = 10 * time.Minute
	}
	return &Store{client: client, ttl: ttl}
}

var _ usecase.TreeCache = (*Store)(nil)

func (s *Store) Get(ctx context.Context) ([]usecase.TreeNodeDTO, bool, error) {
	raw, err := s.client.Get(ctx, treeKey).Bytes()
	if errors.Is(err, redis.Nil) {
		return nil, false, nil
	}
	if err != nil {
		return nil, false, fmt.Errorf("tree cache get: %w", err)
	}
	var tree []usecase.TreeNodeDTO
	if err := json.Unmarshal(raw, &tree); err != nil {
		return nil, false, fmt.Errorf("tree cache decode: %w", err)
	}
	return tree, true, nil
}

func (s *Store) Set(ctx context.Context, tree []usecase.TreeNodeDTO) error {
	raw, err := json.Marshal(tree)
	if err != nil {
		return fmt.Errorf("tree cache encode: %w", err)
	}
	if err := s.client.Set(ctx, treeKey, raw, s.ttl).Err(); err != nil {
		return fmt.Errorf("tree cache set: %w", err)
	}
	return nil
}

func (s *Store) Invalidate(ctx context.Context) error {
	if err := s.client.Del(ctx, treeKey).Err(); err != nil {
		return fmt.Errorf("tree cache invalidate: %w", err)
	}
	return nil
}

// Noop is used when Redis is not configured or unavailable.
type Noop struct{}

var _ usecase.TreeCache = Noop{}

func (Noop) Get(context.Context) ([]usecase.TreeNodeDTO, bool, error) { return nil, false, nil }
func (Noop) Set(context.Context, []usecase.TreeNodeDTO) error         { return nil }
func (Noop) Invalidate(context.Context) error                         { return nil }
