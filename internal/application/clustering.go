package application

import (
	"context"
	"fmt"
	"math"
	"sort"

	"github.com/fsvxavier/nexs-mcp/internal/domain"
	"github.com/fsvxavier/nexs-mcp/internal/embeddings"
	"github.com/fsvxavier/nexs-mcp/internal/vectorstore"
)

// Cluster represents a group of similar memories.
type Cluster struct {
	ID          int              `json:"id"`
	Centroid    []float32        `json:"centroid"`
	Members     []*domain.Memory `json:"members"`
	Size        int              `json:"size"`
	AvgDistance float32          `json:"avg_distance"`
	Label       string           `json:"label,omitempty"` // Auto-generated label
}

// ClusteringConfig contains configuration for clustering algorithm.
type ClusteringConfig struct {
	Algorithm       string  // "dbscan" or "kmeans"
	MinClusterSize  int     // Minimum memories per cluster (DBSCAN)
	EpsilonDistance float32 // Distance threshold (DBSCAN)
	NumClusters     int     // Number of clusters (K-means)
	MaxIterations   int     // Max iterations for K-means
}

// ClusteringService provides memory clustering capabilities.
type ClusteringService struct {
	store      *vectorstore.Store
	provider   embeddings.Provider
	repository ElementRepository
	config     ClusteringConfig
}

// NewClusteringService creates a new clustering service.
func NewClusteringService(provider embeddings.Provider, repository ElementRepository, config ClusteringConfig) *ClusteringService {
	// Default configuration
	if config.Algorithm == "" {
		config.Algorithm = "dbscan"
	}
	if config.MinClusterSize == 0 {
		config.MinClusterSize = 3 // At least 3 memories per cluster
	}
	if config.EpsilonDistance == 0 {
		config.EpsilonDistance = 0.15 // 15% distance threshold
	}
	if config.NumClusters == 0 {
		config.NumClusters = 10 // Default 10 clusters for K-means
	}
	if config.MaxIterations == 0 {
		config.MaxIterations = 100
	}

	return &ClusteringService{
		store:      vectorstore.NewStore(provider),
		provider:   provider,
		repository: repository,
		config:     config,
	}
}

// ClusterMemories clusters memories using the configured algorithm.
func (c *ClusteringService) ClusterMemories(ctx context.Context) ([]Cluster, error) {
	// Get all memories
	memories, err := c.getMemories(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get memories: %w", err)
	}

	if len(memories) < c.config.MinClusterSize {
		return []Cluster{}, nil // Not enough memories to cluster
	}

	// Generate embeddings for all memories
	embeddings := make([][]float32, len(memories))
	for i, memory := range memories {
		embedding, err := c.provider.Embed(ctx, memory.Content)
		if err != nil {
			return nil, fmt.Errorf("failed to embed memory %s: %w", memory.GetID(), err)
		}
		embeddings[i] = embedding
	}

	// Choose clustering algorithm
	var clusters []Cluster
	switch c.config.Algorithm {
	case "dbscan":
		clusters, err = c.dbscan(memories, embeddings)
	case "kmeans":
		clusters, err = c.kmeans(memories, embeddings)
	default:
		return nil, fmt.Errorf("unknown clustering algorithm: %s", c.config.Algorithm)
	}

	if err != nil {
		return nil, fmt.Errorf("clustering failed: %w", err)
	}

	// Generate labels for clusters
	for i := range clusters {
		clusters[i].Label = c.generateClusterLabel(&clusters[i])
	}

	return clusters, nil
}

// dbscan implements the DBSCAN (Density-Based Spatial Clustering) algorithm.
func (c *ClusteringService) dbscan(memories []*domain.Memory, embeddings [][]float32) ([]Cluster, error) {
	n := len(memories)
	visited := make([]bool, n)
	clusterID := make([]int, n)
	for i := range clusterID {
		clusterID[i] = -1 // -1 means noise/unassigned
	}

	currentCluster := 0

	for i := 0; i < n; i++ {
		if visited[i] {
			continue
		}
		visited[i] = true

		// Find neighbors within epsilon distance
		neighbors := c.findNeighbors(i, embeddings, c.config.EpsilonDistance)

		// If not enough neighbors, mark as noise
		if len(neighbors) < c.config.MinClusterSize {
			continue
		}

		// Start new cluster
		clusterID[i] = currentCluster
		queue := neighbors

		// Expand cluster
		for len(queue) > 0 {
			neighborIdx := queue[0]
			queue = queue[1:]

			if visited[neighborIdx] {
				continue
			}
			visited[neighborIdx] = true
			clusterID[neighborIdx] = currentCluster

			// Find neighbors of neighbor
			neighborNeighbors := c.findNeighbors(neighborIdx, embeddings, c.config.EpsilonDistance)
			if len(neighborNeighbors) >= c.config.MinClusterSize {
				queue = append(queue, neighborNeighbors...)
			}
		}

		currentCluster++
	}

	// Convert cluster assignments to Cluster objects
	clusterMap := make(map[int]*Cluster)
	for i := 0; i < n; i++ {
		cid := clusterID[i]
		if cid == -1 {
			continue // Skip noise
		}

		if _, exists := clusterMap[cid]; !exists {
			clusterMap[cid] = &Cluster{
				ID:      cid,
				Members: []*domain.Memory{},
			}
		}
		clusterMap[cid].Members = append(clusterMap[cid].Members, memories[i])
	}

	// Compute centroids and statistics
	clusters := make([]Cluster, 0, len(clusterMap))
	for _, cluster := range clusterMap {
		cluster.Size = len(cluster.Members)
		cluster.Centroid = c.computeCentroid(cluster.Members, embeddings, clusterID)
		cluster.AvgDistance = c.computeAvgDistance(cluster.Members, cluster.Centroid, embeddings, clusterID)
		clusters = append(clusters, *cluster)
	}

	// Sort by size (largest first)
	sort.Slice(clusters, func(i, j int) bool {
		return clusters[i].Size > clusters[j].Size
	})

	return clusters, nil
}

// kmeans implements the K-means clustering algorithm.
func (c *ClusteringService) kmeans(memories []*domain.Memory, embeddings [][]float32) ([]Cluster, error) {
	n := len(memories)
	k := c.config.NumClusters
	if k > n {
		k = n
	}

	dim := len(embeddings[0])

	// Initialize centroids randomly
	centroids := make([][]float32, k)
	for i := 0; i < k; i++ {
		centroids[i] = make([]float32, dim)
		copy(centroids[i], embeddings[i%n])
	}

	assignments := make([]int, n)
	converged := false
	iteration := 0

	for !converged && iteration < c.config.MaxIterations {
		iteration++
		changed := false

		// Assignment step: assign each point to nearest centroid
		for i := 0; i < n; i++ {
			minDist := float32(math.MaxFloat32)
			bestCluster := 0

			for j := 0; j < k; j++ {
				dist := c.euclideanDistance(embeddings[i], centroids[j])
				if dist < minDist {
					minDist = dist
					bestCluster = j
				}
			}

			if assignments[i] != bestCluster {
				assignments[i] = bestCluster
				changed = true
			}
		}

		if !changed {
			converged = true
			break
		}

		// Update step: recompute centroids
		counts := make([]int, k)
		newCentroids := make([][]float32, k)
		for i := 0; i < k; i++ {
			newCentroids[i] = make([]float32, dim)
		}

		for i := 0; i < n; i++ {
			cluster := assignments[i]
			counts[cluster]++
			for d := 0; d < dim; d++ {
				newCentroids[cluster][d] += embeddings[i][d]
			}
		}

		for i := 0; i < k; i++ {
			if counts[i] > 0 {
				for d := 0; d < dim; d++ {
					newCentroids[i][d] /= float32(counts[i])
				}
				centroids[i] = newCentroids[i]
			}
		}
	}

	// Convert assignments to Cluster objects
	clusterMap := make(map[int]*Cluster)
	for i := 0; i < n; i++ {
		cid := assignments[i]
		if _, exists := clusterMap[cid]; !exists {
			clusterMap[cid] = &Cluster{
				ID:       cid,
				Centroid: centroids[cid],
				Members:  []*domain.Memory{},
			}
		}
		clusterMap[cid].Members = append(clusterMap[cid].Members, memories[i])
	}

	// Compute statistics
	clusters := make([]Cluster, 0, len(clusterMap))
	for _, cluster := range clusterMap {
		cluster.Size = len(cluster.Members)
		cluster.AvgDistance = c.computeAvgDistanceKmeans(cluster.Members, cluster.Centroid, embeddings, assignments)
		clusters = append(clusters, *cluster)
	}

	// Sort by size
	sort.Slice(clusters, func(i, j int) bool {
		return clusters[i].Size > clusters[j].Size
	})

	return clusters, nil
}

// findNeighbors finds all neighbors within epsilon distance.
func (c *ClusteringService) findNeighbors(idx int, embeddings [][]float32, epsilon float32) []int {
	neighbors := []int{}
	for i := 0; i < len(embeddings); i++ {
		if i == idx {
			continue
		}
		dist := c.euclideanDistance(embeddings[idx], embeddings[i])
		if dist <= epsilon {
			neighbors = append(neighbors, i)
		}
	}
	return neighbors
}

// euclideanDistance computes Euclidean distance between two vectors.
func (c *ClusteringService) euclideanDistance(a, b []float32) float32 {
	if len(a) != len(b) {
		return float32(math.MaxFloat32)
	}

	sum := float32(0)
	for i := 0; i < len(a); i++ {
		diff := a[i] - b[i]
		sum += diff * diff
	}
	return float32(math.Sqrt(float64(sum)))
}

// computeCentroid computes the centroid of a cluster.
func (c *ClusteringService) computeCentroid(members []*domain.Memory, embeddings [][]float32, clusterID []int) []float32 {
	if len(members) == 0 {
		return nil
	}

	dim := len(embeddings[0])
	centroid := make([]float32, dim)

	// Sum all embeddings
	for i := range embeddings {
		if i < len(clusterID) {
			for d := 0; d < dim; d++ {
				centroid[d] += embeddings[i][d]
			}
		}
	}

	// Average
	for d := 0; d < dim; d++ {
		centroid[d] /= float32(len(members))
	}

	return centroid
}

// computeAvgDistance computes average distance from members to centroid.
func (c *ClusteringService) computeAvgDistance(members []*domain.Memory, centroid []float32, embeddings [][]float32, clusterID []int) float32 {
	if len(members) == 0 || centroid == nil {
		return 0
	}

	totalDist := float32(0)
	count := 0

	for i, emb := range embeddings {
		if i < len(clusterID) {
			dist := c.euclideanDistance(emb, centroid)
			totalDist += dist
			count++
		}
	}

	if count == 0 {
		return 0
	}
	return totalDist / float32(count)
}

// computeAvgDistanceKmeans computes average distance for k-means.
func (c *ClusteringService) computeAvgDistanceKmeans(members []*domain.Memory, centroid []float32, embeddings [][]float32, assignments []int) float32 {
	if len(members) == 0 {
		return 0
	}

	totalDist := float32(0)
	for i, emb := range embeddings {
		if i < len(assignments) {
			dist := c.euclideanDistance(emb, centroid)
			totalDist += dist
		}
	}

	return totalDist / float32(len(members))
}

// generateClusterLabel generates a descriptive label for a cluster.
func (c *ClusteringService) generateClusterLabel(cluster *Cluster) string {
	if cluster.Size == 0 {
		return "Empty Cluster"
	}

	// Extract keywords from member names
	keywords := make(map[string]int)
	for _, member := range cluster.Members {
		name := member.GetMetadata().Name
		keywords[name]++
	}

	// Find most common keyword
	maxCount := 0
	commonKeyword := "Memories"
	for keyword, count := range keywords {
		if count > maxCount {
			maxCount = count
			commonKeyword = keyword
		}
	}

	return fmt.Sprintf("Cluster %d: %s (%d memories)", cluster.ID, commonKeyword, cluster.Size)
}

// GetClusterByID retrieves a specific cluster by ID.
func (c *ClusteringService) GetClusterByID(ctx context.Context, clusterID int) (*Cluster, error) {
	clusters, err := c.ClusterMemories(ctx)
	if err != nil {
		return nil, err
	}

	for _, cluster := range clusters {
		if cluster.ID == clusterID {
			return &cluster, nil
		}
	}

	return nil, fmt.Errorf("cluster %d not found", clusterID)
}

// getMemories retrieves all memories from the repository.
func (c *ClusteringService) getMemories(ctx context.Context) ([]*domain.Memory, error) {
	memoryType := domain.MemoryElement
	elements, err := c.repository.List(domain.ElementFilter{
		Type: &memoryType,
	})
	if err != nil {
		return nil, err
	}

	memories := make([]*domain.Memory, 0, len(elements))
	for _, element := range elements {
		if memory, ok := element.(*domain.Memory); ok {
			memories = append(memories, memory)
		}
	}

	return memories, nil
}
