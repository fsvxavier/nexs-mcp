package mcp

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/fsvxavier/nexs-mcp/internal/collection"
	"github.com/fsvxavier/nexs-mcp/internal/collection/sources"
)

// CollectionTools provides MCP tools for managing collections.
type CollectionTools struct {
	registry  *collection.Registry
	installer *collection.Installer
	manager   *collection.Manager
}

// NewCollectionTools creates a new collection tools manager.
func NewCollectionTools(registry *collection.Registry, installer *collection.Installer) *CollectionTools {
	return &CollectionTools{
		registry:  registry,
		installer: installer,
		manager:   collection.NewManager(installer, registry),
	}
}

// ToolDefinitions returns the MCP tool definitions for collections.
func (ct *CollectionTools) ToolDefinitions() []ToolDef {
	return []ToolDef{
		{
			Name:        "browse_collections",
			Description: "Discover and search for available collections from configured sources",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"source": map[string]interface{}{
						"type":        "string",
						"description": "Optional: Filter by specific source (github, local, http)",
					},
					"category": map[string]interface{}{
						"type":        "string",
						"description": "Filter by category (devops, creative-writing, etc.)",
					},
					"author": map[string]interface{}{
						"type":        "string",
						"description": "Filter by author name",
					},
					"tags": map[string]interface{}{
						"type":        "array",
						"items":       map[string]interface{}{"type": "string"},
						"description": "Filter by tags (must have ALL specified tags)",
					},
					"query": map[string]interface{}{
						"type":        "string",
						"description": "Text search across name, description, and keywords",
					},
					"limit": map[string]interface{}{
						"type":        "number",
						"description": "Maximum number of results (default: 50)",
					},
					"offset": map[string]interface{}{
						"type":        "number",
						"description": "Number of results to skip (for pagination)",
					},
				},
			},
		},
		{
			Name:        "install_collection",
			Description: "Install a collection from a URI (github://, file://, or https://)",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"uri": map[string]interface{}{
						"type":        "string",
						"description": "Collection URI (e.g., github://owner/repo, file:///path, https://...)",
					},
					"force": map[string]interface{}{
						"type":        "boolean",
						"description": "Force reinstallation if already installed",
					},
					"skip_dependencies": map[string]interface{}{
						"type":        "boolean",
						"description": "Skip installing dependencies",
					},
					"skip_validation": map[string]interface{}{
						"type":        "boolean",
						"description": "Skip validation checks",
					},
					"skip_hooks": map[string]interface{}{
						"type":        "boolean",
						"description": "Skip pre/post install hooks",
					},
				},
				"required": []string{"uri"},
			},
		},
		{
			Name:        "uninstall_collection",
			Description: "Uninstall a collection by its ID (author/name)",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"id": map[string]interface{}{
						"type":        "string",
						"description": "Collection ID (author/name)",
					},
					"force": map[string]interface{}{
						"type":        "boolean",
						"description": "Force uninstall even if other collections depend on it",
					},
				},
				"required": []string{"id"},
			},
		},
		{
			Name:        "list_installed_collections",
			Description: "List all installed collections with their status and metadata",
			InputSchema: map[string]interface{}{
				"type":       "object",
				"properties": map[string]interface{}{},
			},
		},
		{
			Name:        "get_collection_info",
			Description: "Get detailed information about a collection (installed or available)",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"uri": map[string]interface{}{
						"type":        "string",
						"description": "Collection URI or ID (for installed collections)",
					},
				},
				"required": []string{"uri"},
			},
		},
		{
			Name:        "export_collection",
			Description: "Export a collection to a tar.gz archive",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"id": map[string]interface{}{
						"type":        "string",
						"description": "Collection ID (author/name)",
					},
					"output": map[string]interface{}{
						"type":        "string",
						"description": "Output path for the tar.gz file",
					},
					"include_backups": map[string]interface{}{
						"type":        "boolean",
						"description": "Include backup files in export",
					},
					"compression": map[string]interface{}{
						"type":        "string",
						"description": "Compression level: none, fast, best (default: best)",
					},
				},
				"required": []string{"id", "output"},
			},
		},
		{
			Name:        "update_collection",
			Description: "Update a specific collection to the latest version",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"id": map[string]interface{}{
						"type":        "string",
						"description": "Collection ID (author/name)",
					},
					"skip_dependencies": map[string]interface{}{
						"type":        "boolean",
						"description": "Skip updating dependencies",
					},
					"skip_validation": map[string]interface{}{
						"type":        "boolean",
						"description": "Skip validation checks",
					},
					"skip_hooks": map[string]interface{}{
						"type":        "boolean",
						"description": "Skip pre/post update hooks",
					},
				},
				"required": []string{"id"},
			},
		},
		{
			Name:        "update_all_collections",
			Description: "Update all installed collections to their latest versions",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"skip_dependencies": map[string]interface{}{
						"type":        "boolean",
						"description": "Skip updating dependencies",
					},
					"skip_validation": map[string]interface{}{
						"type":        "boolean",
						"description": "Skip validation checks",
					},
					"skip_hooks": map[string]interface{}{
						"type":        "boolean",
						"description": "Skip pre/post update hooks",
					},
				},
			},
		},
		{
			Name:        "check_collection_updates",
			Description: "Check for available updates for all installed collections",
			InputSchema: map[string]interface{}{
				"type":       "object",
				"properties": map[string]interface{}{},
			},
		},
		{
			Name:        "publish_collection",
			Description: "Publish a collection to GitHub repository",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"id": map[string]interface{}{
						"type":        "string",
						"description": "Collection ID (author/name)",
					},
					"github_repo": map[string]interface{}{
						"type":        "string",
						"description": "GitHub repository (owner/repo)",
					},
					"branch": map[string]interface{}{
						"type":        "string",
						"description": "Target branch (default: main)",
					},
					"commit_message": map[string]interface{}{
						"type":        "string",
						"description": "Commit message",
					},
					"create_release": map[string]interface{}{
						"type":        "boolean",
						"description": "Create a GitHub release",
					},
					"release_tag": map[string]interface{}{
						"type":        "string",
						"description": "Release tag (defaults to version)",
					},
					"release_notes": map[string]interface{}{
						"type":        "string",
						"description": "Release notes",
					},
					"force": map[string]interface{}{
						"type":        "boolean",
						"description": "Force push changes",
					},
				},
				"required": []string{"id"},
			},
		},
	}
}

// HandleTool routes tool calls to the appropriate handler.
func (ct *CollectionTools) HandleTool(ctx context.Context, name string, arguments map[string]interface{}) (interface{}, error) {
	switch name {
	case "browse_collections":
		return ct.handleBrowse(ctx, arguments)
	case "install_collection":
		return ct.handleInstall(ctx, arguments)
	case "uninstall_collection":
		return ct.handleUninstall(ctx, arguments)
	case "list_installed_collections":
		return ct.handleListInstalled(ctx, arguments)
	case "get_collection_info":
		return ct.handleGetInfo(ctx, arguments)
	case "export_collection":
		return ct.handleExport(ctx, arguments)
	case "update_collection":
		return ct.handleUpdate(ctx, arguments)
	case "update_all_collections":
		return ct.handleUpdateAll(ctx, arguments)
	case "check_collection_updates":
		return ct.handleCheckUpdates(ctx, arguments)
	case "publish_collection":
		return ct.handlePublish(ctx, arguments)
	default:
		return nil, fmt.Errorf("unknown tool: %s", name)
	}
}

// handleBrowse implements browse_collections tool.
func (ct *CollectionTools) handleBrowse(ctx context.Context, args map[string]interface{}) (interface{}, error) {
	// Parse filter
	filter := &sources.BrowseFilter{}

	if category, ok := args["category"].(string); ok {
		filter.Category = category
	}
	if author, ok := args["author"].(string); ok {
		filter.Author = author
	}
	if query, ok := args["query"].(string); ok {
		filter.Query = query
	}
	if limit, ok := args["limit"].(float64); ok {
		filter.Limit = int(limit)
	} else {
		filter.Limit = 50 // Default
	}
	if offset, ok := args["offset"].(float64); ok {
		filter.Offset = int(offset)
	}
	if tags, ok := args["tags"].([]interface{}); ok {
		filter.Tags = make([]string, 0, len(tags))
		for _, tag := range tags {
			if tagStr, ok := tag.(string); ok {
				filter.Tags = append(filter.Tags, tagStr)
			}
		}
	}

	sourceName := ""
	if source, ok := args["source"].(string); ok {
		sourceName = source
	}

	// Browse collections
	collections, err := ct.registry.Browse(ctx, filter, sourceName)
	if err != nil {
		return nil, fmt.Errorf("failed to browse collections: %w", err)
	}

	return map[string]interface{}{
		"collections": collections,
		"count":       len(collections),
		"filter":      filter,
	}, nil
}

// handleInstall implements install_collection tool.
func (ct *CollectionTools) handleInstall(ctx context.Context, args map[string]interface{}) (interface{}, error) {
	uri, ok := args["uri"].(string)
	if !ok {
		return nil, errors.New("uri parameter is required")
	}

	options := &collection.InstallOptions{}
	if force, ok := args["force"].(bool); ok {
		options.Force = force
	}
	if skipDeps, ok := args["skip_dependencies"].(bool); ok {
		options.SkipDependencies = skipDeps
	}
	if skipVal, ok := args["skip_validation"].(bool); ok {
		options.SkipValidation = skipVal
	}
	if skipHooks, ok := args["skip_hooks"].(bool); ok {
		options.SkipHooks = skipHooks
	}

	// Install collection
	if err := ct.installer.Install(ctx, uri, options); err != nil {
		return nil, fmt.Errorf("installation failed: %w", err)
	}

	// Get installed record
	collection, err := ct.registry.Get(ctx, uri)
	if err != nil {
		return map[string]interface{}{
			"status":  "installed",
			"message": "Collection installed successfully",
		}, nil
	}

	manifestMap, _ := collection.Manifest.(map[string]interface{})
	manifest, _ := parseManifestFromMap(manifestMap)
	if manifest != nil {
		record, _ := ct.installer.GetInstalled(manifest.ID())
		return map[string]interface{}{
			"status":       "installed",
			"message":      "Collection installed successfully",
			"collection":   collection.Metadata,
			"installation": record,
		}, nil
	}

	return map[string]interface{}{
		"status":  "installed",
		"message": "Collection installed successfully",
	}, nil
}

// handleUninstall implements uninstall_collection tool.
func (ct *CollectionTools) handleUninstall(ctx context.Context, args map[string]interface{}) (interface{}, error) {
	id, ok := args["id"].(string)
	if !ok {
		return nil, errors.New("id parameter is required")
	}

	options := &collection.UninstallOptions{}
	if force, ok := args["force"].(bool); ok {
		options.Force = force
	}

	if err := ct.installer.Uninstall(ctx, id, options); err != nil {
		return nil, fmt.Errorf("uninstallation failed: %w", err)
	}

	return map[string]interface{}{
		"status":  "uninstalled",
		"message": fmt.Sprintf("Collection %s uninstalled successfully", id),
	}, nil
}

// handleListInstalled implements list_installed_collections tool.
func (ct *CollectionTools) handleListInstalled(ctx context.Context, args map[string]interface{}) (interface{}, error) {
	installed := ct.installer.ListInstalled()

	return map[string]interface{}{
		"collections": installed,
		"count":       len(installed),
	}, nil
}

// handleGetInfo implements get_collection_info tool.
func (ct *CollectionTools) handleGetInfo(ctx context.Context, args map[string]interface{}) (interface{}, error) {
	uri, ok := args["uri"].(string)
	if !ok {
		return nil, errors.New("uri parameter is required")
	}

	// Try to get from installed first
	if record, exists := ct.installer.GetInstalled(uri); exists {
		return map[string]interface{}{
			"source":       "installed",
			"installation": record,
		}, nil
	}

	// Get from registry
	collection, err := ct.registry.Get(ctx, uri)
	if err != nil {
		return nil, fmt.Errorf("collection not found: %w", err)
	}

	return map[string]interface{}{
		"source":     "registry",
		"metadata":   collection.Metadata,
		"manifest":   collection.Manifest,
		"sourceData": collection.SourceData,
	}, nil
}

// handleExport implements export_collection tool.
func (ct *CollectionTools) handleExport(ctx context.Context, args map[string]interface{}) (interface{}, error) {
	id, ok := args["id"].(string)
	if !ok {
		return nil, errors.New("id parameter is required")
	}

	output, ok := args["output"].(string)
	if !ok {
		return nil, errors.New("output parameter is required")
	}

	options := &collection.ExportOptions{}
	if includeBackups, ok := args["include_backups"].(bool); ok {
		options.IncludeBackups = includeBackups
	}
	if compression, ok := args["compression"].(string); ok {
		options.Compression = compression
	}

	// Export collection
	if err := ct.manager.Export(ctx, id, output, options); err != nil {
		return nil, fmt.Errorf("export failed: %w", err)
	}

	return map[string]interface{}{
		"status":  "exported",
		"message": "Collection exported to " + output,
		"path":    output,
	}, nil
}

// handleUpdate implements update_collection tool.
func (ct *CollectionTools) handleUpdate(ctx context.Context, args map[string]interface{}) (interface{}, error) {
	id, ok := args["id"].(string)
	if !ok {
		return nil, errors.New("id parameter is required")
	}

	options := &collection.UpdateOptions{}
	if skipDeps, ok := args["skip_dependencies"].(bool); ok {
		options.SkipDependencies = skipDeps
	}
	if skipVal, ok := args["skip_validation"].(bool); ok {
		options.SkipValidation = skipVal
	}
	if skipHooks, ok := args["skip_hooks"].(bool); ok {
		options.SkipHooks = skipHooks
	}

	result, err := ct.manager.Update(ctx, id, options)
	if err != nil {
		return nil, fmt.Errorf("update failed: %w", err)
	}

	return map[string]interface{}{
		"result": result,
		"status": "completed",
	}, nil
}

// handleUpdateAll implements update_all_collections tool.
func (ct *CollectionTools) handleUpdateAll(ctx context.Context, args map[string]interface{}) (interface{}, error) {
	options := &collection.UpdateOptions{}
	if skipDeps, ok := args["skip_dependencies"].(bool); ok {
		options.SkipDependencies = skipDeps
	}
	if skipVal, ok := args["skip_validation"].(bool); ok {
		options.SkipValidation = skipVal
	}
	if skipHooks, ok := args["skip_hooks"].(bool); ok {
		options.SkipHooks = skipHooks
	}

	results, err := ct.manager.UpdateAll(ctx, options)
	if err != nil {
		return nil, fmt.Errorf("update all failed: %w", err)
	}

	successCount := 0
	for _, result := range results {
		if result.Updated {
			successCount++
		}
	}

	return map[string]interface{}{
		"results": results,
		"total":   len(results),
		"updated": successCount,
		"status":  "completed",
	}, nil
}

// handleCheckUpdates implements check_collection_updates tool.
func (ct *CollectionTools) handleCheckUpdates(ctx context.Context, args map[string]interface{}) (interface{}, error) {
	results, err := ct.manager.CheckUpdates(ctx)
	if err != nil {
		return nil, fmt.Errorf("check updates failed: %w", err)
	}

	updateCount := 0
	for _, result := range results {
		if result.UpdateAvailable {
			updateCount++
		}
	}

	return map[string]interface{}{
		"results":           results,
		"total":             len(results),
		"updates_available": updateCount,
	}, nil
}

// handlePublish implements publish_collection tool.
func (ct *CollectionTools) handlePublish(ctx context.Context, args map[string]interface{}) (interface{}, error) {
	id, ok := args["id"].(string)
	if !ok {
		return nil, errors.New("id parameter is required")
	}

	options := &collection.PublishOptions{}
	if repo, ok := args["github_repo"].(string); ok {
		options.GitHubRepo = repo
	}
	if branch, ok := args["branch"].(string); ok {
		options.Branch = branch
	}
	if commitMsg, ok := args["commit_message"].(string); ok {
		options.CommitMessage = commitMsg
	}
	if createRelease, ok := args["create_release"].(bool); ok {
		options.CreateRelease = createRelease
	}
	if releaseTag, ok := args["release_tag"].(string); ok {
		options.ReleaseTag = releaseTag
	}
	if releaseNotes, ok := args["release_notes"].(string); ok {
		options.ReleaseNotes = releaseNotes
	}
	if force, ok := args["force"].(bool); ok {
		options.Force = force
	}

	if err := ct.manager.Publish(ctx, id, options); err != nil {
		return nil, fmt.Errorf("publish failed: %w", err)
	}

	return map[string]interface{}{
		"status":  "published",
		"message": fmt.Sprintf("Collection %s published successfully", id),
	}, nil
}

// Helper function to parse manifest from map.
func parseManifestFromMap(m map[string]interface{}) (*collection.Manifest, error) {
	// Marshal to JSON and unmarshal to Manifest
	data, err := json.Marshal(m)
	if err != nil {
		return nil, err
	}

	var manifest collection.Manifest
	if err := json.Unmarshal(data, &manifest); err != nil {
		return nil, err
	}

	return &manifest, nil
}

// ToolDef represents an MCP tool definition.
type ToolDef struct {
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	InputSchema map[string]interface{} `json:"inputSchema"`
}
