package service

import (
	"context"
	"errors"
	"io"
	"strings"
	"testing"

	"github.com/LimeOnTop/gamedev-workspace/backend/internal/apperr"
	"github.com/LimeOnTop/gamedev-workspace/backend/internal/entity"
	"github.com/LimeOnTop/gamedev-workspace/backend/internal/usecase"
)

// memModels is an in-memory usecase.ModelRepository keyed by node ID.
type memModels struct{ byNode map[string]entity.Model }

func (m *memModels) Set(_ context.Context, model entity.Model) (entity.Model, string, error) {
	old := m.byNode[model.NodeID].StoredName
	m.byNode[model.NodeID] = model
	return model, old, nil
}

func (m *memModels) GetByNodeID(_ context.Context, nodeID string) (entity.Model, error) {
	model, ok := m.byNode[nodeID]
	if !ok {
		return entity.Model{}, apperr.ErrNotFound
	}
	return model, nil
}

func (m *memModels) GetByStoredName(context.Context, string) (entity.Model, error) {
	return entity.Model{}, apperr.ErrNotFound
}

func (m *memModels) Delete(_ context.Context, nodeID string) (string, error) {
	model, ok := m.byNode[nodeID]
	if !ok {
		return "", apperr.ErrNotFound
	}
	delete(m.byNode, nodeID)
	return model.StoredName, nil
}

// sizedFiles reports the real number of bytes read on Save.
type sizedFiles struct{ fakeFiles }

func (f *sizedFiles) Save(_ context.Context, _ string, r io.Reader, _ int64) (int64, error) {
	return io.Copy(io.Discard, r)
}

func newTestModelService() (*ModelService, *memModels, *sizedFiles) {
	nodes, _, _, _ := newTestService()
	models := &memModels{byNode: map[string]entity.Model{}}
	files := &sizedFiles{}
	return NewModelService(nodes.nodes, models, files, &fakeCache{}, 1<<20), models, files
}

func upload(s *ModelService, nodeID, content string) (usecase.ModelDTO, error) {
	return s.Upload(context.Background(), usecase.UploadModelInput{
		NodeID: nodeID, Filename: "castle.glb", Content: strings.NewReader(content),
	})
}

func TestModelUploadRejectsNonGLB(t *testing.T) {
	s, _, _ := newTestModelService()
	if _, err := upload(s, "castle", `{"asset":{"version":"2.0"}}`); !errors.Is(err, apperr.ErrValidation) {
		t.Fatalf("expected validation error for non-GLB content, got %v", err)
	}
	if _, err := upload(s, "map", "glTF\x02\x00\x00\x00"); !errors.Is(err, apperr.ErrValidation) {
		t.Fatalf("expected validation error for a folder, got %v", err)
	}
}

func TestModelUploadReplacesPreviousFile(t *testing.T) {
	s, models, files := newTestModelService()
	first, err := upload(s, "castle", "glTF\x02\x00\x00\x00first")
	if err != nil {
		t.Fatalf("first upload: %v", err)
	}
	if first.Size != 13 || first.ContentType != modelContentType || !strings.HasSuffix(first.StoredName, ".glb") {
		t.Fatalf("unexpected model: %+v", first)
	}
	second, err := upload(s, "castle", "glTF\x02\x00\x00\x00second")
	if err != nil {
		t.Fatalf("second upload: %v", err)
	}
	if len(files.removed) != 1 || files.removed[0] != first.StoredName {
		t.Fatalf("expected the replaced file to be removed, removed=%v", files.removed)
	}
	if models.byNode["castle"].StoredName != second.StoredName {
		t.Fatal("expected the new model to be stored")
	}

	if err := s.Delete(context.Background(), "castle"); err != nil {
		t.Fatalf("delete: %v", err)
	}
	if len(files.removed) != 2 || files.removed[1] != second.StoredName {
		t.Fatalf("expected the model file to be removed on delete, removed=%v", files.removed)
	}
	if err := s.Delete(context.Background(), "castle"); !errors.Is(err, apperr.ErrNotFound) {
		t.Fatalf("expected not found on second delete, got %v", err)
	}
}
