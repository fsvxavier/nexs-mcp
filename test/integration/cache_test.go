package integration_test

import (
	"testing"
	"time"

	"github.com/fsvxavier/nexs-mcp/internal/application"
	"github.com/fsvxavier/nexs-mcp/internal/domain"
	"github.com/fsvxavier/nexs-mcp/internal/infrastructure"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCacheIntegration(t *testing.T) {
	tmpDir := t.TempDir()
	repo, err := infrastructure.NewFileElementRepository(tmpDir)
	require.NoError(t, err)

	cache := application.NewAdaptiveCacheService(application.AdaptiveCacheConfig{
		Enabled: true,
		MinTTL:  1 * time.Hour,
		MaxTTL:  168 * time.Hour,
		BaseTTL: 24 * time.Hour,
	})

	repo.SetAdaptiveCache(cache)

	persona := domain.NewPersona("Test Persona", "Test description", "1.0.0", "test-author")
	err = repo.Create(persona)
	require.NoError(t, err)

	personaID := persona.GetMetadata().ID
	statsInitial := cache.GetStats()

	// First access - miss
	_, err = repo.GetByID(personaID)
	require.NoError(t, err)

	// Second access - hit
	_, err = repo.GetByID(personaID)
	require.NoError(t, err)

	statsFinal := cache.GetStats()
	assert.Greater(t, statsFinal.TotalHits, statsInitial.TotalHits)
	t.Logf("Cache stats: hits=%d, misses=%d, hit_rate=%.2f%%",
		statsFinal.TotalHits, statsFinal.TotalMisses, cache.GetHitRate()*100)
}
