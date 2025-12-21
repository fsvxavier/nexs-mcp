package mcp

import (
	"context"
	"fmt"
	"sort"
	"strconv"
	"strings"
	"time"

	sdk "github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/fsvxavier/nexs-mcp/internal/collection/sources"
)

// SearchCollectionsInput defines input for search_collections tool.
type SearchCollectionsInput struct {
	Query      string   `json:"query"                 jsonschema:"text search across name, description, keywords"`
	Category   string   `json:"category,omitempty"    jsonschema:"filter by category"`
	Author     string   `json:"author,omitempty"      jsonschema:"filter by author name"`
	Tags       []string `json:"tags,omitempty"        jsonschema:"filter by tags (must have ALL specified tags)"`
	MinStars   int      `json:"min_stars,omitempty"   jsonschema:"minimum number of stars"`
	Source     string   `json:"source,omitempty"      jsonschema:"filter by source (github, local, http)"`
	SortBy     string   `json:"sort_by,omitempty"     jsonschema:"sort by: relevance (default), stars, downloads, updated, created, name"`
	SortOrder  string   `json:"sort_order,omitempty"  jsonschema:"sort order: desc (default), asc"`
	Limit      int      `json:"limit,omitempty"       jsonschema:"maximum number of results (default: 20)"`
	Offset     int      `json:"offset,omitempty"      jsonschema:"number of results to skip (for pagination)"`
	RichFormat bool     `json:"rich_format,omitempty" jsonschema:"use rich formatting with emojis and stats"`
}

// SearchCollectionsOutput defines output for search_collections tool.
type SearchCollectionsOutput struct {
	Collections []CollectionResult     `json:"collections"       jsonschema:"list of matching collections"`
	Total       int                    `json:"total"             jsonschema:"total number of results (before pagination)"`
	Query       string                 `json:"query,omitempty"   jsonschema:"search query used"`
	Filters     map[string]interface{} `json:"filters,omitempty" jsonschema:"filters applied"`
	Timing      map[string]string      `json:"timing,omitempty"  jsonschema:"performance timing information"`
}

// CollectionResult represents a single collection result with enhanced formatting.
type CollectionResult struct {
	ID            string       `json:"id"                      jsonschema:"collection ID (author/name)"`
	Name          string       `json:"name"                    jsonschema:"collection name"`
	Version       string       `json:"version"                 jsonschema:"collection version"`
	Author        string       `json:"author"                  jsonschema:"author name"`
	Category      string       `json:"category"                jsonschema:"collection category"`
	Description   string       `json:"description"             jsonschema:"collection description"`
	Tags          []string     `json:"tags,omitempty"          jsonschema:"collection tags"`
	Stars         int          `json:"stars,omitempty"         jsonschema:"number of stars/favorites"`
	Downloads     int          `json:"downloads,omitempty"     jsonschema:"download count"`
	Elements      ElementStats `json:"elements"                jsonschema:"element statistics"`
	Repository    string       `json:"repository,omitempty"    jsonschema:"repository URL"`
	Homepage      string       `json:"homepage,omitempty"      jsonschema:"homepage URL"`
	Documentation string       `json:"documentation,omitempty" jsonschema:"documentation URL"`
	License       string       `json:"license,omitempty"       jsonschema:"license identifier"`
	UpdatedAt     string       `json:"updated_at,omitempty"    jsonschema:"last update timestamp"`
	CreatedAt     string       `json:"created_at,omitempty"    jsonschema:"creation timestamp"`
	Source        string       `json:"source,omitempty"        jsonschema:"source type (github, local, http)"`
	URI           string       `json:"uri,omitempty"           jsonschema:"collection URI"`
	Relevance     float64      `json:"relevance,omitempty"     jsonschema:"search relevance score (0-100)"`
	Display       string       `json:"display,omitempty"       jsonschema:"rich formatted display (if rich_format=true)"`
}

// ElementStats contains statistics about collection elements.
type ElementStats struct {
	Total     int `json:"total"               jsonschema:"total number of elements"`
	Personas  int `json:"personas,omitempty"  jsonschema:"number of personas"`
	Skills    int `json:"skills,omitempty"    jsonschema:"number of skills"`
	Templates int `json:"templates,omitempty" jsonschema:"number of templates"`
	Agents    int `json:"agents,omitempty"    jsonschema:"number of agents"`
	Memories  int `json:"memories,omitempty"  jsonschema:"number of memories"`
	Ensembles int `json:"ensembles,omitempty" jsonschema:"number of ensembles"`
}

// handleSearchCollections implements the search_collections tool.
func (s *MCPServer) handleSearchCollections(ctx context.Context, req *sdk.CallToolRequest, input SearchCollectionsInput) (*sdk.CallToolResult, SearchCollectionsOutput, error) {
	startTime := time.Now()

	output := SearchCollectionsOutput{
		Query: input.Query,
		Filters: map[string]interface{}{
			"category":  input.Category,
			"author":    input.Author,
			"tags":      input.Tags,
			"min_stars": input.MinStars,
			"source":    input.Source,
		},
		Timing: make(map[string]string),
	}

	// Set defaults
	if input.Limit == 0 {
		input.Limit = 20
	}
	if input.SortBy == "" {
		input.SortBy = "relevance"
	}
	if input.SortOrder == "" {
		input.SortOrder = "desc"
	}

	// Build filter
	filter := &sources.BrowseFilter{
		Category: input.Category,
		Author:   input.Author,
		Tags:     input.Tags,
		Query:    input.Query,
		Limit:    input.Limit * 2, // Fetch extra for post-filtering
		Offset:   input.Offset,
	}

	// Get registry (assuming we have access through server)
	// For now, we'll use a basic browse since we need registry integration
	// This would need to be wired up properly in production

	// Search using registry
	results := s.searchCollectionsWithRegistry(ctx, filter, input.Source)

	output.Timing["search"] = time.Since(startTime).String()
	sortStart := time.Now()

	// Post-filter by min_stars
	if input.MinStars > 0 {
		filtered := make([]*sources.CollectionMetadata, 0, len(results))
		for _, result := range results {
			if result.Stars >= input.MinStars {
				filtered = append(filtered, result)
			}
		}
		results = filtered
	}

	output.Total = len(results)

	// Sort results
	sortCollections(results, input.SortBy, input.SortOrder)
	output.Timing["sort"] = time.Since(sortStart).String()

	// Apply pagination
	if input.Offset < len(results) {
		end := input.Offset + input.Limit
		if end > len(results) {
			end = len(results)
		}
		results = results[input.Offset:end]
	} else {
		results = nil
	}

	// Convert to output format
	formatStart := time.Now()
	output.Collections = make([]CollectionResult, 0, len(results))
	for _, meta := range results {
		result := collectionMetadataToResult(meta, input.RichFormat)
		output.Collections = append(output.Collections, result)
	}
	output.Timing["format"] = time.Since(formatStart).String()
	output.Timing["total"] = time.Since(startTime).String()

	return nil, output, nil
}

// searchCollectionsWithRegistry performs the actual search.
func (s *MCPServer) searchCollectionsWithRegistry(ctx context.Context, filter *sources.BrowseFilter, sourceName string) []*sources.CollectionMetadata {
	// This is a placeholder - in production this would use the registry's Search method
	// For now, return empty results
	// TODO: Wire up registry access from MCPServer
	return []*sources.CollectionMetadata{}
}

// sortCollections sorts collection results by the specified field and order.
func sortCollections(collections []*sources.CollectionMetadata, sortBy, sortOrder string) {
	descending := sortOrder == "desc"

	sort.Slice(collections, func(i, j int) bool {
		var less bool

		switch sortBy {
		case "stars":
			less = collections[i].Stars < collections[j].Stars
		case "downloads":
			less = collections[i].Downloads < collections[j].Downloads
		case "name":
			less = collections[i].Name < collections[j].Name
		case "relevance":
			// For now, use stars as proxy for relevance
			// In production, this would use actual relevance scoring
			less = collections[i].Stars < collections[j].Stars
		default:
			// Default to stars
			less = collections[i].Stars < collections[j].Stars
		}

		if descending {
			return !less
		}
		return less
	})
}

// collectionMetadataToResult converts CollectionMetadata to CollectionResult.
func collectionMetadataToResult(meta *sources.CollectionMetadata, richFormat bool) CollectionResult {
	result := CollectionResult{
		ID:          fmt.Sprintf("%s/%s", meta.Author, meta.Name),
		Name:        meta.Name,
		Version:     meta.Version,
		Author:      meta.Author,
		Category:    meta.Category,
		Description: meta.Description,
		Tags:        meta.Tags,
		Stars:       meta.Stars,
		Downloads:   meta.Downloads,
		Repository:  meta.Repository,
		Source:      meta.SourceName,
		URI:         meta.URI,
	}

	// Element stats - would need to be extracted from manifest if available
	// For now, set total to 0 as we don't have this in metadata
	result.Elements = ElementStats{
		Total: 0, // Would need full manifest to get this
	}

	// Rich formatting
	if richFormat {
		result.Display = formatCollectionRich(meta)
	}

	return result
}

// formatCollectionRich creates a rich formatted display string.
func formatCollectionRich(meta *sources.CollectionMetadata) string {
	var builder strings.Builder

	// Header with emoji based on category
	emoji := getCategoryEmoji(meta.Category)
	builder.WriteString(fmt.Sprintf("%s **%s** v%s\n", emoji, meta.Name, meta.Version))

	// Author and stars
	builder.WriteString("👤 " + meta.Author)
	if meta.Stars > 0 {
		builder.WriteString(fmt.Sprintf(" • ⭐ %d", meta.Stars))
	}
	if meta.Downloads > 0 {
		builder.WriteString(" • 📥 " + formatNumber(meta.Downloads))
	}
	builder.WriteString("\n")

	// Category and tags
	builder.WriteString("📂 " + meta.Category)
	if len(meta.Tags) > 0 {
		builder.WriteString(" • 🏷️  ")
		builder.WriteString(strings.Join(meta.Tags, ", "))
	}
	builder.WriteString("\n")

	// Description
	if meta.Description != "" {
		builder.WriteString(fmt.Sprintf("📝 %s\n", meta.Description))
	}

	// Links
	if meta.Repository != "" {
		builder.WriteString(fmt.Sprintf("🔗 %s\n", meta.Repository))
	}

	return builder.String()
}

// getCategoryEmoji returns an emoji for a category.
func getCategoryEmoji(category string) string {
	emojis := map[string]string{
		"devops":           "⚙️",
		"creative-writing": "✍️",
		"data-science":     "📊",
		"web-development":  "🌐",
		"mobile":           "📱",
		"ai-ml":            "🤖",
		"security":         "🔒",
		"testing":          "🧪",
		"productivity":     "⚡",
		"education":        "📚",
		"gaming":           "🎮",
		"finance":          "💰",
		"healthcare":       "🏥",
		"iot":              "🌡️",
		"blockchain":       "⛓️",
	}

	if emoji, ok := emojis[strings.ToLower(category)]; ok {
		return emoji
	}
	return "📦" // Default
}

// formatNumber formats a number with K, M, B suffixes.
func formatNumber(n int) string {
	if n >= 1_000_000_000 {
		return fmt.Sprintf("%.1fB", float64(n)/1_000_000_000)
	}
	if n >= 1_000_000 {
		return fmt.Sprintf("%.1fM", float64(n)/1_000_000)
	}
	if n >= 1_000 {
		return fmt.Sprintf("%.1fK", float64(n)/1_000)
	}
	return strconv.Itoa(n)
}

// formatTimestamp formats a timestamp in a human-readable way.
func formatTimestamp(t time.Time) string {
	now := time.Now()
	diff := now.Sub(t)

	if diff < time.Minute {
		return "just now"
	}
	if diff < time.Hour {
		mins := int(diff.Minutes())
		if mins == 1 {
			return "1 minute ago"
		}
		return fmt.Sprintf("%d minutes ago", mins)
	}
	if diff < 24*time.Hour {
		hours := int(diff.Hours())
		if hours == 1 {
			return "1 hour ago"
		}
		return fmt.Sprintf("%d hours ago", hours)
	}
	if diff < 7*24*time.Hour {
		days := int(diff.Hours() / 24)
		if days == 1 {
			return "1 day ago"
		}
		return fmt.Sprintf("%d days ago", days)
	}
	if diff < 30*24*time.Hour {
		weeks := int(diff.Hours() / 24 / 7)
		if weeks == 1 {
			return "1 week ago"
		}
		return fmt.Sprintf("%d weeks ago", weeks)
	}
	if diff < 365*24*time.Hour {
		months := int(diff.Hours() / 24 / 30)
		if months == 1 {
			return "1 month ago"
		}
		return fmt.Sprintf("%d months ago", months)
	}

	years := int(diff.Hours() / 24 / 365)
	if years == 1 {
		return "1 year ago"
	}
	return fmt.Sprintf("%d years ago", years)
}

// ListCollectionsInput defines input for enhanced list_collections tool.
type ListCollectionsInput struct {
	Category   string   `json:"category,omitempty"    jsonschema:"filter by category"`
	Author     string   `json:"author,omitempty"      jsonschema:"filter by author name"`
	Tags       []string `json:"tags,omitempty"        jsonschema:"filter by tags"`
	Source     string   `json:"source,omitempty"      jsonschema:"filter by source"`
	Limit      int      `json:"limit,omitempty"       jsonschema:"maximum number of results (default: 50)"`
	Offset     int      `json:"offset,omitempty"      jsonschema:"number of results to skip"`
	RichFormat bool     `json:"rich_format,omitempty" jsonschema:"use rich formatting with emojis and stats"`
	GroupBy    string   `json:"group_by,omitempty"    jsonschema:"group results by: category, author, source"`
}

// ListCollectionsOutput defines output for enhanced list_collections tool.
type ListCollectionsOutput struct {
	Collections []CollectionResult            `json:"collections,omitempty" jsonschema:"list of collections (if not grouped)"`
	Groups      map[string][]CollectionResult `json:"groups,omitempty"      jsonschema:"grouped collections (if group_by specified)"`
	Total       int                           `json:"total"                 jsonschema:"total number of collections"`
	Summary     CollectionsSummary            `json:"summary"               jsonschema:"summary statistics"`
}

// CollectionsSummary provides aggregate statistics.
type CollectionsSummary struct {
	TotalCollections int            `json:"total_collections" jsonschema:"total number of collections"`
	TotalElements    int            `json:"total_elements"    jsonschema:"total number of elements across all collections"`
	ByCategory       map[string]int `json:"by_category"       jsonschema:"count by category"`
	ByAuthor         map[string]int `json:"by_author"         jsonschema:"count by author"`
	BySource         map[string]int `json:"by_source"         jsonschema:"count by source"`
	TotalDownloads   int            `json:"total_downloads"   jsonschema:"total downloads across all collections"`
	AverageStars     float64        `json:"average_stars"     jsonschema:"average star rating"`
}

// handleListCollections implements enhanced list_collections tool.
func (s *MCPServer) handleListCollections(ctx context.Context, req *sdk.CallToolRequest, input ListCollectionsInput) (*sdk.CallToolResult, ListCollectionsOutput, error) {
	output := ListCollectionsOutput{
		Summary: CollectionsSummary{
			ByCategory: make(map[string]int),
			ByAuthor:   make(map[string]int),
			BySource:   make(map[string]int),
		},
	}

	// Set defaults
	if input.Limit == 0 {
		input.Limit = 50
	}

	// Build filter
	filter := &sources.BrowseFilter{
		Category: input.Category,
		Author:   input.Author,
		Tags:     input.Tags,
		Limit:    input.Limit,
		Offset:   input.Offset,
	}

	// Search using registry
	results := s.searchCollectionsWithRegistry(ctx, filter, input.Source)

	output.Total = len(results)

	// Build summary
	totalStars := 0
	for _, meta := range results {
		output.Summary.TotalCollections++
		// Note: TotalElements would need full manifest, not available in metadata
		output.Summary.TotalDownloads += meta.Downloads
		totalStars += meta.Stars

		output.Summary.ByCategory[meta.Category]++
		output.Summary.ByAuthor[meta.Author]++
		output.Summary.BySource[meta.SourceName]++
	}

	if output.Summary.TotalCollections > 0 {
		output.Summary.AverageStars = float64(totalStars) / float64(output.Summary.TotalCollections)
	}

	// Convert to output format
	if input.GroupBy != "" {
		// Group results
		output.Groups = make(map[string][]CollectionResult)
		for _, meta := range results {
			result := collectionMetadataToResult(meta, input.RichFormat)

			var groupKey string
			switch input.GroupBy {
			case "category":
				groupKey = meta.Category
			case "author":
				groupKey = meta.Author
			case "source":
				groupKey = meta.SourceName
			default:
				groupKey = "other"
			}

			output.Groups[groupKey] = append(output.Groups[groupKey], result)
		}
	} else {
		// Flat list
		output.Collections = make([]CollectionResult, 0, len(results))
		for _, meta := range results {
			result := collectionMetadataToResult(meta, input.RichFormat)
			output.Collections = append(output.Collections, result)
		}
	}

	return nil, output, nil
}
