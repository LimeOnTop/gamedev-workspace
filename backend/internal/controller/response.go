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

type pathItemResponse struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type nodeResponse struct {
	ID              string                  `json:"id"`
	ParentID        *string                 `json:"parent_id"`
	Kind            string                  `json:"kind"`
	Name            string                  `json:"name"`
	Description     string                  `json:"description"`
	Mechanics       string                  `json:"mechanics"`
	Characteristics []entity.Characteristic `json:"characteristics"`
	References      []referenceResponse     `json:"references"`
	Path            []pathItemResponse      `json:"path"`
	CreatedAt       string                  `json:"created_at"`
	UpdatedAt       string                  `json:"updated_at"`
}

type treeNodeResponse struct {
	ID       string             `json:"id"`
	ParentID *string            `json:"parent_id"`
	Kind     string             `json:"kind"`
	Name     string             `json:"name"`
	Summary  string             `json:"summary"`
	Preview  *string            `json:"preview"`
	Children []treeNodeResponse `json:"children"`
}

type searchHitResponse struct {
	ID       string  `json:"id"`
	ParentID *string `json:"parent_id"`
	Kind     string  `json:"kind"`
	Name     string  `json:"name"`
	Summary  string  `json:"summary"`
}

// UploadURL is the public path a stored reference is served from.
func UploadURL(storedName string) string {
	return "/uploads/" + url.PathEscape(storedName)
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

func toNodeResponse(node usecase.NodeDTO) nodeResponse {
	refs := make([]referenceResponse, 0, len(node.References))
	for _, ref := range node.References {
		refs = append(refs, toReferenceResponse(ref))
	}
	path := make([]pathItemResponse, 0, len(node.Path))
	for _, p := range node.Path {
		path = append(path, pathItemResponse{ID: p.ID, Name: p.Name})
	}
	return nodeResponse{
		ID:              node.ID,
		ParentID:        node.ParentID,
		Kind:            node.Kind,
		Name:            node.Name,
		Description:     node.Description,
		Mechanics:       node.Mechanics,
		Characteristics: node.Characteristics,
		References:      refs,
		Path:            path,
		CreatedAt:       node.CreatedAt.Format(time.RFC3339),
		UpdatedAt:       node.UpdatedAt.Format(time.RFC3339),
	}
}

func toTreeResponse(nodes []usecase.TreeNodeDTO) []treeNodeResponse {
	result := make([]treeNodeResponse, 0, len(nodes))
	for _, n := range nodes {
		var preview *string
		if n.Preview != nil {
			u := UploadURL(*n.Preview)
			preview = &u
		}
		result = append(result, treeNodeResponse{
			ID:       n.ID,
			ParentID: n.ParentID,
			Kind:     n.Kind,
			Name:     n.Name,
			Summary:  n.Summary,
			Preview:  preview,
			Children: toTreeResponse(n.Children),
		})
	}
	return result
}

func toSearchResponse(hits []usecase.SearchHitDTO) []searchHitResponse {
	result := make([]searchHitResponse, 0, len(hits))
	for _, h := range hits {
		result = append(result, searchHitResponse{
			ID: h.ID, ParentID: h.ParentID, Kind: h.Kind, Name: h.Name, Summary: h.Summary,
		})
	}
	return result
}
