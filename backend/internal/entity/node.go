package entity

import "time"

const (
	KindFolder = "folder"
	KindFile   = "file"
)

// File types: an object card (references, 3D model, stats) or a scenario
// document that only has a structured Markdown description. Folders have none.
const (
	FileTypeObject   = "object"
	FileTypeScenario = "scenario"
)

type Characteristic struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

type Node struct {
	ID       string
	ParentID *string
	Kind     string
	// FileType is set for files only; "" for folders.
	FileType        string
	Name            string
	Description     string
	Mechanics       string
	Characteristics []Characteristic
	// ReferencePrompt is a text-to-image prompt for generating references of the object.
	ReferencePrompt string
	// AssetCategory is set for files describing objects that get a 3D model.
	AssetCategory *string
	CreatedAt     time.Time
	UpdatedAt     time.Time
}

// NodeBrief is the lightweight projection used to build the folder tree.
type NodeBrief struct {
	ID            string
	ParentID      *string
	Kind          string
	FileType      string
	Name          string
	Summary       string
	Preview       *string
	AssetCategory *string
}

type PathItem struct {
	ID   string
	Name string
}
