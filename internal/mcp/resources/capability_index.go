// Copyright 2025 NEXS-MCP Authors. All rights reserved.
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

package resources

import (
	"context"
	"encoding/json"
	"fmt"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/fsvxavier/nexs-mcp/internal/domain"
	"github.com/fsvxavier/nexs-mcp/internal/indexing"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// URIs for the capability index resources
const (
	URISummary = "capability-index://summary"
	URIFull    = "capability-index://full"
	URIStats   = "capability-index://stats"
)

// CachedResource stores a cached resource with its generation time
type CachedResource struct {
	Content   string
	Timestamp time.Time
}

// CapabilityIndexResource generates resource variants for the capability index
type CapabilityIndexResource struct {
	repository domain.ElementRepository
	index      *indexing.TFIDFIndex
	cache      map[string]CachedResource
	cacheTTL   time.Duration
	mu         sync.RWMutex
}

// NewCapabilityIndexResource creates a new CapabilityIndexResource
func NewCapabilityIndexResource(repo domain.ElementRepository, index *indexing.TFIDFIndex, cacheTTL time.Duration) *CapabilityIndexResource {
	return &CapabilityIndexResource{
		repository: repo,
		index:      index,
		cache:      make(map[string]CachedResource),
		cacheTTL:   cacheTTL,
	}
}

// Handler returns a ResourceHandler for the capability index
func (r *CapabilityIndexResource) Handler() mcp.ResourceHandler {
	return func(ctx context.Context, req *mcp.ReadResourceRequest) (*mcp.ReadResourceResult, error) {
		uri := req.Params.URI

		// Check cache first
		if cached, ok := r.getCachedResource(uri); ok {
			return &mcp.ReadResourceResult{
				Contents: []*mcp.ResourceContents{
					{
						URI:      uri,
						MIMEType: r.getMIMEType(uri),
						Text:     cached.Content,
					},
				},
			}, nil
		}

		// Generate resource based on URI
		var content string
		var err error

		switch uri {
		case URISummary:
			content, err = r.GenerateSummary(ctx)
		case URIFull:
			content, err = r.GenerateFull(ctx)
		case URIStats:
			content, err = r.GenerateStats(ctx)
		default:
			return nil, mcp.ResourceNotFoundError(uri)
		}

		if err != nil {
			return nil, fmt.Errorf("generating resource %s: %w", uri, err)
		}

		// Cache the result
		r.setCachedResource(uri, content)

		return &mcp.ReadResourceResult{
			Contents: []*mcp.ResourceContents{
				{
					URI:      uri,
					MIMEType: r.getMIMEType(uri),
					Text:     content,
				},
			},
		}, nil
	}
}

// GenerateSummary generates a ~3K token summary of the capability index
func (r *CapabilityIndexResource) GenerateSummary(ctx context.Context) (string, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var sb strings.Builder

	// Header
	sb.WriteString("# NEXS-MCP Capability Index - Summary\n\n")
	sb.WriteString(fmt.Sprintf("**Generated:** %s\n\n", time.Now().Format(time.RFC3339)))

	// Element counts by type
	sb.WriteString("## Element Counts\n\n")
	counts := r.getElementCounts(ctx)
	total := 0
	for _, count := range counts {
		total += count
	}
	sb.WriteString(fmt.Sprintf("- **Total Elements:** %d\n", total))
	for elemType, count := range counts {
		sb.WriteString(fmt.Sprintf("- **%s:** %d\n", elemType, count))
	}
	sb.WriteString("\n")

	// Index statistics
	stats := r.index.GetStats()
	sb.WriteString("## Index Statistics\n\n")
	sb.WriteString(fmt.Sprintf("- **Indexed Documents:** %v\n", stats["total_documents"]))
	sb.WriteString(fmt.Sprintf("- **Vocabulary Size:** %v unique terms\n", stats["total_unique_terms"]))
	sb.WriteString(fmt.Sprintf("- **Average Document Length:** %v terms\n", stats["avg_terms_per_doc"]))
	sb.WriteString("\n")

	// Top keywords (most common across all documents)
	sb.WriteString("## Top Keywords\n\n")
	topKeywords := r.getTopKeywords(20)
	for i, kw := range topKeywords {
		sb.WriteString(fmt.Sprintf("%d. **%s** (%.2f)\n", i+1, kw.Term, kw.Score))
	}
	sb.WriteString("\n")

	// Recent elements (last 7 days)
	sb.WriteString("## Recent Elements (Last 7 Days)\n\n")
	recentElements := r.getRecentElements(ctx, 7*24*time.Hour)
	if len(recentElements) == 0 {
		sb.WriteString("*No recent elements*\n\n")
	} else {
		for _, elem := range recentElements {
			sb.WriteString(fmt.Sprintf("- **[%s]** %s", elem.GetType(), elem.GetMetadata().Name))
			if desc := elem.GetMetadata().Description; desc != "" {
				sb.WriteString(fmt.Sprintf(" - %s", truncate(desc, 100)))
			}
			sb.WriteString("\n")
		}
		sb.WriteString("\n")
	}

	// Quick statistics overview
	sb.WriteString("## Overview\n\n")
	sb.WriteString(fmt.Sprintf("The capability index currently tracks **%d elements** across **%d element types**. ", total, len(counts)))
	vocabSize, _ := stats["total_unique_terms"].(int)
	sb.WriteString(fmt.Sprintf("The semantic search index has **%d terms** in its vocabulary and can perform similarity matching across all indexed content.\n\n", vocabSize))
	sb.WriteString(fmt.Sprintf("- `%s` - Complete index details (~40K tokens)\n", URIFull))
	sb.WriteString(fmt.Sprintf("- `%s` - Statistics in JSON format\n", URIStats))

	return sb.String(), nil
}

// GenerateFull generates a ~40K token detailed view of the capability index
func (r *CapabilityIndexResource) GenerateFull(ctx context.Context) (string, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var sb strings.Builder

	// Header
	sb.WriteString("# NEXS-MCP Capability Index - Full Details\n\n")
	sb.WriteString(fmt.Sprintf("**Generated:** %s\n\n", time.Now().Format(time.RFC3339)))

	// Index statistics
	stats := r.index.GetStats()
	sb.WriteString("## Index Statistics\n\n")
	sb.WriteString(fmt.Sprintf("- **Indexed Documents:** %v\n", stats["total_documents"]))
	sb.WriteString(fmt.Sprintf("- **Vocabulary Size:** %v unique terms\n", stats["total_unique_terms"]))
	sb.WriteString(fmt.Sprintf("- **Average Document Length:** %v terms\n", stats["avg_terms_per_doc"]))
	sb.WriteString(fmt.Sprintf("- **Index Memory Usage:** ~%d KB\n", r.estimateIndexMemory()))
	sb.WriteString("\n")

	// Element counts
	counts := r.getElementCounts(ctx)
	sb.WriteString("## Element Distribution\n\n")
	for elemType, count := range counts {
		sb.WriteString(fmt.Sprintf("### %s (%d)\n\n", elemType, count))

		// List all elements of this type
		elements := r.getElementsByType(ctx, elemType)
		for _, elem := range elements {
			meta := elem.GetMetadata()

			sb.WriteString(fmt.Sprintf("#### %s\n\n", meta.Name))
			if meta.Description != "" {
				sb.WriteString(fmt.Sprintf("**Description:** %s\n\n", meta.Description))
			}
			sb.WriteString(fmt.Sprintf("- **ID:** `%s`\n", meta.ID))
			sb.WriteString(fmt.Sprintf("- **Version:** %s\n", meta.Version))
			sb.WriteString(fmt.Sprintf("- **Created:** %s\n", meta.CreatedAt.Format(time.RFC3339)))
			sb.WriteString(fmt.Sprintf("- **Updated:** %s\n", meta.UpdatedAt.Format(time.RFC3339)))

			if len(meta.Tags) > 0 {
				sb.WriteString(fmt.Sprintf("- **Tags:** %s\n", strings.Join(meta.Tags, ", ")))
			}

			sb.WriteString("\n")
		}
	}

	// Vocabulary breakdown (top 100 terms)
	sb.WriteString("## Vocabulary Breakdown\n\n")
	sb.WriteString("Top 100 most significant terms:\n\n")
	topTerms := r.getTopKeywords(100)
	for i, kw := range topTerms {
		if i%5 == 0 {
			sb.WriteString("\n")
		}
		sb.WriteString(fmt.Sprintf("`%s` (%.2f) ", kw.Term, kw.Score))
	}
	sb.WriteString("\n\n")

	// Relationships overview
	sb.WriteString("## Relationship Graph\n\n")
	r.writeRelationshipGraph(&sb, ctx)

	return sb.String(), nil
}

// GenerateStats generates statistics in JSON format
func (r *CapabilityIndexResource) GenerateStats(ctx context.Context) (string, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	counts := r.getElementCounts(ctx)
	stats := r.index.GetStats()

	// Build cache statistics
	r.mu.RUnlock()
	r.mu.Lock()
	cacheStats := map[string]interface{}{
		"entries":     len(r.cache),
		"ttl_seconds": r.cacheTTL.Seconds(),
	}
	r.mu.Unlock()
	r.mu.RLock()

	data := map[string]interface{}{
		"generated_at":   time.Now().Format(time.RFC3339),
		"element_counts": counts,
		"index_statistics": map[string]interface{}{
			"document_count":      stats["total_documents"],
			"vocabulary_size":     stats["total_unique_terms"],
			"avg_document_length": stats["avg_terms_per_doc"],
			"memory_usage_kb":     r.estimateIndexMemory(),
		},
		"cache_statistics": cacheStats,
		"resources": map[string]string{
			"summary": URISummary,
			"full":    URIFull,
			"stats":   URIStats,
		},
	}

	jsonBytes, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return "", fmt.Errorf("marshaling stats: %w", err)
	}

	return string(jsonBytes), nil
}

// Helper methods

func (r *CapabilityIndexResource) getCachedResource(uri string) (CachedResource, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	cached, ok := r.cache[uri]
	if !ok {
		return CachedResource{}, false
	}

	// Check if cache is still valid
	if time.Since(cached.Timestamp) > r.cacheTTL {
		return CachedResource{}, false
	}

	return cached, true
}

func (r *CapabilityIndexResource) setCachedResource(uri, content string) {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.cache[uri] = CachedResource{
		Content:   content,
		Timestamp: time.Now(),
	}
}

func (r *CapabilityIndexResource) getMIMEType(uri string) string {
	if uri == URIStats {
		return "application/json"
	}
	return "text/markdown"
}

func (r *CapabilityIndexResource) getElementCounts(ctx context.Context) map[string]int {
	counts := make(map[string]int)

	// Count each type
	for _, elemType := range []domain.ElementType{
		domain.PersonaElement,
		domain.SkillElement,
		domain.TemplateElement,
		domain.AgentElement,
		domain.EnsembleElement,
		domain.MemoryElement,
	} {
		filter := domain.ElementFilter{Type: &elemType}
		elements, _ := r.repository.List(filter)
		counts[string(elemType)] = len(elements)
	}

	return counts
}

func (r *CapabilityIndexResource) getElementsByType(ctx context.Context, elemType string) []domain.Element {
	et := domain.ElementType(elemType)
	filter := domain.ElementFilter{Type: &et}
	elements, _ := r.repository.List(filter)
	return elements
}

type KeywordScore struct {
	Term  string
	Score float64
}

func (r *CapabilityIndexResource) getTopKeywords(limit int) []KeywordScore {
	// For now, return empty slice
	// In a full implementation, we'd analyze IDF scores
	// This would require exposing IDF scores from TFIDFIndex
	return []KeywordScore{}
}

func (r *CapabilityIndexResource) getRecentElements(ctx context.Context, duration time.Duration) []domain.Element {
	cutoff := time.Now().Add(-duration)
	var recent []domain.Element

	// Get all elements
	allElements, _ := r.repository.List(domain.ElementFilter{})

	for _, elem := range allElements {
		if elem.GetMetadata().CreatedAt.After(cutoff) {
			recent = append(recent, elem)
		}
	}

	// Sort by creation time (newest first)
	sort.Slice(recent, func(i, j int) bool {
		return recent[i].GetMetadata().CreatedAt.After(recent[j].GetMetadata().CreatedAt)
	})

	// Limit to 10 most recent
	if len(recent) > 10 {
		recent = recent[:10]
	}

	return recent
}

func (r *CapabilityIndexResource) estimateIndexMemory() int {
	stats := r.index.GetStats()

	// Rough estimate: vocabulary size * 100 bytes per term + document vectors
	vocabSize, _ := stats["total_unique_terms"].(int)
	docCount, _ := stats["total_documents"].(int)
	estimate := (vocabSize * 100) + (docCount * 500)
	return estimate / 1024 // Convert to KB
}

func (r *CapabilityIndexResource) writeRelationshipGraph(sb *strings.Builder, ctx context.Context) {
	// Get all agents (they typically have relationships)
	agentType := domain.AgentElement
	filter := domain.ElementFilter{Type: &agentType}
	allAgents, _ := r.repository.List(filter)

	if len(allAgents) == 0 {
		sb.WriteString("*No agents with relationships defined*\n\n")
		return
	}

	for _, elem := range allAgents {
		agent, ok := elem.(*domain.Agent)
		if !ok {
			continue
		}

		sb.WriteString(fmt.Sprintf("### %s\n\n", agent.GetMetadata().Name))

		if len(agent.Goals) > 0 {
			sb.WriteString(fmt.Sprintf("- **Goals:** %s\n", strings.Join(agent.Goals, ", ")))
		}

		if len(agent.Actions) > 0 {
			actionNames := make([]string, len(agent.Actions))
			for i, action := range agent.Actions {
				actionNames[i] = action.Name
			}
			sb.WriteString(fmt.Sprintf("- **Actions:** %s\n", strings.Join(actionNames, ", ")))
		}

		sb.WriteString("\n")
	}
}

func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen-3] + "..."
}
