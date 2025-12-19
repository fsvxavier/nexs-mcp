package mcp

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/fsvxavier/nexs-mcp/internal/collection"
	"github.com/fsvxavier/nexs-mcp/internal/collection/sources"
)

// Mock source for testing
type MockToolSource struct {
	name        string
	collections map[string]*sources.Collection
	browseFunc  func(ctx context.Context, filter *sources.BrowseFilter) ([]*sources.CollectionMetadata, error)
}

func (m *MockToolSource) Name() string {
	return m.name
}

func (m *MockToolSource) Supports(uri string) bool {
	_, exists := m.collections[uri]
	return exists
}

func (m *MockToolSource) Browse(ctx context.Context, filter *sources.BrowseFilter) ([]*sources.CollectionMetadata, error) {
	if m.browseFunc != nil {
		return m.browseFunc(ctx, filter)
	}
	var results []*sources.CollectionMetadata
	for _, coll := range m.collections {
		// Apply author filter
		if filter != nil && filter.Author != "" && coll.Metadata.Author != filter.Author {
			continue
		}
		results = append(results, coll.Metadata)
	}
	return results, nil
}

func (m *MockToolSource) Get(ctx context.Context, uri string) (*sources.Collection, error) {
	coll, exists := m.collections[uri]
	if !exists {
		return nil, os.ErrNotExist
	}
	return coll, nil
}

// Helper to create test collection for MCP tests
func createMCPTestCollection(t *testing.T, tempDir, name, author, version string) (string, *sources.Collection) {
	t.Helper()

	collectionDir := filepath.Join(tempDir, author, name)
	if err := os.MkdirAll(collectionDir, 0755); err != nil {
		t.Fatal(err)
	}

	// Create skill file
	skillsDir := filepath.Join(collectionDir, "skills")
	if err := os.MkdirAll(skillsDir, 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(skillsDir, "test.skill"), []byte("test skill"), 0644); err != nil {
		t.Fatal(err)
	}

	// Create manifest file
	manifestPath := filepath.Join(collectionDir, "collection.yaml")
	manifestContent := "name: " + name + "\nversion: " + version + "\nauthor: " + author + "\ndescription: Test collection\ncategory: testing\n"
	if err := os.WriteFile(manifestPath, []byte(manifestContent), 0644); err != nil {
		t.Fatal(err)
	}

	collection := &sources.Collection{
		Metadata: &sources.CollectionMetadata{
			Name:        name,
			Version:     version,
			Author:      author,
			Description: "Test collection",
			Category:    "testing",
			SourceName:  "mock",
			URI:         "mock://" + author + "/" + name,
		},
		Manifest: map[string]interface{}{
			"name":        name,
			"version":     version,
			"author":      author,
			"description": "Test collection",
			"category":    "testing",
			"elements": []interface{}{
				map[string]interface{}{
					"type": "skill",
					"path": "skills/test.skill",
				},
			},
		},
		SourceData: map[string]interface{}{
			"path": collectionDir,
		},
	}

	return collectionDir, collection
}

func TestCollectionTools_ToolDefinitions(t *testing.T) {
	registry := collection.NewRegistry()
	installer, _ := collection.NewInstaller(registry, t.TempDir())
	tools := NewCollectionTools(registry, installer)

	defs := tools.ToolDefinitions()
	if len(defs) != 6 {
		t.Errorf("Expected 6 tool definitions, got %d", len(defs))
	}

	expectedTools := map[string]bool{
		"browse_collections":         false,
		"install_collection":         false,
		"uninstall_collection":       false,
		"list_installed_collections": false,
		"get_collection_info":        false,
		"export_collection":          false,
	}

	for _, def := range defs {
		if _, exists := expectedTools[def.Name]; !exists {
			t.Errorf("Unexpected tool: %s", def.Name)
		}
		expectedTools[def.Name] = true

		// Check for description
		if def.Description == "" {
			t.Errorf("Tool %s missing description", def.Name)
		}

		// Check for input schema
		if def.InputSchema == nil {
			t.Errorf("Tool %s missing input schema", def.Name)
		}
	}

	for name, found := range expectedTools {
		if !found {
			t.Errorf("Missing tool definition: %s", name)
		}
	}
}

func TestCollectionTools_BrowseCollections(t *testing.T) {
	registry := collection.NewRegistry()
	installer, _ := collection.NewInstaller(registry, t.TempDir())

	// Create mock source
	_, coll1 := createMCPTestCollection(t, t.TempDir(), "collection1", "author1", "1.0.0")
	_, coll2 := createMCPTestCollection(t, t.TempDir(), "collection2", "author2", "1.0.0")

	mockSource := &MockToolSource{
		name: "mock",
		collections: map[string]*sources.Collection{
			"mock://author1/collection1": coll1,
			"mock://author2/collection2": coll2,
		},
	}
	registry.AddSource(mockSource)

	tools := NewCollectionTools(registry, installer)
	ctx := context.Background()

	// Test: Browse all
	args := map[string]interface{}{}
	result, err := tools.handleBrowse(ctx, args)
	if err != nil {
		t.Fatalf("Browse failed: %v", err)
	}

	resultMap, ok := result.(map[string]interface{})
	if !ok {
		t.Fatal("Result is not a map")
	}

	collections, ok := resultMap["collections"].([]*sources.CollectionMetadata)
	if !ok {
		t.Fatal("collections field is not []*CollectionMetadata")
	}

	if len(collections) != 2 {
		t.Errorf("Expected 2 collections, got %d", len(collections))
	}

	// Test: Browse with filter
	args = map[string]interface{}{
		"author": "author1",
		"limit":  float64(10),
	}
	result, err = tools.handleBrowse(ctx, args)
	if err != nil {
		t.Fatalf("Browse with filter failed: %v", err)
	}

	resultMap = result.(map[string]interface{})
	collections = resultMap["collections"].([]*sources.CollectionMetadata)

	if len(collections) != 1 {
		t.Errorf("Expected 1 collection with filter, got %d", len(collections))
	}
	if collections[0].Author != "author1" {
		t.Errorf("Expected author1, got %s", collections[0].Author)
	}
}

func TestCollectionTools_InstallCollection(t *testing.T) {
	tempDir := t.TempDir()
	sourceDir := filepath.Join(tempDir, "source")
	installDir := filepath.Join(tempDir, "install")

	_, coll := createMCPTestCollection(t, sourceDir, "test-collection", "test-author", "1.0.0")

	registry := collection.NewRegistry()
	mockSource := &MockToolSource{
		name: "mock",
		collections: map[string]*sources.Collection{
			"mock://test-author/test-collection": coll,
		},
	}
	registry.AddSource(mockSource)

	installer, _ := collection.NewInstaller(registry, installDir)
	tools := NewCollectionTools(registry, installer)
	ctx := context.Background()

	// Test: Install collection
	args := map[string]interface{}{
		"uri":               "mock://test-author/test-collection",
		"skip_validation":   true,
		"skip_hooks":        true,
		"skip_dependencies": true,
	}

	result, err := tools.handleInstall(ctx, args)
	if err != nil {
		t.Fatalf("Install failed: %v", err)
	}

	resultMap, ok := result.(map[string]interface{})
	if !ok {
		t.Fatal("Result is not a map")
	}

	status, ok := resultMap["status"].(string)
	if !ok || status != "installed" {
		t.Errorf("Expected status 'installed', got %v", resultMap["status"])
	}

	// Verify installation
	if _, exists := installer.GetInstalled("test-author/test-collection"); !exists {
		t.Error("Collection not installed")
	}
}

func TestCollectionTools_InstallCollection_MissingURI(t *testing.T) {
	registry := collection.NewRegistry()
	installer, _ := collection.NewInstaller(registry, t.TempDir())
	tools := NewCollectionTools(registry, installer)
	ctx := context.Background()

	args := map[string]interface{}{}
	_, err := tools.handleInstall(ctx, args)
	if err == nil {
		t.Error("Expected error for missing URI")
	}
}

func TestCollectionTools_UninstallCollection(t *testing.T) {
	tempDir := t.TempDir()
	sourceDir := filepath.Join(tempDir, "source")
	installDir := filepath.Join(tempDir, "install")

	_, coll := createMCPTestCollection(t, sourceDir, "test-collection", "test-author", "1.0.0")

	registry := collection.NewRegistry()
	mockSource := &MockToolSource{
		name: "mock",
		collections: map[string]*sources.Collection{
			"mock://test-author/test-collection": coll,
		},
	}
	registry.AddSource(mockSource)

	installer, _ := collection.NewInstaller(registry, installDir)
	tools := NewCollectionTools(registry, installer)
	ctx := context.Background()

	// Install first
	installArgs := map[string]interface{}{
		"uri":               "mock://test-author/test-collection",
		"skip_validation":   true,
		"skip_hooks":        true,
		"skip_dependencies": true,
	}
	if _, err := tools.handleInstall(ctx, installArgs); err != nil {
		t.Fatalf("Install failed: %v", err)
	}

	// Test: Uninstall collection
	args := map[string]interface{}{
		"id": "test-author/test-collection",
	}

	result, err := tools.handleUninstall(ctx, args)
	if err != nil {
		t.Fatalf("Uninstall failed: %v", err)
	}

	resultMap, ok := result.(map[string]interface{})
	if !ok {
		t.Fatal("Result is not a map")
	}

	status, ok := resultMap["status"].(string)
	if !ok || status != "uninstalled" {
		t.Errorf("Expected status 'uninstalled', got %v", resultMap["status"])
	}

	// Verify uninstallation
	if _, exists := installer.GetInstalled("test-author/test-collection"); exists {
		t.Error("Collection still installed")
	}
}

func TestCollectionTools_ListInstalled(t *testing.T) {
	tempDir := t.TempDir()
	sourceDir := filepath.Join(tempDir, "source")
	installDir := filepath.Join(tempDir, "install")

	_, coll1 := createMCPTestCollection(t, sourceDir, "collection1", "author1", "1.0.0")
	_, coll2 := createMCPTestCollection(t, sourceDir, "collection2", "author2", "1.0.0")

	registry := collection.NewRegistry()
	mockSource := &MockToolSource{
		name: "mock",
		collections: map[string]*sources.Collection{
			"mock://author1/collection1": coll1,
			"mock://author2/collection2": coll2,
		},
	}
	registry.AddSource(mockSource)

	installer, _ := collection.NewInstaller(registry, installDir)
	tools := NewCollectionTools(registry, installer)
	ctx := context.Background()

	// Install both collections
	for _, uri := range []string{"mock://author1/collection1", "mock://author2/collection2"} {
		args := map[string]interface{}{
			"uri":               uri,
			"skip_validation":   true,
			"skip_hooks":        true,
			"skip_dependencies": true,
		}
		if _, err := tools.handleInstall(ctx, args); err != nil {
			t.Fatalf("Install failed: %v", err)
		}
	}

	// Test: List installed
	args := map[string]interface{}{}
	result, err := tools.handleListInstalled(ctx, args)
	if err != nil {
		t.Fatalf("List installed failed: %v", err)
	}

	resultMap, ok := result.(map[string]interface{})
	if !ok {
		t.Fatal("Result is not a map")
	}

	count, ok := resultMap["count"].(int)
	if !ok || count != 2 {
		t.Errorf("Expected count 2, got %v", resultMap["count"])
	}
}

func TestCollectionTools_GetCollectionInfo_Installed(t *testing.T) {
	tempDir := t.TempDir()
	sourceDir := filepath.Join(tempDir, "source")
	installDir := filepath.Join(tempDir, "install")

	_, coll := createMCPTestCollection(t, sourceDir, "test-collection", "test-author", "1.0.0")

	registry := collection.NewRegistry()
	mockSource := &MockToolSource{
		name: "mock",
		collections: map[string]*sources.Collection{
			"mock://test-author/test-collection": coll,
		},
	}
	registry.AddSource(mockSource)

	installer, _ := collection.NewInstaller(registry, installDir)
	tools := NewCollectionTools(registry, installer)
	ctx := context.Background()

	// Install collection
	installArgs := map[string]interface{}{
		"uri":               "mock://test-author/test-collection",
		"skip_validation":   true,
		"skip_hooks":        true,
		"skip_dependencies": true,
	}
	if _, err := tools.handleInstall(ctx, installArgs); err != nil {
		t.Fatalf("Install failed: %v", err)
	}

	// Test: Get info for installed collection
	args := map[string]interface{}{
		"uri": "test-author/test-collection",
	}

	result, err := tools.handleGetInfo(ctx, args)
	if err != nil {
		t.Fatalf("Get info failed: %v", err)
	}

	resultMap, ok := result.(map[string]interface{})
	if !ok {
		t.Fatal("Result is not a map")
	}

	source, ok := resultMap["source"].(string)
	if !ok || source != "installed" {
		t.Errorf("Expected source 'installed', got %v", resultMap["source"])
	}
}

func TestCollectionTools_GetCollectionInfo_Registry(t *testing.T) {
	tempDir := t.TempDir()
	sourceDir := filepath.Join(tempDir, "source")

	_, coll := createMCPTestCollection(t, sourceDir, "test-collection", "test-author", "1.0.0")

	registry := collection.NewRegistry()
	mockSource := &MockToolSource{
		name: "mock",
		collections: map[string]*sources.Collection{
			"mock://test-author/test-collection": coll,
		},
	}
	registry.AddSource(mockSource)

	installer, _ := collection.NewInstaller(registry, t.TempDir())
	tools := NewCollectionTools(registry, installer)
	ctx := context.Background()

	// Test: Get info from registry (not installed)
	args := map[string]interface{}{
		"uri": "mock://test-author/test-collection",
	}

	result, err := tools.handleGetInfo(ctx, args)
	if err != nil {
		t.Fatalf("Get info failed: %v", err)
	}

	resultMap, ok := result.(map[string]interface{})
	if !ok {
		t.Fatal("Result is not a map")
	}

	source, ok := resultMap["source"].(string)
	if !ok || source != "registry" {
		t.Errorf("Expected source 'registry', got %v", resultMap["source"])
	}
}

func TestCollectionTools_ExportCollection(t *testing.T) {
	tempDir := t.TempDir()
	sourceDir := filepath.Join(tempDir, "source")
	installDir := filepath.Join(tempDir, "install")
	outputFile := filepath.Join(tempDir, "export.tar.gz")

	_, coll := createMCPTestCollection(t, sourceDir, "test-collection", "test-author", "1.0.0")

	registry := collection.NewRegistry()
	mockSource := &MockToolSource{
		name: "mock",
		collections: map[string]*sources.Collection{
			"mock://test-author/test-collection": coll,
		},
	}
	registry.AddSource(mockSource)

	installer, _ := collection.NewInstaller(registry, installDir)
	tools := NewCollectionTools(registry, installer)
	ctx := context.Background()

	// Install collection
	installArgs := map[string]interface{}{
		"uri":               "mock://test-author/test-collection",
		"skip_validation":   true,
		"skip_hooks":        true,
		"skip_dependencies": true,
	}
	if _, err := tools.handleInstall(ctx, installArgs); err != nil {
		t.Fatalf("Install failed: %v", err)
	}

	// Test: Export collection
	args := map[string]interface{}{
		"id":     "test-author/test-collection",
		"output": outputFile,
	}

	result, err := tools.handleExport(ctx, args)
	if err != nil {
		t.Fatalf("Export failed: %v", err)
	}

	resultMap, ok := result.(map[string]interface{})
	if !ok {
		t.Fatal("Result is not a map")
	}

	status, ok := resultMap["status"].(string)
	if !ok || status != "exported" {
		t.Errorf("Expected status 'exported', got %v", resultMap["status"])
	}

	// Verify export file exists
	if _, err := os.Stat(outputFile); os.IsNotExist(err) {
		t.Error("Export file was not created")
	}
}

func TestCollectionTools_HandleTool_UnknownTool(t *testing.T) {
	registry := collection.NewRegistry()
	installer, _ := collection.NewInstaller(registry, t.TempDir())
	tools := NewCollectionTools(registry, installer)
	ctx := context.Background()

	_, err := tools.HandleTool(ctx, "unknown_tool", map[string]interface{}{})
	if err == nil {
		t.Error("Expected error for unknown tool")
	}
}

func TestCollectionTools_ParseManifestFromMap(t *testing.T) {
	manifestMap := map[string]interface{}{
		"name":        "test",
		"version":     "1.0.0",
		"author":      "test-author",
		"description": "Test manifest",
		"category":    "testing",
	}

	manifest, err := parseManifestFromMap(manifestMap)
	if err != nil {
		t.Fatalf("Parse manifest failed: %v", err)
	}

	if manifest.Name != "test" {
		t.Errorf("Expected name 'test', got %s", manifest.Name)
	}
	if manifest.Version != "1.0.0" {
		t.Errorf("Expected version '1.0.0', got %s", manifest.Version)
	}
	if manifest.Author != "test-author" {
		t.Errorf("Expected author 'test-author', got %s", manifest.Author)
	}
}

func TestCollectionTools_ToolDefinitions_Schema(t *testing.T) {
	registry := collection.NewRegistry()
	installer, _ := collection.NewInstaller(registry, t.TempDir())
	tools := NewCollectionTools(registry, installer)

	defs := tools.ToolDefinitions()

	// Test install_collection schema
	var installDef *ToolDef
	for i := range defs {
		if defs[i].Name == "install_collection" {
			installDef = &defs[i]
			break
		}
	}

	if installDef == nil {
		t.Fatal("install_collection tool not found")
	}

	// Verify required fields
	schema := installDef.InputSchema
	props, ok := schema["properties"].(map[string]interface{})
	if !ok {
		t.Fatal("properties not found in schema")
	}

	if _, ok := props["uri"]; !ok {
		t.Error("uri property missing")
	}

	required, ok := schema["required"].([]string)
	if !ok {
		t.Fatal("required field not found or wrong type")
	}

	if len(required) != 1 || required[0] != "uri" {
		t.Errorf("Expected required: [uri], got %v", required)
	}
}

func TestParseManifestFromMap(t *testing.T) {
	tests := []struct {
		name    string
		input   map[string]interface{}
		wantErr bool
	}{
		{
			name: "valid manifest",
			input: map[string]interface{}{
				"name":        "test",
				"version":     "1.0.0",
				"author":      "author",
				"description": "desc",
			},
			wantErr: false,
		},
		{
			name: "with elements",
			input: map[string]interface{}{
				"name":    "test",
				"version": "1.0.0",
				"author":  "author",
				"elements": []interface{}{
					map[string]interface{}{
						"type": "skill",
						"path": "skills/test.skill",
					},
				},
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			manifest, err := parseManifestFromMap(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("parseManifestFromMap() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && manifest == nil {
				t.Error("Expected manifest, got nil")
			}
		})
	}
}

func TestToolDef_JSONMarshaling(t *testing.T) {
	def := ToolDef{
		Name:        "test_tool",
		Description: "Test tool description",
		InputSchema: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"param1": map[string]interface{}{
					"type": "string",
				},
			},
		},
	}

	// Marshal to JSON
	data, err := json.Marshal(def)
	if err != nil {
		t.Fatalf("Marshal failed: %v", err)
	}

	// Unmarshal back
	var unmarshaled ToolDef
	if err := json.Unmarshal(data, &unmarshaled); err != nil {
		t.Fatalf("Unmarshal failed: %v", err)
	}

	if unmarshaled.Name != def.Name {
		t.Errorf("Expected name %s, got %s", def.Name, unmarshaled.Name)
	}
}
