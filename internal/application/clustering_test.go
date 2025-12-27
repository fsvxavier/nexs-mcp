package application

import (
	"context"
	"testing"

	"github.com/fsvxavier/nexs-mcp/internal/domain"
	"github.com/fsvxavier/nexs-mcp/internal/embeddings"
)

func TestNewClusteringService(t *testing.T) {
	repo := newMockRepoForIndex()
	provider := embeddings.NewMockProvider("mock", 128)
	config := ClusteringConfig{}

	service := NewClusteringService(provider, repo, config)

	if service == nil {
		t.Fatal("Expected non-nil service")
	}
	if service.config.Algorithm != "dbscan" {
		t.Errorf("Expected default algorithm 'dbscan', got: %s", service.config.Algorithm)
	}
	if service.config.MinClusterSize != 3 {
		t.Errorf("Expected default min cluster size 3, got: %d", service.config.MinClusterSize)
	}
	if service.config.EpsilonDistance != 0.15 {
		t.Errorf("Expected default epsilon 0.15, got: %f", service.config.EpsilonDistance)
	}
	if service.config.NumClusters != 10 {
		t.Errorf("Expected default num clusters 10, got: %d", service.config.NumClusters)
	}
	if service.config.MaxIterations != 100 {
		t.Errorf("Expected default max iterations 100, got: %d", service.config.MaxIterations)
	}
}

func TestClusterMemories_DBSCAN(t *testing.T) {
	repo := newMockRepoForIndex()
	provider := embeddings.NewMockProvider("mock", 128)
	config := ClusteringConfig{
		Algorithm:      "dbscan",
		MinClusterSize: 2,
	}

	service := NewClusteringService(provider, repo, config)

	// Create memories with similar content
	mem1 := domain.NewMemory("Mem1", "Machine learning algorithms for classification", "1.0.0", "test")
	mem2 := domain.NewMemory("Mem2", "Deep learning neural networks", "1.0.0", "test")
	mem3 := domain.NewMemory("Mem3", "Database optimization techniques", "1.0.0", "test")
	mem4 := domain.NewMemory("Mem4", "SQL query performance tuning", "1.0.0", "test")

	repo.Create(mem1)
	repo.Create(mem2)
	repo.Create(mem3)
	repo.Create(mem4)

	ctx := context.Background()
	clusters, err := service.ClusterMemories(ctx)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	// Should create some clusters
	if clusters == nil {
		t.Error("Expected non-nil clusters slice")
	}
}

func TestClusterMemories_KMeans(t *testing.T) {
	repo := newMockRepoForIndex()
	provider := embeddings.NewMockProvider("mock", 128)
	config := ClusteringConfig{
		Algorithm:   "kmeans",
		NumClusters: 2,
	}

	service := NewClusteringService(provider, repo, config)

	// Create enough memories for clustering
	for i := range 10 {
		mem := domain.NewMemory("Mem"+string(rune(i)), "Content "+string(rune(i)), "1.0.0", "test")
		repo.Create(mem)
	}

	ctx := context.Background()
	clusters, err := service.ClusterMemories(ctx)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	// K-means should create requested number of clusters (if enough data)
	if clusters == nil {
		t.Error("Expected non-nil clusters slice")
	}
}

func TestClusterMemories_InsufficientData(t *testing.T) {
	repo := newMockRepoForIndex()
	provider := embeddings.NewMockProvider("mock", 128)
	config := ClusteringConfig{
		MinClusterSize: 5,
	}

	service := NewClusteringService(provider, repo, config)

	// Create too few memories
	mem1 := domain.NewMemory("Mem1", "Content", "1.0.0", "test")
	mem2 := domain.NewMemory("Mem2", "Content", "1.0.0", "test")

	repo.Create(mem1)
	repo.Create(mem2)

	ctx := context.Background()
	clusters, err := service.ClusterMemories(ctx)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if len(clusters) != 0 {
		t.Errorf("Expected 0 clusters for insufficient data, got: %d", len(clusters))
	}
}

func TestClusterMemories_EmptyRepository(t *testing.T) {
	repo := newMockRepoForIndex()
	provider := embeddings.NewMockProvider("mock", 128)
	config := ClusteringConfig{}

	service := NewClusteringService(provider, repo, config)

	ctx := context.Background()
	clusters, err := service.ClusterMemories(ctx)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if len(clusters) != 0 {
		t.Errorf("Expected 0 clusters for empty repo, got: %d", len(clusters))
	}
}

func TestClusterMemories_InvalidAlgorithm(t *testing.T) {
	repo := newMockRepoForIndex()
	provider := embeddings.NewMockProvider("mock", 128)
	config := ClusteringConfig{
		Algorithm: "invalid-algorithm",
	}

	service := NewClusteringService(provider, repo, config)

	// Create some memories
	mem := domain.NewMemory("Mem", "Content", "1.0.0", "test")
	repo.Create(mem)

	ctx := context.Background()
	// Service may default to DBSCAN for invalid algorithm
	clusters, err := service.ClusterMemories(ctx)
	if err != nil {
		t.Logf("ClusterMemories returned error (acceptable): %v", err)
	}
	if clusters != nil {
		t.Logf("Service handled invalid algorithm gracefully")
	}
}

func TestCluster_Structure(t *testing.T) {
	mem1 := domain.NewMemory("Mem1", "Content 1", "1.0.0", "test")
	mem2 := domain.NewMemory("Mem2", "Content 2", "1.0.0", "test")

	cluster := Cluster{
		ID:          1,
		Centroid:    []float32{0.1, 0.2, 0.3},
		Members:     []*domain.Memory{mem1, mem2},
		Size:        2,
		AvgDistance: 0.15,
		Label:       "Test Cluster",
	}

	if cluster.ID != 1 {
		t.Errorf("Expected cluster ID 1, got %d", cluster.ID)
	}
	if cluster.Size != 2 {
		t.Errorf("Expected cluster size 2, got %d", cluster.Size)
	}
	if len(cluster.Members) != 2 {
		t.Errorf("Expected 2 members, got %d", len(cluster.Members))
	}
	if len(cluster.Centroid) != 3 {
		t.Errorf("Expected centroid with 3 dimensions, got %d", len(cluster.Centroid))
	}
	if cluster.AvgDistance != 0.15 {
		t.Errorf("Expected avg distance 0.15, got %f", cluster.AvgDistance)
	}
	if cluster.Label != "Test Cluster" {
		t.Errorf("Expected label 'Test Cluster', got %s", cluster.Label)
	}
}

func TestClusteringConfig_Defaults(t *testing.T) {
	tests := []struct {
		name     string
		config   ClusteringConfig
		expected ClusteringConfig
	}{
		{
			name:   "empty config applies defaults",
			config: ClusteringConfig{},
			expected: ClusteringConfig{
				Algorithm:       "dbscan",
				MinClusterSize:  3,
				EpsilonDistance: 0.15,
				NumClusters:     10,
				MaxIterations:   100,
			},
		},
		{
			name: "partial config preserves custom values",
			config: ClusteringConfig{
				Algorithm:   "kmeans",
				NumClusters: 5,
			},
			expected: ClusteringConfig{
				Algorithm:       "kmeans",
				MinClusterSize:  3,
				EpsilonDistance: 0.15,
				NumClusters:     5,
				MaxIterations:   100,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := newMockRepoForIndex()
			provider := embeddings.NewMockProvider("mock", 128)

			service := NewClusteringService(provider, repo, tt.config)

			if service.config.Algorithm != tt.expected.Algorithm {
				t.Errorf("Expected algorithm %s, got %s", tt.expected.Algorithm, service.config.Algorithm)
			}
			if service.config.MinClusterSize != tt.expected.MinClusterSize {
				t.Errorf("Expected min cluster size %d, got %d", tt.expected.MinClusterSize, service.config.MinClusterSize)
			}
			if service.config.EpsilonDistance != tt.expected.EpsilonDistance {
				t.Errorf("Expected epsilon %f, got %f", tt.expected.EpsilonDistance, service.config.EpsilonDistance)
			}
			if service.config.NumClusters != tt.expected.NumClusters {
				t.Errorf("Expected num clusters %d, got %d", tt.expected.NumClusters, service.config.NumClusters)
			}
			if service.config.MaxIterations != tt.expected.MaxIterations {
				t.Errorf("Expected max iterations %d, got %d", tt.expected.MaxIterations, service.config.MaxIterations)
			}
		})
	}
}

func TestClusterMemories_DBSCAN_MinSize(t *testing.T) {
	repo := newMockRepoForIndex()
	provider := embeddings.NewMockProvider("mock", 128)
	config := ClusteringConfig{
		Algorithm:      "dbscan",
		MinClusterSize: 3, // Require at least 3 members
	}

	service := NewClusteringService(provider, repo, config)

	// Create memories
	for i := range 5 {
		mem := domain.NewMemory("Mem"+string(rune(i)), "Similar content", "1.0.0", "test")
		repo.Create(mem)
	}

	ctx := context.Background()
	clusters, err := service.ClusterMemories(ctx)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	// Each cluster should have at least MinClusterSize members
	for _, cluster := range clusters {
		if cluster.Size < config.MinClusterSize {
			t.Errorf("Cluster %d has size %d, expected at least %d", cluster.ID, cluster.Size, config.MinClusterSize)
		}
	}
}

func TestClusterMemories_KMeans_NumClusters(t *testing.T) {
	repo := newMockRepoForIndex()
	provider := embeddings.NewMockProvider("mock", 128)
	config := ClusteringConfig{
		Algorithm:   "kmeans",
		NumClusters: 3,
	}

	service := NewClusteringService(provider, repo, config)

	// Create enough memories
	for i := range 15 {
		mem := domain.NewMemory("Mem"+string(rune(i)), "Content "+string(rune(i)), "1.0.0", "test")
		repo.Create(mem)
	}

	ctx := context.Background()
	clusters, err := service.ClusterMemories(ctx)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	// K-means should attempt to create requested number of clusters
	// (actual count may vary depending on convergence)
	if len(clusters) > config.NumClusters+1 {
		t.Errorf("Expected around %d clusters, got %d", config.NumClusters, len(clusters))
	}
}

func TestClusterMemories_WithLabels(t *testing.T) {
	repo := newMockRepoForIndex()
	provider := embeddings.NewMockProvider("mock", 128)
	config := ClusteringConfig{
		Algorithm: "kmeans",
	}

	service := NewClusteringService(provider, repo, config)

	// Create memories with keywords
	mem1 := domain.NewMemory("Mem1", "machine learning algorithms", "1.0.0", "test")
	mem2 := domain.NewMemory("Mem2", "database optimization", "1.0.0", "test")

	repo.Create(mem1)
	repo.Create(mem2)

	ctx := context.Background()
	clusters, err := service.ClusterMemories(ctx)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	// Clusters may have auto-generated labels
	for _, cluster := range clusters {
		// Label can be empty or auto-generated
		if cluster.Label != "" {
			t.Logf("Cluster %d has label: %s", cluster.ID, cluster.Label)
		}
	}
}

func TestClusterMemories_AvgDistance(t *testing.T) {
	repo := newMockRepoForIndex()
	provider := embeddings.NewMockProvider("mock", 128)
	config := ClusteringConfig{
		Algorithm: "kmeans",
	}

	service := NewClusteringService(provider, repo, config)

	// Create memories
	for i := range 10 {
		mem := domain.NewMemory("Mem"+string(rune(i)), "Content", "1.0.0", "test")
		repo.Create(mem)
	}

	ctx := context.Background()
	clusters, err := service.ClusterMemories(ctx)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	// Average distance should be non-negative
	for _, cluster := range clusters {
		if cluster.AvgDistance < 0 {
			t.Errorf("Cluster %d has negative avg distance: %f", cluster.ID, cluster.AvgDistance)
		}
	}
}

func TestClusterMemories_CentroidDimensions(t *testing.T) {
	repo := newMockRepoForIndex()
	provider := embeddings.NewMockProvider("mock", 128)
	config := ClusteringConfig{
		Algorithm: "kmeans",
	}

	service := NewClusteringService(provider, repo, config)

	// Create memories
	for i := range 5 {
		mem := domain.NewMemory("Mem"+string(rune(i)), "Content", "1.0.0", "test")
		repo.Create(mem)
	}

	ctx := context.Background()
	clusters, err := service.ClusterMemories(ctx)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	expectedDim := provider.Dimensions()

	for _, cluster := range clusters {
		if len(cluster.Centroid) != expectedDim {
			t.Errorf("Expected centroid dimension %d, got %d", expectedDim, len(cluster.Centroid))
		}
	}
}

func TestClusterMemories_MemberConsistency(t *testing.T) {
	repo := newMockRepoForIndex()
	provider := embeddings.NewMockProvider("mock", 128)
	config := ClusteringConfig{
		Algorithm: "kmeans",
	}

	service := NewClusteringService(provider, repo, config)

	// Create memories
	for i := range 8 {
		mem := domain.NewMemory("Mem"+string(rune(i)), "Content", "1.0.0", "test")
		repo.Create(mem)
	}

	ctx := context.Background()
	clusters, err := service.ClusterMemories(ctx)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	// Size should match number of members
	for _, cluster := range clusters {
		if cluster.Size != len(cluster.Members) {
			t.Errorf("Cluster %d: Size=%d but Members=%d", cluster.ID, cluster.Size, len(cluster.Members))
		}
	}
}
