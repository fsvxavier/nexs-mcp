package mcp

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	sdk "github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/fsvxavier/nexs-mcp/internal/domain"
)

// --- SimpleElement implementation ---

// SimpleElement is a basic implementation of Element for MCP operations
type SimpleElement struct {
	metadata domain.ElementMetadata
}

func (s *SimpleElement) GetMetadata() domain.ElementMetadata { return s.metadata }
func (s *SimpleElement) Validate() error                     { return nil }
func (s *SimpleElement) GetType() domain.ElementType         { return s.metadata.Type }
func (s *SimpleElement) GetID() string                       { return s.metadata.ID }
func (s *SimpleElement) IsActive() bool                      { return s.metadata.IsActive }
func (s *SimpleElement) Activate() error {
	s.metadata.IsActive = true
	s.metadata.UpdatedAt = time.Now()
	return nil
}
func (s *SimpleElement) Deactivate() error {
	s.metadata.IsActive = false
	s.metadata.UpdatedAt = time.Now()
	return nil
}

// --- Input/Output structures for tools ---

// ListElementsInput defines input for list_elements tool
type ListElementsInput struct {
	Type       string `json:"type,omitempty" jsonschema:"element type filter (persona, skill, template, agent, memory, ensemble)"`
	IsActive   *bool  `json:"is_active,omitempty" jsonschema:"active status filter"`
	ActiveOnly bool   `json:"active_only,omitempty" jsonschema:"if true, return only active elements (shortcut for is_active=true)"`
	Tags       string `json:"tags,omitempty" jsonschema:"comma-separated tags to filter"`
	User       string `json:"user,omitempty" jsonschema:"authenticated username for access control (optional)"`
}

// ListElementsOutput defines output for list_elements tool
type ListElementsOutput struct {
	Elements []map[string]interface{} `json:"elements" jsonschema:"list of elements"`
	Total    int                      `json:"total" jsonschema:"total number of elements"`
}

// GetElementInput defines input for get_element tool
type GetElementInput struct {
	ID   string `json:"id" jsonschema:"the element ID"`
	User string `json:"user,omitempty" jsonschema:"authenticated username for access control (optional)"`
}

// GetElementOutput defines output for get_element tool
type GetElementOutput struct {
	Element map[string]interface{} `json:"element" jsonschema:"the element details"`
}

// CreateElementInput defines input for create_element tool
type CreateElementInput struct {
	Type        string   `json:"type" jsonschema:"element type (persona, skill, template, agent, memory, ensemble)"`
	Name        string   `json:"name" jsonschema:"element name (3-100 characters)"`
	Description string   `json:"description,omitempty" jsonschema:"element description (max 500 characters)"`
	Version     string   `json:"version" jsonschema:"element version (semver)"`
	Author      string   `json:"author" jsonschema:"element author"`
	Tags        []string `json:"tags,omitempty" jsonschema:"element tags"`
	IsActive    bool     `json:"is_active,omitempty" jsonschema:"active status (default: true)"`
	User        string   `json:"user,omitempty" jsonschema:"authenticated username for access control (optional)"`
}

// CreateElementOutput defines output for create_element tool
type CreateElementOutput struct {
	ID      string                 `json:"id" jsonschema:"the created element ID"`
	Element map[string]interface{} `json:"element" jsonschema:"the created element details"`
}

// UpdateElementInput defines input for update_element tool
type UpdateElementInput struct {
	ID          string   `json:"id" jsonschema:"the element ID"`
	Name        string   `json:"name,omitempty" jsonschema:"element name"`
	Description string   `json:"description,omitempty" jsonschema:"element description"`
	Tags        []string `json:"tags,omitempty" jsonschema:"element tags"`
	IsActive    *bool    `json:"is_active,omitempty" jsonschema:"active status"`
	User        string   `json:"user,omitempty" jsonschema:"authenticated username for access control (optional)"`
}

// UpdateElementOutput defines output for update_element tool
type UpdateElementOutput struct {
	Element map[string]interface{} `json:"element" jsonschema:"the updated element details"`
}

// DeleteElementInput defines input for delete_element tool
type DeleteElementInput struct {
	ID   string `json:"id" jsonschema:"the element ID to delete"`
	User string `json:"user,omitempty" jsonschema:"authenticated username for access control (optional)"`
}

// DeleteElementOutput defines output for delete_element tool
type DeleteElementOutput struct {
	Success bool   `json:"success" jsonschema:"deletion success status"`
	Message string `json:"message" jsonschema:"deletion result message"`
}

// DuplicateElementInput defines input for duplicate_element tool
type DuplicateElementInput struct {
	ID      string `json:"id" jsonschema:"the element ID to duplicate"`
	NewName string `json:"new_name,omitempty" jsonschema:"optional new name for the duplicate (default: 'Copy of {original_name}')"`
	User    string `json:"user,omitempty" jsonschema:"authenticated username for access control (optional)"`
}

// DuplicateElementOutput defines output for duplicate_element tool
type DuplicateElementOutput struct {
	ID      string                 `json:"id" jsonschema:"the duplicated element ID"`
	Element map[string]interface{} `json:"element" jsonschema:"the duplicated element details"`
	Message string                 `json:"message" jsonschema:"duplication result message"`
}

// GetUsageStatsInput defines input for get_usage_stats tool
type GetUsageStatsInput struct {
	Period string `json:"period,omitempty" jsonschema:"time period for statistics (last_hour, last_24h, last_7_days, last_30_days, all)"`
}

// GetUsageStatsOutput defines output for get_usage_stats tool
type GetUsageStatsOutput struct {
	TotalOperations    int                      `json:"total_operations" jsonschema:"total number of operations"`
	SuccessfulOps      int                      `json:"successful_ops" jsonschema:"number of successful operations"`
	FailedOps          int                      `json:"failed_ops" jsonschema:"number of failed operations"`
	SuccessRate        float64                  `json:"success_rate" jsonschema:"success rate percentage"`
	OperationsByTool   map[string]int           `json:"operations_by_tool" jsonschema:"operation count by tool name"`
	ErrorsByTool       map[string]int           `json:"errors_by_tool" jsonschema:"error count by tool name"`
	AvgDurationByTool  map[string]float64       `json:"avg_duration_by_tool_ms" jsonschema:"average duration in milliseconds by tool"`
	MostUsedTools      []map[string]interface{} `json:"most_used_tools" jsonschema:"top 10 most used tools"`
	SlowestOperations  []map[string]interface{} `json:"slowest_operations" jsonschema:"top 10 slowest operations"`
	RecentErrors       []map[string]interface{} `json:"recent_errors" jsonschema:"most recent errors"`
	ActiveUsers        []string                 `json:"active_users" jsonschema:"list of active users"`
	OperationsByPeriod map[string]int           `json:"operations_by_period" jsonschema:"operations grouped by date"`
	Period             string                   `json:"period" jsonschema:"period queried"`
	StartTime          string                   `json:"start_time" jsonschema:"period start time (ISO 8601)"`
	EndTime            string                   `json:"end_time" jsonschema:"period end time (ISO 8601)"`
}

// --- Tool handlers ---

// handleListElements handles list_elements tool calls
func (s *MCPServer) handleListElements(ctx context.Context, req *sdk.CallToolRequest, input ListElementsInput) (*sdk.CallToolResult, ListElementsOutput, error) {
	// Build filter
	filter := domain.ElementFilter{}

	if input.Type != "" {
		elementType := domain.ElementType(input.Type)
		filter.Type = &elementType
	}

	// Handle active_only shortcut - takes priority over is_active
	if input.ActiveOnly {
		isActive := true
		filter.IsActive = &isActive
	} else if input.IsActive != nil {
		filter.IsActive = input.IsActive
	}

	if input.Tags != "" {
		// Parse comma-separated tags
		// For simplicity, we'll add basic tag parsing here
		filter.Tags = []string{input.Tags}
	}

	// List elements
	elements, err := s.repo.List(filter)
	if err != nil {
		return nil, ListElementsOutput{}, fmt.Errorf("failed to list elements: %w", err)
	}

	// Apply access control filtering
	userCtx := GetUserContext(input.User)
	accessControl := domain.NewAccessControl()
	filteredElements := accessControl.FilterByPermissions(userCtx, elements)

	// Convert to map format
	result := make([]map[string]interface{}, 0, len(filteredElements))
	for _, elem := range filteredElements {
		result = append(result, elem.GetMetadata().ToMap())
	}

	output := ListElementsOutput{
		Elements: result,
		Total:    len(result),
	}

	return nil, output, nil
}

// handleGetElement handles get_element tool calls
func (s *MCPServer) handleGetElement(ctx context.Context, req *sdk.CallToolRequest, input GetElementInput) (*sdk.CallToolResult, GetElementOutput, error) {
	if input.ID == "" {
		return nil, GetElementOutput{}, fmt.Errorf("id is required")
	}

	element, err := s.repo.GetByID(input.ID)
	if err != nil {
		return nil, GetElementOutput{}, fmt.Errorf("failed to get element: %w", err)
	}

	// Check read permission
	userCtx := GetUserContext(input.User)
	accessControl := domain.NewAccessControl()

	// Extract privacy fields from element (if it's a Persona, otherwise allow public access)
	owner := element.GetMetadata().Author
	privacyLevel := domain.PrivacyLevelPublic
	var sharedWith []string

	if persona, ok := element.(*domain.Persona); ok {
		privacyLevel = domain.PrivacyLevel(persona.PrivacyLevel)
		sharedWith = persona.SharedWith
	}

	if !accessControl.CheckReadPermission(userCtx, owner, privacyLevel, sharedWith) {
		return nil, GetElementOutput{}, fmt.Errorf("access denied: user does not have read permission")
	}

	output := GetElementOutput{
		Element: element.GetMetadata().ToMap(),
	}

	return nil, output, nil
}

// handleCreateElement handles create_element tool calls
func (s *MCPServer) handleCreateElement(ctx context.Context, req *sdk.CallToolRequest, input CreateElementInput) (*sdk.CallToolResult, CreateElementOutput, error) {
	// Validate input
	if input.Type == "" {
		return nil, CreateElementOutput{}, fmt.Errorf("type is required")
	}
	if input.Name == "" || len(input.Name) < 3 || len(input.Name) > 100 {
		return nil, CreateElementOutput{}, fmt.Errorf("name must be between 3 and 100 characters")
	}
	if len(input.Description) > 500 {
		return nil, CreateElementOutput{}, fmt.Errorf("description must be at most 500 characters")
	}
	if input.Version == "" {
		return nil, CreateElementOutput{}, fmt.Errorf("version is required")
	}
	if input.Author == "" {
		return nil, CreateElementOutput{}, fmt.Errorf("author is required")
	}

	// Validate element type
	validTypes := map[string]bool{
		"persona":  true,
		"skill":    true,
		"template": true,
		"agent":    true,
		"memory":   true,
		"ensemble": true,
	}
	if !validTypes[input.Type] {
		return nil, CreateElementOutput{}, fmt.Errorf("invalid element type: %s", input.Type)
	}

	// Generate ID
	id := uuid.New().String()
	now := time.Now()

	// Create metadata
	metadata := domain.ElementMetadata{
		ID:          id,
		Type:        domain.ElementType(input.Type),
		Name:        input.Name,
		Description: input.Description,
		Version:     input.Version,
		Author:      input.Author,
		Tags:        input.Tags,
		IsActive:    input.IsActive,
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	// Create SimpleElement
	element := &SimpleElement{metadata: metadata}

	// Save to repository
	if err := s.repo.Create(element); err != nil {
		return nil, CreateElementOutput{}, fmt.Errorf("failed to create element: %w", err)
	}

	output := CreateElementOutput{
		ID:      id,
		Element: metadata.ToMap(),
	}

	return nil, output, nil
}

// handleUpdateElement handles update_element tool calls
func (s *MCPServer) handleUpdateElement(ctx context.Context, req *sdk.CallToolRequest, input UpdateElementInput) (*sdk.CallToolResult, UpdateElementOutput, error) {
	if input.ID == "" {
		return nil, UpdateElementOutput{}, fmt.Errorf("id is required")
	}

	// Get existing element
	element, err := s.repo.GetByID(input.ID)
	if err != nil {
		return nil, UpdateElementOutput{}, fmt.Errorf("failed to get element: %w", err)
	}

	// Check write permission
	userCtx := GetUserContext(input.User)
	accessControl := domain.NewAccessControl()
	owner := element.GetMetadata().Author

	if !accessControl.CheckWritePermission(userCtx, owner) {
		return nil, UpdateElementOutput{}, fmt.Errorf("access denied: only the owner can update this element")
	}

	metadata := element.GetMetadata()

	// Update fields
	updated := false

	if input.Name != "" && input.Name != metadata.Name {
		metadata.Name = input.Name
		updated = true
	}

	if input.Description != "" && input.Description != metadata.Description {
		metadata.Description = input.Description
		updated = true
	}

	if len(input.Tags) > 0 {
		metadata.Tags = input.Tags
		updated = true
	}

	if input.IsActive != nil && *input.IsActive != metadata.IsActive {
		metadata.IsActive = *input.IsActive
		updated = true
	}

	if updated {
		metadata.UpdatedAt = time.Now()

		// Create updated element
		updatedElement := &SimpleElement{metadata: metadata}

		if err := s.repo.Update(updatedElement); err != nil {
			return nil, UpdateElementOutput{}, fmt.Errorf("failed to update element: %w", err)
		}
	}

	output := UpdateElementOutput{
		Element: metadata.ToMap(),
	}

	return nil, output, nil
}

// handleDeleteElement handles delete_element tool calls
func (s *MCPServer) handleDeleteElement(ctx context.Context, req *sdk.CallToolRequest, input DeleteElementInput) (*sdk.CallToolResult, DeleteElementOutput, error) {
	if input.ID == "" {
		return nil, DeleteElementOutput{}, fmt.Errorf("id is required")
	}

	// Get element to check permissions
	element, err := s.repo.GetByID(input.ID)
	if err != nil {
		return nil, DeleteElementOutput{
			Success: false,
			Message: fmt.Sprintf("failed to get element: %v", err),
		}, nil
	}

	// Check delete permission
	userCtx := GetUserContext(input.User)
	accessControl := domain.NewAccessControl()
	owner := element.GetMetadata().Author

	if !accessControl.CheckDeletePermission(userCtx, owner) {
		return nil, DeleteElementOutput{
			Success: false,
			Message: "access denied: only the owner can delete this element",
		}, nil
	}

	if err := s.repo.Delete(input.ID); err != nil {
		return nil, DeleteElementOutput{
			Success: false,
			Message: fmt.Sprintf("failed to delete element: %v", err),
		}, nil
	}

	output := DeleteElementOutput{
		Success: true,
		Message: fmt.Sprintf("Element %s deleted successfully", input.ID),
	}

	return nil, output, nil
}

// handleDuplicateElement handles duplicate_element tool calls
func (s *MCPServer) handleDuplicateElement(ctx context.Context, req *sdk.CallToolRequest, input DuplicateElementInput) (*sdk.CallToolResult, DuplicateElementOutput, error) {
	if input.ID == "" {
		return nil, DuplicateElementOutput{}, fmt.Errorf("id is required")
	}

	// Get original element
	original, err := s.repo.GetByID(input.ID)
	if err != nil {
		return nil, DuplicateElementOutput{}, fmt.Errorf("failed to get original element: %w", err)
	}

	// Check read permission on original
	userCtx := GetUserContext(input.User)
	accessControl := domain.NewAccessControl()
	owner := original.GetMetadata().Author
	privacyLevel := domain.PrivacyLevelPublic
	var sharedWith []string

	if persona, ok := original.(*domain.Persona); ok {
		privacyLevel = domain.PrivacyLevel(persona.PrivacyLevel)
		sharedWith = persona.SharedWith
	}

	if !accessControl.CheckReadPermission(userCtx, owner, privacyLevel, sharedWith) {
		return nil, DuplicateElementOutput{}, fmt.Errorf("access denied: user does not have read permission on original element")
	}

	// Create duplicate metadata
	originalMeta := original.GetMetadata()
	timestamp := time.Now().Format("20060102-150405")
	newID := fmt.Sprintf("%s-copy-%s", originalMeta.ID, timestamp)

	newName := input.NewName
	if newName == "" {
		newName = fmt.Sprintf("Copy of %s", originalMeta.Name)
	}

	duplicateMeta := domain.ElementMetadata{
		ID:          newID,
		Type:        originalMeta.Type,
		Name:        newName,
		Description: originalMeta.Description,
		Version:     originalMeta.Version,
		Author:      originalMeta.Author,
		Tags:        append([]string{}, originalMeta.Tags...), // Copy tags
		IsActive:    originalMeta.IsActive,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	// Create duplicate element
	duplicate := &SimpleElement{metadata: duplicateMeta}

	// Save duplicate
	if err := s.repo.Create(duplicate); err != nil {
		return nil, DuplicateElementOutput{}, fmt.Errorf("failed to create duplicate: %w", err)
	}

	output := DuplicateElementOutput{
		ID:      newID,
		Element: duplicateMeta.ToMap(),
		Message: fmt.Sprintf("Element duplicated successfully: %s -> %s", input.ID, newID),
	}

	return nil, output, nil
}
