package mcp

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	sdk "github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/fsvxavier/nexs-mcp/internal/application"
	"github.com/fsvxavier/nexs-mcp/internal/logger"
)

// UserSession manages user identity and context for the MCP session.
type UserSession struct {
	mu       sync.RWMutex
	username string
	metadata map[string]string
}

// globalUserSession is the global user session instance.
var globalUserSession = &UserSession{
	metadata: make(map[string]string),
}

// GetUserSession returns the global user session.
func GetUserSession() *UserSession {
	return globalUserSession
}

// SetUser sets the current user for the session.
func (us *UserSession) SetUser(username string, metadata map[string]string) {
	us.mu.Lock()
	defer us.mu.Unlock()

	us.username = username
	if metadata != nil {
		us.metadata = make(map[string]string)
		for k, v := range metadata {
			us.metadata[k] = v
		}
	}

	logger.Info("User context updated", "user", username)
}

// GetUser returns the current user and metadata.
func (us *UserSession) GetUser() (string, map[string]string) {
	us.mu.RLock()
	defer us.mu.RUnlock()

	// Create a copy of metadata to avoid race conditions
	metaCopy := make(map[string]string)
	for k, v := range us.metadata {
		metaCopy[k] = v
	}

	return us.username, metaCopy
}

// ClearUser clears the current user context.
func (us *UserSession) ClearUser() {
	us.mu.Lock()
	defer us.mu.Unlock()

	us.username = ""
	us.metadata = make(map[string]string)

	logger.Info("User context cleared")
}

// IsAuthenticated returns true if a user is set.
func (us *UserSession) IsAuthenticated() bool {
	us.mu.RLock()
	defer us.mu.RUnlock()

	return us.username != ""
}

// --- Get Current User Input/Output structures ---

// GetCurrentUserInput defines input for get_current_user tool.
type GetCurrentUserInput struct {
	// No input parameters needed
}

// GetCurrentUserOutput defines output for get_current_user tool.
type GetCurrentUserOutput struct {
	Username        string            `json:"username"           jsonschema:"current username (empty if not authenticated)"`
	IsAuthenticated bool              `json:"is_authenticated"   jsonschema:"whether a user is authenticated"`
	Metadata        map[string]string `json:"metadata,omitempty" jsonschema:"user metadata"`
	Message         string            `json:"message"            jsonschema:"status message"`
}

// handleGetCurrentUser handles the get_current_user tool.
func (s *MCPServer) handleGetCurrentUser(ctx context.Context, req *sdk.CallToolRequest, input GetCurrentUserInput) (*sdk.CallToolResult, GetCurrentUserOutput, error) {
	startTime := time.Now()
	var handlerErr error
	defer func() {
		s.metrics.RecordToolCall(application.ToolCallMetric{
			ToolName:  "get_current_user",
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

	session := GetUserSession()
	username, metadata := session.GetUser()
	isAuth := session.IsAuthenticated()

	var message string
	if isAuth {
		message = fmt.Sprintf("User '%s' is authenticated", username)
	} else {
		message = "No user authenticated"
	}

	logger.InfoContext(ctx, "Retrieved current user context",
		"user", username,
		"is_authenticated", isAuth)

	output := GetCurrentUserOutput{
		Username:        username,
		IsAuthenticated: isAuth,
		Metadata:        metadata,
		Message:         message,
	}

	// Measure response size and record token metrics
	s.responseMiddleware.MeasureResponseSize(ctx, "get_current_user", output)

	return nil, output, nil
}

// --- Set User Context Input/Output structures ---

// SetUserContextInput defines input for set_user_context tool.
type SetUserContextInput struct {
	Username string            `json:"username"           jsonschema:"username to set in context"`
	Metadata map[string]string `json:"metadata,omitempty" jsonschema:"user metadata (email, role, etc.)"`
}

// SetUserContextOutput defines output for set_user_context tool.
type SetUserContextOutput struct {
	Success  bool   `json:"success"  jsonschema:"whether user context was set successfully"`
	Username string `json:"username" jsonschema:"username that was set"`
	Message  string `json:"message"  jsonschema:"status message"`
}

// handleSetUserContext handles the set_user_context tool.
func (s *MCPServer) handleSetUserContext(ctx context.Context, req *sdk.CallToolRequest, input SetUserContextInput) (*sdk.CallToolResult, SetUserContextOutput, error) {
	startTime := time.Now()
	var handlerErr error
	defer func() {
		s.metrics.RecordToolCall(application.ToolCallMetric{
			ToolName:  "set_user_context",
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

	// Validate input
	if input.Username == "" {
		handlerErr = errors.New("username is required")
		return nil, SetUserContextOutput{}, handlerErr
	}

	// Set user session
	session := GetUserSession()
	session.SetUser(input.Username, input.Metadata)

	logger.InfoContext(ctx, "User context set",
		"user", input.Username,
		"metadata_count", len(input.Metadata))

	output := SetUserContextOutput{
		Success:  true,
		Username: input.Username,
		Message:  fmt.Sprintf("User context set to '%s'", input.Username),
	}

	// Measure response size and record token metrics
	s.responseMiddleware.MeasureResponseSize(ctx, "set_user_context", output)

	return nil, output, nil
}

// --- Clear User Context Input/Output structures ---

// ClearUserContextInput defines input for clear_user_context tool.
type ClearUserContextInput struct {
	Confirm bool `json:"confirm" jsonschema:"confirmation flag (must be true)"`
}

// ClearUserContextOutput defines output for clear_user_context tool.
type ClearUserContextOutput struct {
	Success bool   `json:"success" jsonschema:"whether user context was cleared"`
	Message string `json:"message" jsonschema:"status message"`
}

// handleClearUserContext handles the clear_user_context tool.
func (s *MCPServer) handleClearUserContext(ctx context.Context, req *sdk.CallToolRequest, input ClearUserContextInput) (*sdk.CallToolResult, ClearUserContextOutput, error) {
	startTime := time.Now()
	var handlerErr error
	defer func() {
		s.metrics.RecordToolCall(application.ToolCallMetric{
			ToolName:  "clear_user_context",
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

	// Require confirmation
	if !input.Confirm {
		handlerErr = errors.New("confirmation required: set confirm=true to proceed")
		return nil, ClearUserContextOutput{}, handlerErr
	}

	// Clear session
	session := GetUserSession()
	previousUser, _ := session.GetUser()
	session.ClearUser()

	logger.InfoContext(ctx, "User context cleared",
		"previous_user", previousUser)

	output := ClearUserContextOutput{
		Success: true,
		Message: "User context cleared successfully",
	}

	// Measure response size and record token metrics
	s.responseMiddleware.MeasureResponseSize(ctx, "clear_user_context", output)

	return nil, output, nil
}
