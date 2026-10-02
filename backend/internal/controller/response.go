package controller

import (
	"net/url"
	"time"

	"github.com/LimeOnTop/gamedev-workspace/backend/internal/entity"
	"github.com/LimeOnTop/gamedev-workspace/backend/internal/usecase"
)

type referenceResponse struct {
	ID           string `json:"id"`
	NodeID       string `json:"node_id"`
	OriginalName string `json:"original_name"`
	URL          string `json:"url"`
	ContentType  string `json:"content_type"`
	Size         int64  `json:"size"`
	CreatedAt    string `json:"created_at"`
}

type modelResponse struct {
	ID           string `json:"id"`
	NodeID       string `json:"node_id"`
	OriginalName string `json:"original_name"`
	URL          string `json:"url"`
	ContentType  string `json:"content_type"`
	Size         int64  `json:"size"`
	CreatedAt    string `json:"created_at"`
}

type pathItemResponse struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type nodeResponse struct {
	ID              string                  `json:"id"`
	ParentID        *string                 `json:"parent_id"`
	Kind            string                  `json:"kind"`
	FileType        *string                 `json:"file_type"` // null for folders
	Name            string                  `json:"name"`
	Description     string                  `json:"description"`
	Mechanics       string                  `json:"mechanics"`
	Characteristics []entity.Characteristic `json:"characteristics"`
	ReferencePrompt string                  `json:"reference_prompt"`
	AssetCategory   *string                 `json:"asset_category"`
	References      []referenceResponse     `json:"references"`
	Model           *modelResponse          `json:"model"`
	Path            []pathItemResponse      `json:"path"`
	CreatedAt       string                  `json:"created_at"`
	UpdatedAt       string                  `json:"updated_at"`
}

type treeNodeResponse struct {
	ID            string             `json:"id"`
	ParentID      *string            `json:"parent_id"`
	Kind          string             `json:"kind"`
	FileType      *string            `json:"file_type"`
	Name          string             `json:"name"`
	Summary       string             `json:"summary"`
	Preview       *string            `json:"preview"`
	AssetCategory *string            `json:"asset_category"`
	Children      []treeNodeResponse `json:"children"`
}

type assetCategoryResponse struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	AssetCount  int    `json:"asset_count"`
}

type assetResponse struct {
	ID       string             `json:"id"`
	ParentID *string            `json:"parent_id"`
	Category string             `json:"category"`
	Name     string             `json:"name"`
	Summary  string             `json:"summary"`
	Preview  *string            `json:"preview"`
	HasModel bool               `json:"has_model"`
	Path     []pathItemResponse `json:"path"`
}

type searchHitResponse struct {
	ID       string  `json:"id"`
	ParentID *string `json:"parent_id"`
	Kind     string  `json:"kind"`
	FileType *string `json:"file_type"`
	Name     string  `json:"name"`
	Summary  string  `json:"summary"`
}

// UploadURL is the public path a stored reference is served from.
func UploadURL(storedName string) string {
	return "/uploads/" + url.PathEscape(storedName)
}

// ModelURL is the public path a stored 3D model is served from.
func ModelURL(storedName string) string {
	return "/uploads/models/" + url.PathEscape(storedName)
}

func previewURL(storedName *string) *string {
	if storedName == nil {
		return nil
	}
	u := UploadURL(*storedName)
	return &u
}

// fileType renders the file type of folders as null.
func fileType(t string) *string {
	if t == "" {
		return nil
	}
	return &t
}

func toPathResponse(path []entity.PathItem) []pathItemResponse {
	result := make([]pathItemResponse, 0, len(path))
	for _, p := range path {
		result = append(result, pathItemResponse{ID: p.ID, Name: p.Name})
	}
	return result
}

func toAssetCategoriesResponse(categories []usecase.AssetCategoryDTO) []assetCategoryResponse {
	result := make([]assetCategoryResponse, 0, len(categories))
	for _, c := range categories {
		result = append(result, assetCategoryResponse{ID: c.ID, Name: c.Name, Description: c.Description, AssetCount: c.AssetCount})
	}
	return result
}

func toAssetsResponse(assets []usecase.AssetDTO) []assetResponse {
	result := make([]assetResponse, 0, len(assets))
	for _, a := range assets {
		result = append(result, assetResponse{
			ID: a.ID, ParentID: a.ParentID, Category: a.Category, Name: a.Name,
			Summary: a.Summary, Preview: previewURL(a.Preview), HasModel: a.HasModel, Path: toPathResponse(a.Path),
		})
	}
	return result
}

func toReferenceResponse(ref usecase.ReferenceDTO) referenceResponse {
	return referenceResponse{
		ID:           ref.ID,
		NodeID:       ref.NodeID,
		OriginalName: ref.OriginalName,
		URL:          UploadURL(ref.StoredName),
		ContentType:  ref.ContentType,
		Size:         ref.Size,
		CreatedAt:    ref.CreatedAt.Format(time.RFC3339),
	}
}

func toModelResponse(model usecase.ModelDTO) modelResponse {
	return modelResponse{
		ID:           model.ID,
		NodeID:       model.NodeID,
		OriginalName: model.OriginalName,
		URL:          ModelURL(model.StoredName),
		ContentType:  model.ContentType,
		Size:         model.Size,
		CreatedAt:    model.CreatedAt.Format(time.RFC3339),
	}
}

func toNodeResponse(node usecase.NodeDTO) nodeResponse {
	refs := make([]referenceResponse, 0, len(node.References))
	for _, ref := range node.References {
		refs = append(refs, toReferenceResponse(ref))
	}
	path := make([]pathItemResponse, 0, len(node.Path))
	for _, p := range node.Path {
		path = append(path, pathItemResponse{ID: p.ID, Name: p.Name})
	}
	var model *modelResponse
	if node.Model != nil {
		m := toModelResponse(*node.Model)
		model = &m
	}
	return nodeResponse{
		ID:              node.ID,
		ParentID:        node.ParentID,
		Kind:            node.Kind,
		FileType:        fileType(node.FileType),
		Name:            node.Name,
		Description:     node.Description,
		Mechanics:       node.Mechanics,
		Characteristics: node.Characteristics,
		ReferencePrompt: node.ReferencePrompt,
		AssetCategory:   node.AssetCategory,
		References:      refs,
		Model:           model,
		Path:            path,
		CreatedAt:       node.CreatedAt.Format(time.RFC3339),
		UpdatedAt:       node.UpdatedAt.Format(time.RFC3339),
	}
}

func toTreeResponse(nodes []usecase.TreeNodeDTO) []treeNodeResponse {
	result := make([]treeNodeResponse, 0, len(nodes))
	for _, n := range nodes {
		result = append(result, treeNodeResponse{
			ID:            n.ID,
			ParentID:      n.ParentID,
			Kind:          n.Kind,
			FileType:      fileType(n.FileType),
			Name:          n.Name,
			Summary:       n.Summary,
			Preview:       previewURL(n.Preview),
			AssetCategory: n.AssetCategory,
			Children:      toTreeResponse(n.Children),
		})
	}
	return result
}

func toSearchResponse(hits []usecase.SearchHitDTO) []searchHitResponse {
	result := make([]searchHitResponse, 0, len(hits))
	for _, h := range hits {
		result = append(result, searchHitResponse{
			ID: h.ID, ParentID: h.ParentID, Kind: h.Kind, FileType: fileType(h.FileType), Name: h.Name, Summary: h.Summary,
		})
	}
	return result
}
