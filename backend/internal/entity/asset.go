package entity

// AssetCategory groups files that describe objects with a 3D model
// (weapons, map locations, characters, ...), independently of the folder tree.
type AssetCategory struct {
	ID          string
	Name        string
	Description string
	SortOrder   int
	AssetCount  int
}

// Asset is a file node assigned to an asset category.
type Asset struct {
	ID       string
	ParentID *string
	Category string
	Name     string
	Summary  string
	Preview  *string
	Path     []PathItem
}
