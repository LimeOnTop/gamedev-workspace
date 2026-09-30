package service

import (
	"context"
	"errors"
	"io"
	"testing"

	"github.com/LimeOnTop/gamedev-workspace/backend/internal/apperr"
	"github.com/LimeOnTop/gamedev-workspace/backend/internal/entity"
	"github.com/LimeOnTop/gamedev-workspace/backend/internal/usecase"
)

// fakeNodes is an in-memory usecase.NodeRepository.
type fakeNodes struct {
	nodes   map[string]entity.Node
	deleted []string
	files   []string
}

func (f *fakeNodes) Create(_ context.Context, n entity.Node) (entity.Node, error) {
	n.ID = n.Name
	f.nodes[n.ID] = n
	return n, nil
}

func (f *fakeNodes) GetByID(_ context.Context, id string) (entity.Node, error) {
	n, ok := f.nodes[id]
	if !ok {
		return entity.Node{}, apperr.ErrNotFound
	}
	return n, nil
}

func (f *fakeNodes) GetAllBrief(context.Context) ([]entity.NodeBrief, error) { return nil, nil }
func (f *fakeNodes) GetPath(context.Context, string) ([]entity.PathItem, error) {
	return []entity.PathItem{}, nil
}

func (f *fakeNodes) Update(_ context.Context, n entity.Node) (entity.Node, error) {
	f.nodes[n.ID] = n
	return n, nil
}

func (f *fakeNodes) SetParent(_ context.Context, id string, parentID *string) error {
	n := f.nodes[id]
	n.ParentID = parentID
	f.nodes[id] = n
	return nil
}

func (f *fakeNodes) IsDescendant(_ context.Context, id, candidate string) (bool, error) {
	for cur := &candidate; cur != nil; cur = f.nodes[*cur].ParentID {
		if *cur == id {
			return true, nil
		}
	}
	return false, nil
}

func (f *fakeNodes) Delete(_ context.Context, id string) ([]string, error) {
	f.deleted = append(f.deleted, id)
	return f.files, nil
}

func (f *fakeNodes) Search(context.Context, string, int64) ([]entity.NodeBrief, error) {
	return nil, nil
}

type fakeRefs struct{}

func (fakeRefs) Create(_ context.Context, r entity.Reference) (entity.Reference, error) {
	return r, nil
}
func (fakeRefs) GetByNodeID(context.Context, string) ([]entity.Reference, error) {
	return []entity.Reference{}, nil
}
func (fakeRefs) GetByStoredName(context.Context, string) (entity.Reference, error) {
	return entity.Reference{}, apperr.ErrNotFound
}
func (fakeRefs) Delete(context.Context, string) (string, error) { return "", nil }

type fakeFiles struct{ removed []string }

func (f *fakeFiles) Save(context.Context, string, io.Reader, int64) (int64, error) { return 0, nil }
func (f *fakeFiles) Open(context.Context, string) (io.ReadSeekCloser, error)       { return nil, nil }
func (f *fakeFiles) Remove(_ context.Context, name string) error {
	f.removed = append(f.removed, name)
	return nil
}

type fakeCache struct{ invalidations int }

func (c *fakeCache) Get(context.Context) ([]usecase.TreeNodeDTO, bool, error) { return nil, false, nil }
func (c *fakeCache) Set(context.Context, []usecase.TreeNodeDTO) error         { return nil }
func (c *fakeCache) Invalidate(context.Context) error {
	c.invalidations++
	return nil
}

func ptr(s string) *string { return &s }

func newTestService() (*NodeService, *fakeNodes, *fakeFiles, *fakeCache) {
	nodes := &fakeNodes{nodes: map[string]entity.Node{
		"map":    {ID: "map", Kind: entity.KindFolder, Name: "Map"},
		"region": {ID: "region", ParentID: ptr("map"), Kind: entity.KindFolder, Name: "Region"},
		"castle": {ID: "castle", ParentID: ptr("region"), Kind: entity.KindFile, Name: "Castle"},
	}}
	files := &fakeFiles{}
	cache := &fakeCache{}
	return NewNodeService(nodes, fakeRefs{}, files, cache), nodes, files, cache
}

func TestCreateValidatesParent(t *testing.T) {
	svc, _, _, cache := newTestService()
	ctx := context.Background()

	if _, err := svc.Create(ctx, usecase.CreateNodeInput{ParentID: ptr("castle"), Kind: entity.KindFile, Name: "x"}); !errors.Is(err, apperr.ErrValidation) {
		t.Fatalf("file parent: want validation error, got %v", err)
	}
	if _, err := svc.Create(ctx, usecase.CreateNodeInput{ParentID: ptr("missing"), Kind: entity.KindFile, Name: "x"}); !errors.Is(err, apperr.ErrValidation) {
		t.Fatalf("missing parent: want validation error, got %v", err)
	}
	if _, err := svc.Create(ctx, usecase.CreateNodeInput{Kind: entity.KindFolder, Name: "   "}); !errors.Is(err, apperr.ErrValidation) {
		t.Fatalf("blank name: want validation error, got %v", err)
	}

	got, err := svc.Create(ctx, usecase.CreateNodeInput{
		ParentID: ptr("map"), Kind: entity.KindFile, Name: "  Village ",
		Characteristics: []entity.Characteristic{{Key: " Size ", Value: "10"}, {Key: " ", Value: ""}},
	})
	if err != nil {
		t.Fatal(err)
	}
	if got.Name != "Village" || len(got.Characteristics) != 1 || got.Characteristics[0].Key != "Size" {
		t.Fatalf("input not cleaned: %+v", got)
	}
	if cache.invalidations != 1 {
		t.Fatalf("want 1 cache invalidation, got %d", cache.invalidations)
	}
}

func TestMoveRejectsCycles(t *testing.T) {
	svc, _, _, _ := newTestService()
	ctx := context.Background()

	for _, target := range []string{"map", "region"} {
		if _, err := svc.Move(ctx, "map", ptr(target)); !errors.Is(err, apperr.ErrValidation) {
			t.Fatalf("move map into %s: want validation error, got %v", target, err)
		}
	}
	if _, err := svc.Move(ctx, "castle", ptr("map")); err != nil {
		t.Fatalf("valid move failed: %v", err)
	}
	if _, err := svc.Move(ctx, "region", nil); err != nil {
		t.Fatalf("move to root failed: %v", err)
	}
}

func TestDeleteRemovesStoredFiles(t *testing.T) {
	svc, nodes, files, cache := newTestService()
	nodes.files = []string{"a.png", "b.jpg"}

	if err := svc.Delete(context.Background(), "map"); err != nil {
		t.Fatal(err)
	}
	if len(files.removed) != 2 || files.removed[0] != "a.png" || files.removed[1] != "b.jpg" {
		t.Fatalf("stored files not removed: %v", files.removed)
	}
	if cache.invalidations != 1 {
		t.Fatalf("want 1 cache invalidation, got %d", cache.invalidations)
	}
}

func TestBuildTree(t *testing.T) {
	tree := buildTree([]entity.NodeBrief{
		{ID: "a", Kind: entity.KindFolder, Name: "A"},
		{ID: "b", ParentID: ptr("a"), Kind: entity.KindFile, Name: "B"},
		{ID: "c", Kind: entity.KindFolder, Name: "C"},
	})
	if len(tree) != 2 || len(tree[0].Children) != 1 || tree[0].Children[0].ID != "b" || tree[1].Children == nil {
		t.Fatalf("unexpected tree: %+v", tree)
	}
}
