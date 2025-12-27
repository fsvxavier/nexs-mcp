package application

import (
	"context"
	"fmt"
	"time"

	"github.com/fsvxavier/nexs-mcp/internal/domain"
	"github.com/fsvxavier/nexs-mcp/internal/embeddings"
)

// ConsolidationReport contains the results of memory consolidation.
type ConsolidationReport struct {
	TotalMemories     int                   `json:"total_memories"`
	DuplicateGroups   []DuplicateGroup      `json:"duplicate_groups"`
	Clusters          []Cluster             `json:"clusters"`
	KnowledgeGraphs   []*KnowledgeGraph     `json:"knowledge_graphs,omitempty"`
	RecommendedMerges []MergeRecommendation `json:"recommended_merges"`
	ProcessingTime    time.Duration         `json:"processing_time"`
	QualityScore      float32               `json:"quality_score"`
	Timestamp         time.Time             `json:"timestamp"`
}

// MergeRecommendation suggests merging duplicate memories.
type MergeRecommendation struct {
	RepresentativeID string   `json:"representative_id"`
	DuplicateIDs     []string `json:"duplicate_ids"`
	Similarity       float32  `json:"similarity"`
	Confidence       float32  `json:"confidence"`
	Reason           string   `json:"reason"`
}

// ConsolidationOptions configures the consolidation process.
type ConsolidationOptions struct {
	DetectDuplicates      bool                     // Run duplicate detection
	ClusterMemories       bool                     // Run clustering
	ExtractKnowledge      bool                     // Extract knowledge graphs
	AutoMerge             bool                     // Automatically merge duplicates
	DuplicateConfig       DuplicateDetectionConfig // Duplicate detection settings
	ClusteringConfig      ClusteringConfig         // Clustering settings
	MinSimilarityForMerge float32                  // Minimum similarity to auto-merge
}

// MemoryConsolidationService orchestrates memory consolidation operations.
type MemoryConsolidationService struct {
	duplicateDetector  *DuplicateDetectionService
	clusteringService  *ClusteringService
	knowledgeExtractor *KnowledgeGraphExtractor
	repository         ElementRepository
	provider           embeddings.Provider
}

// NewMemoryConsolidationService creates a new memory consolidation service.
func NewMemoryConsolidationService(
	provider embeddings.Provider,
	repository ElementRepository,
	duplicateConfig DuplicateDetectionConfig,
	clusteringConfig ClusteringConfig,
) *MemoryConsolidationService {
	return &MemoryConsolidationService{
		duplicateDetector:  NewDuplicateDetectionService(provider, repository, duplicateConfig),
		clusteringService:  NewClusteringService(provider, repository, clusteringConfig),
		knowledgeExtractor: NewKnowledgeGraphExtractor(repository),
		repository:         repository,
		provider:           provider,
	}
}

// ConsolidateMemories performs comprehensive memory consolidation.
func (m *MemoryConsolidationService) ConsolidateMemories(ctx context.Context, options ConsolidationOptions) (*ConsolidationReport, error) {
	startTime := time.Now()

	report := &ConsolidationReport{
		Timestamp:         startTime,
		DuplicateGroups:   []DuplicateGroup{},
		Clusters:          []Cluster{},
		KnowledgeGraphs:   []*KnowledgeGraph{},
		RecommendedMerges: []MergeRecommendation{},
	}

	// Count total memories
	memories, err := m.getMemories(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get memories: %w", err)
	}
	report.TotalMemories = len(memories)

	// Step 1: Detect duplicates
	if options.DetectDuplicates {
		duplicateGroups, err := m.duplicateDetector.DetectDuplicates(ctx)
		if err != nil {
			return nil, fmt.Errorf("duplicate detection failed: %w", err)
		}
		report.DuplicateGroups = duplicateGroups

		// Generate merge recommendations
		report.RecommendedMerges = m.generateMergeRecommendations(duplicateGroups, options.MinSimilarityForMerge)

		// Auto-merge if enabled
		if options.AutoMerge {
			for _, rec := range report.RecommendedMerges {
				if rec.Confidence >= 0.95 { // High confidence only
					_, err := m.duplicateDetector.MergeDuplicates(ctx, rec.RepresentativeID, rec.DuplicateIDs)
					if err != nil {
						// Log error but continue
						continue
					}
				}
			}
		}
	}

	// Step 2: Cluster memories
	if options.ClusterMemories {
		clusters, err := m.clusteringService.ClusterMemories(ctx)
		if err != nil {
			return nil, fmt.Errorf("clustering failed: %w", err)
		}
		report.Clusters = clusters

		// Step 3: Extract knowledge graphs from clusters
		if options.ExtractKnowledge {
			for i := range clusters {
				graph, err := m.knowledgeExtractor.ExtractFromCluster(ctx, &clusters[i])
				if err != nil {
					continue // Skip on error
				}
				report.KnowledgeGraphs = append(report.KnowledgeGraphs, graph)
			}
		}
	}

	// Compute quality score
	report.QualityScore = m.computeQualityScore(report)

	report.ProcessingTime = time.Since(startTime)

	return report, nil
}

// DetectDuplicatesOnly runs only duplicate detection.
func (m *MemoryConsolidationService) DetectDuplicatesOnly(ctx context.Context) ([]DuplicateGroup, error) {
	return m.duplicateDetector.DetectDuplicates(ctx)
}

// ClusterMemoriesOnly runs only clustering.
func (m *MemoryConsolidationService) ClusterMemoriesOnly(ctx context.Context) ([]Cluster, error) {
	return m.clusteringService.ClusterMemories(ctx)
}

// ExtractKnowledgeOnly runs only knowledge extraction.
func (m *MemoryConsolidationService) ExtractKnowledgeOnly(ctx context.Context, memoryIDs []string) (*KnowledgeGraph, error) {
	return m.knowledgeExtractor.ExtractFromMultipleMemories(ctx, memoryIDs)
}

// MergeDuplicates merges duplicate memories.
func (m *MemoryConsolidationService) MergeDuplicates(ctx context.Context, representativeID string, duplicateIDs []string) (*domain.Memory, error) {
	return m.duplicateDetector.MergeDuplicates(ctx, representativeID, duplicateIDs)
}

// FindSimilarMemories finds memories similar to a given memory.
func (m *MemoryConsolidationService) FindSimilarMemories(ctx context.Context, memoryID string, threshold float32) ([]domain.Memory, error) {
	// Temporarily update config
	originalThreshold := m.duplicateDetector.config.SimilarityThreshold
	m.duplicateDetector.config.SimilarityThreshold = threshold
	defer func() {
		m.duplicateDetector.config.SimilarityThreshold = originalThreshold
	}()

	return m.duplicateDetector.FindDuplicatesForMemory(ctx, memoryID)
}

// GetClusterDetails retrieves detailed information about a specific cluster.
func (m *MemoryConsolidationService) GetClusterDetails(ctx context.Context, clusterID int) (*ClusterDetails, error) {
	cluster, err := m.clusteringService.GetClusterByID(ctx, clusterID)
	if err != nil {
		return nil, err
	}

	// Extract knowledge graph for this cluster
	graph, err := m.knowledgeExtractor.ExtractFromCluster(ctx, cluster)
	if err != nil {
		return nil, err
	}

	return &ClusterDetails{
		Cluster:        *cluster,
		KnowledgeGraph: graph,
	}, nil
}

// ClusterDetails combines cluster information with knowledge graph.
type ClusterDetails struct {
	Cluster        Cluster         `json:"cluster"`
	KnowledgeGraph *KnowledgeGraph `json:"knowledge_graph"`
}

// ComputeSimilarity computes similarity between two memories.
func (m *MemoryConsolidationService) ComputeSimilarity(ctx context.Context, memoryID1, memoryID2 string) (float32, error) {
	return m.duplicateDetector.ComputeSimilarity(ctx, memoryID1, memoryID2)
}

// generateMergeRecommendations generates merge recommendations from duplicate groups.
func (m *MemoryConsolidationService) generateMergeRecommendations(groups []DuplicateGroup, minSimilarity float32) []MergeRecommendation {
	recommendations := []MergeRecommendation{}

	for _, group := range groups {
		if group.Similarity < minSimilarity {
			continue
		}

		duplicateIDs := make([]string, len(group.Duplicates))
		for i, dup := range group.Duplicates {
			duplicateIDs[i] = dup.GetID()
		}

		rec := MergeRecommendation{
			RepresentativeID: group.Representative.GetID(),
			DuplicateIDs:     duplicateIDs,
			Similarity:       group.Similarity,
			Confidence:       m.computeMergeConfidence(group),
			Reason:           fmt.Sprintf("%d highly similar memories detected", group.Count),
		}
		recommendations = append(recommendations, rec)
	}

	return recommendations
}

// computeMergeConfidence computes confidence score for a merge recommendation.
func (m *MemoryConsolidationService) computeMergeConfidence(group DuplicateGroup) float32 {
	// Base confidence from similarity
	confidence := group.Similarity

	// Adjust based on group size (larger groups = more confident)
	if group.Count >= 5 {
		confidence *= 1.1
	} else if group.Count == 2 {
		confidence *= 0.9
	}

	// Cap at 0.99
	if confidence > 0.99 {
		confidence = 0.99
	}

	return confidence
}

// computeQualityScore computes overall quality score for the consolidation report.
func (m *MemoryConsolidationService) computeQualityScore(report *ConsolidationReport) float32 {
	if report.TotalMemories == 0 {
		return 0
	}

	score := float32(1.0)

	// Penalize for duplicates
	totalDuplicates := 0
	for _, group := range report.DuplicateGroups {
		totalDuplicates += group.Count - 1 // -1 because representative isn't duplicate
	}
	duplicateRatio := float32(totalDuplicates) / float32(report.TotalMemories)
	score -= duplicateRatio * 0.3 // Max 30% penalty for duplicates

	// Reward for good clustering (not too many small clusters)
	if len(report.Clusters) > 0 {
		avgClusterSize := float32(report.TotalMemories) / float32(len(report.Clusters))
		if avgClusterSize >= 5 {
			score += 0.1 // Bonus for good cluster sizes
		}
	}

	// Reward for knowledge extraction
	if len(report.KnowledgeGraphs) > 0 {
		score += 0.1
	}

	// Cap between 0 and 1
	if score < 0 {
		score = 0
	}
	if score > 1 {
		score = 1
	}

	return score
}

// GetConsolidationStatistics returns statistics about memory consolidation.
func (m *MemoryConsolidationService) GetConsolidationStatistics(ctx context.Context) (*ConsolidationStatistics, error) {
	memories, err := m.getMemories(ctx)
	if err != nil {
		return nil, err
	}

	stats := &ConsolidationStatistics{
		TotalMemories: len(memories),
		Timestamp:     time.Now(),
	}

	// Count duplicates
	duplicateGroups, err := m.duplicateDetector.DetectDuplicates(ctx)
	if err == nil {
		stats.DuplicateCount = 0
		for _, group := range duplicateGroups {
			stats.DuplicateCount += group.Count - 1
		}
		stats.DuplicateGroups = len(duplicateGroups)
	}

	// Count clusters
	clusters, err := m.clusteringService.ClusterMemories(ctx)
	if err == nil {
		stats.ClusterCount = len(clusters)
		if len(clusters) > 0 {
			stats.AvgClusterSize = float32(len(memories)) / float32(len(clusters))
		}
	}

	return stats, nil
}

// ConsolidationStatistics contains statistics about memory consolidation.
type ConsolidationStatistics struct {
	TotalMemories   int       `json:"total_memories"`
	DuplicateCount  int       `json:"duplicate_count"`
	DuplicateGroups int       `json:"duplicate_groups"`
	ClusterCount    int       `json:"cluster_count"`
	AvgClusterSize  float32   `json:"avg_cluster_size"`
	Timestamp       time.Time `json:"timestamp"`
}

// getMemories retrieves all memories from the repository.
func (m *MemoryConsolidationService) getMemories(ctx context.Context) ([]*domain.Memory, error) {
	elements, err := m.repository.List(domain.ElementFilter{
		Types: []domain.ElementType{domain.MemoryElement},
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
