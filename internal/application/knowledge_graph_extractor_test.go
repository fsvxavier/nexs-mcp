package application

import (
	"context"
	"strings"
	"testing"

	"github.com/fsvxavier/nexs-mcp/internal/domain"
)

func TestNewKnowledgeGraphExtractor(t *testing.T) {
	repo := newMockRepoForIndex()

	extractor := NewKnowledgeGraphExtractor(repo)

	if extractor == nil {
		t.Fatal("Expected non-nil extractor")
	}
	if extractor.repository == nil {
		t.Error("Expected non-nil repository")
	}
}

func TestExtractFromMemory(t *testing.T) {
	repo := newMockRepoForIndex()
	extractor := NewKnowledgeGraphExtractor(repo)

	// Create memory with rich content
	content := `John Smith works at Google on machine learning projects. 
	He collaborates with Sarah Johnson on neural network research. 
	They use TensorFlow and PyTorch frameworks for deep learning.
	Contact: john@example.com, https://example.com/project`

	mem := domain.NewMemory("TestMem", content, "1.0.0", "test")
	repo.Create(mem)

	ctx := context.Background()
	graph, err := extractor.ExtractFromMemory(ctx, mem.GetID())
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if graph == nil {
		t.Fatal("Expected non-nil graph")
	}

	// Extraction results depend on NLP implementation
	t.Logf("Extracted: %d entities, %d concepts, %d keywords",
		len(graph.Entities), len(graph.Concepts), len(graph.Keywords))

	// Graph should at least be initialized
	if graph.Entities == nil {
		t.Error("Expected non-nil entities slice")
	}
	if graph.Concepts == nil {
		t.Error("Expected non-nil concepts map")
	}
	if graph.Keywords == nil {
		t.Error("Expected non-nil keywords slice")
	}
}

func TestExtractFromMemory_NotFound(t *testing.T) {
	repo := newMockRepoForIndex()
	extractor := NewKnowledgeGraphExtractor(repo)

	ctx := context.Background()
	_, err := extractor.ExtractFromMemory(ctx, "nonexistent-id")
	if err == nil {
		t.Error("Expected error for nonexistent memory")
	}
}

func TestExtractFromMemory_InvalidType(t *testing.T) {
	repo := newMockRepoForIndex()
	extractor := NewKnowledgeGraphExtractor(repo)

	// Create non-memory element
	persona := domain.NewPersona("Test", "Test persona", "1.0.0", "test")
	repo.Create(persona)

	ctx := context.Background()
	_, err := extractor.ExtractFromMemory(ctx, persona.GetID())
	if err == nil {
		t.Error("Expected error for non-memory element")
	}
}

func TestExtractFromCluster(t *testing.T) {
	repo := newMockRepoForIndex()
	extractor := NewKnowledgeGraphExtractor(repo)

	// Create cluster with memories
	mem1 := domain.NewMemory("Mem1", "Machine learning algorithms for classification", "1.0.0", "test")
	mem2 := domain.NewMemory("Mem2", "Deep learning neural networks", "1.0.0", "test")

	cluster := &Cluster{
		ID:      1,
		Members: []*domain.Memory{mem1, mem2},
		Size:    2,
	}

	ctx := context.Background()
	graph, err := extractor.ExtractFromCluster(ctx, cluster)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if graph == nil {
		t.Fatal("Expected non-nil graph")
	}

	// Should have summary
	if graph.Summary == "" {
		t.Error("Expected non-empty summary")
	}

	if !strings.Contains(graph.Summary, "cluster 1") {
		t.Error("Expected summary to mention cluster ID")
	}
}

func TestExtractFromCluster_Empty(t *testing.T) {
	repo := newMockRepoForIndex()
	extractor := NewKnowledgeGraphExtractor(repo)

	// Empty cluster
	cluster := &Cluster{
		ID:      1,
		Members: []*domain.Memory{},
		Size:    0,
	}

	ctx := context.Background()
	graph, err := extractor.ExtractFromCluster(ctx, cluster)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if len(graph.Entities) != 0 {
		t.Error("Expected 0 entities for empty cluster")
	}
	if len(graph.Concepts) != 0 {
		t.Error("Expected 0 concepts for empty cluster")
	}
}

func TestExtractFromMultipleMemories(t *testing.T) {
	repo := newMockRepoForIndex()
	extractor := NewKnowledgeGraphExtractor(repo)

	// Create memories
	mem1 := domain.NewMemory("Mem1", "Python programming language", "1.0.0", "test")
	mem2 := domain.NewMemory("Mem2", "JavaScript frameworks", "1.0.0", "test")
	mem3 := domain.NewMemory("Mem3", "Database optimization", "1.0.0", "test")

	repo.Create(mem1)
	repo.Create(mem2)
	repo.Create(mem3)

	memoryIDs := []string{mem1.GetID(), mem2.GetID(), mem3.GetID()}

	ctx := context.Background()
	graph, err := extractor.ExtractFromMultipleMemories(ctx, memoryIDs)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if graph == nil {
		t.Fatal("Expected non-nil graph")
	}

	// Graph should be initialized even if extraction returns few results
	t.Logf("Extracted from %d memories: %d entities, %d concepts",
		len(memoryIDs), len(graph.Entities), len(graph.Concepts))
}

func TestExtractFromMultipleMemories_WithMissing(t *testing.T) {
	repo := newMockRepoForIndex()
	extractor := NewKnowledgeGraphExtractor(repo)

	// Create one memory
	mem := domain.NewMemory("Mem", "Valid content", "1.0.0", "test")
	repo.Create(mem)

	// Include nonexistent IDs
	memoryIDs := []string{mem.GetID(), "nonexistent-1", "nonexistent-2"}

	ctx := context.Background()
	graph, err := extractor.ExtractFromMultipleMemories(ctx, memoryIDs)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	// Should skip missing memories and continue
	if graph == nil {
		t.Fatal("Expected non-nil graph even with missing memories")
	}
}

func TestExtractFromMultipleMemories_Empty(t *testing.T) {
	repo := newMockRepoForIndex()
	extractor := NewKnowledgeGraphExtractor(repo)

	ctx := context.Background()
	graph, err := extractor.ExtractFromMultipleMemories(ctx, []string{})
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if len(graph.Entities) != 0 {
		t.Error("Expected 0 entities for empty list")
	}
}

func TestEntity_Structure(t *testing.T) {
	entity := Entity{
		Type:  "person",
		Value: "John Smith",
		Count: 3,
	}

	if entity.Type != "person" {
		t.Errorf("Expected type 'person', got %s", entity.Type)
	}
	if entity.Value != "John Smith" {
		t.Errorf("Expected value 'John Smith', got %s", entity.Value)
	}
	if entity.Count != 3 {
		t.Errorf("Expected count 3, got %d", entity.Count)
	}
}

func TestKnowledgeGraph_Structure(t *testing.T) {
	graph := KnowledgeGraph{
		Entities: []Entity{
			{Type: "person", Value: "Alice", Count: 1},
			{Type: "organization", Value: "Google", Count: 2},
		},
		Relationships: []domain.Relationship{
			{SourceID: "alice", TargetID: "google", Type: "works_at"},
		},
		Concepts: map[string]int{
			"machine-learning": 5,
			"neural-networks":  3,
		},
		Keywords: []string{"AI", "ML", "Python"},
		Summary:  "Test summary",
	}

	if len(graph.Entities) != 2 {
		t.Errorf("Expected 2 entities, got %d", len(graph.Entities))
	}
	if len(graph.Relationships) != 1 {
		t.Errorf("Expected 1 relationship, got %d", len(graph.Relationships))
	}
	if len(graph.Concepts) != 2 {
		t.Errorf("Expected 2 concepts, got %d", len(graph.Concepts))
	}
	if len(graph.Keywords) != 3 {
		t.Errorf("Expected 3 keywords, got %d", len(graph.Keywords))
	}
	if graph.Summary != "Test summary" {
		t.Errorf("Expected summary 'Test summary', got %s", graph.Summary)
	}
}

func TestExtractEntities_Persons(t *testing.T) {
	repo := newMockRepoForIndex()
	extractor := NewKnowledgeGraphExtractor(repo)

	content := "John Smith and Sarah Johnson are working together. Dr. Robert Brown joined them."

	mem := domain.NewMemory("Mem", content, "1.0.0", "test")
	repo.Create(mem)

	ctx := context.Background()
	graph, err := extractor.ExtractFromMemory(ctx, mem.GetID())
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	// Count person entities if any (NLP-dependent)
	personCount := 0
	for _, entity := range graph.Entities {
		if entity.Type == "person" {
			personCount++
		}
	}

	t.Logf("Extracted %d person entities (NLP-dependent)", personCount)
}

func TestExtractEntities_Organizations(t *testing.T) {
	repo := newMockRepoForIndex()
	extractor := NewKnowledgeGraphExtractor(repo)

	content := "Google, Microsoft, and Amazon are leading tech companies."

	mem := domain.NewMemory("Mem", content, "1.0.0", "test")
	repo.Create(mem)

	ctx := context.Background()
	graph, err := extractor.ExtractFromMemory(ctx, mem.GetID())
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	// Count organization entities (NLP-dependent)
	orgCount := 0
	for _, entity := range graph.Entities {
		if entity.Type == "organization" {
			orgCount++
		}
	}

	t.Logf("Extracted %d organization entities (NLP-dependent)", orgCount)
}

func TestExtractEntities_URLs(t *testing.T) {
	repo := newMockRepoForIndex()
	extractor := NewKnowledgeGraphExtractor(repo)

	content := "Visit https://example.com and http://test.org for more info."

	mem := domain.NewMemory("Mem", content, "1.0.0", "test")
	repo.Create(mem)

	ctx := context.Background()
	graph, err := extractor.ExtractFromMemory(ctx, mem.GetID())
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	// Count URL entities (regex-dependent)
	urlCount := 0
	for _, entity := range graph.Entities {
		if entity.Type == "url" {
			urlCount++
		}
	}

	t.Logf("Extracted %d URL entities (regex-dependent)", urlCount)
}

func TestExtractEntities_Emails(t *testing.T) {
	repo := newMockRepoForIndex()
	extractor := NewKnowledgeGraphExtractor(repo)

	content := "Contact john@example.com or sarah@test.org for details."

	mem := domain.NewMemory("Mem", content, "1.0.0", "test")
	repo.Create(mem)

	ctx := context.Background()
	graph, err := extractor.ExtractFromMemory(ctx, mem.GetID())
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	// Count email entities (regex-dependent)
	emailCount := 0
	for _, entity := range graph.Entities {
		if entity.Type == "email" {
			emailCount++
		}
	}

	t.Logf("Extracted %d email entities (regex-dependent)", emailCount)
}

func TestExtractConcepts(t *testing.T) {
	repo := newMockRepoForIndex()
	extractor := NewKnowledgeGraphExtractor(repo)

	content := `Machine learning is a subset of artificial intelligence. 
	Neural networks are fundamental to deep learning. 
	Training models requires large datasets and computational power.`

	mem := domain.NewMemory("Mem", content, "1.0.0", "test")
	repo.Create(mem)

	ctx := context.Background()
	graph, err := extractor.ExtractFromMemory(ctx, mem.GetID())
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	// Concepts extraction is NLP-dependent
	t.Logf("Extracted %d concepts", len(graph.Concepts))
	if graph.Concepts == nil {
		t.Error("Expected non-nil concepts map")
	}

	// Concepts should have frequency counts
	for concept, count := range graph.Concepts {
		if count <= 0 {
			t.Errorf("Concept '%s' has invalid count: %d", concept, count)
		}
	}
}

func TestGraphExtractKeywords(t *testing.T) {
	repo := newMockRepoForIndex()
	extractor := NewKnowledgeGraphExtractor(repo)

	content := "Python programming language is widely used for machine learning and data analysis."

	mem := domain.NewMemory("Mem", content, "1.0.0", "test")
	repo.Create(mem)

	ctx := context.Background()
	graph, err := extractor.ExtractFromMemory(ctx, mem.GetID())
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	// Keywords extraction may vary by implementation
	t.Logf("Extracted %d keywords", len(graph.Keywords))
	if graph.Keywords == nil {
		t.Error("Expected non-nil keywords slice")
	}

	// Keywords should be non-empty
	for _, keyword := range graph.Keywords {
		if keyword == "" {
			t.Error("Found empty keyword")
		}
	}
}

func TestExtractRelationships(t *testing.T) {
	repo := newMockRepoForIndex()
	extractor := NewKnowledgeGraphExtractor(repo)

	content := "Python is a programming language. TensorFlow uses Python. Keras is built on TensorFlow."

	mem := domain.NewMemory("Mem", content, "1.0.0", "test")
	repo.Create(mem)

	ctx := context.Background()
	graph, err := extractor.ExtractFromMemory(ctx, mem.GetID())
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	// Relationship extraction is NLP-dependent
	t.Logf("Extracted %d relationships", len(graph.Relationships))
	if graph.Relationships == nil {
		t.Error("Expected non-nil relationships slice")
	}

	// Relationships should have source, target, and type
	for _, rel := range graph.Relationships {
		if rel.SourceID == "" {
			t.Error("Relationship missing 'source'")
		}
		if rel.TargetID == "" {
			t.Error("Relationship missing 'target'")
		}
		if rel.Type == "" {
			t.Error("Relationship missing 'type'")
		}
	}
}

func TestExtractFromContent_EmptyContent(t *testing.T) {
	repo := newMockRepoForIndex()
	extractor := NewKnowledgeGraphExtractor(repo)

	mem := domain.NewMemory("Mem", "", "1.0.0", "test")
	repo.Create(mem)

	ctx := context.Background()
	graph, err := extractor.ExtractFromMemory(ctx, mem.GetID())
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	// Empty content should produce empty graph
	if len(graph.Entities) != 0 {
		t.Error("Expected 0 entities for empty content")
	}
	if len(graph.Concepts) != 0 {
		t.Error("Expected 0 concepts for empty content")
	}
	if len(graph.Keywords) != 0 {
		t.Error("Expected 0 keywords for empty content")
	}
}

func TestExtractFromContent_ShortContent(t *testing.T) {
	repo := newMockRepoForIndex()
	extractor := NewKnowledgeGraphExtractor(repo)

	mem := domain.NewMemory("Mem", "Test", "1.0.0", "test")
	repo.Create(mem)

	ctx := context.Background()
	graph, err := extractor.ExtractFromMemory(ctx, mem.GetID())
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	// Short content may produce minimal extraction
	if graph == nil {
		t.Fatal("Expected non-nil graph")
	}
}
