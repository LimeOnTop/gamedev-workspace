package usecase

import (
	"io"
	"time"

	"github.com/LimeOnTop/gamedev-workspace/backend/internal/entity"
)

type ReferenceDTO struct {
	ID           string
	NodeID       string
	OriginalName string
	StoredName   string
	ContentType  string
	Size         int64
	CreatedAt    time.Time
}

type ModelDTO struct {
	ID           string
	NodeID       string
	OriginalName string
	StoredName   string
	ContentType  string
	Size         int64
	CreatedAt    time.Time
}

type NodeDTO struct {
	ID              string
	ParentID        *string
	Kind            string
	FileType        string // "" for folders
	Name            string
	Description     string
	Mechanics       string
	Characteristics []entity.Characteristic
	ReferencePrompt string
	AssetCategory   *string
	References      []ReferenceDTO
	Model           *ModelDTO // nil when the file has no 3D model
	Path            []entity.PathItem
	CreatedAt       time.Time
	UpdatedAt       time.Time
}

// TreeNodeDTO is a node of the folder tree shown in the sidebar and overview cards.
// It is also the shape stored in the tree cache.
type TreeNodeDTO struct {
	ID            string
	ParentID      *string
	Kind          string
	FileType      string
	Name          string
	Summary       string
	Preview       *string
	AssetCategory *string
	Children      []TreeNodeDTO
}

type SearchHitDTO struct {
	ID       string
	ParentID *string
	Kind     string
	FileType string
	Name     string
	Summary  string
}

type CreateNodeInput struct {
	ParentID *string
	Kind     string
	// FileType applies to files only: "" means entity.FileTypeObject.
	FileType        string
	Name            string
	Description     string
	Mechanics       string
	Characteristics []entity.Characteristic
	ReferencePrompt string
	// AssetCategory marks the file as a 3D asset; nil or "" means none.
	AssetCategory *string
}

// UpdateNodeInput is a partial update: nil fields are left unchanged.
type UpdateNodeInput struct {
	Name            *string
	Description     *string
	Mechanics       *string
	Characteristics *[]entity.Characteristic
	ReferencePrompt *string
	// AssetCategory: nil leaves it unchanged, "" removes the file from the asset catalog.
	AssetCategory *string
	// FileType converts a file between object and scenario. Fields hidden for
	// scenarios (characteristics, references, model, ...) are kept, so it is reversible.
	FileType *string
}

type AssetCategoryDTO struct {
	ID          string
	Name        string
	Description string
	AssetCount  int
}

type AssetDTO struct {
	ID       string
	ParentID *string
	Category string
	Name     string
	Summary  string
	Preview  *string
	HasModel bool
	Path     []entity.PathItem
}

type UploadReferenceInput struct {
	NodeID   string
	Filename string
	Content  io.Reader
}

// ReferenceFile is an opened stored reference ready to be streamed to a client.
type ReferenceFile struct {
	Content     io.ReadSeekCloser
	ContentType string
	ModTime     time.Time
}

type UploadModelInput struct {
	NodeID   string
	Filename string
	Content  io.Reader
}

// ModelFile is an opened stored model ready to be streamed to a client.
type ModelFile struct {
	Content     io.ReadSeekCloser
	ContentType string
	ModTime     time.Time
}
