package application

import (
	"context"
	"fmt"
	"strings"
	"testing"

	"github.com/fsvxavier/nexs-mcp/internal/domain"
	"github.com/fsvxavier/nexs-mcp/internal/infrastructure"
)

// Helper function to create test memory with unique name.
func createTestMemory(content string) *domain.Memory {
	// Use content hash to create unique names
	memory := domain.NewMemory(content[:min(20, len(content))], content, "1.0", "test")
	memory.Content = content
	return memory
}

// TestNewTopicModeler verifies topic modeler initialization.
func TestNewTopicModeler(t *testing.T) {
	config := TopicModelingConfig{
		Algorithm:        "lda",
		NumTopics:        5,
		MaxIterations:    100,
		MinWordFrequency: 2,
		MaxWordFrequency: 0.8,
		TopKeywords:      10,
		RandomSeed:       42,
		Alpha:            0.1,
		Beta:             0.01,
		UseONNX:          false,
	}

	repo := infrastructure.NewInMemoryElementRepository()
	modeler := NewTopicModeler(config, repo, nil)

	if modeler == nil {
		t.Fatal("Expected topic modeler instance, got nil")
	}

	if modeler.config.Algorithm != "lda" {
		t.Errorf("Expected algorithm 'lda', got %s", modeler.config.Algorithm)
	}

	if modeler.config.NumTopics != 5 {
		t.Errorf("Expected 5 topics, got %d", modeler.config.NumTopics)
	}
}

// TestTokenize tests the tokenization function.
func TestTokenize(t *testing.T) {
	config := TopicModelingConfig{
		Algorithm:        "lda",
		NumTopics:        5,
		MinWordFrequency: 1,
		UseONNX:          false,
	}

	modeler := NewTopicModeler(config, nil, nil)

	tests := []struct {
		name             string
		text             string
		shouldContain    []string
		shouldNotContain []string
	}{
		{
			name:          "simple text",
			text:          "hello world",
			shouldContain: []string{"hello", "world"},
		},
		{
			name:             "with stopwords",
			text:             "the quick brown fox jumps the lazy dog",
			shouldContain:    []string{"quick", "brown", "fox", "jumps", "lazy", "dog"},
			shouldNotContain: []string{"the"},
		},
		{
			name:          "with punctuation",
			text:          "Hello, World! Programming.",
			shouldContain: []string{"hello", "world", "programming"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tokens := modeler.tokenize(tt.text)

			// Check for expected tokens
			for _, expected := range tt.shouldContain {
				found := false
				for _, token := range tokens {
					if token == expected {
						found = true
						break
					}
				}
				if !found {
					t.Errorf("Expected token '%s' not found in %v", expected, tokens)
				}
			}

			// Check tokens that shouldn't be there
			for _, notExpected := range tt.shouldNotContain {
				for _, token := range tokens {
					if token == notExpected {
						t.Errorf("Unexpected token '%s' found in %v", notExpected, tokens)
					}
				}
			}
		})
	}
}

// TestExtractTopicsLDA tests basic LDA topic extraction.
func TestExtractTopicsLDA(t *testing.T) {
	config := TopicModelingConfig{
		Algorithm:        "lda",
		NumTopics:        2,
		MaxIterations:    30,
		MinWordFrequency: 1,
		MaxWordFrequency: 1.0,
		TopKeywords:      5,
		RandomSeed:       42,
		Alpha:            0.1,
		Beta:             0.01,
		UseONNX:          false,
	}

	repo := infrastructure.NewInMemoryElementRepository()
	ctx := context.Background()

	// Create test memories with distinct topics
	testContents := []string{
		"machine learning algorithms neural networks deep learning artificial intelligence training models",
		"python programming code development software engineering testing debugging implementation",
		"database sql queries optimization indexing performance tuning transactions",
		"neural networks training backpropagation gradient descent optimization algorithms",
		"software development agile methodology testing deployment continuous integration",
		"database management systems transactions ACID properties normalization",
	}

	memoryIDs := make([]string, 0, len(testContents))
	for _, content := range testContents {
		memory := createTestMemory(content)
		if err := repo.Create(memory); err != nil {
			t.Fatalf("Failed to create memory: %v", err)
		}
		memoryIDs = append(memoryIDs, memory.GetID())
	}

	modeler := NewTopicModeler(config, repo, nil)
	topics, err := modeler.ExtractTopics(ctx, memoryIDs)
	if err != nil {
		t.Fatalf("ExtractTopics failed: %v", err)
	}

	if len(topics) == 0 {
		t.Fatal("No topics returned")
	}

	// Verify each topic has keywords
	for i, topic := range topics {
		if len(topic.Keywords) == 0 {
			t.Errorf("Topic %d has no keywords", i)
		}

		// Verify keywords have weights
		for j, kw := range topic.Keywords {
			if kw.Weight <= 0 {
				t.Errorf("Topic %d keyword %d (%s) has invalid weight: %f", i, j, kw.Word, kw.Weight)
			}
		}

		// Verify topic has coherence score
		if topic.Coherence < 0 || topic.Coherence > 1 {
			t.Errorf("Topic %d coherence %f out of range [0, 1]", i, topic.Coherence)
		}

		// Verify topic has diversity score
		if topic.Diversity < 0 || topic.Diversity > 1 {
			t.Errorf("Topic %d diversity %f out of range [0, 1]", i, topic.Diversity)
		}

		t.Logf("Topic %d: %d keywords, coherence=%.2f, diversity=%.2f",
			i, len(topic.Keywords), topic.Coherence, topic.Diversity)
	}
}

// TestExtractTopicsNMF tests NMF topic extraction.
func TestExtractTopicsNMF(t *testing.T) {
	config := TopicModelingConfig{
		Algorithm:        "nmf",
		NumTopics:        2,
		MaxIterations:    30,
		MinWordFrequency: 1,
		MaxWordFrequency: 1.0,
		TopKeywords:      5,
		RandomSeed:       42,
		UseONNX:          false,
	}

	repo := infrastructure.NewInMemoryElementRepository()
	ctx := context.Background()

	// Create test memories
	testContents := []string{
		"cloud computing infrastructure scalability kubernetes docker containers orchestration",
		"cloud services aws azure deployment microservices architecture patterns",
		"data science analytics machine learning predictions models statistics",
		"data analysis visualization statistics regression classification algorithms",
	}

	memoryIDs := make([]string, 0, len(testContents))
	for _, content := range testContents {
		memory := createTestMemory(content)
		if err := repo.Create(memory); err != nil {
			t.Fatalf("Failed to create memory: %v", err)
		}
		memoryIDs = append(memoryIDs, memory.GetID())
	}

	modeler := NewTopicModeler(config, repo, nil)
	topics, err := modeler.ExtractTopics(ctx, memoryIDs)
	if err != nil {
		t.Fatalf("ExtractTopics failed: %v", err)
	}

	if len(topics) == 0 {
		t.Fatal("No topics returned")
	}

	// NMF should produce valid topics
	for i, topic := range topics {
		if len(topic.Keywords) == 0 {
			t.Errorf("Topic %d has no keywords", i)
		}
		t.Logf("NMF Topic %d: %d keywords", i, len(topic.Keywords))
	}
}

// TestExtractTopicsEmptyMemories tests behavior with empty memories.
func TestExtractTopicsEmptyMemories(t *testing.T) {
	config := TopicModelingConfig{
		Algorithm: "lda",
		NumTopics: 3,
		UseONNX:   false,
	}

	repo := infrastructure.NewInMemoryElementRepository()
	ctx := context.Background()

	modeler := NewTopicModeler(config, repo, nil)

	// Test with empty memory list
	_, err := modeler.ExtractTopics(ctx, []string{})
	if err == nil {
		t.Error("Expected error with empty memory list")
	}

	// Test with non-existent memories
	_, err = modeler.ExtractTopics(ctx, []string{"nonexistent"})
	if err == nil {
		t.Error("Expected error with non-existent memories")
	}
}

// TestLDAvsNMF compares LDA and NMF results.
func TestLDAvsNMF(t *testing.T) {
	repo := infrastructure.NewInMemoryElementRepository()
	ctx := context.Background()

	// Create test memories with substantive content
	testContents := []string{
		"machine learning algorithms neural networks deep learning models training artificial intelligence data science predictive analytics computational models statistical methods supervised unsupervised reinforcement learning gradient descent backpropagation optimization techniques feature engineering model evaluation",
		"software development programming code testing deployment methodologies agile scrum kanban continuous integration continuous deployment version control testing frameworks unit tests integration tests test driven development behavior driven development acceptance testing automated testing manual testing debugging profiling performance optimization refactoring code review",
		"database query optimization performance systems management relational databases nosql mongodb postgresql mysql database design normalization indexing query planning execution plans transaction management acid properties consistency availability partition tolerance distributed databases sharding replication architecture administration backup recovery disaster recovery",
		"cloud computing infrastructure services platform containers kubernetes docker orchestration microservices serverless functions event driven architecture message queues pub patterns scalability horizontal vertical balancers reverse proxies gateways service mesh distributed tracing monitoring observability logging metrics alerting dashboards visualization analytics reporting",
		"data engineering pipelines extract transform batch processing stream processing apache spark flink kafka real-time analytics warehousing lakes governance metadata quality validation cleansing transformation aggregation dimensional modeling star schema snowflake marts business intelligence reporting visualization tableau power looker studio",
		"cybersecurity encryption authentication authorization password hashing factor biometric certificate authorities public infrastructure transport security firewalls intrusion detection prevention penetration vulnerability scanning threat modeling audits compliance regulations privacy protection standards frameworks policies procedures incident response forensics malware ransomware phishing",
	}

	memoryIDs := make([]string, 0, len(testContents))
	for i, content := range testContents {
		// Create unique memory with index to avoid ID collisions
		memory := domain.NewMemory(
			fmt.Sprintf("test-memory-compare-%d", i),
			"Test memory for topic modeling comparison",
			"v1.0.0",
			"test-user",
		)
		memory.Content = content // Set the actual content
		if err := repo.Create(memory); err != nil {
			t.Fatalf("Failed to store memory: %v", err)
		}
		memoryIDs = append(memoryIDs, memory.GetID())
	}

	// Test LDA
	ldaConfig := TopicModelingConfig{
		Algorithm:        "lda",
		NumTopics:        3,
		MaxIterations:    50,
		MinWordFrequency: 1,
		MaxWordFrequency: 1.0,
		UseONNX:          false,
	}

	ldaModeler := NewTopicModeler(ldaConfig, repo, nil)
	ldaTopics, err := ldaModeler.ExtractTopics(ctx, memoryIDs)
	if err != nil {
		t.Fatalf("LDA failed: %v", err)
	}

	// Test NMF
	nmfConfig := TopicModelingConfig{
		Algorithm:        "nmf",
		NumTopics:        3,
		MaxIterations:    50,
		MinWordFrequency: 1,
		MaxWordFrequency: 1.0,
		UseONNX:          false,
	}

	nmfModeler := NewTopicModeler(nmfConfig, repo, nil)
	nmfTopics, err := nmfModeler.ExtractTopics(ctx, memoryIDs)
	if err != nil {
		t.Fatalf("NMF failed: %v", err)
	}

	// Both should produce topics
	if len(ldaTopics) == 0 {
		t.Error("LDA produced no topics")
	}

	if len(nmfTopics) == 0 {
		t.Error("NMF produced no topics")
	}

	t.Logf("LDA produced %d topics, NMF produced %d topics", len(ldaTopics), len(nmfTopics))

	// Log first keywords from each
	if len(ldaTopics) > 0 && len(ldaTopics[0].Keywords) > 0 {
		var ldaWords []string
		for _, kw := range ldaTopics[0].Keywords {
			ldaWords = append(ldaWords, kw.Word)
		}
		t.Logf("LDA Topic 0 keywords: %s", strings.Join(ldaWords, ", "))
	}

	if len(nmfTopics) > 0 && len(nmfTopics[0].Keywords) > 0 {
		var nmfWords []string
		for _, kw := range nmfTopics[0].Keywords {
			nmfWords = append(nmfWords, kw.Word)
		}
		t.Logf("NMF Topic 0 keywords: %s", strings.Join(nmfWords, ", "))
	}
}
