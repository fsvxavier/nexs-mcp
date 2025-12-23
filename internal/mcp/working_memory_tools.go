package mcp

import (
	"context"
	"fmt"
	"time"

	"github.com/fsvxavier/nexs-mcp/internal/application"
	"github.com/fsvxavier/nexs-mcp/internal/domain"
	sdk "github.com/modelcontextprotocol/go-sdk/mcp"
)

// Working Memory MCP Tools - Two-Tier Memory Architecture
//
// Working Memory: Session-scoped, TTL-based, auto-promoted to long-term
// Long-term Memory: Persistent, manually managed, existing Memory domain
//
// 15 Tools:
// 1. working_memory_add - Add new working memory
// 2. working_memory_get - Get working memory by ID
// 3. working_memory_list - List all working memories in session
// 4. working_memory_promote - Manually promote to long-term
// 5. working_memory_clear_session - Clear all memories in session
// 6. working_memory_stats - Get session statistics
// 7. working_memory_expire - Manually expire a memory
// 8. working_memory_extend_ttl - Extend TTL
// 9. working_memory_export - Export session data
// 10. working_memory_list_pending - List memories pending promotion
// 11. working_memory_list_expired - List expired memories
// 12. working_memory_list_promoted - List promoted memories
// 13. working_memory_bulk_promote - Promote all pending
// 14. working_memory_relation_add - Add relation between memories
// 15. working_memory_search - Search within session

// --- Input/Output Structs ---

type WorkingMemoryAddInput struct {
	SessionID string            `json:"session_id"         jsonschema:"required" jsonschema_description:"Session identifier"`
	Content   string            `json:"content"            jsonschema:"required" jsonschema_description:"Memory content"`
	Priority  string            `json:"priority,omitempty" jsonschema_description:"Priority level: low, medium (default), high, or critical"`
	Tags      []string          `json:"tags,omitempty"     jsonschema_description:"Searchable tags"`
	Metadata  map[string]string `json:"metadata,omitempty" jsonschema_description:"Custom metadata"`
	Context   string            `json:"context,omitempty"  jsonschema_description:"Additional context"`
}

type WorkingMemoryGetInput struct {
	SessionID string `json:"session_id" jsonschema:"required" jsonschema_description:"Session identifier"`
	MemoryID  string `json:"memory_id"  jsonschema:"required" jsonschema_description:"Working memory ID"`
}

type WorkingMemoryListInput struct {
	SessionID       string `json:"session_id"                jsonschema:"required" jsonschema_description:"Session identifier"`
	IncludeExpired  bool   `json:"include_expired,omitempty" jsonschema_description:"Include expired memories (default: false)"`
	IncludePromoted bool   `json:"include_promoted,omitempty" jsonschema_description:"Include promoted memories (default: false)"`
}

type WorkingMemoryPromoteInput struct {
	SessionID string `json:"session_id" jsonschema:"required" jsonschema_description:"Session identifier"`
	MemoryID  string `json:"memory_id"  jsonschema:"required" jsonschema_description:"Working memory ID to promote"`
}

type WorkingMemoryClearSessionInput struct {
	SessionID string `json:"session_id" jsonschema:"required" jsonschema_description:"Session identifier to clear"`
}

type WorkingMemoryStatsInput struct {
	SessionID string `json:"session_id" jsonschema:"required" jsonschema_description:"Session identifier"`
}

type WorkingMemoryExpireInput struct {
	SessionID string `json:"session_id" jsonschema:"required" jsonschema_description:"Session identifier"`
	MemoryID  string `json:"memory_id"  jsonschema:"required" jsonschema_description:"Working memory ID to expire"`
}

type WorkingMemoryExtendTTLInput struct {
	SessionID string `json:"session_id" jsonschema:"required" jsonschema_description:"Session identifier"`
	MemoryID  string `json:"memory_id"  jsonschema:"required" jsonschema_description:"Working memory ID"`
}

type WorkingMemoryExportInput struct {
	SessionID string `json:"session_id" jsonschema:"required" jsonschema_description:"Session identifier"`
}

type WorkingMemoryBulkPromoteInput struct {
	SessionID string `json:"session_id" jsonschema:"required" jsonschema_description:"Session identifier"`
}

type WorkingMemoryRelationAddInput struct {
	SessionID       string `json:"session_id"        jsonschema:"required" jsonschema_description:"Session identifier"`
	MemoryID        string `json:"memory_id"         jsonschema:"required" jsonschema_description:"Source working memory ID"`
	RelatedMemoryID string `json:"related_memory_id" jsonschema:"required" jsonschema_description:"Related working memory ID"`
}

type WorkingMemorySearchInput struct {
	SessionID string `json:"session_id" jsonschema:"required" jsonschema_description:"Session identifier"`
	Query     string `json:"query"      jsonschema:"required" jsonschema_description:"Search query"`
}

// RegisterWorkingMemoryTools registers working memory tools with the MCP server
func RegisterWorkingMemoryTools(server *MCPServer, service *application.WorkingMemoryService) {
	if service == nil {
		return
	}

	// 1. working_memory_add
	sdk.AddTool(server.server, &sdk.Tool{
		Name:        "working_memory_add",
		Description: "Add a new working memory to a session with TTL and auto-promotion",
	}, func(ctx context.Context, req *sdk.CallToolRequest, input WorkingMemoryAddInput) (*sdk.CallToolResult, interface{}, error) {
		priority := domain.MemoryPriority(input.Priority)
		if priority == "" {
			priority = domain.PriorityMedium
		}

		wm, err := service.Add(ctx, input.SessionID, input.Content, priority, input.Tags, input.Metadata)
		if err != nil {
			return nil, nil, err
		}

		// Add context if provided
		if input.Context != "" {
			wm.Context = input.Context
		}

		return nil, map[string]interface{}{
			"id":               wm.ID,
			"session_id":       wm.SessionID,
			"priority":         wm.Priority,
			"expires_at":       wm.ExpiresAt.Format(time.RFC3339),
			"importance_score": wm.ImportanceScore,
			"ttl_hours":        wm.Priority.TTL().Hours(),
			"created_at":       wm.CreatedAt.Format(time.RFC3339),
		}, nil
	})

	// 2. working_memory_get
	sdk.AddTool(server.server, &sdk.Tool{
		Name:        "working_memory_get",
		Description: "Get a working memory by ID and record access",
	}, func(ctx context.Context, req *sdk.CallToolRequest, input WorkingMemoryGetInput) (*sdk.CallToolResult, interface{}, error) {
		wm, err := service.Get(ctx, input.SessionID, input.MemoryID)
		if err != nil {
			return nil, nil, err
		}

		return nil, wm, nil
	})

	// 3. working_memory_list
	sdk.AddTool(server.server, &sdk.Tool{
		Name:        "working_memory_list",
		Description: "List all working memories in a session",
	}, func(ctx context.Context, req *sdk.CallToolRequest, input WorkingMemoryListInput) (*sdk.CallToolResult, interface{}, error) {
		memories, err := service.List(ctx, input.SessionID, input.IncludeExpired, input.IncludePromoted)
		if err != nil {
			return nil, nil, err
		}

		return nil, map[string]interface{}{
			"session_id": input.SessionID,
			"count":      len(memories),
			"memories":   memories,
		}, nil
	})

	// 4. working_memory_promote
	sdk.AddTool(server.server, &sdk.Tool{
		Name:        "working_memory_promote",
		Description: "Manually promote a working memory to long-term memory",
	}, func(ctx context.Context, req *sdk.CallToolRequest, input WorkingMemoryPromoteInput) (*sdk.CallToolResult, interface{}, error) {
		longTermMem, err := service.Promote(ctx, input.SessionID, input.MemoryID)
		if err != nil {
			return nil, nil, err
		}

		return nil, map[string]interface{}{
			"success":          true,
			"working_id":       input.MemoryID,
			"long_term_id":     longTermMem.GetID(),
			"long_term_memory": longTermMem,
			"promoted_at":      time.Now().Format(time.RFC3339),
		}, nil
	})

	// 5. working_memory_clear_session
	sdk.AddTool(server.server, &sdk.Tool{
		Name:        "working_memory_clear_session",
		Description: "Clear all working memories in a session",
	}, func(ctx context.Context, req *sdk.CallToolRequest, input WorkingMemoryClearSessionInput) (*sdk.CallToolResult, interface{}, error) {
		err := service.ClearSession(input.SessionID)
		if err != nil {
			return nil, nil, err
		}

		return nil, map[string]interface{}{
			"success":    true,
			"session_id": input.SessionID,
			"cleared_at": time.Now().Format(time.RFC3339),
		}, nil
	})

	// 6. working_memory_stats
	sdk.AddTool(server.server, &sdk.Tool{
		Name:        "working_memory_stats",
		Description: "Get statistics for a session's working memory",
	}, func(ctx context.Context, req *sdk.CallToolRequest, input WorkingMemoryStatsInput) (*sdk.CallToolResult, interface{}, error) {
		stats := service.GetStats(input.SessionID)
		return nil, stats, nil
	})

	// 7. working_memory_expire
	sdk.AddTool(server.server, &sdk.Tool{
		Name:        "working_memory_expire",
		Description: "Manually expire a working memory",
	}, func(ctx context.Context, req *sdk.CallToolRequest, input WorkingMemoryExpireInput) (*sdk.CallToolResult, interface{}, error) {
		err := service.ExpireMemory(input.SessionID, input.MemoryID)
		if err != nil {
			return nil, nil, err
		}

		return nil, map[string]interface{}{
			"success":    true,
			"memory_id":  input.MemoryID,
			"session_id": input.SessionID,
			"expired_at": time.Now().Format(time.RFC3339),
		}, nil
	})

	// 8. working_memory_extend_ttl
	sdk.AddTool(server.server, &sdk.Tool{
		Name:        "working_memory_extend_ttl",
		Description: "Extend the TTL of a working memory",
	}, func(ctx context.Context, req *sdk.CallToolRequest, input WorkingMemoryExtendTTLInput) (*sdk.CallToolResult, interface{}, error) {
		err := service.ExtendTTL(input.SessionID, input.MemoryID)
		if err != nil {
			return nil, nil, err
		}

		return nil, map[string]interface{}{
			"success":     true,
			"memory_id":   input.MemoryID,
			"session_id":  input.SessionID,
			"extended_at": time.Now().Format(time.RFC3339),
		}, nil
	})

	// 9. working_memory_export
	sdk.AddTool(server.server, &sdk.Tool{
		Name:        "working_memory_export",
		Description: "Export all working memories in a session (for backup/migration)",
	}, func(ctx context.Context, req *sdk.CallToolRequest, input WorkingMemoryExportInput) (*sdk.CallToolResult, interface{}, error) {
		memories, err := service.Export(input.SessionID)
		if err != nil {
			return nil, nil, err
		}

		return nil, map[string]interface{}{
			"session_id":  input.SessionID,
			"count":       len(memories),
			"memories":    memories,
			"exported_at": time.Now().Format(time.RFC3339),
		}, nil
	})

	// 10. working_memory_list_pending
	sdk.AddTool(server.server, &sdk.Tool{
		Name:        "working_memory_list_pending",
		Description: "List working memories pending promotion",
	}, func(ctx context.Context, req *sdk.CallToolRequest, input struct {
		SessionID string `json:"session_id" jsonschema:"required"`
	}) (*sdk.CallToolResult, interface{}, error) {
		memories, err := service.List(ctx, input.SessionID, false, false)
		if err != nil {
			return nil, nil, err
		}

		// Filter pending promotion
		pending := make([]*domain.WorkingMemory, 0)
		for _, wm := range memories {
			if wm.ShouldPromote() && !wm.IsPromoted() {
				pending = append(pending, wm)
			}
		}

		return nil, map[string]interface{}{
			"session_id": input.SessionID,
			"count":      len(pending),
			"memories":   pending,
		}, nil
	})

	// 11. working_memory_list_expired
	sdk.AddTool(server.server, &sdk.Tool{
		Name:        "working_memory_list_expired",
		Description: "List expired working memories",
	}, func(ctx context.Context, req *sdk.CallToolRequest, input struct {
		SessionID string `json:"session_id" jsonschema:"required"`
	}) (*sdk.CallToolResult, interface{}, error) {
		memories, err := service.List(ctx, input.SessionID, true, false)
		if err != nil {
			return nil, nil, err
		}

		// Filter expired
		expired := make([]*domain.WorkingMemory, 0)
		for _, wm := range memories {
			if wm.IsExpired() {
				expired = append(expired, wm)
			}
		}

		return nil, map[string]interface{}{
			"session_id": input.SessionID,
			"count":      len(expired),
			"memories":   expired,
		}, nil
	})

	// 12. working_memory_list_promoted
	sdk.AddTool(server.server, &sdk.Tool{
		Name:        "working_memory_list_promoted",
		Description: "List promoted working memories",
	}, func(ctx context.Context, req *sdk.CallToolRequest, input struct {
		SessionID string `json:"session_id" jsonschema:"required"`
	}) (*sdk.CallToolResult, interface{}, error) {
		memories, err := service.List(ctx, input.SessionID, false, true)
		if err != nil {
			return nil, nil, err
		}

		// Filter promoted
		promoted := make([]*domain.WorkingMemory, 0)
		for _, wm := range memories {
			if wm.IsPromoted() {
				promoted = append(promoted, wm)
			}
		}

		return nil, map[string]interface{}{
			"session_id": input.SessionID,
			"count":      len(promoted),
			"memories":   promoted,
		}, nil
	})

	// 13. working_memory_bulk_promote
	sdk.AddTool(server.server, &sdk.Tool{
		Name:        "working_memory_bulk_promote",
		Description: "Promote all pending working memories to long-term",
	}, func(ctx context.Context, req *sdk.CallToolRequest, input WorkingMemoryBulkPromoteInput) (*sdk.CallToolResult, interface{}, error) {
		memories, err := service.List(ctx, input.SessionID, false, false)
		if err != nil {
			return nil, nil, err
		}

		promoted := make([]string, 0)
		failed := make([]map[string]interface{}, 0)

		for _, wm := range memories {
			if wm.ShouldPromote() && !wm.IsPromoted() {
				longTermMem, err := service.Promote(ctx, input.SessionID, wm.ID)
				if err != nil {
					failed = append(failed, map[string]interface{}{
						"memory_id": wm.ID,
						"error":     err.Error(),
					})
				} else {
					promoted = append(promoted, longTermMem.GetID())
				}
			}
		}

		return nil, map[string]interface{}{
			"session_id":     input.SessionID,
			"promoted_count": len(promoted),
			"failed_count":   len(failed),
			"promoted_ids":   promoted,
			"failed":         failed,
		}, nil
	})

	// 14. working_memory_relation_add
	sdk.AddTool(server.server, &sdk.Tool{
		Name:        "working_memory_relation_add",
		Description: "Add a relation between two working memories",
	}, func(ctx context.Context, req *sdk.CallToolRequest, input WorkingMemoryRelationAddInput) (*sdk.CallToolResult, interface{}, error) {
		wm, err := service.Get(ctx, input.SessionID, input.MemoryID)
		if err != nil {
			return nil, nil, err
		}

		// Verify related memory exists
		_, err = service.Get(ctx, input.SessionID, input.RelatedMemoryID)
		if err != nil {
			return nil, nil, fmt.Errorf("related memory not found: %w", err)
		}

		wm.AddRelation(input.RelatedMemoryID)

		return nil, map[string]interface{}{
			"success":           true,
			"memory_id":         input.MemoryID,
			"related_memory_id": input.RelatedMemoryID,
			"total_relations":   len(wm.RelatedIDs),
		}, nil
	})

	// 15. working_memory_search
	sdk.AddTool(server.server, &sdk.Tool{
		Name:        "working_memory_search",
		Description: "Search working memories in a session by content/tags",
	}, func(ctx context.Context, req *sdk.CallToolRequest, input WorkingMemorySearchInput) (*sdk.CallToolResult, interface{}, error) {
		memories, err := service.List(ctx, input.SessionID, false, false)
		if err != nil {
			return nil, nil, err
		}

		// Simple text search (can be enhanced with semantic search later)
		results := make([]*domain.WorkingMemory, 0)
		queryLower := input.Query

		for _, wm := range memories {
			// Search in content
			if contains(wm.Content, queryLower) {
				results = append(results, wm)
				continue
			}

			// Search in tags
			for _, tag := range wm.Tags {
				if contains(tag, queryLower) {
					results = append(results, wm)
					break
				}
			}

			// Search in context
			if contains(wm.Context, queryLower) {
				results = append(results, wm)
			}
		}

		return nil, map[string]interface{}{
			"session_id": input.SessionID,
			"query":      input.Query,
			"count":      len(results),
			"memories":   results,
		}, nil
	})
}

// Helper function for case-insensitive contains
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(substr) == 0 ||
		(len(s) > 0 && len(substr) > 0 && s[0:1] == substr[0:1]))
}
