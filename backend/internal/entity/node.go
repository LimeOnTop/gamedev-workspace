package entity

import "time"

const (
	KindFolder = "folder"
	KindFile   = "file"
)

type Characteristic struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

type Node struct {
	ID              string
	ParentID        *string
	Kind            string
	Name            string
	Description     string
	Mechanics       string
	Characteristics []Characteristic
	CreatedAt       time.Time
	UpdatedAt       time.Time
}

// NodeBrief is the lightweight projection used to build the folder tree.
type NodeBrief struct {
	ID       string
	ParentID *string
	Kind     string
	Name     string
	Summary  string
	Preview  *string
}

type PathItem struct {
	ID   string
	Name string
}
