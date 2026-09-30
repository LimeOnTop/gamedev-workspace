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

type NodeDTO struct {
	ID              string
	ParentID        *string
	Kind            string
	Name            string
	Description     string
	Mechanics       string
	Characteristics []entity.Characteristic
	References      []ReferenceDTO
	Path            []entity.PathItem
	CreatedAt       time.Time
	UpdatedAt       time.Time
}

// TreeNodeDTO is a node of the folder tree shown in the sidebar and overview cards.
// It is also the shape stored in the tree cache.
type TreeNodeDTO struct {
	ID       string
	ParentID *string
	Kind     string
	Name     string
	Summary  string
	Preview  *string
	Children []TreeNodeDTO
}

type SearchHitDTO struct {
	ID       string
	ParentID *string
	Kind     string
	Name     string
	Summary  string
}

type CreateNodeInput struct {
	ParentID        *string
	Kind            string
	Name            string
	Description     string
	Mechanics       string
	Characteristics []entity.Characteristic
}

// UpdateNodeInput is a partial update: nil fields are left unchanged.
type UpdateNodeInput struct {
	Name            *string
	Description     *string
	Mechanics       *string
	Characteristics *[]entity.Characteristic
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
