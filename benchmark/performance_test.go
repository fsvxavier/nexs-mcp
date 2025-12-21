package benchmark

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/fsvxavier/nexs-mcp/internal/domain"
	"github.com/fsvxavier/nexs-mcp/internal/infrastructure"
)

// setupTestRepo creates a test repository with cleanup.
func setupTestRepo(b *testing.B) (domain.ElementRepository, func()) {
	b.Helper()
	testDir := filepath.Join(os.TempDir(), fmt.Sprintf("nexs-benchmark-%d", time.Now().UnixNano()))
	repo, err := infrastructure.NewFileElementRepository(testDir)
	if err != nil {
		b.Fatalf("Failed to create repository: %v", err)
	}
	cleanup := func() {
		os.RemoveAll(testDir)
	}
	return repo, cleanup
}

// createTestPersona creates a test persona element.
func createTestPersona(id string) *domain.Persona {
	name := "Test Persona " + id
	persona := domain.NewPersona(name, "Test persona for benchmarking", "1.0.0", "benchmark")

	persona.BehavioralTraits = []domain.BehavioralTrait{
		{Name: "analytical", Intensity: 8},
	}
	persona.ExpertiseAreas = []domain.ExpertiseArea{
		{Domain: "testing", Level: "expert"},
	}
	persona.ResponseStyle = domain.ResponseStyle{
		Tone:      "professional",
		Formality: "neutral",
		Verbosity: "balanced",
	}
	persona.SystemPrompt = "You are a test persona for benchmarking purposes."
	persona.PrivacyLevel = domain.PrivacyPublic

	return persona
}

// BenchmarkElementCreate benchmarks element creation.
func BenchmarkElementCreate(b *testing.B) {
	repo, cleanup := setupTestRepo(b)
	defer cleanup()

	b.ResetTimer()
	for i := range b.N {
		persona := createTestPersona(fmt.Sprintf("bench-create-%d", i))
		_ = repo.Create(persona)
	}
}

// BenchmarkElementRead benchmarks element retrieval.
func BenchmarkElementRead(b *testing.B) {
	repo, cleanup := setupTestRepo(b)
	defer cleanup()

	// Setup: create test element
	persona := createTestPersona("benchmark-read")
	_ = repo.Create(persona)

	b.ResetTimer()
	for range b.N {
		_, _ = repo.GetByID("benchmark-read")
	}
}

// BenchmarkElementUpdate benchmarks element updates.
func BenchmarkElementUpdate(b *testing.B) {
	repo, cleanup := setupTestRepo(b)
	defer cleanup()

	// Setup
	persona := createTestPersona("benchmark-update")
	_ = repo.Create(persona)

	b.ResetTimer()
	for i := range b.N {
		metadata := persona.GetMetadata()
		metadata.Version = fmt.Sprintf("1.0.%d", i)
		_ = repo.Update(persona)
	}
}

// BenchmarkElementDelete benchmarks element deletion.
func BenchmarkElementDelete(b *testing.B) {
	repo, cleanup := setupTestRepo(b)
	defer cleanup()

	b.ResetTimer()
	for i := range b.N {
		b.StopTimer()
		persona := createTestPersona(fmt.Sprintf("benchmark-delete-%d", i))
		_ = repo.Create(persona)
		b.StartTimer()

		_ = repo.Delete(persona.GetID())
	}
}

// BenchmarkElementList benchmarks listing elements.
func BenchmarkElementList(b *testing.B) {
	repo, cleanup := setupTestRepo(b)
	defer cleanup()

	// Setup: create 100 test elements
	for i := range 100 {
		persona := createTestPersona(fmt.Sprintf("benchmark-list-%d", i))
		_ = repo.Create(persona)
	}

	filter := domain.ElementFilter{
		Limit: 100,
	}

	b.ResetTimer()
	for range b.N {
		_, _ = repo.List(filter)
	}
}

// BenchmarkSearchByType benchmarks search by element type.
func BenchmarkSearchByType(b *testing.B) {
	repo, cleanup := setupTestRepo(b)
	defer cleanup()

	// Setup: create 100 personas
	for i := range 100 {
		persona := createTestPersona(fmt.Sprintf("persona-%d", i))
		_ = repo.Create(persona)
	}

	personaType := domain.PersonaElement
	filter := domain.ElementFilter{
		Type:  &personaType,
		Limit: 100,
	}

	b.ResetTimer()
	for range b.N {
		_, _ = repo.List(filter)
	}
}

// BenchmarkSearchByTags benchmarks search by tags.
func BenchmarkSearchByTags(b *testing.B) {
	repo, cleanup := setupTestRepo(b)
	defer cleanup()

	// Setup: create elements with tags
	tags := [][]string{
		{"python", "development"},
		{"javascript", "frontend"},
		{"golang", "backend"},
		{"data", "analysis"},
	}

	for i := range 100 {
		persona := createTestPersona(fmt.Sprintf("persona-%d", i))
		metadata := persona.GetMetadata()
		metadata.Tags = tags[i%len(tags)]
		_ = repo.Create(persona)
	}

	filter := domain.ElementFilter{
		Tags:  []string{"python"},
		Limit: 100,
	}

	b.ResetTimer()
	for range b.N {
		_, _ = repo.List(filter)
	}
}

// BenchmarkValidation benchmarks element validation.
func BenchmarkValidation(b *testing.B) {
	persona := createTestPersona("benchmark-validation")

	b.ResetTimer()
	for range b.N {
		_ = persona.Validate()
	}
}

// BenchmarkMemoryUsage benchmarks memory usage for different operations.
func BenchmarkMemoryUsage(b *testing.B) {
	b.Run("CreateElements", func(b *testing.B) {
		b.ReportAllocs()
		for i := range b.N {
			_ = createTestPersona(fmt.Sprintf("mem-%d", i))
		}
	})

	b.Run("ListElements", func(b *testing.B) {
		repo, cleanup := setupTestRepo(b)
		defer cleanup()

		// Setup
		for i := range 100 {
			persona := createTestPersona(fmt.Sprintf("mem-list-%d", i))
			_ = repo.Create(persona)
		}

		filter := domain.ElementFilter{
			Limit: 100,
		}

		b.ReportAllocs()
		b.ResetTimer()
		for range b.N {
			_, _ = repo.List(filter)
		}
	})
}

// BenchmarkConcurrentReads benchmarks concurrent read operations.
func BenchmarkConcurrentReads(b *testing.B) {
	repo, cleanup := setupTestRepo(b)
	defer cleanup()

	// Setup
	persona := createTestPersona("concurrent-read")
	_ = repo.Create(persona)

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			_, _ = repo.GetByID("concurrent-read")
		}
	})
}

// BenchmarkConcurrentWrites benchmarks concurrent write operations.
func BenchmarkConcurrentWrites(b *testing.B) {
	repo, cleanup := setupTestRepo(b)
	defer cleanup()

	b.RunParallel(func(pb *testing.PB) {
		i := 0
		for pb.Next() {
			persona := createTestPersona(fmt.Sprintf("concurrent-write-%d-%d", b.N, i))
			_ = repo.Create(persona)
			i++
		}
	})
}

// BenchmarkStartupTime benchmarks application startup time.
func BenchmarkStartupTime(b *testing.B) {
	testDir := filepath.Join(os.TempDir(), "nexs-benchmark-startup")
	defer os.RemoveAll(testDir)

	b.ResetTimer()
	for range b.N {
		// Simulate startup operations
		_, _ = infrastructure.NewFileElementRepository(testDir)
		time.Sleep(1 * time.Millisecond) // Simulate other initialization
	}
}
