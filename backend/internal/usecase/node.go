package usecase

import "context"

type Node interface {
	Tree(ctx context.Context) ([]TreeNodeDTO, error)
	GetByID(ctx context.Context, id string) (NodeDTO, error)
	Search(ctx context.Context, query string, limit int64) ([]SearchHitDTO, error)
	Create(ctx context.Context, input CreateNodeInput) (NodeDTO, error)
	Update(ctx context.Context, id string, input UpdateNodeInput) (NodeDTO, error)
	Move(ctx context.Context, id string, parentID *string) (NodeDTO, error)
	Delete(ctx context.Context, id string) error
}
