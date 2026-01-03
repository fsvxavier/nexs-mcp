package mcp

import (
	"context"
	"errors"
	"fmt"
	"time"

	sdk "github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/fsvxavier/nexs-mcp/internal/application"
)

// --- Consolidation Tool Input/Output Structures ---

// ConsolidateMemoriesInput defines input for consolidate_memories tool.
type ConsolidateMemoriesInput struct {
	DetectDuplicates      bool    `json:"detect_duplicates"        jsonschema:"run duplicate detection (default: true)"`
	ClusterMemories       bool    `json:"cluster_memories"         jsonschema:"run clustering analysis (default: true)"`
	ExtractKnowledge      bool    `json:"extract_knowledge"        jsonschema:"extract knowledge graphs (default: true)"`
	AutoMerge             bool    `json:"auto_merge"               jsonschema:"automatically merge high-confidence duplicates (default: false)"`
	SimilarityThreshold   float32 `json:"similarity_threshold"     jsonschema:"similarity threshold for duplicates (0.0-1.0, default: 0.95)"`
	MinSimilarityForMerge float32 `json:"min_similarity_for_merge" jsonschema:"minimum similarity for auto-merge (default: 0.98)"`
	ClusteringAlgorithm   string  `json:"clustering_algorithm"     jsonschema:"clustering algorithm: 'dbscan' or 'kmeans' (default: 'dbscan')"`
	NumClusters           int     `json:"num_clusters"             jsonschema:"number of clusters for kmeans (default: 10)"`
}

// ConsolidateMemoriesOutput defines output for consolidate_memories tool.
type ConsolidateMemoriesOutput struct {
	Report application.ConsolidationReport `json:"report" jsonschema:"comprehensive consolidation report"`
}

// DetectDuplicatesInput defines input for detect_duplicates tool.
type DetectDuplicatesInput struct {
	SimilarityThreshold float32 `json:"similarity_threshold" jsonschema:"similarity threshold (0.0-1.0, default: 0.95)"`
	MinContentLength    int     `json:"min_content_length"   jsonschema:"minimum content length to consider (default: 20)"`
	MaxResults          int     `json:"max_results"          jsonschema:"maximum duplicate groups to return (default: 100)"`
}

// DetectDuplicatesOutput defines output for detect_duplicates tool.
type DetectDuplicatesOutput struct {
	DuplicateGroups []application.MemoryDuplicateGroup `json:"duplicate_groups" jsonschema:"groups of duplicate memories"`
	TotalGroups     int                                `json:"total_groups"     jsonschema:"total number of duplicate groups"`
	TotalDuplicates int                                `json:"total_duplicates" jsonschema:"total duplicate memories found"`
}

// MergeDuplicatesInput defines input for merge_duplicates tool.
type MergeDuplicatesInput struct {
	RepresentativeID string   `json:"representative_id" jsonschema:"ID of memory to keep as representative"`
	DuplicateIDs     []string `json:"duplicate_ids"     jsonschema:"IDs of duplicate memories to merge"`
}

// MergeDuplicatesOutput defines output for merge_duplicates tool.
type MergeDuplicatesOutput struct {
	MergedMemoryID string `json:"merged_memory_id" jsonschema:"ID of the merged memory"`
	MergedMemory   string `json:"merged_memory"    jsonschema:"content of merged memory"`
	MergedCount    int    `json:"merged_count"     jsonschema:"number of memories merged"`
}

// ClusterMemoriesInput defines input for cluster_memories tool.
type ClusterMemoriesInput struct {
	Algorithm       string  `json:"algorithm"        jsonschema:"clustering algorithm: 'dbscan' or 'kmeans' (default: 'dbscan')"`
	MinClusterSize  int     `json:"min_cluster_size" jsonschema:"minimum memories per cluster for DBSCAN (default: 3)"`
	EpsilonDistance float32 `json:"epsilon_distance" jsonschema:"distance threshold for DBSCAN (default: 0.15)"`
	NumClusters     int     `json:"num_clusters"     jsonschema:"number of clusters for kmeans (default: 10)"`
}

// ClusterMemoriesOutput defines output for cluster_memories tool.
type ClusterMemoriesOutput struct {
	Clusters      []application.Cluster `json:"clusters"       jsonschema:"memory clusters"`
	TotalClusters int                   `json:"total_clusters" jsonschema:"total number of clusters"`
	TotalMemories int                   `json:"total_memories" jsonschema:"total memories clustered"`
}

// ExtractKnowledgeInput defines input for extract_knowledge tool.
type ExtractKnowledgeInput struct {
	MemoryIDs []string `json:"memory_ids" jsonschema:"IDs of memories to extract knowledge from"`
}

// ExtractKnowledgeOutput defines output for extract_knowledge tool.
type ExtractKnowledgeOutput struct {
	KnowledgeGraph application.KnowledgeGraph `json:"knowledge_graph" jsonschema:"extracted knowledge graph"`
}

// ConsolidationFindSimilarInput defines input for find_similar_memories tool (consolidation).
type ConsolidationFindSimilarInput struct {
	MemoryID  string  `json:"memory_id" jsonschema:"ID of memory to find similar ones for"`
	Threshold float32 `json:"threshold" jsonschema:"similarity threshold (0.0-1.0, default: 0.85)"`
}

// ConsolidationFindSimilarOutput defines output for find_similar_memories tool (consolidation).
type ConsolidationFindSimilarOutput struct {
	OriginalMemoryID string          `json:"original_memory_id" jsonschema:"the original memory ID"`
	SimilarMemories  []MemorySummary `json:"similar_memories"   jsonschema:"similar memories found"`
	Count            int             `json:"count"              jsonschema:"number of similar memories"`
}

// GetClusterDetailsInput defines input for get_cluster_details tool.
type GetClusterDetailsInput struct {
	ClusterID int `json:"cluster_id" jsonschema:"ID of cluster to get details for"`
}

// GetClusterDetailsOutput defines output for get_cluster_details tool.
type GetClusterDetailsOutput struct {
	Details application.ClusterDetails `json:"details" jsonschema:"detailed cluster information with knowledge graph"`
}

// GetConsolidationStatsOutput defines output for get_consolidation_stats tool.
type GetConsolidationStatsOutput struct {
	Statistics application.ConsolidationStatistics `json:"statistics" jsonschema:"consolidation statistics"`
}

// ComputeSimilarityInput defines input for compute_similarity tool.
type ComputeSimilarityInput struct {
	MemoryID1 string `json:"memory_id_1" jsonschema:"ID of first memory"`
	MemoryID2 string `json:"memory_id_2" jsonschema:"ID of second memory"`
}

// ComputeSimilarityOutput defines output for compute_similarity tool.
type ComputeSimilarityOutput struct {
	MemoryID1  string  `json:"memory_id_1" jsonschema:"ID of first memory"`
	MemoryID2  string  `json:"memory_id_2" jsonschema:"ID of second memory"`
	Similarity float32 `json:"similarity"  jsonschema:"cosine similarity score (0.0-1.0)"`
}

// --- MCP Tool Registration ---

// RegisterConsolidationTools registers all memory consolidation tools with the MCP server.
func (s *MCPServer) RegisterConsolidationTools() {
	// consolidate_memories
	sdk.AddTool(s.server, &sdk.Tool{
		Name:        "consolidate_memories",
		Description: "Performs comprehensive memory consolidation: detects duplicates, clusters memories, extracts knowledge graphs, and generates merge recommendations",
	}, s.handleConsolidateMemories)

	// detect_duplicates
	sdk.AddTool(s.server, &sdk.Tool{
		Name:        "detect_duplicates",
		Description: "Detects duplicate and near-duplicate memories using HNSW semantic similarity",
	}, s.handleDetectDuplicates)

	// merge_duplicates
	sdk.AddTool(s.server, &sdk.Tool{
		Name:        "merge_duplicates",
		Description: "Merges duplicate memories into a single consolidated memory",
	}, s.handleMergeDuplicates)

	// cluster_memories
	sdk.AddTool(s.server, &sdk.Tool{
		Name:        "cluster_memories",
		Description: "Clusters memories by semantic similarity using DBSCAN or K-means algorithms",
	}, s.handleClusterMemories)

	// extract_knowledge
	sdk.AddTool(s.server, &sdk.Tool{
		Name:        "extract_knowledge",
		Description: "Extracts entities, relationships, concepts, and keywords from memory content to build knowledge graphs",
	}, s.handleExtractKnowledge)

	// find_similar_memories
	sdk.AddTool(s.server, &sdk.Tool{
		Name:        "find_similar_memories",
		Description: "Finds memories similar to a given memory using semantic similarity",
	}, s.handleFindSimilarMemories)

	// get_cluster_details
	sdk.AddTool(s.server, &sdk.Tool{
		Name:        "get_cluster_details",
		Description: "Retrieves detailed information about a specific memory cluster including knowledge graph",
	}, s.handleGetClusterDetails)

	// get_consolidation_stats
	sdk.AddTool(s.server, &sdk.Tool{
		Name:        "get_consolidation_stats",
		Description: "Retrieves statistics about memory consolidation (duplicates, clusters, etc.)",
	}, s.handleGetConsolidationStats)

	// compute_similarity
	sdk.AddTool(s.server, &sdk.Tool{
		Name:        "compute_similarity",
		Description: "Computes cosine similarity between two memories",
	}, s.handleComputeSimilarity)
}

// --- Tool Handlers ---

func (s *MCPServer) handleConsolidateMemories(ctx context.Context, req *sdk.CallToolRequest, input ConsolidateMemoriesInput) (*sdk.CallToolResult, ConsolidateMemoriesOutput, error) {
	startTime := time.Now()
	var err error
	defer func() {
		s.metrics.RecordToolCall(application.ToolCallMetric{
			ToolName:  "consolidate_memories",
			Timestamp: startTime,
			Duration:  time.Since(startTime),
			Success:   err == nil,
			ErrorMessage: func() string {
				if err != nil {
					return err.Error()
				}
				return ""
			}(),
		})
	}()

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

	// Create consolidation service
	provider := s.hybridSearch.Provider()
	service := application.NewMemoryConsolidationService(
		provider,
		s.repo,
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

	report, errConsolidate := service.ConsolidateMemories(ctx, options)
	err = errConsolidate
	if err != nil {
		return nil, ConsolidateMemoriesOutput{}, fmt.Errorf("consolidation failed: %w", err)
	}

	output := ConsolidateMemoriesOutput{Report: *report}

	// Measure response size and record token metrics
	s.responseMiddleware.MeasureResponseSize(ctx, "consolidate_memories", output)

	return nil, output, nil
}

func (s *MCPServer) handleDetectDuplicates(ctx context.Context, req *sdk.CallToolRequest, input DetectDuplicatesInput) (*sdk.CallToolResult, DetectDuplicatesOutput, error) {
	startTime := time.Now()
	var err error
	defer func() {
		s.metrics.RecordToolCall(application.ToolCallMetric{
			ToolName:  "detect_duplicates",
			Timestamp: startTime,
			Duration:  time.Since(startTime),
			Success:   err == nil,
			ErrorMessage: func() string {
				if err != nil {
					return err.Error()
				}
				return ""
			}(),
		})
	}()

	provider := s.hybridSearch.Provider()
	service := application.NewDuplicateDetectionService(
		provider,
		s.repo,
		application.DuplicateDetectionConfig{
			SimilarityThreshold: input.SimilarityThreshold,
			MinContentLength:    input.MinContentLength,
			MaxResults:          input.MaxResults,
		},
	)

	groups, errDetect := service.DetectDuplicates(ctx)
	err = errDetect
	if err != nil {
		return nil, DetectDuplicatesOutput{}, fmt.Errorf("duplicate detection failed: %w", err)
	}

	totalDuplicates := 0
	for _, group := range groups {
		totalDuplicates += group.Count - 1
	}

	output := DetectDuplicatesOutput{
		DuplicateGroups: groups,
		TotalGroups:     len(groups),
		TotalDuplicates: totalDuplicates,
	}

	// Measure response size and record token metrics
	s.responseMiddleware.MeasureResponseSize(ctx, "detect_duplicates", output)

	return nil, output, nil
}

func (s *MCPServer) handleMergeDuplicates(ctx context.Context, req *sdk.CallToolRequest, input MergeDuplicatesInput) (*sdk.CallToolResult, MergeDuplicatesOutput, error) {
	startTime := time.Now()
	var handlerErr error
	defer func() {
		s.metrics.RecordToolCall(application.ToolCallMetric{
			ToolName:  "merge_duplicates",
			Timestamp: startTime,
			Duration:  time.Since(startTime),
			Success:   handlerErr == nil,
			ErrorMessage: func() string {
				if handlerErr != nil {
					return handlerErr.Error()
				}
				return ""
			}(),
		})
	}()

	if input.RepresentativeID == "" {
		handlerErr = errors.New("representative_id is required")
		return nil, MergeDuplicatesOutput{}, handlerErr
	}
	if len(input.DuplicateIDs) == 0 {
		handlerErr = errors.New("duplicate_ids is required")
		return nil, MergeDuplicatesOutput{}, handlerErr
	}

	provider := s.hybridSearch.Provider()
	service := application.NewDuplicateDetectionService(provider, s.repo, application.DuplicateDetectionConfig{})

	merged, err := service.MergeDuplicates(ctx, input.RepresentativeID, input.DuplicateIDs)
	if err != nil {
		handlerErr = fmt.Errorf("merge failed: %w", err)
		return nil, MergeDuplicatesOutput{}, handlerErr
	}

	output := MergeDuplicatesOutput{
		MergedMemoryID: merged.GetID(),
		MergedMemory:   merged.Content,
		MergedCount:    len(input.DuplicateIDs) + 1,
	}

	// Measure response size and record token metrics
	s.responseMiddleware.MeasureResponseSize(ctx, "merge_duplicates", output)

	return nil, output, nil
}

func (s *MCPServer) handleClusterMemories(ctx context.Context, req *sdk.CallToolRequest, input ClusterMemoriesInput) (*sdk.CallToolResult, ClusterMemoriesOutput, error) {
	startTime := time.Now()
	var err error
	defer func() {
		s.metrics.RecordToolCall(application.ToolCallMetric{
			ToolName:  "cluster_memories",
			Timestamp: startTime,
			Duration:  time.Since(startTime),
			Success:   err == nil,
			ErrorMessage: func() string {
				if err != nil {
					return err.Error()
				}
				return ""
			}(),
		})
	}()

	provider := s.hybridSearch.Provider()
	service := application.NewClusteringService(
		provider,
		s.repo,
		application.ClusteringConfig{
			Algorithm:       input.Algorithm,
			MinClusterSize:  input.MinClusterSize,
			EpsilonDistance: input.EpsilonDistance,
			NumClusters:     input.NumClusters,
		},
	)

	clusters, errCluster := service.ClusterMemories(ctx)
	err = errCluster
	if err != nil {
		return nil, ClusterMemoriesOutput{}, fmt.Errorf("clustering failed: %w", err)
	}

	totalMemories := 0
	for _, cluster := range clusters {
		totalMemories += cluster.Size
	}

	output := ClusterMemoriesOutput{
		Clusters:      clusters,
		TotalClusters: len(clusters),
		TotalMemories: totalMemories,
	}

	// Measure response size and record token metrics
	s.responseMiddleware.MeasureResponseSize(ctx, "cluster_memories", output)

	return nil, output, nil
}

func (s *MCPServer) handleExtractKnowledge(ctx context.Context, req *sdk.CallToolRequest, input ExtractKnowledgeInput) (*sdk.CallToolResult, ExtractKnowledgeOutput, error) {
	startTime := time.Now()
	var handlerErr error
	defer func() {
		s.metrics.RecordToolCall(application.ToolCallMetric{
			ToolName:  "extract_knowledge",
			Timestamp: startTime,
			Duration:  time.Since(startTime),
			Success:   handlerErr == nil,
			ErrorMessage: func() string {
				if handlerErr != nil {
					return handlerErr.Error()
				}
				return ""
			}(),
		})
	}()

	if len(input.MemoryIDs) == 0 {
		handlerErr = errors.New("memory_ids is required")
		return nil, ExtractKnowledgeOutput{}, handlerErr
	}

	extractor := application.NewKnowledgeGraphExtractor(s.repo)
	graph, err := extractor.ExtractFromMultipleMemories(ctx, input.MemoryIDs)
	if err != nil {
		handlerErr = fmt.Errorf("knowledge extraction failed: %w", err)
		return nil, ExtractKnowledgeOutput{}, handlerErr
	}

	output := ExtractKnowledgeOutput{KnowledgeGraph: *graph}

	// Measure response size and record token metrics
	s.responseMiddleware.MeasureResponseSize(ctx, "extract_knowledge", output)

	return nil, output, nil
}

func (s *MCPServer) handleFindSimilarMemories(ctx context.Context, req *sdk.CallToolRequest, input ConsolidationFindSimilarInput) (*sdk.CallToolResult, ConsolidationFindSimilarOutput, error) {
	startTime := time.Now()
	var handlerErr error
	defer func() {
		s.metrics.RecordToolCall(application.ToolCallMetric{
			ToolName:  "find_similar_memories",
			Timestamp: startTime,
			Duration:  time.Since(startTime),
			Success:   handlerErr == nil,
			ErrorMessage: func() string {
				if handlerErr != nil {
					return handlerErr.Error()
				}
				return ""
			}(),
		})
	}()

	if input.MemoryID == "" {
		handlerErr = errors.New("memory_id is required")
		return nil, ConsolidationFindSimilarOutput{}, handlerErr
	}

	if input.Threshold == 0 {
		input.Threshold = 0.85
	}

	provider := s.hybridSearch.Provider()
	consolidation := application.NewMemoryConsolidationService(
		provider,
		s.repo,
		application.DuplicateDetectionConfig{},
		application.ClusteringConfig{},
	)

	similar, err := consolidation.FindSimilarMemories(ctx, input.MemoryID, input.Threshold)
	if err != nil {
		handlerErr = fmt.Errorf("similar search failed: %w", err)
		return nil, ConsolidationFindSimilarOutput{}, handlerErr
	}

	similarMemories := make([]MemorySummary, len(similar))
	for i, mem := range similar {
		similarMemories[i] = MemorySummary{
			ID:          mem.GetID(),
			Name:        mem.GetMetadata().Name,
			Content:     mem.Content,
			DateCreated: mem.DateCreated,
			Author:      mem.GetMetadata().Author,
			IsActive:    mem.IsActive(),
		}
	}

	output := ConsolidationFindSimilarOutput{
		OriginalMemoryID: input.MemoryID,
		SimilarMemories:  similarMemories,
		Count:            len(similar),
	}

	// Measure response size and record token metrics
	s.responseMiddleware.MeasureResponseSize(ctx, "find_similar_memories", output)

	return nil, output, nil
}

func (s *MCPServer) handleGetClusterDetails(ctx context.Context, req *sdk.CallToolRequest, input GetClusterDetailsInput) (*sdk.CallToolResult, GetClusterDetailsOutput, error) {
	startTime := time.Now()
	var handlerErr error
	defer func() {
		s.metrics.RecordToolCall(application.ToolCallMetric{
			ToolName:  "get_cluster_details",
			Timestamp: startTime,
			Duration:  time.Since(startTime),
			Success:   handlerErr == nil,
			ErrorMessage: func() string {
				if handlerErr != nil {
					return handlerErr.Error()
				}
				return ""
			}(),
		})
	}()

	provider := s.hybridSearch.Provider()
	consolidation := application.NewMemoryConsolidationService(
		provider,
		s.repo,
		application.DuplicateDetectionConfig{},
		application.ClusteringConfig{},
	)

	details, err := consolidation.GetClusterDetails(ctx, input.ClusterID)
	if err != nil {
		handlerErr = fmt.Errorf("failed to get cluster details: %w", err)
		return nil, GetClusterDetailsOutput{}, handlerErr
	}

	output := GetClusterDetailsOutput{Details: *details}

	// Measure response size and record token metrics
	s.responseMiddleware.MeasureResponseSize(ctx, "get_cluster_details", output)

	return nil, output, nil
}

func (s *MCPServer) handleGetConsolidationStats(ctx context.Context, req *sdk.CallToolRequest, _ struct{}) (*sdk.CallToolResult, GetConsolidationStatsOutput, error) {
	startTime := time.Now()
	var handlerErr error
	defer func() {
		s.metrics.RecordToolCall(application.ToolCallMetric{
			ToolName:  "get_consolidation_stats",
			Timestamp: startTime,
			Duration:  time.Since(startTime),
			Success:   handlerErr == nil,
			ErrorMessage: func() string {
				if handlerErr != nil {
					return handlerErr.Error()
				}
				return ""
			}(),
		})
	}()

	provider := s.hybridSearch.Provider()
	consolidation := application.NewMemoryConsolidationService(
		provider,
		s.repo,
		application.DuplicateDetectionConfig{},
		application.ClusteringConfig{},
	)

	stats, err := consolidation.GetConsolidationStatistics(ctx)
	if err != nil {
		handlerErr = fmt.Errorf("failed to get statistics: %w", err)
		return nil, GetConsolidationStatsOutput{}, handlerErr
	}

	output := GetConsolidationStatsOutput{Statistics: *stats}

	// Measure response size and record token metrics
	s.responseMiddleware.MeasureResponseSize(ctx, "get_consolidation_stats", output)

	return nil, output, nil
}

func (s *MCPServer) handleComputeSimilarity(ctx context.Context, req *sdk.CallToolRequest, input ComputeSimilarityInput) (*sdk.CallToolResult, ComputeSimilarityOutput, error) {
	startTime := time.Now()
	var handlerErr error
	defer func() {
		s.metrics.RecordToolCall(application.ToolCallMetric{
			ToolName:  "compute_similarity",
			Timestamp: startTime,
			Duration:  time.Since(startTime),
			Success:   handlerErr == nil,
			ErrorMessage: func() string {
				if handlerErr != nil {
					return handlerErr.Error()
				}
				return ""
			}(),
		})
	}()

	if input.MemoryID1 == "" || input.MemoryID2 == "" {
		handlerErr = errors.New("both memory_id_1 and memory_id_2 are required")
		return nil, ComputeSimilarityOutput{}, handlerErr
	}

	provider := s.hybridSearch.Provider()
	consolidation := application.NewMemoryConsolidationService(
		provider,
		s.repo,
		application.DuplicateDetectionConfig{},
		application.ClusteringConfig{},
	)

	similarity, err := consolidation.ComputeSimilarity(ctx, input.MemoryID1, input.MemoryID2)
	if err != nil {
		handlerErr = fmt.Errorf("similarity computation failed: %w", err)
		return nil, ComputeSimilarityOutput{}, handlerErr
	}

	output := ComputeSimilarityOutput{
		MemoryID1:  input.MemoryID1,
		MemoryID2:  input.MemoryID2,
		Similarity: similarity,
	}

	// Measure response size and record token metrics
	s.responseMiddleware.MeasureResponseSize(ctx, "compute_similarity", output)

	return nil, output, nil
}
