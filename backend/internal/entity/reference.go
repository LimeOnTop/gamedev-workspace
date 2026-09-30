package entity

import "time"

type Reference struct {
	ID           string
	NodeID       string
	OriginalName string
	StoredName   string
	ContentType  string
	Size         int64
	CreatedAt    time.Time
}
