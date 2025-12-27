package mcp

import (
	"context"
	"encoding/json"
	"fmt"

	sdk "github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/fsvxavier/nexs-mcp/internal/application"
)

// --- Consolidation Tool Input/Output Structures ---

// ConsolidateMemoriesInput defines input for consolidate_memories tool.
type ConsolidateMemoriesInput struct {
	DetectDuplicates      bool    `json:"detect_duplicates"`
	ClusterMemories       bool    `json:"cluster_memories"`
	ExtractKnowledge      bool    `json:"extract_knowledge"`
	AutoMerge             bool    `json:"auto_merge"`
	SimilarityThreshold   float32 `json:"similarity_threshold"`
	MinSimilarityForMerge float32 `json:"min_similarity_for_merge"`
	ClusteringAlgorithm   string  `json:"clustering_algorithm"`
	NumClusters           int     `json:"num_clusters"`
}

// DetectDuplicatesInput defines input for detect_duplicates tool.
type DetectDuplicatesInput struct {
	SimilarityThreshold float32 `json:"similarity_threshold"`
	MinContentLength    int     `json:"min_content_length"`
	MaxResults          int     `json:"max_results"`
}

// MergeDuplicatesInput defines input for merge_duplicates tool.
type MergeDuplicatesInput struct {
	RepresentativeID string   `json:"representative_id"`
	DuplicateIDs     []string `json:"duplicate_ids"`
}

// ClusterMemoriesInput defines input for cluster_memories tool.
type ClusterMemoriesInput struct {
	Algorithm       string  `json:"algorithm"`
	MinClusterSize  int     `json:"min_cluster_size"`
	EpsilonDistance float32 `json:"epsilon_distance"`
	NumClusters     int     `json:"num_clusters"`
}

// ExtractKnowledgeInput defines input for extract_knowledge tool.
type ExtractKnowledgeInput struct {
	MemoryIDs []string `json:"memory_ids"`
}

// FindSimilarMemoriesInput defines input for find_similar_memories tool.
type FindSimilarMemoriesInput struct {
	MemoryID  string  `json:"memory_id"`
	Threshold float32 `json:"threshold"`
}

// GetClusterDetailsInput defines input for get_cluster_details tool.
type GetClusterDetailsInput struct {
	ClusterID int `json:"cluster_id"`
}

// ComputeSimilarityInput defines input for compute_similarity tool.
type ComputeSimilarityInput struct {
	MemoryID1 string `json:"memory_id_1"`
	MemoryID2 string `json:"memory_id_2"`
}

// --- MCP Tool Registration ---

// RegisterConsolidationTools registers all memory consolidation tools with the MCP server.
func (s *MCPServer) RegisterConsolidationTools() {
	provider := s.hybridSearch.Provider()

	sdk.AddTool(s.server, &sdk.Tool{
		Name:        "consolidate_memories",
		Description: "Performs comprehensive memory consolidation: detects duplicates, clusters memories, extracts knowledge graphs, and generates merge recommendations",
	}, func(ctx context.Context, params map[string]interface{}) (*sdk.CallToolResult, error) {
		var input ConsolidateMemoriesInput
		if err := mapToStruct(params, &input); err != nil {
			return createErrorResult(fmt.Sprintf("Invalid input: %v", err))
		}

		// Set defaults
		if input.SimilarityThreshold == 0 {
			input.SimilarityThreshold = 0.95
		}
		if input.MinSimilarityForMerge == 0 {
			input.MinSimilarityForMerge = 0.98
		}
		if input.ClusteringAlgorithm == "" {
			input.ClusteringAlgorithm = "dbscan"
		}
		if input.NumClusters == 0 {
			input.NumClusters = 10
		}

		service := application.NewMemoryConsolidationService(
			provider, s.repo,
			application.DuplicateDetectionConfig{SimilarityThreshold: input.SimilarityThreshold},
			application.ClusteringConfig{Algorithm: input.ClusteringAlgorithm, NumClusters: input.NumClusters},
		)

		options := application.ConsolidationOptions{
			DetectDuplicates:      input.DetectDuplicates || (!input.DetectDuplicates && !input.ClusterMemories && !input.ExtractKnowledge),
			ClusterMemories:       input.ClusterMemories || (!input.DetectDuplicates && !input.ClusterMemories && !input.ExtractKnowledge),
			ExtractKnowledge:      input.ExtractKnowledge || (!input.DetectDuplicates && !input.ClusterMemories && !input.ExtractKnowledge),
			AutoMerge:             input.AutoMerge,
			MinSimilarityForMerge: input.MinSimilarityForMerge,
		}

		report, err := service.ConsolidateMemories(ctx, options)
		if err != nil {
			return createErrorResult(fmt.Sprintf("Consolidation failed: %v", err))
		}

		return createSuccessResult(map[string]interface{}{"report": report})
	})

	sdk.AddTool(s.server, &sdk.Tool{
		Name:        "detect_duplicates",
		Description: "Detects duplicate and near-duplicate memories using HNSW semantic similarity",
	}, func(ctx context.Context, params map[string]interface{}) (*sdk.CallToolResult, error) {
		var input DetectDuplicatesInput
		if err := mapToStruct(params, &input); err != nil {
			return createErrorResult(fmt.Sprintf("Invalid input: %v", err))
		}

		service := application.NewDuplicateDetectionService(
			provider, s.repo,
			application.DuplicateDetectionConfig{
				SimilarityThreshold: input.SimilarityThreshold,
				MinContentLength:    input.MinContentLength,
				MaxResults:          input.MaxResults,
			},
		)

		groups, err := service.DetectDuplicates(ctx)
		if err != nil {
			return createErrorResult(fmt.Sprintf("Duplicate detection failed: %v", err))
		}

		totalDuplicates := 0
		for _, group := range groups {
			totalDuplicates += group.Count - 1
		}

		return createSuccessResult(map[string]interface{}{
			"duplicate_groups": groups,
			"total_groups":     len(groups),
			"total_duplicates": totalDuplicates,
		})
	})

	sdk.AddTool(s.server, &sdk.Tool{
		Name:        "merge_duplicates",
		Description: "Merges duplicate memories into a single consolidated memory",
	}, func(ctx context.Context, params map[string]interface{}) (*sdk.CallToolResult, error) {
		var input MergeDuplicatesInput
		if err := mapToStruct(params, &input); err != nil {
			return createErrorResult(fmt.Sprintf("Invalid input: %v", err))
		}

		service := application.NewDuplicateDetectionService(provider, s.repo, application.DuplicateDetectionConfig{})
		merged, err := service.MergeDuplicates(ctx, input.RepresentativeID, input.DuplicateIDs)
		if err != nil {
			return createErrorResult(fmt.Sprintf("Merge failed: %v", err))
		}

		return createSuccessResult(map[string]interface{}{
			"merged_memory": map[string]interface{}{
				"id":           merged.GetID(),
				"name":         merged.GetMetadata().Name,
				"content":      merged.Content,
				"date_created": merged.DateCreated,
			},
			"merged_count": len(input.DuplicateIDs) + 1,
		})
	})

	sdk.AddTool(s.server, &sdk.Tool{
		Name:        "cluster_memories",
		Description: "Clusters memories by semantic similarity using DBSCAN or K-means algorithms",
	}, func(ctx context.Context, params map[string]interface{}) (*sdk.CallToolResult, error) {
		var input ClusterMemoriesInput
		if err := mapToStruct(params, &input); err != nil {
			return createErrorResult(fmt.Sprintf("Invalid input: %v", err))
		}

		service := application.NewClusteringService(
			provider, s.repo,
			application.ClusteringConfig{
				Algorithm:       input.Algorithm,
				MinClusterSize:  input.MinClusterSize,
				EpsilonDistance: input.EpsilonDistance,
				NumClusters:     input.NumClusters,
			},
		)

		clusters, err := service.ClusterMemories(ctx)
		if err != nil {
			return createErrorResult(fmt.Sprintf("Clustering failed: %v", err))
		}

		totalMemories := 0
		for _, cluster := range clusters {
			totalMemories += cluster.Size
		}

		return createSuccessResult(map[string]interface{}{
			"clusters":       clusters,
			"total_clusters": len(clusters),
			"total_memories": totalMemories,
		})
	})

	sdk.AddTool(s.server, &sdk.Tool{
		Name:        "extract_knowledge",
		Description: "Extracts entities, relationships, concepts, and keywords from memory content to build knowledge graphs",
	}, func(ctx context.Context, params map[string]interface{}) (*sdk.CallToolResult, error) {
		var input ExtractKnowledgeInput
		if err := mapToStruct(params, &input); err != nil {
			return createErrorResult(fmt.Sprintf("Invalid input: %v", err))
		}

		extractor := application.NewKnowledgeGraphExtractor(s.repo)
		graph, err := extractor.ExtractFromMultipleMemories(ctx, input.MemoryIDs)
		if err != nil {
			return createErrorResult(fmt.Sprintf("Knowledge extraction failed: %v", err))
		}

		return createSuccessResult(map[string]interface{}{"knowledge_graph": graph})
	})

	sdk.AddTool(s.server, &sdk.Tool{
		Name:        "find_similar_memories",
		Description: "Finds memories similar to a given memory using semantic similarity",
	}, func(ctx context.Context, params map[string]interface{}) (*sdk.CallToolResult, error) {
		var input FindSimilarMemoriesInput
		if err := mapToStruct(params, &input); err != nil {
			return createErrorResult(fmt.Sprintf("Invalid input: %v", err))
		}

		if input.Threshold == 0 {
			input.Threshold = 0.85
		}

		consolidation := application.NewMemoryConsolidationService(
			provider, s.repo,
			application.DuplicateDetectionConfig{},
			application.ClusteringConfig{},
		)

		similar, err := consolidation.FindSimilarMemories(ctx, input.MemoryID, input.Threshold)
		if err != nil {
			return createErrorResult(fmt.Sprintf("Similar search failed: %v", err))
		}

		return createSuccessResult(map[string]interface{}{
			"similar_memories": similar,
			"count":            len(similar),
		})
	})

	sdk.AddTool(s.server, &sdk.Tool{
		Name:        "get_cluster_details",
		Description: "Retrieves detailed information about a specific memory cluster including knowledge graph",
	}, func(ctx context.Context, params map[string]interface{}) (*sdk.CallToolResult, error) {
		var input GetClusterDetailsInput
		if err := mapToStruct(params, &input); err != nil {
			return createErrorResult(fmt.Sprintf("Invalid input: %v", err))
		}

		consolidation := application.NewMemoryConsolidationService(
			provider, s.repo,
			application.DuplicateDetectionConfig{},
			application.ClusteringConfig{},
		)

		details, err := consolidation.GetClusterDetails(ctx, input.ClusterID)
		if err != nil {
			return createErrorResult(fmt.Sprintf("Failed to get cluster details: %v", err))
		}

		return createSuccessResult(map[string]interface{}{"details": details})
	})

	sdk.AddTool(s.server, &sdk.Tool{
		Name:        "get_consolidation_stats",
		Description: "Retrieves statistics about memory consolidation (duplicates, clusters, etc.)",
	}, func(ctx context.Context, params map[string]interface{}) (*sdk.CallToolResult, error) {
		consolidation := application.NewMemoryConsolidationService(
			provider, s.repo,
			application.DuplicateDetectionConfig{},
			application.ClusteringConfig{},
		)

		stats, err := consolidation.GetConsolidationStatistics(ctx)
		if err != nil {
			return createErrorResult(fmt.Sprintf("Failed to get statistics: %v", err))
		}

		return createSuccessResult(map[string]interface{}{"statistics": stats})
	})

	sdk.AddTool(s.server, &sdk.Tool{
		Name:        "compute_similarity",
		Description: "Computes cosine similarity between two memories",
	}, func(ctx context.Context, params map[string]interface{}) (*sdk.CallToolResult, error) {
		var input ComputeSimilarityInput
		if err := mapToStruct(params, &input); err != nil {
			return createErrorResult(fmt.Sprintf("Invalid input: %v", err))
		}

		consolidation := application.NewMemoryConsolidationService(
			provider, s.repo,
			application.DuplicateDetectionConfig{},
			application.ClusteringConfig{},
		)

		similarity, err := consolidation.ComputeSimilarity(ctx, input.MemoryID1, input.MemoryID2)
		if err != nil {
			return createErrorResult(fmt.Sprintf("Similarity computation failed: %v", err))
		}

		return createSuccessResult(map[string]interface{}{
			"memory_id_1": input.MemoryID1,
			"memory_id_2": input.MemoryID2,
			"similarity":  similarity,
		})
	})
}

// Helper function to convert map to struct
func mapToStruct(m map[string]interface{}, result interface{}) error {
	bytes, err := json.Marshal(m)
	if err != nil {
		return err
	}
	return json.Unmarshal(bytes, result)
}
