package service

import (
	"context"
	"errors"
	"fmt"
	"log"
	"strings"

	"github.com/LimeOnTop/gamedev-workspace/backend/internal/apperr"
	"github.com/LimeOnTop/gamedev-workspace/backend/internal/entity"
	"github.com/LimeOnTop/gamedev-workspace/backend/internal/usecase"
)

const (
	maxNameLength      = 200
	defaultSearchLimit = 50
	maxSearchLimit     = 200
)

type NodeService struct {
	nodes      usecase.NodeRepository
	references usecase.ReferenceRepository
	files      usecase.FileStorage
	cache      usecase.TreeCache
}

func NewNodeService(
	nodes usecase.NodeRepository,
	references usecase.ReferenceRepository,
	files usecase.FileStorage,
	cache usecase.TreeCache,
) *NodeService {
	return &NodeService{nodes: nodes, references: references, files: files, cache: cache}
}

var _ usecase.Node = (*NodeService)(nil)

func (s *NodeService) Tree(ctx context.Context) ([]usecase.TreeNodeDTO, error) {
	tree, found, err := s.cache.Get(ctx)
	if err != nil {
		log.Printf("tree cache read failed, loading from db: %v", err)
	}
	if found {
		return tree, nil
	}

	briefs, err := s.nodes.GetAllBrief(ctx)
	if err != nil {
		return nil, fmt.Errorf("load tree: %w", err)
	}
	tree = buildTree(briefs)

	if err := s.cache.Set(ctx, tree); err != nil {
		log.Printf("tree cache write failed: %v", err)
	}
	return tree, nil
}

func (s *NodeService) GetByID(ctx context.Context, id string) (usecase.NodeDTO, error) {
	node, err := s.nodes.GetByID(ctx, id)
	if err != nil {
		return usecase.NodeDTO{}, fmt.Errorf("get node: %w", err)
	}
	return s.toDetailedDTO(ctx, node)
}

func (s *NodeService) Search(ctx context.Context, query string, limit int64) ([]usecase.SearchHitDTO, error) {
	query = strings.TrimSpace(query)
	if query == "" {
		return nil, apperr.Validation("search query is required")
	}
	if limit <= 0 {
		limit = defaultSearchLimit
	}
	limit = min(limit, maxSearchLimit)

	briefs, err := s.nodes.Search(ctx, query, limit)
	if err != nil {
		return nil, fmt.Errorf("search nodes: %w", err)
	}
	hits := make([]usecase.SearchHitDTO, 0, len(briefs))
	for _, b := range briefs {
		hits = append(hits, usecase.SearchHitDTO{
			ID: b.ID, ParentID: b.ParentID, Kind: b.Kind, Name: b.Name, Summary: b.Summary,
		})
	}
	return hits, nil
}

func (s *NodeService) Create(ctx context.Context, input usecase.CreateNodeInput) (usecase.NodeDTO, error) {
	if input.Kind != entity.KindFolder && input.Kind != entity.KindFile {
		return usecase.NodeDTO{}, apperr.Validation("kind must be %q or %q", entity.KindFolder, entity.KindFile)
	}
	name, err := cleanName(input.Name)
	if err != nil {
		return usecase.NodeDTO{}, err
	}
	if err := s.ensureFolder(ctx, input.ParentID); err != nil {
		return usecase.NodeDTO{}, err
	}

	created, err := s.nodes.Create(ctx, entity.Node{
		ParentID:        input.ParentID,
		Kind:            input.Kind,
		Name:            name,
		Description:     strings.TrimSpace(input.Description),
		Mechanics:       strings.TrimSpace(input.Mechanics),
		Characteristics: cleanCharacteristics(input.Characteristics),
	})
	if err != nil {
		return usecase.NodeDTO{}, fmt.Errorf("create node: %w", err)
	}
	s.invalidate(ctx)
	return s.toDetailedDTO(ctx, created)
}

func (s *NodeService) Update(ctx context.Context, id string, input usecase.UpdateNodeInput) (usecase.NodeDTO, error) {
	node, err := s.nodes.GetByID(ctx, id)
	if err != nil {
		return usecase.NodeDTO{}, fmt.Errorf("get node: %w", err)
	}

	if input.Name != nil {
		if node.Name, err = cleanName(*input.Name); err != nil {
			return usecase.NodeDTO{}, err
		}
	}
	if input.Description != nil {
		node.Description = strings.TrimSpace(*input.Description)
	}
	if input.Mechanics != nil {
		node.Mechanics = strings.TrimSpace(*input.Mechanics)
	}
	if input.Characteristics != nil {
		node.Characteristics = cleanCharacteristics(*input.Characteristics)
	}

	updated, err := s.nodes.Update(ctx, node)
	if err != nil {
		return usecase.NodeDTO{}, fmt.Errorf("update node: %w", err)
	}
	s.invalidate(ctx)
	return s.toDetailedDTO(ctx, updated)
}

func (s *NodeService) Move(ctx context.Context, id string, parentID *string) (usecase.NodeDTO, error) {
	if _, err := s.nodes.GetByID(ctx, id); err != nil {
		return usecase.NodeDTO{}, fmt.Errorf("get node: %w", err)
	}
	if err := s.ensureFolder(ctx, parentID); err != nil {
		return usecase.NodeDTO{}, err
	}
	if parentID != nil {
		cycle, err := s.nodes.IsDescendant(ctx, id, *parentID)
		if err != nil {
			return usecase.NodeDTO{}, err
		}
		if cycle {
			return usecase.NodeDTO{}, apperr.Validation("cannot move a folder into itself or its subfolder")
		}
	}

	if err := s.nodes.SetParent(ctx, id, parentID); err != nil {
		return usecase.NodeDTO{}, fmt.Errorf("move node: %w", err)
	}
	s.invalidate(ctx)
	return s.GetByID(ctx, id)
}

func (s *NodeService) Delete(ctx context.Context, id string) error {
	storedNames, err := s.nodes.Delete(ctx, id)
	if err != nil {
		return fmt.Errorf("delete node: %w", err)
	}
	for _, name := range storedNames {
		if err := s.files.Remove(ctx, name); err != nil {
			log.Printf("remove reference file %s: %v", name, err)
		}
	}
	s.invalidate(ctx)
	return nil
}

func (s *NodeService) ensureFolder(ctx context.Context, id *string) error {
	if id == nil {
		return nil
	}
	parent, err := s.nodes.GetByID(ctx, *id)
	if errors.Is(err, apperr.ErrNotFound) {
		return apperr.Validation("parent folder %s does not exist", *id)
	}
	if err != nil {
		return fmt.Errorf("get parent: %w", err)
	}
	if parent.Kind != entity.KindFolder {
		return apperr.Validation("parent must be a folder")
	}
	return nil
}

func (s *NodeService) invalidate(ctx context.Context) {
	invalidateTree(ctx, s.cache)
}

func (s *NodeService) toDetailedDTO(ctx context.Context, node entity.Node) (usecase.NodeDTO, error) {
	references, err := s.references.GetByNodeID(ctx, node.ID)
	if err != nil {
		return usecase.NodeDTO{}, fmt.Errorf("get node references: %w", err)
	}
	path, err := s.nodes.GetPath(ctx, node.ID)
	if err != nil {
		return usecase.NodeDTO{}, fmt.Errorf("get node path: %w", err)
	}

	dto := usecase.NodeDTO{
		ID:              node.ID,
		ParentID:        node.ParentID,
		Kind:            node.Kind,
		Name:            node.Name,
		Description:     node.Description,
		Mechanics:       node.Mechanics,
		Characteristics: node.Characteristics,
		References:      make([]usecase.ReferenceDTO, 0, len(references)),
		Path:            path,
		CreatedAt:       node.CreatedAt,
		UpdatedAt:       node.UpdatedAt,
	}
	if dto.Characteristics == nil {
		dto.Characteristics = []entity.Characteristic{}
	}
	for _, ref := range references {
		dto.References = append(dto.References, toReferenceDTO(ref))
	}
	return dto, nil
}

func buildTree(briefs []entity.NodeBrief) []usecase.TreeNodeDTO {
	children := make(map[string][]entity.NodeBrief)
	var roots []entity.NodeBrief
	for _, b := range briefs {
		if b.ParentID == nil {
			roots = append(roots, b)
		} else {
			children[*b.ParentID] = append(children[*b.ParentID], b)
		}
	}

	var build func(items []entity.NodeBrief) []usecase.TreeNodeDTO
	build = func(items []entity.NodeBrief) []usecase.TreeNodeDTO {
		result := make([]usecase.TreeNodeDTO, 0, len(items))
		for _, b := range items {
			result = append(result, usecase.TreeNodeDTO{
				ID:       b.ID,
				ParentID: b.ParentID,
				Kind:     b.Kind,
				Name:     b.Name,
				Summary:  b.Summary,
				Preview:  b.Preview,
				Children: build(children[b.ID]),
			})
		}
		return result
	}
	return build(roots)
}

func cleanName(name string) (string, error) {
	name = strings.TrimSpace(name)
	if name == "" {
		return "", apperr.Validation("name is required")
	}
	if len([]rune(name)) > maxNameLength {
		return "", apperr.Validation("name is too long (max %d characters)", maxNameLength)
	}
	return name, nil
}

func cleanCharacteristics(chars []entity.Characteristic) []entity.Characteristic {
	result := make([]entity.Characteristic, 0, len(chars))
	for _, c := range chars {
		c.Key, c.Value = strings.TrimSpace(c.Key), strings.TrimSpace(c.Value)
		if c.Key == "" && c.Value == "" {
			continue
		}
		result = append(result, c)
	}
	return result
}

func invalidateTree(ctx context.Context, cache usecase.TreeCache) {
	if err := cache.Invalidate(ctx); err != nil {
		log.Printf("tree cache invalidate failed: %v", err)
	}
}
