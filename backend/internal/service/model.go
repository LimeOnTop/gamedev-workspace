package service

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"log"
	"path/filepath"
	"strings"

	"github.com/LimeOnTop/gamedev-workspace/backend/internal/apperr"
	"github.com/LimeOnTop/gamedev-workspace/backend/internal/entity"
	"github.com/LimeOnTop/gamedev-workspace/backend/internal/usecase"
)

const modelContentType = "model/gltf-binary"

// glbMagic opens every binary glTF (GLB) file.
var glbMagic = []byte("glTF")

type ModelService struct {
	nodes   usecase.NodeRepository
	models  usecase.ModelRepository
	files   usecase.FileStorage
	cache   usecase.TreeCache
	maxSize int64
}

func NewModelService(
	nodes usecase.NodeRepository,
	models usecase.ModelRepository,
	files usecase.FileStorage,
	cache usecase.TreeCache,
	maxSize int64,
) *ModelService {
	return &ModelService{nodes: nodes, models: models, files: files, cache: cache, maxSize: maxSize}
}

var _ usecase.Model = (*ModelService)(nil)

func (s *ModelService) Upload(ctx context.Context, input usecase.UploadModelInput) (usecase.ModelDTO, error) {
	node, err := s.nodes.GetByID(ctx, input.NodeID)
	if err != nil {
		return usecase.ModelDTO{}, fmt.Errorf("get node: %w", err)
	}
	if node.Kind != entity.KindFile {
		return usecase.ModelDTO{}, apperr.Validation("3D models can only be attached to files, not folders")
	}
	if node.FileType == entity.FileTypeScenario {
		return usecase.ModelDTO{}, apperr.Validation("scenario files cannot have a 3D model; make the file an object first")
	}

	content := bufio.NewReaderSize(input.Content, 512)
	head, _ := content.Peek(len(glbMagic))
	if len(head) == 0 {
		return usecase.ModelDTO{}, apperr.Validation("file is empty")
	}
	if !bytes.Equal(head, glbMagic) {
		return usecase.ModelDTO{}, apperr.Validation("only binary glTF (.glb) models are supported")
	}

	storedName := randomName() + ".glb"
	size, err := s.files.Save(ctx, storedName, content, s.maxSize)
	if err != nil {
		return usecase.ModelDTO{}, fmt.Errorf("save model file: %w", err)
	}

	originalName := filepath.Base(strings.TrimSpace(input.Filename))
	if originalName == "." || originalName == "/" {
		originalName = "model.glb"
	}
	created, replaced, err := s.models.Set(ctx, entity.Model{
		NodeID:       node.ID,
		OriginalName: originalName,
		StoredName:   storedName,
		ContentType:  modelContentType,
		Size:         size,
	})
	if err != nil {
		s.removeFile(ctx, storedName)
		return usecase.ModelDTO{}, fmt.Errorf("set model: %w", err)
	}
	if replaced != "" {
		s.removeFile(ctx, replaced)
	}

	invalidateTree(ctx, s.cache)
	return toModelDTO(created), nil
}

func (s *ModelService) Delete(ctx context.Context, nodeID string) error {
	storedName, err := s.models.Delete(ctx, nodeID)
	if err != nil {
		return fmt.Errorf("delete model: %w", err)
	}
	s.removeFile(ctx, storedName)
	invalidateTree(ctx, s.cache)
	return nil
}

func (s *ModelService) Open(ctx context.Context, storedName string) (usecase.ModelFile, error) {
	model, err := s.models.GetByStoredName(ctx, storedName)
	if err != nil {
		return usecase.ModelFile{}, fmt.Errorf("get model: %w", err)
	}
	content, err := s.files.Open(ctx, model.StoredName)
	if err != nil {
		return usecase.ModelFile{}, fmt.Errorf("open model file: %w", err)
	}
	return usecase.ModelFile{Content: content, ContentType: model.ContentType, ModTime: model.CreatedAt}, nil
}

func (s *ModelService) removeFile(ctx context.Context, storedName string) {
	if err := s.files.Remove(ctx, storedName); err != nil {
		log.Printf("remove model file %s: %v", storedName, err)
	}
}

func toModelDTO(model entity.Model) usecase.ModelDTO {
	return usecase.ModelDTO{
		ID:           model.ID,
		NodeID:       model.NodeID,
		OriginalName: model.OriginalName,
		StoredName:   model.StoredName,
		ContentType:  model.ContentType,
		Size:         model.Size,
		CreatedAt:    model.CreatedAt,
	}
}
