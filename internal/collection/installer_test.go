package collection

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/fsvxavier/nexs-mcp/internal/collection/sources"
)

// Mock source for testing.
type MockInstallSource struct {
	collections map[string]*sources.Collection
}

func (m *MockInstallSource) Name() string {
	return "mock-install"
}

func (m *MockInstallSource) Supports(uri string) bool {
	_, exists := m.collections[uri]
	return exists
}

func (m *MockInstallSource) Browse(ctx context.Context, filter *sources.BrowseFilter) ([]*sources.CollectionMetadata, error) {
	return nil, nil
}

func (m *MockInstallSource) Get(ctx context.Context, uri string) (*sources.Collection, error) {
	collection, exists := m.collections[uri]
	if !exists {
		return nil, os.ErrNotExist
	}
	return collection, nil
}

// Helper to create test collection.
func createTestCollectionForInstall(t *testing.T, tempDir, name, author, version string, deps []Dependency) (string, *sources.Collection) {
	t.Helper()

	// Create collection directory
	collectionDir := filepath.Join(tempDir, author, name)
	if err := os.MkdirAll(collectionDir, 0755); err != nil {
		t.Fatal(err)
	}

	// Create manifest
	manifest := &Manifest{
		Name:         name,
		Version:      version,
		Author:       author,
		Description:  "Test collection",
		Category:     "testing",
		Dependencies: deps,
		Elements: []Element{
			{Type: "skill", Path: "skills/test.skill"},
		},
	}

	// Create manifest file
	manifestPath := filepath.Join(collectionDir, "collection.yaml")
	if err := writeManifestFile(manifestPath, manifest); err != nil {
		t.Fatal(err)
	}

	// Create test skill file
	skillsDir := filepath.Join(collectionDir, "skills")
	if err := os.MkdirAll(skillsDir, 0755); err != nil {
		t.Fatal(err)
	}
	skillPath := filepath.Join(skillsDir, "test.skill")
	if err := os.WriteFile(skillPath, []byte("test skill content"), 0644); err != nil {
		t.Fatal(err)
	}

	// Create collection object
	collection := &sources.Collection{
		Metadata: &sources.CollectionMetadata{
			Name:       name,
			Version:    version,
			Author:     author,
			SourceName: "mock-install",
			URI:        "mock://" + author + "/" + name,
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

	// Add dependencies to manifest map
	if len(deps) > 0 {
		depsInterface := make([]interface{}, len(deps))
		for i, dep := range deps {
			depsInterface[i] = map[string]interface{}{
				"uri":     dep.URI,
				"version": dep.Version,
			}
		}
		collection.Manifest.(map[string]interface{})["dependencies"] = depsInterface
	}

	return collectionDir, collection
}

func writeManifestFile(path string, manifest *Manifest) error {
	// Simple YAML writing for test
	content := "name: " + manifest.Name + "\n"
	content += "version: " + manifest.Version + "\n"
	content += "author: " + manifest.Author + "\n"
	content += "description: " + manifest.Description + "\n"
	content += "category: " + manifest.Category + "\n"
	return os.WriteFile(path, []byte(content), 0644)
}

func TestNewInstaller(t *testing.T) {
	tempDir := t.TempDir()
	registry := NewRegistry()

	installer, err := NewInstaller(registry, tempDir)
	if err != nil {
		t.Fatalf("NewInstaller failed: %v", err)
	}

	if installer.installDir != tempDir {
		t.Errorf("Expected installDir %s, got %s", tempDir, installer.installDir)
	}

	if installer.registry != registry {
		t.Error("Registry not set correctly")
	}

	// Check directory was created
	if _, err := os.Stat(tempDir); os.IsNotExist(err) {
		t.Error("Install directory was not created")
	}
}

func TestInstaller_Install_Simple(t *testing.T) {
	tempDir := t.TempDir()
	sourceDir := filepath.Join(tempDir, "source")
	installDir := filepath.Join(tempDir, "install")

	// Create test collection
	_, collection := createTestCollectionForInstall(t, sourceDir, "test-collection", "test-author", "1.0.0", nil)

	// Create registry and source
	registry := NewRegistry()
	mockSource := &MockInstallSource{
		collections: map[string]*sources.Collection{
			"mock://test-author/test-collection": collection,
		},
	}
	registry.AddSource(mockSource)

	// Create installer
	installer, err := NewInstaller(registry, installDir)
	if err != nil {
		t.Fatalf("NewInstaller failed: %v", err)
	}

	// Install collection
	ctx := context.Background()
	err = installer.Install(ctx, "mock://test-author/test-collection", nil)
	if err != nil {
		t.Fatalf("Install failed: %v", err)
	}

	// Verify installation
	record, exists := installer.GetInstalled("test-author/test-collection")
	if !exists {
		t.Fatal("Collection not recorded as installed")
	}

	if record.Version != "1.0.0" {
		t.Errorf("Expected version 1.0.0, got %s", record.Version)
	}

	// Verify files were copied
	installedPath := filepath.Join(installDir, "test-author", "test-collection")
	if _, err := os.Stat(installedPath); os.IsNotExist(err) {
		t.Error("Collection directory was not created")
	}

	skillPath := filepath.Join(installedPath, "skills", "test.skill")
	if _, err := os.Stat(skillPath); os.IsNotExist(err) {
		t.Error("Skill file was not copied")
	}
}

func TestInstaller_Install_AlreadyInstalled(t *testing.T) {
	tempDir := t.TempDir()
	sourceDir := filepath.Join(tempDir, "source")
	installDir := filepath.Join(tempDir, "install")

	_, collection := createTestCollectionForInstall(t, sourceDir, "test-collection", "test-author", "1.0.0", nil)

	registry := NewRegistry()
	mockSource := &MockInstallSource{
		collections: map[string]*sources.Collection{
			"mock://test-author/test-collection": collection,
		},
	}
	registry.AddSource(mockSource)

	installer, err := NewInstaller(registry, installDir)
	if err != nil {
		t.Fatalf("NewInstaller failed: %v", err)
	}

	ctx := context.Background()

	// Install once
	if err := installer.Install(ctx, "mock://test-author/test-collection", nil); err != nil {
		t.Fatalf("First install failed: %v", err)
	}

	// Try to install again (should fail)
	err = installer.Install(ctx, "mock://test-author/test-collection", nil)
	if err == nil {
		t.Error("Expected error when installing already installed collection")
	}

	// Install with force (should succeed)
	err = installer.Install(ctx, "mock://test-author/test-collection", &InstallOptions{Force: true})
	if err != nil {
		t.Errorf("Force install failed: %v", err)
	}
}

func TestInstaller_Install_WithDependencies(t *testing.T) {
	tempDir := t.TempDir()
	sourceDir := filepath.Join(tempDir, "source")
	installDir := filepath.Join(tempDir, "install")

	// Create dependency collection
	_, depCollection := createTestCollectionForInstall(t, sourceDir, "dependency", "test-author", "1.0.0", nil)

	// Create main collection with dependency
	deps := []Dependency{
		{URI: "mock://test-author/dependency", Version: "^1.0.0"},
	}
	_, mainCollection := createTestCollectionForInstall(t, sourceDir, "main-collection", "test-author", "1.0.0", deps)

	// Setup registry
	registry := NewRegistry()
	mockSource := &MockInstallSource{
		collections: map[string]*sources.Collection{
			"mock://test-author/dependency":      depCollection,
			"mock://test-author/main-collection": mainCollection,
		},
	}
	registry.AddSource(mockSource)

	installer, err := NewInstaller(registry, installDir)
	if err != nil {
		t.Fatalf("NewInstaller failed: %v", err)
	}

	// Install main collection
	ctx := context.Background()
	err = installer.Install(ctx, "mock://test-author/main-collection", &InstallOptions{SkipValidation: true})
	if err != nil {
		t.Fatalf("Install failed: %v", err)
	}

	// Verify both collections are installed
	if _, exists := installer.GetInstalled("test-author/dependency"); !exists {
		t.Error("Dependency not installed")
	}

	if _, exists := installer.GetInstalled("test-author/main-collection"); !exists {
		t.Error("Main collection not installed")
	}
}

func TestInstaller_Uninstall(t *testing.T) {
	tempDir := t.TempDir()
	sourceDir := filepath.Join(tempDir, "source")
	installDir := filepath.Join(tempDir, "install")

	_, collection := createTestCollectionForInstall(t, sourceDir, "test-collection", "test-author", "1.0.0", nil)

	registry := NewRegistry()
	mockSource := &MockInstallSource{
		collections: map[string]*sources.Collection{
			"mock://test-author/test-collection": collection,
		},
	}
	registry.AddSource(mockSource)

	installer, err := NewInstaller(registry, installDir)
	if err != nil {
		t.Fatalf("NewInstaller failed: %v", err)
	}

	ctx := context.Background()

	// Install
	if err := installer.Install(ctx, "mock://test-author/test-collection", nil); err != nil {
		t.Fatalf("Install failed: %v", err)
	}

	// Verify installed
	if _, exists := installer.GetInstalled("test-author/test-collection"); !exists {
		t.Fatal("Collection not installed")
	}

	// Uninstall
	err = installer.Uninstall(ctx, "test-author/test-collection", nil)
	if err != nil {
		t.Fatalf("Uninstall failed: %v", err)
	}

	// Verify not installed
	if _, exists := installer.GetInstalled("test-author/test-collection"); exists {
		t.Error("Collection still listed as installed")
	}

	// Verify files removed
	installedPath := filepath.Join(installDir, "test-author", "test-collection")
	if _, err := os.Stat(installedPath); !os.IsNotExist(err) {
		t.Error("Collection directory still exists")
	}
}

func TestInstaller_Uninstall_WithDependents(t *testing.T) {
	tempDir := t.TempDir()
	sourceDir := filepath.Join(tempDir, "source")
	installDir := filepath.Join(tempDir, "install")

	// Create dependency
	_, depCollection := createTestCollectionForInstall(t, sourceDir, "dependency", "test-author", "1.0.0", nil)

	// Create dependent collection
	deps := []Dependency{
		{URI: "mock://test-author/dependency", Version: "^1.0.0"},
	}
	_, mainCollection := createTestCollectionForInstall(t, sourceDir, "main-collection", "test-author", "1.0.0", deps)

	registry := NewRegistry()
	mockSource := &MockInstallSource{
		collections: map[string]*sources.Collection{
			"mock://test-author/dependency":      depCollection,
			"mock://test-author/main-collection": mainCollection,
		},
	}
	registry.AddSource(mockSource)

	installer, err := NewInstaller(registry, installDir)
	if err != nil {
		t.Fatalf("NewInstaller failed: %v", err)
	}

	ctx := context.Background()

	// Install both
	if err := installer.Install(ctx, "mock://test-author/main-collection", &InstallOptions{SkipValidation: true}); err != nil {
		t.Fatalf("Install failed: %v", err)
	}

	// Verify dependency was installed
	if _, exists := installer.GetInstalled("test-author/dependency"); !exists {
		t.Fatal("Dependency was not installed")
	}

	// Try to uninstall dependency (should fail)
	err = installer.Uninstall(ctx, "test-author/dependency", nil)
	if err == nil {
		t.Error("Expected error when uninstalling dependency with dependents")
	}

	// Force uninstall (should succeed)
	err = installer.Uninstall(ctx, "test-author/dependency", &UninstallOptions{Force: true})
	if err != nil {
		t.Errorf("Force uninstall failed: %v", err)
	}
}

func TestInstaller_ListInstalled(t *testing.T) {
	tempDir := t.TempDir()
	sourceDir := filepath.Join(tempDir, "source")
	installDir := filepath.Join(tempDir, "install")

	// Create multiple collections
	_, col1 := createTestCollectionForInstall(t, sourceDir, "collection1", "author1", "1.0.0", nil)
	_, col2 := createTestCollectionForInstall(t, sourceDir, "collection2", "author2", "1.0.0", nil)

	registry := NewRegistry()
	mockSource := &MockInstallSource{
		collections: map[string]*sources.Collection{
			"mock://author1/collection1": col1,
			"mock://author2/collection2": col2,
		},
	}
	registry.AddSource(mockSource)

	installer, err := NewInstaller(registry, installDir)
	if err != nil {
		t.Fatalf("NewInstaller failed: %v", err)
	}

	ctx := context.Background()

	// Initially empty
	if len(installer.ListInstalled()) != 0 {
		t.Error("Expected empty list initially")
	}

	// Install collections
	if err := installer.Install(ctx, "mock://author1/collection1", nil); err != nil {
		t.Fatalf("Install 1 failed: %v", err)
	}
	if err := installer.Install(ctx, "mock://author2/collection2", nil); err != nil {
		t.Fatalf("Install 2 failed: %v", err)
	}

	// Check list
	installed := installer.ListInstalled()
	if len(installed) != 2 {
		t.Errorf("Expected 2 installed collections, got %d", len(installed))
	}
}

func TestInstaller_StatePersistence(t *testing.T) {
	tempDir := t.TempDir()
	sourceDir := filepath.Join(tempDir, "source")
	installDir := filepath.Join(tempDir, "install")

	_, collection := createTestCollectionForInstall(t, sourceDir, "test-collection", "test-author", "1.0.0", nil)

	registry := NewRegistry()
	mockSource := &MockInstallSource{
		collections: map[string]*sources.Collection{
			"mock://test-author/test-collection": collection,
		},
	}
	registry.AddSource(mockSource)

	// Create installer and install
	installer1, err := NewInstaller(registry, installDir)
	if err != nil {
		t.Fatalf("NewInstaller failed: %v", err)
	}

	ctx := context.Background()
	if err := installer1.Install(ctx, "mock://test-author/test-collection", nil); err != nil {
		t.Fatalf("Install failed: %v", err)
	}

	// Create new installer instance (simulating restart)
	installer2, err := NewInstaller(registry, installDir)
	if err != nil {
		t.Fatalf("Second NewInstaller failed: %v", err)
	}

	// Verify state was loaded
	if _, exists := installer2.GetInstalled("test-author/test-collection"); !exists {
		t.Error("Installation state not persisted")
	}
}

func TestInstaller_InstallationRecord(t *testing.T) {
	tempDir := t.TempDir()
	sourceDir := filepath.Join(tempDir, "source")
	installDir := filepath.Join(tempDir, "install")

	_, collection := createTestCollectionForInstall(t, sourceDir, "test-collection", "test-author", "1.0.0", nil)

	registry := NewRegistry()
	mockSource := &MockInstallSource{
		collections: map[string]*sources.Collection{
			"mock://test-author/test-collection": collection,
		},
	}
	registry.AddSource(mockSource)

	installer, err := NewInstaller(registry, installDir)
	if err != nil {
		t.Fatalf("NewInstaller failed: %v", err)
	}

	ctx := context.Background()
	beforeInstall := time.Now()
	if err := installer.Install(ctx, "mock://test-author/test-collection", nil); err != nil {
		t.Fatalf("Install failed: %v", err)
	}

	record, exists := installer.GetInstalled("test-author/test-collection")
	if !exists {
		t.Fatal("Installation record not found")
	}

	// Verify record fields
	if record.ID != "test-author/test-collection" {
		t.Errorf("Expected ID test-author/test-collection, got %s", record.ID)
	}

	if record.Version != "1.0.0" {
		t.Errorf("Expected version 1.0.0, got %s", record.Version)
	}

	if record.URI != "mock://test-author/test-collection" {
		t.Errorf("Expected URI mock://test-author/test-collection, got %s", record.URI)
	}

	if record.SourceName != "mock-install" {
		t.Errorf("Expected source mock-install, got %s", record.SourceName)
	}

	if record.InstalledAt.Before(beforeInstall) {
		t.Error("Installation timestamp is before install was called")
	}

	expectedLocation := filepath.Join(installDir, "test-author", "test-collection")
	if record.InstallLocation != expectedLocation {
		t.Errorf("Expected location %s, got %s", expectedLocation, record.InstallLocation)
	}
}

func TestInstaller_SkipValidation(t *testing.T) {
	tempDir := t.TempDir()
	sourceDir := filepath.Join(tempDir, "source")
	installDir := filepath.Join(tempDir, "install")

	_, collection := createTestCollectionForInstall(t, sourceDir, "test-collection", "test-author", "1.0.0", nil)

	registry := NewRegistry()
	mockSource := &MockInstallSource{
		collections: map[string]*sources.Collection{
			"mock://test-author/test-collection": collection,
		},
	}
	registry.AddSource(mockSource)

	installer, err := NewInstaller(registry, installDir)
	if err != nil {
		t.Fatalf("NewInstaller failed: %v", err)
	}

	// Install with skip validation
	ctx := context.Background()
	err = installer.Install(ctx, "mock://test-author/test-collection", &InstallOptions{SkipValidation: true})
	if err != nil {
		t.Fatalf("Install with skip validation failed: %v", err)
	}

	// Verify installed
	if _, exists := installer.GetInstalled("test-author/test-collection"); !exists {
		t.Error("Collection not installed")
	}
}
