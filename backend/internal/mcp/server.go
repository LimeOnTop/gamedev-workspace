// Package mcp exposes the workspace to AI agents (Claude Code, Claude Desktop, ...)
// over the Model Context Protocol, using the same use cases as the HTTP API.
package mcp

import (
	"bytes"
	"context"
	"encoding/base64"
	"errors"
	"fmt"
	"mime"
	"net/http"
	"net/url"
	"path"
	"strings"
	"time"

	"github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/LimeOnTop/gamedev-workspace/backend/internal/apperr"
	"github.com/LimeOnTop/gamedev-workspace/backend/internal/entity"
	"github.com/LimeOnTop/gamedev-workspace/backend/internal/usecase"
)

const instructions = `Gamedev Workspace stores game design documentation as a tree of folders and files.
- Folders group related entities (e.g. "Weapons", "Map", "Characters").
- Files describe a single game object (e.g. "Castle", "Pistol") and hold: description,
  characteristics (ordered key/value pairs such as damage or size), interaction mechanics,
  and references (images, videos, PDFs).
Workflow: call get_tree first to learn the structure and node IDs, then read/create/update nodes.
Write content in the language the existing workspace uses. Every result includes web_url
so you can link the user to the page you changed.

3D asset catalog: files describing physical objects that will get a 3D model (weapons, siege engines,
locations/buildings, characters, items, props, gear) should have asset_category set
(see list_asset_categories). Design documents (story, rules, mechanics overviews, UI) must NOT
get a category. Folders never have one.

reference_prompt: a text-to-image prompt used to generate reference art for the object. Write it in
English, as one self-contained paragraph: the object itself (shape, materials, colours, wear, scale
cues), the game's art style, and a presentation useful for 3D modelling (single object, neutral
background, three-quarter view, even lighting, no text).`

type Server struct {
	node      usecase.Node
	reference usecase.Reference
	asset     usecase.Asset
	publicURL string
	http      *http.Client
}

// NewHandler returns an http.Handler serving MCP over Streamable HTTP.
func NewHandler(node usecase.Node, reference usecase.Reference, asset usecase.Asset, publicURL string) http.Handler {
	s := &Server{
		node:      node,
		reference: reference,
		asset:     asset,
		publicURL: strings.TrimRight(publicURL, "/"),
		http:      &http.Client{Timeout: 60 * time.Second},
	}
	server := mcp.NewServer(&mcp.Implementation{Name: "gamedev-workspace", Version: "1.0.0"}, &mcp.ServerOptions{
		Instructions: instructions,
	})
	s.register(server)
	return mcp.NewStreamableHTTPHandler(func(*http.Request) *mcp.Server { return server }, nil)
}

func ptr[T any](v T) *T { return &v }

func (s *Server) register(server *mcp.Server) {
	readOnly := &mcp.ToolAnnotations{ReadOnlyHint: true}
	additive := &mcp.ToolAnnotations{DestructiveHint: ptr(false)}
	destructive := &mcp.ToolAnnotations{DestructiveHint: ptr(true)}

	mcp.AddTool(server, &mcp.Tool{
		Name:        "get_tree",
		Title:       "Get workspace tree",
		Description: "Returns the whole folder/file tree as an indented outline with node IDs; 3D assets are marked [3D: category].",
		Annotations: readOnly,
	}, s.getTree)

	mcp.AddTool(server, &mcp.Tool{
		Name:        "get_node",
		Title:       "Get node",
		Description: "Returns a folder or file with its description, characteristics, mechanics, reference prompt, asset category, references and breadcrumb path.",
		Annotations: readOnly,
	}, s.getNode)

	mcp.AddTool(server, &mcp.Tool{
		Name:        "list_asset_categories",
		Title:       "List 3D asset categories",
		Description: "Returns the categories of the 3D asset catalog (weapons, map, characters, ...) with asset counts.",
		Annotations: readOnly,
	}, s.listAssetCategories)

	mcp.AddTool(server, &mcp.Tool{
		Name:        "list_assets",
		Title:       "List 3D assets",
		Description: "Returns files in the 3D asset catalog, optionally filtered by category, with their folder path.",
		Annotations: readOnly,
	}, s.listAssets)

	mcp.AddTool(server, &mcp.Tool{
		Name:        "search_nodes",
		Title:       "Search nodes",
		Description: "Case-insensitive substring search over names, descriptions, mechanics and characteristics.",
		Annotations: readOnly,
	}, s.searchNodes)

	mcp.AddTool(server, &mcp.Tool{
		Name:        "create_folder",
		Title:       "Create folder",
		Description: "Creates a folder at the root or inside another folder.",
		Annotations: additive,
	}, s.createFolder)

	mcp.AddTool(server, &mcp.Tool{
		Name:        "create_file",
		Title:       "Create file",
		Description: "Creates a file (a game object card) with optional description, characteristics and mechanics.",
		Annotations: additive,
	}, s.createFile)

	mcp.AddTool(server, &mcp.Tool{
		Name:  "update_node",
		Title: "Update node",
		Description: "Updates a folder or file. Omitted fields are unchanged. By default `characteristics` " +
			"replaces the whole list; set merge_characteristics=true to upsert by key instead.",
		Annotations: &mcp.ToolAnnotations{DestructiveHint: ptr(true), IdempotentHint: true},
	}, s.updateNode)

	mcp.AddTool(server, &mcp.Tool{
		Name:        "move_node",
		Title:       "Move node",
		Description: "Moves a folder or file into another folder, or to the root when parent_id is omitted.",
		Annotations: &mcp.ToolAnnotations{DestructiveHint: ptr(false), IdempotentHint: true},
	}, s.moveNode)

	mcp.AddTool(server, &mcp.Tool{
		Name:        "delete_node",
		Title:       "Delete node",
		Description: "Permanently deletes a folder or file together with all nested nodes and their reference files.",
		Annotations: destructive,
	}, s.deleteNode)

	mcp.AddTool(server, &mcp.Tool{
		Name:        "add_reference_from_url",
		Title:       "Add reference from URL",
		Description: "Downloads an image/video/PDF from a public http(s) URL and attaches it to a file as a reference.",
		Annotations: &mcp.ToolAnnotations{DestructiveHint: ptr(false), OpenWorldHint: ptr(true)},
	}, s.addReferenceFromURL)

	mcp.AddTool(server, &mcp.Tool{
		Name:        "add_reference_from_base64",
		Title:       "Add reference from base64",
		Description: "Attaches base64-encoded file content (optionally a data: URL) to a file as a reference.",
		Annotations: additive,
	}, s.addReferenceFromBase64)

	mcp.AddTool(server, &mcp.Tool{
		Name:        "delete_reference",
		Title:       "Delete reference",
		Description: "Permanently deletes a reference and its stored file.",
		Annotations: destructive,
	}, s.deleteReference)
}

// ---- inputs ----

type idInput struct {
	ID string `json:"id" jsonschema:"node ID (UUID) from get_tree or search_nodes"`
}

type searchInput struct {
	Query string `json:"query" jsonschema:"text to search for"`
	Limit int64  `json:"limit,omitempty" jsonschema:"max results, default 50"`
}

type createFolderInput struct {
	ParentID    string `json:"parent_id,omitempty" jsonschema:"parent folder ID; omit to create at the root"`
	Name        string `json:"name" jsonschema:"folder name"`
	Description string `json:"description,omitempty" jsonschema:"short description shown on the folder card"`
}

type createFileInput struct {
	ParentID        string                  `json:"parent_id,omitempty" jsonschema:"parent folder ID; omit to create at the root"`
	Name            string                  `json:"name" jsonschema:"object name"`
	Description     string                  `json:"description,omitempty" jsonschema:"what the object is: look, role, lore"`
	Mechanics       string                  `json:"mechanics,omitempty" jsonschema:"how the player interacts with the object"`
	Characteristics []entity.Characteristic `json:"characteristics,omitempty" jsonschema:"ordered key/value stats, e.g. Damage=18"`
	ReferencePrompt string                  `json:"reference_prompt,omitempty" jsonschema:"English text-to-image prompt for generating reference art of the object"`
	AssetCategory   string                  `json:"asset_category,omitempty" jsonschema:"3D asset category ID from list_asset_categories; only for objects that get a 3D model"`
}

type updateNodeInput struct {
	ID                       string                  `json:"id" jsonschema:"node ID"`
	Name                     *string                 `json:"name,omitempty" jsonschema:"new name"`
	Description              *string                 `json:"description,omitempty" jsonschema:"new description (replaces the old one)"`
	Mechanics                *string                 `json:"mechanics,omitempty" jsonschema:"new mechanics text (replaces the old one)"`
	Characteristics          []entity.Characteristic `json:"characteristics,omitempty" jsonschema:"characteristics to set"`
	MergeCharacteristics     bool                    `json:"merge_characteristics,omitempty" jsonschema:"true: upsert given keys and keep others; false (default): replace the whole list"`
	RemoveCharacteristicKeys []string                `json:"remove_characteristic_keys,omitempty" jsonschema:"characteristic keys to remove (case-insensitive)"`
	ReferencePrompt          *string                 `json:"reference_prompt,omitempty" jsonschema:"new reference generation prompt (replaces the old one)"`
	AssetCategory            *string                 `json:"asset_category,omitempty" jsonschema:"3D asset category ID; empty string removes the file from the asset catalog"`
}

type listAssetsInput struct {
	Category string `json:"category,omitempty" jsonschema:"category ID; omit for all categories"`
}

type moveNodeInput struct {
	ID       string `json:"id" jsonschema:"node ID to move"`
	ParentID string `json:"parent_id,omitempty" jsonschema:"target folder ID; omit to move to the root"`
}

type urlReferenceInput struct {
	NodeID   string `json:"node_id" jsonschema:"file node ID"`
	URL      string `json:"url" jsonschema:"public http(s) URL of the file"`
	Filename string `json:"filename,omitempty" jsonschema:"file name to store; derived from the URL when omitted"`
}

type base64ReferenceInput struct {
	NodeID   string `json:"node_id" jsonschema:"file node ID"`
	Filename string `json:"filename" jsonschema:"file name with extension, e.g. castle.png"`
	Data     string `json:"data" jsonschema:"base64 content or a data: URL"`
}

// ---- outputs ----

type referenceView struct {
	ID           string `json:"id"`
	OriginalName string `json:"original_name"`
	ContentType  string `json:"content_type"`
	Size         int64  `json:"size"`
	URL          string `json:"url"`
}

type nodeView struct {
	ID              string                  `json:"id"`
	ParentID        *string                 `json:"parent_id"`
	Kind            string                  `json:"kind"`
	Name            string                  `json:"name"`
	Path            string                  `json:"path"`
	Description     string                  `json:"description"`
	Mechanics       string                  `json:"mechanics,omitempty"`
	Characteristics []entity.Characteristic `json:"characteristics,omitempty"`
	ReferencePrompt string                  `json:"reference_prompt,omitempty"`
	AssetCategory   *string                 `json:"asset_category,omitempty"`
	References      []referenceView         `json:"references,omitempty"`
	Children        []childView             `json:"children,omitempty"`
	WebURL          string                  `json:"web_url"`
	UpdatedAt       string                  `json:"updated_at"`
}

type assetView struct {
	ID       string `json:"id"`
	Category string `json:"category"`
	Name     string `json:"name"`
	Path     string `json:"path"`
	Summary  string `json:"summary,omitempty"`
	WebURL   string `json:"web_url"`
}

type childView struct {
	ID   string `json:"id"`
	Kind string `json:"kind"`
	Name string `json:"name"`
}

type searchHitView struct {
	ID      string `json:"id"`
	Kind    string `json:"kind"`
	Name    string `json:"name"`
	Summary string `json:"summary,omitempty"`
	WebURL  string `json:"web_url"`
}

func (s *Server) webURL(id string) string { return s.publicURL + "/node/" + id }

func (s *Server) uploadURL(storedName string) string {
	return s.publicURL + "/uploads/" + url.PathEscape(storedName)
}

func (s *Server) referenceView(ref usecase.ReferenceDTO) referenceView {
	return referenceView{
		ID:           ref.ID,
		OriginalName: ref.OriginalName,
		ContentType:  ref.ContentType,
		Size:         ref.Size,
		URL:          s.uploadURL(ref.StoredName),
	}
}

func (s *Server) nodeView(ctx context.Context, node usecase.NodeDTO) nodeView {
	names := make([]string, 0, len(node.Path)+1)
	for _, p := range node.Path {
		names = append(names, p.Name)
	}
	names = append(names, node.Name)

	v := nodeView{
		ID:              node.ID,
		ParentID:        node.ParentID,
		Kind:            node.Kind,
		Name:            node.Name,
		Path:            "/" + strings.Join(names, "/"),
		Description:     node.Description,
		Mechanics:       node.Mechanics,
		Characteristics: node.Characteristics,
		ReferencePrompt: node.ReferencePrompt,
		AssetCategory:   node.AssetCategory,
		WebURL:          s.webURL(node.ID),
		UpdatedAt:       node.UpdatedAt.Format(time.RFC3339),
	}
	for _, ref := range node.References {
		v.References = append(v.References, s.referenceView(ref))
	}
	if node.Kind == entity.KindFolder {
		if tree, err := s.node.Tree(ctx); err == nil {
			if found := findTreeNode(tree, node.ID); found != nil {
				for _, c := range found.Children {
					v.Children = append(v.Children, childView{ID: c.ID, Kind: c.Kind, Name: c.Name})
				}
			}
		}
	}
	return v
}

// ---- handlers ----

func (s *Server) getTree(ctx context.Context, _ *mcp.CallToolRequest, _ struct{}) (*mcp.CallToolResult, any, error) {
	tree, err := s.node.Tree(ctx)
	if err != nil {
		return nil, nil, publicErr(err)
	}
	var b strings.Builder
	if len(tree) == 0 {
		b.WriteString("(workspace is empty)\n")
	}
	var walk func(nodes []usecase.TreeNodeDTO, depth int)
	walk = func(nodes []usecase.TreeNodeDTO, depth int) {
		for _, n := range nodes {
			icon := "📄"
			if n.Kind == entity.KindFolder {
				icon = "📁"
			}
			asset := ""
			if n.AssetCategory != nil {
				asset = "  [3D: " + *n.AssetCategory + "]"
			}
			fmt.Fprintf(&b, "%s%s %s  [id=%s]%s\n", strings.Repeat("  ", depth), icon, n.Name, n.ID, asset)
			walk(n.Children, depth+1)
		}
	}
	walk(tree, 0)
	return &mcp.CallToolResult{Content: []mcp.Content{&mcp.TextContent{Text: b.String()}}}, nil, nil
}

func (s *Server) getNode(ctx context.Context, _ *mcp.CallToolRequest, in idInput) (*mcp.CallToolResult, any, error) {
	node, err := s.node.GetByID(ctx, in.ID)
	if err != nil {
		return nil, nil, publicErr(err)
	}
	return nil, s.nodeView(ctx, node), nil
}

func (s *Server) listAssetCategories(ctx context.Context, _ *mcp.CallToolRequest, _ struct{}) (*mcp.CallToolResult, any, error) {
	categories, err := s.asset.Categories(ctx)
	if err != nil {
		return nil, nil, publicErr(err)
	}
	type categoryView struct {
		ID          string `json:"id"`
		Name        string `json:"name"`
		Description string `json:"description"`
		AssetCount  int    `json:"asset_count"`
	}
	views := make([]categoryView, 0, len(categories))
	for _, c := range categories {
		views = append(views, categoryView{ID: c.ID, Name: c.Name, Description: c.Description, AssetCount: c.AssetCount})
	}
	return nil, map[string]any{"categories": views}, nil
}

func (s *Server) listAssets(ctx context.Context, _ *mcp.CallToolRequest, in listAssetsInput) (*mcp.CallToolResult, any, error) {
	assets, err := s.asset.List(ctx, in.Category)
	if err != nil {
		return nil, nil, publicErr(err)
	}
	views := make([]assetView, 0, len(assets))
	for _, a := range assets {
		names := make([]string, 0, len(a.Path)+1)
		for _, p := range a.Path {
			names = append(names, p.Name)
		}
		names = append(names, a.Name)
		views = append(views, assetView{
			ID: a.ID, Category: a.Category, Name: a.Name, Summary: a.Summary,
			Path: "/" + strings.Join(names, "/"), WebURL: s.webURL(a.ID),
		})
	}
	return nil, map[string]any{"assets": views}, nil
}

func (s *Server) searchNodes(ctx context.Context, _ *mcp.CallToolRequest, in searchInput) (*mcp.CallToolResult, any, error) {
	hits, err := s.node.Search(ctx, in.Query, in.Limit)
	if err != nil {
		return nil, nil, publicErr(err)
	}
	views := make([]searchHitView, 0, len(hits))
	for _, h := range hits {
		views = append(views, searchHitView{ID: h.ID, Kind: h.Kind, Name: h.Name, Summary: h.Summary, WebURL: s.webURL(h.ID)})
	}
	return nil, map[string]any{"results": views}, nil
}

func optional(s string) *string {
	if strings.TrimSpace(s) == "" {
		return nil
	}
	return &s
}

func (s *Server) createFolder(ctx context.Context, _ *mcp.CallToolRequest, in createFolderInput) (*mcp.CallToolResult, any, error) {
	node, err := s.node.Create(ctx, usecase.CreateNodeInput{
		ParentID:    optional(in.ParentID),
		Kind:        entity.KindFolder,
		Name:        in.Name,
		Description: in.Description,
	})
	if err != nil {
		return nil, nil, publicErr(err)
	}
	return nil, s.nodeView(ctx, node), nil
}

func (s *Server) createFile(ctx context.Context, _ *mcp.CallToolRequest, in createFileInput) (*mcp.CallToolResult, any, error) {
	node, err := s.node.Create(ctx, usecase.CreateNodeInput{
		ParentID:        optional(in.ParentID),
		Kind:            entity.KindFile,
		Name:            in.Name,
		Description:     in.Description,
		Mechanics:       in.Mechanics,
		Characteristics: in.Characteristics,
		ReferencePrompt: in.ReferencePrompt,
		AssetCategory:   optional(in.AssetCategory),
	})
	if err != nil {
		return nil, nil, publicErr(err)
	}
	return nil, s.nodeView(ctx, node), nil
}

func (s *Server) updateNode(ctx context.Context, _ *mcp.CallToolRequest, in updateNodeInput) (*mcp.CallToolResult, any, error) {
	input := usecase.UpdateNodeInput{
		Name:            in.Name,
		Description:     in.Description,
		Mechanics:       in.Mechanics,
		ReferencePrompt: in.ReferencePrompt,
		AssetCategory:   in.AssetCategory,
	}

	if in.Characteristics != nil || len(in.RemoveCharacteristicKeys) > 0 {
		var chars []entity.Characteristic
		if in.MergeCharacteristics || in.Characteristics == nil {
			current, err := s.node.GetByID(ctx, in.ID)
			if err != nil {
				return nil, nil, publicErr(err)
			}
			chars = mergeCharacteristics(current.Characteristics, in.Characteristics)
		} else {
			chars = in.Characteristics
		}
		chars = removeCharacteristics(chars, in.RemoveCharacteristicKeys)
		input.Characteristics = &chars
	}

	node, err := s.node.Update(ctx, in.ID, input)
	if err != nil {
		return nil, nil, publicErr(err)
	}
	return nil, s.nodeView(ctx, node), nil
}

func (s *Server) moveNode(ctx context.Context, _ *mcp.CallToolRequest, in moveNodeInput) (*mcp.CallToolResult, any, error) {
	node, err := s.node.Move(ctx, in.ID, optional(in.ParentID))
	if err != nil {
		return nil, nil, publicErr(err)
	}
	return nil, s.nodeView(ctx, node), nil
}

func (s *Server) deleteNode(ctx context.Context, _ *mcp.CallToolRequest, in idInput) (*mcp.CallToolResult, any, error) {
	if err := s.node.Delete(ctx, in.ID); err != nil {
		return nil, nil, publicErr(err)
	}
	return nil, map[string]any{"deleted": in.ID}, nil
}

func (s *Server) addReferenceFromURL(ctx context.Context, _ *mcp.CallToolRequest, in urlReferenceInput) (*mcp.CallToolResult, any, error) {
	u, err := url.Parse(in.URL)
	if err != nil || (u.Scheme != "http" && u.Scheme != "https") || u.Host == "" {
		return nil, nil, fmt.Errorf("url must be an absolute http(s) URL")
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u.String(), nil)
	if err != nil {
		return nil, nil, publicErr(err)
	}
	req.Header.Set("User-Agent", "gamedev-workspace/1.0 (+reference importer)")
	resp, err := s.http.Do(req)
	if err != nil {
		return nil, nil, fmt.Errorf("download failed: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, nil, fmt.Errorf("download failed: HTTP %d", resp.StatusCode)
	}

	filename := in.Filename
	if filename == "" {
		if _, params, err := mime.ParseMediaType(resp.Header.Get("Content-Disposition")); err == nil {
			filename = params["filename"]
		}
	}
	if filename == "" {
		filename = path.Base(u.Path)
	}
	if filename == "" || filename == "/" || filename == "." {
		filename = "reference"
	}
	if path.Ext(filename) == "" {
		if exts, _ := mime.ExtensionsByType(resp.Header.Get("Content-Type")); len(exts) > 0 {
			filename += exts[0]
		}
	}

	ref, err := s.reference.Upload(ctx, usecase.UploadReferenceInput{NodeID: in.NodeID, Filename: filename, Content: resp.Body})
	if err != nil {
		return nil, nil, publicErr(err)
	}
	return nil, s.referenceView(ref), nil
}

func (s *Server) addReferenceFromBase64(ctx context.Context, _ *mcp.CallToolRequest, in base64ReferenceInput) (*mcp.CallToolResult, any, error) {
	data := strings.TrimSpace(in.Data)
	if strings.HasPrefix(data, "data:") {
		if i := strings.Index(data, ","); i >= 0 {
			data = data[i+1:]
		}
	}
	raw, err := base64.StdEncoding.DecodeString(data)
	if err != nil {
		if raw, err = base64.RawStdEncoding.DecodeString(data); err != nil {
			return nil, nil, fmt.Errorf("data is not valid base64")
		}
	}
	ref, err := s.reference.Upload(ctx, usecase.UploadReferenceInput{NodeID: in.NodeID, Filename: in.Filename, Content: bytes.NewReader(raw)})
	if err != nil {
		return nil, nil, publicErr(err)
	}
	return nil, s.referenceView(ref), nil
}

func (s *Server) deleteReference(ctx context.Context, _ *mcp.CallToolRequest, in idInput) (*mcp.CallToolResult, any, error) {
	if err := s.reference.Delete(ctx, in.ID); err != nil {
		return nil, nil, publicErr(err)
	}
	return nil, map[string]any{"deleted": in.ID}, nil
}

// ---- helpers ----

func mergeCharacteristics(current, updates []entity.Characteristic) []entity.Characteristic {
	result := append([]entity.Characteristic(nil), current...)
	for _, u := range updates {
		replaced := false
		for i := range result {
			if strings.EqualFold(result[i].Key, u.Key) {
				result[i].Value = u.Value
				replaced = true
				break
			}
		}
		if !replaced {
			result = append(result, u)
		}
	}
	return result
}

func removeCharacteristics(chars []entity.Characteristic, keys []string) []entity.Characteristic {
	if len(keys) == 0 {
		return chars
	}
	result := make([]entity.Characteristic, 0, len(chars))
outer:
	for _, c := range chars {
		for _, k := range keys {
			if strings.EqualFold(c.Key, k) {
				continue outer
			}
		}
		result = append(result, c)
	}
	return result
}

func findTreeNode(nodes []usecase.TreeNodeDTO, id string) *usecase.TreeNodeDTO {
	for i := range nodes {
		if nodes[i].ID == id {
			return &nodes[i]
		}
		if found := findTreeNode(nodes[i].Children, id); found != nil {
			return found
		}
	}
	return nil
}

// publicErr hides infrastructure details from tool results while keeping business messages.
func publicErr(err error) error {
	return errors.New(apperr.Message(err))
}
