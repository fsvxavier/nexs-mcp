package collection

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/fsvxavier/nexs-mcp/internal/collection/sources"
	"gopkg.in/yaml.v3"
)

func TestManager_CheckUpdate(t *testing.T) {
	ctx := context.Background()

	// Create temp directories
	tmpDir := t.TempDir()
	installDir := filepath.Join(tmpDir, "collections")
	sourceDir := filepath.Join(tmpDir, "source")

	// Create test collection
	createTestCollection(t, sourceDir, "test-author", "test-collection", "1.0.0")

	// Create registry and installer
	registry := NewRegistry()
	localSource := createLocalSource(t, sourceDir)
	registry.AddSource(localSource)

	installer, err := NewInstaller(registry, installDir)
	if err != nil {
		t.Fatalf("Failed to create installer: %v", err)
	}

	// Install collection
	uri := "file://" + filepath.Join(sourceDir, "test-author", "test-collection")
	if err := installer.Install(ctx, uri, nil); err != nil {
		t.Fatalf("Failed to install collection: %v", err)
	}

	// Create manager
	manager := NewManager(installer, registry)

	// Test: Check update (no update available)
	result, err := manager.CheckUpdate(ctx, "test-author/test-collection")
	if err != nil {
		t.Fatalf("CheckUpdate failed: %v", err)
	}

	if result.UpdateAvailable {
		t.Errorf("Expected no update available, got update available")
	}

	// Update source collection to new version
	createTestCollection(t, sourceDir, "test-author", "test-collection", "2.0.0")

	// Test: Check update (update available)
	result, err = manager.CheckUpdate(ctx, "test-author/test-collection")
	if err != nil {
		t.Fatalf("CheckUpdate failed: %v", err)
	}

	if !result.UpdateAvailable {
		t.Errorf("Expected update available, got no update")
	}

	if result.OldVersion != "1.0.0" {
		t.Errorf("Expected old version 1.0.0, got %s", result.OldVersion)
	}

	if result.NewVersion != "2.0.0" {
		t.Errorf("Expected new version 2.0.0, got %s", result.NewVersion)
	}
}

func TestManager_Update(t *testing.T) {
	ctx := context.Background()

	// Create temp directories
	tmpDir := t.TempDir()
	installDir := filepath.Join(tmpDir, "collections")
	sourceDir := filepath.Join(tmpDir, "source")

	// Create test collection v1.0.0
	createTestCollection(t, sourceDir, "test-author", "test-collection", "1.0.0")

	// Create registry and installer
	registry := NewRegistry()
	localSource := createLocalSource(t, sourceDir)
	registry.AddSource(localSource)

	installer, err := NewInstaller(registry, installDir)
	if err != nil {
		t.Fatalf("Failed to create installer: %v", err)
	}

	// Install collection v1.0.0
	uri := "file://" + filepath.Join(sourceDir, "test-author", "test-collection")
	if err := installer.Install(ctx, uri, nil); err != nil {
		t.Fatalf("Failed to install collection: %v", err)
	}

	// Update source to v2.0.0
	createTestCollection(t, sourceDir, "test-author", "test-collection", "2.0.0")

	// Create manager
	manager := NewManager(installer, registry)

	// Test: Update collection
	result, err := manager.Update(ctx, "test-author/test-collection", nil)
	if err != nil {
		t.Fatalf("Update failed: %v", err)
	}

	if !result.Updated {
		t.Errorf("Expected collection to be updated")
	}

	if result.OldVersion != "1.0.0" || result.NewVersion != "2.0.0" {
		t.Errorf("Expected update from 1.0.0 to 2.0.0, got %s to %s", result.OldVersion, result.NewVersion)
	}

	// Verify installed version
	record, exists := installer.GetInstalled("test-author/test-collection")
	if !exists {
		t.Fatalf("Collection not found after update")
	}

	if record.Version != "2.0.0" {
		t.Errorf("Expected version 2.0.0, got %s", record.Version)
	}
}

func TestManager_Export(t *testing.T) {
	ctx := context.Background()

	// Create temp directories
	tmpDir := t.TempDir()
	installDir := filepath.Join(tmpDir, "collections")
	sourceDir := filepath.Join(tmpDir, "source")
	outputPath := filepath.Join(tmpDir, "export.tar.gz")

	// Create test collection
	createTestCollection(t, sourceDir, "test-author", "test-collection", "1.0.0")

	// Create registry and installer
	registry := NewRegistry()
	localSource := createLocalSource(t, sourceDir)
	registry.AddSource(localSource)

	installer, err := NewInstaller(registry, installDir)
	if err != nil {
		t.Fatalf("Failed to create installer: %v", err)
	}

	// Install collection
	uri := "file://" + filepath.Join(sourceDir, "test-author", "test-collection")
	if err := installer.Install(ctx, uri, nil); err != nil {
		t.Fatalf("Failed to install collection: %v", err)
	}

	// Create manager
	manager := NewManager(installer, registry)

	// Test: Export collection
	if err := manager.Export(ctx, "test-author/test-collection", outputPath, nil); err != nil {
		t.Fatalf("Export failed: %v", err)
	}

	// Verify export file exists
	if _, err := os.Stat(outputPath); os.IsNotExist(err) {
		t.Errorf("Export file not created: %s", outputPath)
	}

	// Verify file size > 0
	info, err := os.Stat(outputPath)
	if err != nil {
		t.Fatalf("Failed to stat export file: %v", err)
	}

	if info.Size() == 0 {
		t.Errorf("Export file is empty")
	}
}

func TestManager_CheckUpdates(t *testing.T) {
	ctx := context.Background()

	// Create temp directories
	tmpDir := t.TempDir()
	installDir := filepath.Join(tmpDir, "collections")
	sourceDir := filepath.Join(tmpDir, "source")

	// Create multiple test collections
	createTestCollection(t, sourceDir, "author1", "collection1", "1.0.0")
	createTestCollection(t, sourceDir, "author2", "collection2", "1.0.0")

	// Create registry and installer
	registry := NewRegistry()
	localSource := createLocalSource(t, sourceDir)
	registry.AddSource(localSource)

	installer, err := NewInstaller(registry, installDir)
	if err != nil {
		t.Fatalf("Failed to create installer: %v", err)
	}

	// Install collections
	uri1 := "file://" + filepath.Join(sourceDir, "author1", "collection1")
	uri2 := "file://" + filepath.Join(sourceDir, "author2", "collection2")

	if err := installer.Install(ctx, uri1, nil); err != nil {
		t.Fatalf("Failed to install collection1: %v", err)
	}
	if err := installer.Install(ctx, uri2, nil); err != nil {
		t.Fatalf("Failed to install collection2: %v", err)
	}

	// Update collection1 to v2.0.0 (update available)
	createTestCollection(t, sourceDir, "author1", "collection1", "2.0.0")

	// Create manager
	manager := NewManager(installer, registry)

	// Test: Check all updates
	results, err := manager.CheckUpdates(ctx)
	if err != nil {
		t.Fatalf("CheckUpdates failed: %v", err)
	}

	if len(results) != 2 {
		t.Errorf("Expected 2 results, got %d", len(results))
	}

	// Verify one has update available
	updateCount := 0
	for _, result := range results {
		if result.UpdateAvailable {
			updateCount++
		}
	}

	if updateCount != 1 {
		t.Errorf("Expected 1 update available, got %d", updateCount)
	}
}

func TestManager_UpdateAll(t *testing.T) {
	ctx := context.Background()

	// Create temp directories
	tmpDir := t.TempDir()
	installDir := filepath.Join(tmpDir, "collections")
	sourceDir := filepath.Join(tmpDir, "source")

	// Create multiple test collections
	createTestCollection(t, sourceDir, "author1", "collection1", "1.0.0")
	createTestCollection(t, sourceDir, "author2", "collection2", "1.0.0")

	// Create registry and installer
	registry := NewRegistry()
	localSource := createLocalSource(t, sourceDir)
	registry.AddSource(localSource)

	installer, err := NewInstaller(registry, installDir)
	if err != nil {
		t.Fatalf("Failed to create installer: %v", err)
	}

	// Install collections
	uri1 := "file://" + filepath.Join(sourceDir, "author1", "collection1")
	uri2 := "file://" + filepath.Join(sourceDir, "author2", "collection2")

	if err := installer.Install(ctx, uri1, nil); err != nil {
		t.Fatalf("Failed to install collection1: %v", err)
	}
	if err := installer.Install(ctx, uri2, nil); err != nil {
		t.Fatalf("Failed to install collection2: %v", err)
	}

	// Update both collections in source
	createTestCollection(t, sourceDir, "author1", "collection1", "2.0.0")
	createTestCollection(t, sourceDir, "author2", "collection2", "2.0.0")

	// Create manager
	manager := NewManager(installer, registry)

	// Test: Update all collections
	results, err := manager.UpdateAll(ctx, nil)
	if err != nil {
		t.Fatalf("UpdateAll failed: %v", err)
	}

	if len(results) != 2 {
		t.Errorf("Expected 2 results, got %d", len(results))
	}

	// Verify both were updated
	updatedCount := 0
	for _, result := range results {
		if result.Updated {
			updatedCount++
		}
	}

	if updatedCount != 2 {
		t.Errorf("Expected 2 collections updated, got %d", updatedCount)
	}

	// Verify installed versions
	record1, _ := installer.GetInstalled("author1/collection1")
	if record1.Version != "2.0.0" {
		t.Errorf("Expected collection1 version 2.0.0, got %s", record1.Version)
	}

	record2, _ := installer.GetInstalled("author2/collection2")
	if record2.Version != "2.0.0" {
		t.Errorf("Expected collection2 version 2.0.0, got %s", record2.Version)
	}
}

func TestManager_ExportWithOptions(t *testing.T) {
	ctx := context.Background()

	// Create temp directories
	tmpDir := t.TempDir()
	installDir := filepath.Join(tmpDir, "collections")
	sourceDir := filepath.Join(tmpDir, "source")

	// Create test collection
	createTestCollection(t, sourceDir, "test-author", "test-collection", "1.0.0")

	// Create registry and installer
	registry := NewRegistry()
	localSource := createLocalSource(t, sourceDir)
	registry.AddSource(localSource)

	installer, err := NewInstaller(registry, installDir)
	if err != nil {
		t.Fatalf("Failed to create installer: %v", err)
	}

	// Install collection
	uri := "file://" + filepath.Join(sourceDir, "test-author", "test-collection")
	if err := installer.Install(ctx, uri, nil); err != nil {
		t.Fatalf("Failed to install collection: %v", err)
	}

	// Create manager
	manager := NewManager(installer, registry)

	tests := []struct {
		name        string
		options     *ExportOptions
		expectError bool
	}{
		{
			name:        "Default options",
			options:     nil,
			expectError: false,
		},
		{
			name: "Fast compression",
			options: &ExportOptions{
				Compression: "fast",
			},
			expectError: false,
		},
		{
			name: "No compression",
			options: &ExportOptions{
				Compression: "none",
			},
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			outputPath := filepath.Join(tmpDir, "export-"+tt.name+".tar.gz")
			err := manager.Export(ctx, "test-author/test-collection", outputPath, tt.options)

			if tt.expectError && err == nil {
				t.Errorf("Expected error, got nil")
			}
			if !tt.expectError && err != nil {
				t.Errorf("Unexpected error: %v", err)
			}

			if !tt.expectError {
				// Verify file exists
				if _, err := os.Stat(outputPath); os.IsNotExist(err) {
					t.Errorf("Export file not created")
				}
			}
		})
	}
}

// Helper functions

func createTestCollection(t *testing.T, baseDir, author, name, version string) {
	t.Helper()

	collectionDir := filepath.Join(baseDir, author, name)
	if err := os.MkdirAll(collectionDir, 0755); err != nil {
		t.Fatalf("Failed to create collection directory: %v", err)
	}

	// Create manifest
	manifest := &Manifest{
		Name:        name,
		Version:     version,
		Author:      author,
		Description: "Test collection",
		Elements: []Element{
			{Path: "test.yaml", Type: "persona"},
		},
	}

	manifestData, err := yaml.Marshal(manifest)
	if err != nil {
		t.Fatalf("Failed to marshal manifest: %v", err)
	}

	manifestPath := filepath.Join(collectionDir, "collection.yaml")
	if err := os.WriteFile(manifestPath, manifestData, 0644); err != nil {
		t.Fatalf("Failed to write manifest: %v", err)
	}

	// Create test element
	testElement := filepath.Join(collectionDir, "test.yaml")
	if err := os.WriteFile(testElement, []byte("name: test\ntype: persona\n"), 0644); err != nil {
		t.Fatalf("Failed to write test element: %v", err)
	}
}

func createLocalSource(t *testing.T, baseDir string) sources.CollectionSource {
	t.Helper()

	// Use the mock from registry_test.go
	return &testLocalSource{
		name:    "local",
		baseDir: baseDir,
	}
}

// testLocalSource for testing (simplified mock)
type testLocalSource struct {
	name    string
	baseDir string
}

func (m *testLocalSource) Name() string {
	return m.name
}

func (m *testLocalSource) Browse(ctx context.Context, filter *sources.BrowseFilter) ([]*sources.CollectionMetadata, error) {
	return nil, nil
}

func (m *testLocalSource) Get(ctx context.Context, uri string) (*sources.Collection, error) {
	// Parse file:// URI
	path := uri
	if filepath.IsAbs(path) || strings.HasPrefix(uri, "file://") {
		path = strings.TrimPrefix(uri, "file://")
	}

	// Read manifest
	manifestPath := filepath.Join(path, "collection.yaml")
	manifestData, err := os.ReadFile(manifestPath)
	if err != nil {
		return nil, err
	}

	var manifest map[string]interface{}
	if err := yaml.Unmarshal(manifestData, &manifest); err != nil {
		return nil, err
	}

	metadata := &sources.CollectionMetadata{
		Name:       manifest["name"].(string),
		Version:    manifest["version"].(string),
		Author:     manifest["author"].(string),
		SourceName: m.name,
		URI:        uri,
	}

	return &sources.Collection{
		Metadata:   metadata,
		Manifest:   manifest,
		SourceData: map[string]interface{}{"path": path},
	}, nil
}

func (m *testLocalSource) Supports(uri string) bool {
	return strings.HasPrefix(uri, "file://") || filepath.IsAbs(uri)
}
