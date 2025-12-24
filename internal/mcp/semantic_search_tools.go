package mcp

import (
	"context"

	"github.com/fsvxavier/nexs-mcp/internal/application"
	sdk "github.com/modelcontextprotocol/go-sdk/mcp"
)

// SemanticSearchInput is the input for semantic search.
type SemanticSearchInput struct {
	Query       string `json:"query"                  jsonschema:"required,description=Natural language search query"`
	ElementType string `json:"element_type,omitempty" jsonschema:"description=Filter by element type (persona skill agent memory template ensemble),enum=persona,enum=skill,enum=agent,enum=memory,enum=template,enum=ensemble"`
	Limit       int    `json:"limit,omitempty"        jsonschema:"description=Maximum number of results,default=10,minimum=1,maximum=100"`
}

// SemanticSearchOutput is the output for semantic search.
type SemanticSearchOutput struct {
	Results []interface{} `json:"results"`
	Count   int           `json:"count"`
	Query   string        `json:"query"`
	Filter  string        `json:"filter,omitempty"`
}

// FindSimilarMemoriesInput is the input for finding similar memories.
type FindSimilarMemoriesInput struct {
	Query string `json:"query"           jsonschema:"required,description=Query text to find similar memories"`
	Limit int    `json:"limit,omitempty" jsonschema:"description=Maximum number of memories,default=10,minimum=1,maximum=50"`
}

// FindSimilarMemoriesOutput is the output for finding similar memories.
type FindSimilarMemoriesOutput struct {
	Memories []interface{} `json:"memories"`
	Count    int           `json:"count"`
	Query    string        `json:"query"`
}

// IndexElementInput is the input for indexing an element.
type IndexElementInput struct {
	ElementID   string `json:"element_id"   jsonschema:"required,description=ID of the element to index"`
	ElementType string `json:"element_type" jsonschema:"required,description=Type of element,enum=persona,enum=skill,enum=agent,enum=memory,enum=template,enum=ensemble"`
}

// IndexElementOutput is the output for indexing an element.
type IndexElementOutput struct {
	Success     bool   `json:"success"`
	ElementID   string `json:"element_id"`
	ElementType string `json:"element_type"`
	Message     string `json:"message"`
}

// RebuildSearchIndexOutput is the output for rebuilding the search index.
type RebuildSearchIndexOutput struct {
	Success bool                   `json:"success"`
	Message string                 `json:"message"`
	Stats   map[string]interface{} `json:"stats"`
}

// GetSearchStatsOutput is the output for getting search stats.
type GetSearchStatsOutput struct {
	Stats map[string]interface{} `json:"stats"`
}

// RegisterSemanticSearchTools registers semantic search tools with the MCP server
// NOTE: Semantic search tools are currently disabled pending full implementation
// of vector embedding providers (OpenAI, Transformers, ONNX).
// To enable: uncomment the tool registrations below and ensure embeddings are configured.
func RegisterSemanticSearchTools(server *MCPServer, service *application.SemanticSearchService) {
	if service == nil {
		return // Semantic search not enabled
	}

	// semantic_search tool
	sdk.AddTool(server.server, &sdk.Tool{
		Name:        "semantic_search",
		Description: "Search for elements using semantic similarity (vector embeddings)",
	}, func(ctx context.Context, req *sdk.CallToolRequest, input SemanticSearchInput) (*sdk.CallToolResult, SemanticSearchOutput, error) {
		if input.Limit == 0 {
			input.Limit = 10
		}

		results, err := service.Search(ctx, input.Query, input.Limit, input.ElementType)
		if err != nil {
			return nil, SemanticSearchOutput{}, err
		}

		// Convert results to interface{}
		interfaceResults := make([]interface{}, len(results))
		for i, r := range results {
			interfaceResults[i] = r
		}

		output := SemanticSearchOutput{
			Results: interfaceResults,
			Count:   len(results),
			Query:   input.Query,
			Filter:  input.ElementType,
		}

		return nil, output, nil
	})
}
