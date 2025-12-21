package collection

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/fsvxavier/nexs-mcp/internal/collection/sources"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// MockCollectionSource is a mock implementation of CollectionSource for testing.
type MockCollectionSource struct {
	name            string
	browseResult    []*sources.CollectionMetadata
	browseError     error
	getResult       *sources.Collection
	getError        error
	supportsPattern string
}

func (m *MockCollectionSource) Name() string {
	return m.name
}

func (m *MockCollectionSource) Browse(ctx context.Context, filter *sources.BrowseFilter) ([]*sources.CollectionMetadata, error) {
	return m.browseResult, m.browseError
}

func (m *MockCollectionSource) Get(ctx context.Context, uri string) (*sources.Collection, error) {
	return m.getResult, m.getError
}

func (m *MockCollectionSource) Supports(uri string) bool {
	if m.supportsPattern == "" {
		return false
	}
	return len(uri) >= len(m.supportsPattern) && uri[:len(m.supportsPattern)] == m.supportsPattern
}

func TestNewRegistry(t *testing.T) {
	registry := NewRegistry()
	assert.NotNil(t, registry)
	assert.NotNil(t, registry.sources)
	assert.Equal(t, 0, len(registry.sources))
}

func TestRegistry_AddSource(t *testing.T) {
	registry := NewRegistry()

	source1 := &MockCollectionSource{name: "mock1"}
	source2 := &MockCollectionSource{name: "mock2"}

	registry.AddSource(source1)
	assert.Equal(t, 1, len(registry.sources))

	registry.AddSource(source2)
	assert.Equal(t, 2, len(registry.sources))
}

func TestRegistry_GetSources(t *testing.T) {
	registry := NewRegistry()

	source1 := &MockCollectionSource{name: "mock1"}
	source2 := &MockCollectionSource{name: "mock2"}

	registry.AddSource(source1)
	registry.AddSource(source2)

	sources := registry.GetSources()
	assert.Equal(t, 2, len(sources))
	assert.Equal(t, "mock1", sources[0].Name())
	assert.Equal(t, "mock2", sources[1].Name())

	// Verify it returns a copy (not the original slice)
	sources[0] = nil
	assert.NotNil(t, registry.sources[0])
}

func TestRegistry_Browse_EmptyRegistry(t *testing.T) {
	registry := NewRegistry()
	ctx := context.Background()

	results, err := registry.Browse(ctx, nil, "")
	assert.NoError(t, err)
	assert.NotNil(t, results)
	assert.Equal(t, 0, len(results))
}

func TestRegistry_Browse_SingleSource(t *testing.T) {
	registry := NewRegistry()
	ctx := context.Background()

	mockSource := &MockCollectionSource{
		name: "test-source",
		browseResult: []*sources.CollectionMetadata{
			{
				SourceName:  "test-source",
				URI:         "test://collection1",
				Name:        "collection1",
				Version:     "1.0.0",
				Author:      "author1",
				Description: "Test collection 1",
			},
			{
				SourceName:  "test-source",
				URI:         "test://collection2",
				Name:        "collection2",
				Version:     "1.0.0",
				Author:      "author1",
				Description: "Test collection 2",
			},
		},
	}

	registry.AddSource(mockSource)

	results, err := registry.Browse(ctx, nil, "")
	assert.NoError(t, err)
	assert.Equal(t, 2, len(results))
	assert.Equal(t, "collection1", results[0].Name)
	assert.Equal(t, "collection2", results[1].Name)
}

func TestRegistry_Browse_MultipleSources(t *testing.T) {
	registry := NewRegistry()
	ctx := context.Background()

	source1 := &MockCollectionSource{
		name: "source1",
		browseResult: []*sources.CollectionMetadata{
			{
				SourceName: "source1",
				URI:        "source1://collection1",
				Name:       "collection1",
				Version:    "1.0.0",
				Author:     "author1",
			},
		},
	}

	source2 := &MockCollectionSource{
		name: "source2",
		browseResult: []*sources.CollectionMetadata{
			{
				SourceName: "source2",
				URI:        "source2://collection2",
				Name:       "collection2",
				Version:    "1.0.0",
				Author:     "author2",
			},
		},
	}

	registry.AddSource(source1)
	registry.AddSource(source2)

	results, err := registry.Browse(ctx, nil, "")
	assert.NoError(t, err)
	assert.Equal(t, 2, len(results))
}

func TestRegistry_Browse_Deduplication(t *testing.T) {
	registry := NewRegistry()
	ctx := context.Background()

	// Both sources return the same collection
	source1 := &MockCollectionSource{
		name: "source1",
		browseResult: []*sources.CollectionMetadata{
			{
				SourceName: "source1",
				URI:        "source1://collection",
				Name:       "collection",
				Version:    "1.0.0",
				Author:     "author",
			},
		},
	}

	source2 := &MockCollectionSource{
		name: "source2",
		browseResult: []*sources.CollectionMetadata{
			{
				SourceName: "source2",
				URI:        "source2://collection",
				Name:       "collection",
				Version:    "1.0.0",
				Author:     "author",
			},
		},
	}

	registry.AddSource(source1)
	registry.AddSource(source2)

	results, err := registry.Browse(ctx, nil, "")
	assert.NoError(t, err)
	// Should be deduplicated to 1 result (same author/name/version)
	assert.Equal(t, 1, len(results))
}

func TestRegistry_Browse_DifferentVersions(t *testing.T) {
	registry := NewRegistry()
	ctx := context.Background()

	// Same collection but different versions
	source1 := &MockCollectionSource{
		name: "source1",
		browseResult: []*sources.CollectionMetadata{
			{
				SourceName: "source1",
				URI:        "source1://collection",
				Name:       "collection",
				Version:    "1.0.0",
				Author:     "author",
			},
		},
	}

	source2 := &MockCollectionSource{
		name: "source2",
		browseResult: []*sources.CollectionMetadata{
			{
				SourceName: "source2",
				URI:        "source2://collection",
				Name:       "collection",
				Version:    "2.0.0",
				Author:     "author",
			},
		},
	}

	registry.AddSource(source1)
	registry.AddSource(source2)

	results, err := registry.Browse(ctx, nil, "")
	assert.NoError(t, err)
	// Should NOT be deduplicated (different versions)
	assert.Equal(t, 2, len(results))
}

func TestRegistry_Browse_SpecificSource(t *testing.T) {
	registry := NewRegistry()
	ctx := context.Background()

	source1 := &MockCollectionSource{
		name: "source1",
		browseResult: []*sources.CollectionMetadata{
			{Name: "collection1", Version: "1.0.0", Author: "author1"},
		},
	}

	source2 := &MockCollectionSource{
		name: "source2",
		browseResult: []*sources.CollectionMetadata{
			{Name: "collection2", Version: "1.0.0", Author: "author2"},
		},
	}

	registry.AddSource(source1)
	registry.AddSource(source2)

	// Query only source1
	results, err := registry.Browse(ctx, nil, "source1")
	assert.NoError(t, err)
	assert.Equal(t, 1, len(results))
	assert.Equal(t, "collection1", results[0].Name)
}

func TestRegistry_Browse_SourceNotFound(t *testing.T) {
	registry := NewRegistry()
	ctx := context.Background()

	source := &MockCollectionSource{name: "source1"}
	registry.AddSource(source)

	// Query non-existent source
	results, err := registry.Browse(ctx, nil, "nonexistent")
	assert.Error(t, err)
	assert.Nil(t, results)
	assert.Contains(t, err.Error(), "source not found: nonexistent")
}

func TestRegistry_Browse_SourceError(t *testing.T) {
	registry := NewRegistry()
	ctx := context.Background()

	source := &MockCollectionSource{
		name:        "source1",
		browseError: errors.New("source error"),
	}
	registry.AddSource(source)

	_, err := registry.Browse(ctx, nil, "")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "all sources failed")
}

func TestRegistry_Browse_PartialFailure(t *testing.T) {
	registry := NewRegistry()
	ctx := context.Background()

	source1 := &MockCollectionSource{
		name:        "source1",
		browseError: errors.New("source1 error"),
	}

	source2 := &MockCollectionSource{
		name: "source2",
		browseResult: []*sources.CollectionMetadata{
			{Name: "collection2", Version: "1.0.0", Author: "author2"},
		},
	}

	registry.AddSource(source1)
	registry.AddSource(source2)

	// Should succeed with results from source2
	results, err := registry.Browse(ctx, nil, "")
	assert.NoError(t, err)
	assert.Equal(t, 1, len(results))
	assert.Equal(t, "collection2", results[0].Name)
}

func TestRegistry_Get(t *testing.T) {
	registry := NewRegistry()
	ctx := context.Background()

	expectedCollection := &sources.Collection{
		Metadata: &sources.CollectionMetadata{
			Name:    "test-collection",
			Version: "1.0.0",
			Author:  "test-author",
		},
	}

	source := &MockCollectionSource{
		name:            "test-source",
		supportsPattern: "test://",
		getResult:       expectedCollection,
	}

	registry.AddSource(source)

	result, err := registry.Get(ctx, "test://collection")
	assert.NoError(t, err)
	assert.Equal(t, expectedCollection, result)
}

func TestRegistry_Get_NoSupportingSource(t *testing.T) {
	registry := NewRegistry()
	ctx := context.Background()

	source := &MockCollectionSource{
		name:            "test-source",
		supportsPattern: "test://",
	}

	registry.AddSource(source)

	result, err := registry.Get(ctx, "other://collection")
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "no source supports URI")
}

func TestRegistry_Get_SourceError(t *testing.T) {
	registry := NewRegistry()
	ctx := context.Background()

	source := &MockCollectionSource{
		name:            "test-source",
		supportsPattern: "test://",
		getError:        errors.New("get error"),
	}

	registry.AddSource(source)

	result, err := registry.Get(ctx, "test://collection")
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "get error")
}

func TestRegistry_FindSource(t *testing.T) {
	registry := NewRegistry()

	source1 := &MockCollectionSource{
		name:            "source1",
		supportsPattern: "github://",
	}

	source2 := &MockCollectionSource{
		name:            "source2",
		supportsPattern: "file://",
	}

	registry.AddSource(source1)
	registry.AddSource(source2)

	// Find GitHub source
	found := registry.FindSource("github://owner/repo")
	assert.NotNil(t, found)
	assert.Equal(t, "source1", found.Name())

	// Find file source
	found = registry.FindSource("file:///path/to/collection")
	assert.NotNil(t, found)
	assert.Equal(t, "source2", found.Name())

	// No matching source
	found = registry.FindSource("http://example.com")
	assert.Nil(t, found)
}

func TestRegistry_ThreadSafety(t *testing.T) {
	registry := NewRegistry()

	// Add sources concurrently
	done := make(chan bool)
	for i := range 10 {
		go func(n int) {
			source := &MockCollectionSource{
				name: fmt.Sprintf("source%d", n),
			}
			registry.AddSource(source)
			done <- true
		}(i)
	}

	// Wait for all goroutines
	for range 10 {
		<-done
	}

	sources := registry.GetSources()
	assert.Equal(t, 10, len(sources))
}

func TestRegistry_Browse_WithFilter(t *testing.T) {
	registry := NewRegistry()
	ctx := context.Background()

	filter := &sources.BrowseFilter{
		Category: "devops",
		Author:   "test-author",
		Limit:    10,
	}

	source := &MockCollectionSource{
		name: "test-source",
		browseResult: []*sources.CollectionMetadata{
			{Name: "collection1", Version: "1.0.0", Author: "test-author", Category: "devops"},
		},
	}

	registry.AddSource(source)

	results, err := registry.Browse(ctx, filter, "")
	require.NoError(t, err)
	require.NotNil(t, results)
	assert.Equal(t, 1, len(results))
}
