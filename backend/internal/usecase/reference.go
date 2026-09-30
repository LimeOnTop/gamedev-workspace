package usecase

import "context"

type Reference interface {
	Upload(ctx context.Context, input UploadReferenceInput) (ReferenceDTO, error)
	Delete(ctx context.Context, id string) error
	Open(ctx context.Context, storedName string) (ReferenceFile, error)
}
