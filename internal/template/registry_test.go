package template

import (
	"testing"
	"time"

	"github.com/fsvxavier/nexs-mcp/internal/domain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTemplateCache_GetSet(t *testing.T) {
	cache := &TemplateCache{
		templates: make(map[string]*domain.Template),
		expires:   make(map[string]time.Time),
		ttl:       5 * time.Minute,
	}

	tmpl := domain.NewTemplate("test-id", "Test", "1.0.0", "test-user")
	actualID := tmpl.GetID() // Get the auto-generated ID

	// Set template with its actual ID
	cache.Set(actualID, tmpl)

	// Get template
	retrieved, found := cache.Get(actualID)
	assert.True(t, found)
	require.NotNil(t, retrieved)
	assert.Equal(t, actualID, retrieved.GetID())
}

func TestTemplateCache_Miss(t *testing.T) {
	cache := &TemplateCache{
		templates: make(map[string]*domain.Template),
		expires:   make(map[string]time.Time),
		ttl:       5 * time.Minute,
	}

	// Get non-existent template
	_, found := cache.Get("non-existent")
	assert.False(t, found)
}

func TestTemplateCache_TTLExpiration(t *testing.T) {
	cache := &TemplateCache{
		templates: make(map[string]*domain.Template),
		expires:   make(map[string]time.Time),
		ttl:       10 * time.Millisecond,
	}

	tmpl := domain.NewTemplate("test-id", "Test", "1.0.0", "test-user")
	cache.Set("test-id", tmpl)

	// Should exist immediately
	_, found := cache.Get("test-id")
	assert.True(t, found)

	// Wait for expiration
	time.Sleep(15 * time.Millisecond)

	// Should be expired
	_, found = cache.Get("test-id")
	assert.False(t, found)
}

func TestTemplateCache_Clear(t *testing.T) {
	cache := &TemplateCache{
		templates: make(map[string]*domain.Template),
		expires:   make(map[string]time.Time),
		ttl:       5 * time.Minute,
	}

	// Add templates
	cache.Set("id1", domain.NewTemplate("id1", "Test1", "1.0.0", "test-user"))
	cache.Set("id2", domain.NewTemplate("id2", "Test2", "1.0.0", "test-user"))

	// Clear cache
	cache.Clear()

	// Both should be gone
	_, found := cache.Get("id1")
	assert.False(t, found)
	_, found = cache.Get("id2")
	assert.False(t, found)
}
