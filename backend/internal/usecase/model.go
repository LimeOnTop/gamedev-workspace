package usecase

import "context"

type Model interface {
	// Upload attaches a GLB model to a file, replacing the previous one.
	Upload(ctx context.Context, input UploadModelInput) (ModelDTO, error)
	// Delete removes the model of the given node.
	Delete(ctx context.Context, nodeID string) error
	Open(ctx context.Context, storedName string) (ModelFile, error)
}
