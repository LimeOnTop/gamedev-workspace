package service

import (
	"bufio"
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"log"
	"mime"
	"net/http"
	"path/filepath"
	"strings"

	"github.com/LimeOnTop/gamedev-workspace/backend/internal/apperr"
	"github.com/LimeOnTop/gamedev-workspace/backend/internal/entity"
	"github.com/LimeOnTop/gamedev-workspace/backend/internal/usecase"
)

type ReferenceService struct {
	nodes      usecase.NodeRepository
	references usecase.ReferenceRepository
	files      usecase.FileStorage
	cache      usecase.TreeCache
	maxSize    int64
}

func NewReferenceService(
	nodes usecase.NodeRepository,
	references usecase.ReferenceRepository,
	files usecase.FileStorage,
	cache usecase.TreeCache,
	maxSize int64,
) *ReferenceService {
	return &ReferenceService{nodes: nodes, references: references, files: files, cache: cache, maxSize: maxSize}
}

var _ usecase.Reference = (*ReferenceService)(nil)

func (s *ReferenceService) Upload(ctx context.Context, input usecase.UploadReferenceInput) (usecase.ReferenceDTO, error) {
	node, err := s.nodes.GetByID(ctx, input.NodeID)
	if err != nil {
		return usecase.ReferenceDTO{}, fmt.Errorf("get node: %w", err)
	}
	if node.Kind != entity.KindFile {
		return usecase.ReferenceDTO{}, apperr.Validation("references can only be attached to files, not folders")
	}

	// Sniff the real type instead of trusting the client-provided one.
	content := bufio.NewReaderSize(input.Content, 512)
	head, _ := content.Peek(512)
	if len(head) == 0 {
		return usecase.ReferenceDTO{}, apperr.Validation("file is empty")
	}
	contentType := detectContentType(head, input.Filename)

	storedName := randomName() + sanitizeExt(filepath.Ext(input.Filename))
	size, err := s.files.Save(ctx, storedName, content, s.maxSize)
	if err != nil {
		return usecase.ReferenceDTO{}, fmt.Errorf("save reference file: %w", err)
	}

	originalName := filepath.Base(strings.TrimSpace(input.Filename))
	if originalName == "." || originalName == "/" {
		originalName = storedName
	}
	created, err := s.references.Create(ctx, entity.Reference{
		NodeID:       node.ID,
		OriginalName: originalName,
		StoredName:   storedName,
		ContentType:  contentType,
		Size:         size,
	})
	if err != nil {
		if rmErr := s.files.Remove(ctx, storedName); rmErr != nil {
			log.Printf("remove orphan reference file %s: %v", storedName, rmErr)
		}
		return usecase.ReferenceDTO{}, fmt.Errorf("create reference: %w", err)
	}

	invalidateTree(ctx, s.cache)
	return toReferenceDTO(created), nil
}

func (s *ReferenceService) Delete(ctx context.Context, id string) error {
	storedName, err := s.references.Delete(ctx, id)
	if err != nil {
		return fmt.Errorf("delete reference: %w", err)
	}
	if err := s.files.Remove(ctx, storedName); err != nil {
		log.Printf("remove reference file %s: %v", storedName, err)
	}
	invalidateTree(ctx, s.cache)
	return nil
}

func (s *ReferenceService) Open(ctx context.Context, storedName string) (usecase.ReferenceFile, error) {
	ref, err := s.references.GetByStoredName(ctx, storedName)
	if err != nil {
		return usecase.ReferenceFile{}, fmt.Errorf("get reference: %w", err)
	}
	content, err := s.files.Open(ctx, ref.StoredName)
	if err != nil {
		return usecase.ReferenceFile{}, fmt.Errorf("open reference file: %w", err)
	}
	return usecase.ReferenceFile{Content: content, ContentType: ref.ContentType, ModTime: ref.CreatedAt}, nil
}

func toReferenceDTO(ref entity.Reference) usecase.ReferenceDTO {
	return usecase.ReferenceDTO{
		ID:           ref.ID,
		NodeID:       ref.NodeID,
		OriginalName: ref.OriginalName,
		StoredName:   ref.StoredName,
		ContentType:  ref.ContentType,
		Size:         ref.Size,
		CreatedAt:    ref.CreatedAt,
	}
}

// detectContentType falls back to the extension for formats the sniffer does
// not know, but never lets an upload be served as HTML or script.
func detectContentType(head []byte, filename string) string {
	ct := http.DetectContentType(head)
	if ct == "application/octet-stream" || strings.HasPrefix(ct, "text/plain") {
		byExt := mime.TypeByExtension(strings.ToLower(filepath.Ext(filename)))
		if byExt != "" && !strings.Contains(byExt, "html") && !strings.Contains(byExt, "javascript") {
			return byExt
		}
	}
	return ct
}

func randomName() string {
	b := make([]byte, 16)
	rand.Read(b)
	return hex.EncodeToString(b)
}

func sanitizeExt(ext string) string {
	ext = strings.ToLower(ext)
	if len(ext) < 2 || len(ext) > 10 {
		return ""
	}
	for _, c := range ext[1:] {
		if !(c >= 'a' && c <= 'z' || c >= '0' && c <= '9') {
			return ""
		}
	}
	return ext
}
