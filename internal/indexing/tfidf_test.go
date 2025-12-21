package indexing

import (
	"testing"

	"github.com/fsvxavier/nexs-mcp/internal/domain"
)

func TestNewTFIDFIndex(t *testing.T) {
	idx := NewTFIDFIndex()

	if idx == nil {
		t.Fatal("NewTFIDFIndex returned nil")
	}

	if idx.totalDocuments != 0 {
		t.Errorf("Expected 0 documents, got %d", idx.totalDocuments)
	}

	if len(idx.documents) != 0 {
		t.Errorf("Expected empty documents map, got %d entries", len(idx.documents))
	}
}

func TestAddDocument(t *testing.T) {
	idx := NewTFIDFIndex()

	doc := &Document{
		ID:      "test-1",
		Type:    domain.PersonaElement,
		Name:    "Test Persona",
		Content: "This is a test document for testing search functionality",
	}

	idx.AddDocument(doc)

	if idx.totalDocuments != 1 {
		t.Errorf("Expected 1 document, got %d", idx.totalDocuments)
	}

	if len(doc.Terms) == 0 {
		t.Error("Document terms should be populated after adding")
	}

	if len(idx.idf) == 0 {
		t.Error("IDF should be built after adding document")
	}
}

func TestRemoveDocument(t *testing.T) {
	idx := NewTFIDFIndex()

	doc := &Document{
		ID:      "test-1",
		Type:    domain.SkillElement,
		Name:    "Test Skill",
		Content: "Test content",
	}

	idx.AddDocument(doc)
	idx.RemoveDocument("test-1")

	if idx.totalDocuments != 0 {
		t.Errorf("Expected 0 documents after removal, got %d", idx.totalDocuments)
	}

	if _, exists := idx.documents["test-1"]; exists {
		t.Error("Document should be removed from index")
	}
}

func TestSearch(t *testing.T) {
	idx := NewTFIDFIndex()

	// Add test documents
	docs := []*Document{
		{
			ID:      "doc-1",
			Type:    domain.PersonaElement,
			Name:    "Go Programming Expert",
			Content: "Expert in Go programming language, concurrency, and microservices architecture",
		},
		{
			ID:      "doc-2",
			Type:    domain.SkillElement,
			Name:    "Python Data Science",
			Content: "Python expert specializing in data science, machine learning, and AI",
		},
		{
			ID:      "doc-3",
			Type:    domain.TemplateElement,
			Name:    "API Documentation",
			Content: "Template for API documentation using Go and OpenAPI specification",
		},
	}

	for _, doc := range docs {
		idx.AddDocument(doc)
	}

	// Test search for "Go"
	results := idx.Search("Go programming", 10)

	if len(results) == 0 {
		t.Fatal("Search should return results for 'Go programming'")
	}

	// First result should be the Go expert document
	if results[0].DocumentID != "doc-1" && results[0].DocumentID != "doc-3" {
		t.Errorf("Expected doc-1 or doc-3 as top result, got %s", results[0].DocumentID)
	}

	// Scores should be descending
	for i := 1; i < len(results); i++ {
		if results[i].Score > results[i-1].Score {
			t.Error("Results should be sorted by score descending")
		}
	}
}

func TestSearchWithLimit(t *testing.T) {
	idx := NewTFIDFIndex()

	// Add multiple documents with matching content
	for i := range 5 {
		doc := &Document{
			ID:      string(rune('a' + i)),
			Type:    domain.PersonaElement,
			Name:    "Test Doc",
			Content: "test content programming code development software",
		}
		idx.AddDocument(doc)
	}

	results := idx.Search("programming code", 2)

	if len(results) > 2 {
		t.Errorf("Expected max 2 results with limit, got %d", len(results))
	}
}

func TestSearchEmptyQuery(t *testing.T) {
	idx := NewTFIDFIndex()

	doc := &Document{
		ID:      "test-1",
		Type:    domain.PersonaElement,
		Name:    "Test",
		Content: "Test content",
	}
	idx.AddDocument(doc)

	results := idx.Search("", 10)

	if len(results) != 0 {
		t.Errorf("Empty query should return no results, got %d", len(results))
	}
}

func TestSearchEmptyIndex(t *testing.T) {
	idx := NewTFIDFIndex()

	results := idx.Search("test query", 10)

	if results != nil {
		t.Error("Search on empty index should return nil")
	}
}

func TestFindSimilar(t *testing.T) {
	idx := NewTFIDFIndex()

	// Add documents with similar content
	docs := []*Document{
		{
			ID:      "doc-1",
			Type:    domain.PersonaElement,
			Name:    "Go Expert",
			Content: "Expert in Go programming, concurrency, and microservices",
		},
		{
			ID:      "doc-2",
			Type:    domain.PersonaElement,
			Name:    "Go Developer",
			Content: "Go developer focusing on microservices and distributed systems",
		},
		{
			ID:      "doc-3",
			Type:    domain.SkillElement,
			Name:    "Python Skills",
			Content: "Python programming for data science and machine learning",
		},
	}

	for _, doc := range docs {
		idx.AddDocument(doc)
	}

	// Find similar to doc-1
	similar := idx.FindSimilar("doc-1", 10)

	if len(similar) == 0 {
		t.Fatal("Should find similar documents")
	}

	// doc-2 should be more similar than doc-3
	if similar[0].DocumentID != "doc-2" {
		t.Errorf("Expected doc-2 as most similar, got %s", similar[0].DocumentID)
	}
}

func TestFindSimilarNonExistent(t *testing.T) {
	idx := NewTFIDFIndex()

	doc := &Document{
		ID:      "test-1",
		Type:    domain.PersonaElement,
		Name:    "Test",
		Content: "Test content",
	}
	idx.AddDocument(doc)

	similar := idx.FindSimilar("non-existent", 10)

	if similar != nil {
		t.Error("Finding similar for non-existent doc should return nil")
	}
}

func TestFindSimilarWithLimit(t *testing.T) {
	idx := NewTFIDFIndex()

	// Add base document
	baseDoc := &Document{
		ID:      "base",
		Type:    domain.PersonaElement,
		Name:    "Base",
		Content: "test programming code similar content development",
	}
	idx.AddDocument(baseDoc)

	// Add multiple similar documents
	for i := range 5 {
		doc := &Document{
			ID:      string(rune('a' + i)),
			Type:    domain.PersonaElement,
			Name:    "Test",
			Content: "test programming code similar content software",
		}
		idx.AddDocument(doc)
	}

	similar := idx.FindSimilar("base", 2)

	if len(similar) > 2 {
		t.Errorf("Expected max 2 similar docs with limit, got %d", len(similar))
	}
}

func TestTokenizeAndCount(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected map[string]int
	}{
		{
			name:  "Simple text",
			input: "Go is a programming language",
			expected: map[string]int{
				"go":          1,
				"is":          1,
				"programming": 1,
				"language":    1,
			},
		},
		{
			name:  "Repeated words",
			input: "test ",
			expected: map[string]int{
				"test": 3,
			},
		},
		{
			name:     "Single character ignored",
			input:    "a b c test",
			expected: map[string]int{"test": 1},
		},
		{
			name:  "Mixed case",
			input: "Go GO go",
			expected: map[string]int{
				"go": 3,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tokenizeAndCount(tt.input)

			if len(result) != len(tt.expected) {
				t.Errorf("Expected %d terms, got %d", len(tt.expected), len(result))
			}

			for term, count := range tt.expected {
				if result[term] != count {
					t.Errorf("Term %s: expected count %d, got %d", term, count, result[term])
				}
			}
		})
	}
}

func TestExtractHighlights(t *testing.T) {
	content := "This is a test. Go is great for programming. Python is also good."
	queryTerms := map[string]int{
		"go":          1,
		"programming": 1,
	}

	highlights := extractHighlights(content, queryTerms, 2)

	if len(highlights) == 0 {
		t.Fatal("Should extract highlights")
	}

	if len(highlights) > 2 {
		t.Errorf("Expected max 2 highlights, got %d", len(highlights))
	}
}

func TestGetStats(t *testing.T) {
	idx := NewTFIDFIndex()

	// Add documents
	docs := []*Document{
		{ID: "1", Type: domain.PersonaElement, Name: "P1", Content: "go programming"},
		{ID: "2", Type: domain.SkillElement, Name: "S1", Content: "python data science"},
		{ID: "3", Type: domain.PersonaElement, Name: "P2", Content: "testing quality"},
	}

	for _, doc := range docs {
		idx.AddDocument(doc)
	}

	stats := idx.GetStats()

	totalDocs := stats["total_documents"].(int)
	if totalDocs != 3 {
		t.Errorf("Expected 3 total documents, got %d", totalDocs)
	}

	typeCount := stats["documents_by_type"].(map[domain.ElementType]int)
	if typeCount[domain.PersonaElement] != 2 {
		t.Errorf("Expected 2 personas, got %d", typeCount[domain.PersonaElement])
	}
	if typeCount[domain.SkillElement] != 1 {
		t.Errorf("Expected 1 skill, got %d", typeCount[domain.SkillElement])
	}
}

func TestCalculateSimilarity(t *testing.T) {
	idx := NewTFIDFIndex()

	// Build index with multiple documents
	idx.AddDocument(&Document{
		ID:      "1",
		Type:    domain.PersonaElement,
		Name:    "Doc1",
		Content: "go programming language development",
	})

	idx.AddDocument(&Document{
		ID:      "2",
		Type:    domain.PersonaElement,
		Name:    "Doc2",
		Content: "python data science machine learning",
	})

	// Test identical terms
	terms1 := map[string]int{"go": 2, "programming": 2, "language": 1}
	terms2 := map[string]int{"go": 2, "programming": 2, "language": 1}

	similarity := idx.calculateSimilarity(terms1, terms2)

	if similarity < 0.99 {
		t.Errorf("Expected very high similarity (>0.99) for identical terms, got %.4f", similarity)
	}

	// Test different terms
	terms3 := map[string]int{"python": 2, "data": 1}
	similarity = idx.calculateSimilarity(terms1, terms3)

	if similarity > 0.01 {
		t.Errorf("Expected near-zero similarity for different terms, got %.4f", similarity)
	}
}

func BenchmarkAddDocument(b *testing.B) {
	idx := NewTFIDFIndex()
	doc := &Document{
		ID:      "bench-1",
		Type:    domain.PersonaElement,
		Name:    "Benchmark Doc",
		Content: "This is a benchmark document with some content for testing performance",
	}

	b.ResetTimer()
	for range b.N {
		idx.AddDocument(doc)
		idx.RemoveDocument("bench-1")
	}
}

func BenchmarkSearch(b *testing.B) {
	idx := NewTFIDFIndex()

	// Populate index
	for i := range 100 {
		doc := &Document{
			ID:      string(rune(i)),
			Type:    domain.PersonaElement,
			Name:    "Doc",
			Content: "programming software development testing quality assurance",
		}
		idx.AddDocument(doc)
	}

	b.ResetTimer()
	for range b.N {
		idx.Search("programming testing", 10)
	}
}

func BenchmarkFindSimilar(b *testing.B) {
	idx := NewTFIDFIndex()

	// Populate index
	for i := range 100 {
		doc := &Document{
			ID:      string(rune(i)),
			Type:    domain.PersonaElement,
			Name:    "Doc",
			Content: "programming software development testing quality",
		}
		idx.AddDocument(doc)
	}

	b.ResetTimer()
	for range b.N {
		idx.FindSimilar(string(rune(0)), 10)
	}
}
