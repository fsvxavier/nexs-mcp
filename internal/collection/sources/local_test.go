package sources

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"
)

func TestNewLocalSource(t *testing.T) {
	t.Run("with custom search paths", func(t *testing.T) {
		tmpDir := t.TempDir()
		path1 := filepath.Join(tmpDir, "path1")
		path2 := filepath.Join(tmpDir, "path2")

		source, err := NewLocalSource([]string{path1, path2})
		require.NoError(t, err)
		assert.Equal(t, 2, len(source.searchPaths))

		// Verify directories were created
		_, err = os.Stat(path1)
		assert.NoError(t, err)
		_, err = os.Stat(path2)
		assert.NoError(t, err)
	})

	t.Run("with default search paths", func(t *testing.T) {
		source, err := NewLocalSource(nil)
		require.NoError(t, err)
		assert.True(t, len(source.searchPaths) > 0)
	})
}

func TestLocalSource_Name(t *testing.T) {
	source, err := NewLocalSource([]string{t.TempDir()})
	require.NoError(t, err)
	assert.Equal(t, "local", source.Name())
}

func TestLocalSource_Supports(t *testing.T) {
	source, err := NewLocalSource([]string{t.TempDir()})
	require.NoError(t, err)

	tests := []struct {
		name     string
		uri      string
		expected bool
	}{
		{
			name:     "file URI",
			uri:      "file:///path/to/collection",
			expected: true,
		},
		{
			name:     "absolute path (Unix)",
			uri:      "/path/to/collection",
			expected: true,
		},
		{
			name:     "github URI",
			uri:      "github://owner/repo",
			expected: false,
		},
		{
			name:     "https URI",
			uri:      "https://example.com/collection.tar.gz",
			expected: false,
		},
		{
			name:     "relative path",
			uri:      "relative/path",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := source.Supports(tt.uri)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestLocalSource_ParseURI(t *testing.T) {
	source, err := NewLocalSource([]string{t.TempDir()})
	require.NoError(t, err)

	tests := []struct {
		name         string
		uri          string
		expectedPath string
		expectError  bool
	}{
		{
			name:         "file URI",
			uri:          "file:///home/user/collection",
			expectedPath: "/home/user/collection",
			expectError:  false,
		},
		{
			name:         "absolute path",
			uri:          "/home/user/collection",
			expectedPath: "/home/user/collection",
			expectError:  false,
		},
		{
			name:        "relative path",
			uri:         "relative/path",
			expectError: true,
		},
		{
			name:        "invalid URI",
			uri:         "github://owner/repo",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			path, err := source.parseURI(tt.uri)

			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedPath, path)
			}
		})
	}
}

func TestLocalSource_Browse(t *testing.T) {
	tmpDir := t.TempDir()
	ctx := context.Background()

	// Create test collections
	createTestCollection(t, tmpDir, "collection1", map[string]interface{}{
		"name":        "Collection 1",
		"version":     "1.0.0",
		"author":      "author1",
		"description": "Test collection 1",
		"category":    "testing",
		"tags":        []string{"test", "demo"},
		"elements":    []map[string]string{},
	})

	createTestCollection(t, tmpDir, "collection2", map[string]interface{}{
		"name":        "Collection 2",
		"version":     "2.0.0",
		"author":      "author2",
		"description": "Test collection 2",
		"category":    "development",
		"tags":        []string{"dev"},
		"elements":    []map[string]string{},
	})

	source, err := NewLocalSource([]string{tmpDir})
	require.NoError(t, err)

	t.Run("browse all collections", func(t *testing.T) {
		collections, err := source.Browse(ctx, nil)
		assert.NoError(t, err)
		assert.Equal(t, 2, len(collections))
	})

	t.Run("filter by category", func(t *testing.T) {
		filter := &BrowseFilter{Category: "testing"}
		collections, err := source.Browse(ctx, filter)
		assert.NoError(t, err)
		assert.Equal(t, 1, len(collections))
		assert.Equal(t, "Collection 1", collections[0].Name)
	})

	t.Run("filter by author", func(t *testing.T) {
		filter := &BrowseFilter{Author: "author2"}
		collections, err := source.Browse(ctx, filter)
		assert.NoError(t, err)
		assert.Equal(t, 1, len(collections))
		assert.Equal(t, "Collection 2", collections[0].Name)
	})

	t.Run("filter by tags", func(t *testing.T) {
		filter := &BrowseFilter{Tags: []string{"test"}}
		collections, err := source.Browse(ctx, filter)
		assert.NoError(t, err)
		assert.Equal(t, 1, len(collections))
		assert.Equal(t, "Collection 1", collections[0].Name)
	})

	t.Run("filter by query", func(t *testing.T) {
		filter := &BrowseFilter{Query: "collection 2"}
		collections, err := source.Browse(ctx, filter)
		assert.NoError(t, err)
		assert.Equal(t, 1, len(collections))
		if len(collections) > 0 {
			assert.Equal(t, "Collection 2", collections[0].Name)
		}
	})

	t.Run("with limit", func(t *testing.T) {
		filter := &BrowseFilter{Limit: 1}
		collections, err := source.Browse(ctx, filter)
		assert.NoError(t, err)
		assert.Equal(t, 1, len(collections))
	})

	t.Run("with offset", func(t *testing.T) {
		filter := &BrowseFilter{Offset: 1}
		collections, err := source.Browse(ctx, filter)
		assert.NoError(t, err)
		assert.Equal(t, 1, len(collections))
	})
}

func TestLocalSource_Get_Directory(t *testing.T) {
	tmpDir := t.TempDir()
	ctx := context.Background()

	// Create test collection
	collectionDir := createTestCollection(t, tmpDir, "test-collection", map[string]interface{}{
		"name":        "Test Collection",
		"version":     "1.0.0",
		"author":      "test-author",
		"description": "A test collection",
		"elements":    []map[string]string{},
	})

	source, err := NewLocalSource([]string{tmpDir})
	require.NoError(t, err)

	t.Run("get by file URI", func(t *testing.T) {
		uri := fmt.Sprintf("file://%s", collectionDir)
		collection, err := source.Get(ctx, uri)
		assert.NoError(t, err)
		assert.NotNil(t, collection)
		assert.Equal(t, "Test Collection", collection.Metadata.Name)
		assert.Equal(t, "1.0.0", collection.Metadata.Version)
		assert.Equal(t, "test-author", collection.Metadata.Author)
	})

	t.Run("get by absolute path", func(t *testing.T) {
		collection, err := source.Get(ctx, collectionDir)
		assert.NoError(t, err)
		assert.NotNil(t, collection)
		assert.Equal(t, "Test Collection", collection.Metadata.Name)
	})

	t.Run("missing collection.yaml", func(t *testing.T) {
		emptyDir := filepath.Join(tmpDir, "empty")
		os.MkdirAll(emptyDir, 0755)

		_, err := source.Get(ctx, emptyDir)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "collection.yaml not found")
	})

	t.Run("invalid manifest", func(t *testing.T) {
		invalidDir := filepath.Join(tmpDir, "invalid")
		os.MkdirAll(invalidDir, 0755)
		os.WriteFile(filepath.Join(invalidDir, "collection.yaml"), []byte("invalid: yaml: content:"), 0644)

		_, err := source.Get(ctx, invalidDir)
		assert.Error(t, err)
	})
}

func TestLocalSource_Get_Archive(t *testing.T) {
	tmpDir := t.TempDir()
	ctx := context.Background()

	// Create test collection
	collectionDir := createTestCollection(t, tmpDir, "test-collection", map[string]interface{}{
		"name":        "Archived Collection",
		"version":     "1.0.0",
		"author":      "test-author",
		"description": "A test collection in archive",
		"elements":    []map[string]string{},
	})

	// Create tar.gz archive
	archivePath := filepath.Join(tmpDir, "collection.tar.gz")
	err := ExportToTarGz(collectionDir, archivePath)
	require.NoError(t, err)

	source, err := NewLocalSource([]string{tmpDir})
	require.NoError(t, err)

	t.Run("get from tar.gz", func(t *testing.T) {
		collection, err := source.Get(ctx, archivePath)
		assert.NoError(t, err)
		assert.NotNil(t, collection)
		assert.Equal(t, "Archived Collection", collection.Metadata.Name)
		assert.Equal(t, "1.0.0", collection.Metadata.Version)
	})

	t.Run("invalid archive", func(t *testing.T) {
		invalidArchive := filepath.Join(tmpDir, "invalid.tar.gz")
		os.WriteFile(invalidArchive, []byte("not a valid tar.gz"), 0644)

		_, err := source.Get(ctx, invalidArchive)
		assert.Error(t, err)
	})
}

func TestExportToTarGz(t *testing.T) {
	tmpDir := t.TempDir()

	// Create test collection
	collectionDir := createTestCollection(t, tmpDir, "export-test", map[string]interface{}{
		"name":        "Export Test",
		"version":     "1.0.0",
		"author":      "test-author",
		"description": "Test export",
		"elements":    []map[string]string{},
	})

	// Create additional files
	os.WriteFile(filepath.Join(collectionDir, "README.md"), []byte("# Test Collection"), 0644)
	os.MkdirAll(filepath.Join(collectionDir, "personas"), 0755)
	os.WriteFile(filepath.Join(collectionDir, "personas", "test.yaml"), []byte("name: test"), 0644)

	// Export to tar.gz
	archivePath := filepath.Join(tmpDir, "exported.tar.gz")
	err := ExportToTarGz(collectionDir, archivePath)
	require.NoError(t, err)

	// Verify archive exists
	_, err = os.Stat(archivePath)
	assert.NoError(t, err)

	// Verify we can load it back
	source, err := NewLocalSource([]string{tmpDir})
	require.NoError(t, err)

	collection, err := source.Get(context.Background(), archivePath)
	assert.NoError(t, err)
	assert.Equal(t, "Export Test", collection.Metadata.Name)
}

func TestLocalSource_MatchesFilter(t *testing.T) {
	source, err := NewLocalSource([]string{t.TempDir()})
	require.NoError(t, err)

	metadata := &CollectionMetadata{
		Name:        "Test Collection",
		Author:      "test-author",
		Description: "A testing collection for development",
		Category:    "development",
		Tags:        []string{"test", "demo", "example"},
	}

	tests := []struct {
		name     string
		filter   *BrowseFilter
		expected bool
	}{
		{
			name:     "nil filter",
			filter:   nil,
			expected: true,
		},
		{
			name:     "matching category",
			filter:   &BrowseFilter{Category: "development"},
			expected: true,
		},
		{
			name:     "non-matching category",
			filter:   &BrowseFilter{Category: "production"},
			expected: false,
		},
		{
			name:     "matching author",
			filter:   &BrowseFilter{Author: "test-author"},
			expected: true,
		},
		{
			name:     "non-matching author",
			filter:   &BrowseFilter{Author: "other-author"},
			expected: false,
		},
		{
			name:     "matching tags",
			filter:   &BrowseFilter{Tags: []string{"test", "demo"}},
			expected: true,
		},
		{
			name:     "partial matching tags",
			filter:   &BrowseFilter{Tags: []string{"test", "missing"}},
			expected: false,
		},
		{
			name:     "matching query in name",
			filter:   &BrowseFilter{Query: "test"},
			expected: true,
		},
		{
			name:     "matching query in description",
			filter:   &BrowseFilter{Query: "development"},
			expected: true,
		},
		{
			name:     "non-matching query",
			filter:   &BrowseFilter{Query: "xyz"},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := source.matchesFilter(metadata, tt.filter)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// Helper function to create a test collection
func createTestCollection(t *testing.T, baseDir, name string, manifest map[string]interface{}) string {
	collectionDir := filepath.Join(baseDir, name)
	err := os.MkdirAll(collectionDir, 0755)
	require.NoError(t, err)

	// Write collection.yaml
	manifestPath := filepath.Join(collectionDir, "collection.yaml")
	data, err := yaml.Marshal(manifest)
	require.NoError(t, err)

	err = os.WriteFile(manifestPath, data, 0644)
	require.NoError(t, err)

	return collectionDir
}
