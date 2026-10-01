package entity

import "time"

// Model is the 3D model (GLB) attached to a file node; a file has at most one.
type Model struct {
	ID           string
	NodeID       string
	OriginalName string
	StoredName   string
	ContentType  string
	Size         int64
	CreatedAt    time.Time
}
