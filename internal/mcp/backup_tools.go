package mcp

import (
	"context"
	"errors"
	"fmt"
	"time"

	sdk "github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/fsvxavier/nexs-mcp/internal/application"
	"github.com/fsvxavier/nexs-mcp/internal/backup"
)

// --- Backup/Restore Input/Output structures ---

// BackupPortfolioInput defines input for backup_portfolio tool.
type BackupPortfolioInput struct {
	Path        string `json:"path"                  jsonschema:"backup file path (.tar.gz)"`
	Compression string `json:"compression,omitempty" jsonschema:"compression level: none, fast, best (default: best)"`
	Description string `json:"description,omitempty" jsonschema:"backup description"`
	Author      string `json:"author,omitempty"      jsonschema:"backup author"`
}

// BackupPortfolioOutput defines output for backup_portfolio tool.
type BackupPortfolioOutput struct {
	Path         string `json:"path"          jsonschema:"backup file path"`
	ElementCount int    `json:"element_count" jsonschema:"number of elements backed up"`
	TotalSize    int64  `json:"total_size"    jsonschema:"total backup size in bytes"`
	Checksum     string `json:"checksum"      jsonschema:"SHA-256 checksum"`
	Version      string `json:"version"       jsonschema:"backup format version"`
}

// RestorePortfolioInput defines input for restore_portfolio tool.
type RestorePortfolioInput struct {
	Path           string `json:"path"                      jsonschema:"backup file path to restore from"`
	Overwrite      bool   `json:"overwrite,omitempty"       jsonschema:"overwrite existing elements (default: false)"`
	SkipValidation bool   `json:"skip_validation,omitempty" jsonschema:"skip backup validation (default: false)"`
	MergeStrategy  string `json:"merge_strategy,omitempty"  jsonschema:"merge strategy: skip, overwrite, merge (default: skip)"`
	BackupBefore   bool   `json:"backup_before,omitempty"   jsonschema:"create backup before restore (default: true)"`
	DryRun         bool   `json:"dry_run,omitempty"         jsonschema:"validate without applying changes (default: false)"`
}

// RestorePortfolioOutput defines output for restore_portfolio tool.
type RestorePortfolioOutput struct {
	Success         bool     `json:"success"               jsonschema:"whether restore was successful"`
	ElementsAdded   int      `json:"elements_added"        jsonschema:"number of elements added"`
	ElementsUpdated int      `json:"elements_updated"      jsonschema:"number of elements updated"`
	ElementsSkipped int      `json:"elements_skipped"      jsonschema:"number of elements skipped"`
	Errors          []string `json:"errors,omitempty"      jsonschema:"list of errors encountered"`
	BackupPath      string   `json:"backup_path,omitempty" jsonschema:"pre-restore backup path"`
	Duration        string   `json:"duration"              jsonschema:"restore duration"`
}

// --- Element Activation Input/Output structures ---

// ActivateElementInput defines input for activate_element tool.
type ActivateElementInput struct {
	ID   string `json:"id"             jsonschema:"the element ID to activate"`
	User string `json:"user,omitempty" jsonschema:"authenticated username for access control (optional)"`
}

// ActivateElementOutput defines output for activate_element tool.
type ActivateElementOutput struct {
	ID        string `json:"id"         jsonschema:"the activated element ID"`
	Name      string `json:"name"       jsonschema:"the element name"`
	Type      string `json:"type"       jsonschema:"the element type"`
	IsActive  bool   `json:"is_active"  jsonschema:"current active status (should be true)"`
	UpdatedAt string `json:"updated_at" jsonschema:"timestamp when element was activated"`
}

// DeactivateElementInput defines input for deactivate_element tool.
type DeactivateElementInput struct {
	ID   string `json:"id"             jsonschema:"the element ID to deactivate"`
	User string `json:"user,omitempty" jsonschema:"authenticated username for access control (optional)"`
}

// DeactivateElementOutput defines output for deactivate_element tool.
type DeactivateElementOutput struct {
	ID        string `json:"id"         jsonschema:"the deactivated element ID"`
	Name      string `json:"name"       jsonschema:"the element name"`
	Type      string `json:"type"       jsonschema:"the element type"`
	IsActive  bool   `json:"is_active"  jsonschema:"current active status (should be false)"`
	UpdatedAt string `json:"updated_at" jsonschema:"timestamp when element was deactivated"`
}

// --- Tool handlers ---

// handleBackupPortfolio handles the backup_portfolio tool.
func (s *MCPServer) handleBackupPortfolio(ctx context.Context, req *sdk.CallToolRequest, input BackupPortfolioInput) (*sdk.CallToolResult, BackupPortfolioOutput, error) {
	startTime := time.Now()
	var handlerErr error
	defer func() {
		s.metrics.RecordToolCall(application.ToolCallMetric{
			ToolName:  "backup_portfolio",
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

	// Validate required fields
	if input.Path == "" {
		handlerErr = errors.New("path is required")
		return nil, BackupPortfolioOutput{}, handlerErr
	}

	// Set defaults
	if input.Compression == "" {
		input.Compression = "best"
	}

	// Get elements directory from context or use default
	// For now, we'll use a default location
	elementsDir := "./data/elements"

	// Create backup service
	backupSvc := backup.NewBackupService(s.repo, elementsDir)

	// Create backup options
	options := backup.BackupOptions{
		Compression: input.Compression,
		Description: input.Description,
		Author:      input.Author,
	}

	// Execute backup
	metadata, err := backupSvc.Backup(input.Path, options)
	if err != nil {
		handlerErr = fmt.Errorf("backup failed: %w", err)
		return nil, BackupPortfolioOutput{}, handlerErr
	}

	// Create output
	output := BackupPortfolioOutput{
		Path:         input.Path,
		ElementCount: metadata.ElementCount,
		TotalSize:    metadata.TotalSize,
		Checksum:     metadata.Checksum,
		Version:      metadata.Version,
	}

	// Measure response size and record token metrics
	s.responseMiddleware.MeasureResponseSize(ctx, "backup_portfolio", output)

	return nil, output, nil
}

// handleRestorePortfolio handles the restore_portfolio tool.
func (s *MCPServer) handleRestorePortfolio(ctx context.Context, req *sdk.CallToolRequest, input RestorePortfolioInput) (*sdk.CallToolResult, RestorePortfolioOutput, error) {
	startTime := time.Now()
	var handlerErr error
	defer func() {
		s.metrics.RecordToolCall(application.ToolCallMetric{
			ToolName:  "restore_portfolio",
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

	// Validate required fields
	if input.Path == "" {
		handlerErr = errors.New("path is required")
		return nil, RestorePortfolioOutput{}, handlerErr
	}

	// Set defaults
	if input.MergeStrategy == "" {
		input.MergeStrategy = "skip"
	}

	// Get elements directory from context or use default
	elementsDir := "./data/elements"

	// Create backup and restore services
	backupSvc := backup.NewBackupService(s.repo, elementsDir)
	restoreSvc := backup.NewRestoreService(s.repo, backupSvc, elementsDir)

	// Create restore options
	options := backup.RestoreOptions{
		Overwrite:      input.Overwrite,
		SkipValidation: input.SkipValidation,
		MergeStrategy:  input.MergeStrategy,
		BackupBefore:   input.BackupBefore,
		DryRun:         input.DryRun,
	}

	// Execute restore
	result, err := restoreSvc.Restore(input.Path, options)
	if err != nil {
		handlerErr = fmt.Errorf("restore failed: %w", err)
		return nil, RestorePortfolioOutput{}, handlerErr
	}

	// Create output
	output := RestorePortfolioOutput{
		Success:         result.Success,
		ElementsAdded:   result.ElementsAdded,
		ElementsUpdated: result.ElementsUpdated,
		ElementsSkipped: result.ElementsSkipped,
		Errors:          result.Errors,
		BackupPath:      result.BackupPath,
		Duration:        result.Duration.String(),
	}

	// Measure response size and record token metrics
	s.responseMiddleware.MeasureResponseSize(ctx, "restore_portfolio", output)

	return nil, output, nil
}

// handleActivateElement handles the activate_element tool.
func (s *MCPServer) handleActivateElement(ctx context.Context, req *sdk.CallToolRequest, input ActivateElementInput) (*sdk.CallToolResult, ActivateElementOutput, error) {
	startTime := time.Now()
	var handlerErr error
	defer func() {
		s.metrics.RecordToolCall(application.ToolCallMetric{
			ToolName:  "activate_element",
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

	// Validate required fields
	if input.ID == "" {
		handlerErr = errors.New("id is required")
		return nil, ActivateElementOutput{}, handlerErr
	}

	// Get element from repository
	element, err := s.repo.GetByID(input.ID)
	if err != nil {
		handlerErr = fmt.Errorf("element not found: %w", err)
		return nil, ActivateElementOutput{}, handlerErr
	}

	// Activate element
	if err := element.Activate(); err != nil {
		handlerErr = fmt.Errorf("activation failed: %w", err)
		return nil, ActivateElementOutput{}, handlerErr
	}

	// Update in repository
	if err := s.repo.Update(element); err != nil {
		handlerErr = fmt.Errorf("failed to save element: %w", err)
		return nil, ActivateElementOutput{}, handlerErr
	}

	// Get updated metadata
	meta := element.GetMetadata()

	// Create output
	output := ActivateElementOutput{
		ID:        meta.ID,
		Name:      meta.Name,
		Type:      string(meta.Type),
		IsActive:  meta.IsActive,
		UpdatedAt: meta.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
	}

	// Measure response size and record token metrics
	s.responseMiddleware.MeasureResponseSize(ctx, "activate_element", output)

	return nil, output, nil
}

// handleDeactivateElement handles the deactivate_element tool.
func (s *MCPServer) handleDeactivateElement(ctx context.Context, req *sdk.CallToolRequest, input DeactivateElementInput) (*sdk.CallToolResult, DeactivateElementOutput, error) {
	startTime := time.Now()
	var handlerErr error
	defer func() {
		s.metrics.RecordToolCall(application.ToolCallMetric{
			ToolName:  "deactivate_element",
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

	// Validate required fields
	if input.ID == "" {
		handlerErr = errors.New("id is required")
		return nil, DeactivateElementOutput{}, handlerErr
	}

	// Get element from repository
	element, err := s.repo.GetByID(input.ID)
	if err != nil {
		handlerErr = fmt.Errorf("element not found: %w", err)
		return nil, DeactivateElementOutput{}, handlerErr
	}

	// Deactivate element
	if err := element.Deactivate(); err != nil {
		handlerErr = fmt.Errorf("deactivation failed: %w", err)
		return nil, DeactivateElementOutput{}, handlerErr
	}

	// Update in repository
	if err := s.repo.Update(element); err != nil {
		handlerErr = fmt.Errorf("failed to save element: %w", err)
		return nil, DeactivateElementOutput{}, handlerErr
	}

	// Get updated metadata
	meta := element.GetMetadata()

	// Create output
	output := DeactivateElementOutput{
		ID:        meta.ID,
		Name:      meta.Name,
		Type:      string(meta.Type),
		IsActive:  meta.IsActive,
		UpdatedAt: meta.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
	}

	// Measure response size and record token metrics
	s.responseMiddleware.MeasureResponseSize(ctx, "deactivate_element", output)

	return nil, output, nil
}
