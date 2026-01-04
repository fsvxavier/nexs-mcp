package mcp

import (
	"context"
	"fmt"
	"time"

	sdk "github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/fsvxavier/nexs-mcp/internal/application"
	"github.com/fsvxavier/nexs-mcp/internal/infrastructure"
	"github.com/fsvxavier/nexs-mcp/internal/logger"
)

// --- Check GitHub Auth Input/Output structures ---

// CheckGitHubAuthInput defines input for check_github_auth tool.
type CheckGitHubAuthInput struct {
	// No input parameters needed
}

// CheckGitHubAuthOutput defines output for check_github_auth tool.
type CheckGitHubAuthOutput struct {
	IsAuthenticated bool     `json:"is_authenticated"     jsonschema:"whether GitHub authentication is valid"`
	Username        string   `json:"username,omitempty"   jsonschema:"authenticated GitHub username"`
	ExpiresAt       string   `json:"expires_at,omitempty" jsonschema:"token expiration time (RFC3339)"`
	Scopes          []string `json:"scopes,omitempty"     jsonschema:"authorized scopes"`
	Message         string   `json:"message"              jsonschema:"status message"`
}

// handleCheckGitHubAuth handles the check_github_auth tool.
func (s *MCPServer) handleCheckGitHubAuth(ctx context.Context, req *sdk.CallToolRequest, input CheckGitHubAuthInput) (*sdk.CallToolResult, CheckGitHubAuthOutput, error) {
	startTime := time.Now()
	var handlerErr error
	defer func() {
		s.metrics.RecordToolCall(application.ToolCallMetric{
			ToolName:  "check_github_auth",
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

	// Get GitHub OAuth client
	oauthClient, err := s.getGitHubOAuthClient()
	if err != nil {
		logger.WarnContext(ctx, "GitHub OAuth client not available", "error", err)
		output := CheckGitHubAuthOutput{
			IsAuthenticated: false,
			Message:         "GitHub authentication not configured",
		}
		s.responseMiddleware.MeasureResponseSize(ctx, "check_github_auth", output)
		return nil, output, nil
	}

	// Check if token exists and is valid
	token, err := oauthClient.GetToken(ctx)
	if err != nil {
		logger.InfoContext(ctx, "No GitHub token found", "error", err)
		output := CheckGitHubAuthOutput{
			IsAuthenticated: false,
			Message:         "No GitHub token found. Please authenticate using init_github_auth tool.",
		}
		s.responseMiddleware.MeasureResponseSize(ctx, "check_github_auth", output)
		return nil, output, nil
	}

	// Check if token is expired
	if !token.Valid() {
		logger.WarnContext(ctx, "GitHub token expired", "expiry", token.Expiry)
		output := CheckGitHubAuthOutput{
			IsAuthenticated: false,
			ExpiresAt:       token.Expiry.Format(time.RFC3339),
			Message:         "GitHub token has expired. Use refresh_github_token to get a new token.",
		}
		s.responseMiddleware.MeasureResponseSize(ctx, "check_github_auth", output)
		return nil, output, nil
	}

	// Get user information using GitHubClient
	githubClient := infrastructure.NewGitHubClient(oauthClient)
	username, err := githubClient.GetUser(ctx)
	if err != nil {
		logger.ErrorContext(ctx, "Failed to get GitHub user", "error", err)
		output := CheckGitHubAuthOutput{
			IsAuthenticated: false,
			Message:         fmt.Sprintf("Failed to verify GitHub authentication: %v", err),
		}
		s.responseMiddleware.MeasureResponseSize(ctx, "check_github_auth", output)
		return nil, output, nil
	}

	logger.InfoContext(ctx, "GitHub authentication verified", "user", username)

	output := CheckGitHubAuthOutput{
		IsAuthenticated: true,
		Username:        username,
		ExpiresAt:       token.Expiry.Format(time.RFC3339),
		Scopes:          []string{"repo", "user"},
		Message:         fmt.Sprintf("Authenticated as GitHub user '%s'", username),
	}

	// Measure response size and record token metrics
	s.responseMiddleware.MeasureResponseSize(ctx, "check_github_auth", output)

	return nil, output, nil
}

// --- Refresh GitHub Token Input/Output structures ---

// RefreshGitHubTokenInput defines input for refresh_github_token tool.
type RefreshGitHubTokenInput struct {
	Force bool `json:"force,omitempty" jsonschema:"force token refresh even if not expired"`
}

// RefreshGitHubTokenOutput defines output for refresh_github_token tool.
type RefreshGitHubTokenOutput struct {
	Success   bool   `json:"success"              jsonschema:"whether token was refreshed successfully"`
	ExpiresAt string `json:"expires_at,omitempty" jsonschema:"new token expiration time (RFC3339)"`
	Message   string `json:"message"              jsonschema:"status message"`
}

// handleRefreshGitHubToken handles the refresh_github_token tool.
func (s *MCPServer) handleRefreshGitHubToken(ctx context.Context, req *sdk.CallToolRequest, input RefreshGitHubTokenInput) (*sdk.CallToolResult, RefreshGitHubTokenOutput, error) {
	startTime := time.Now()
	var handlerErr error
	defer func() {
		s.metrics.RecordToolCall(application.ToolCallMetric{
			ToolName:  "refresh_github_token",
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

	// Get GitHub OAuth client
	oauthClient, err := s.getGitHubOAuthClient()
	if err != nil {
		handlerErr = fmt.Errorf("GitHub OAuth client not available: %w", err)
		return nil, RefreshGitHubTokenOutput{}, handlerErr
	}

	// Get current token
	token, err := oauthClient.GetToken(ctx)
	if err != nil {
		handlerErr = fmt.Errorf("no GitHub token found: %w", err)
		return nil, RefreshGitHubTokenOutput{}, handlerErr
	}

	// Check if refresh is needed
	if !input.Force && token.Valid() {
		timeUntilExpiry := time.Until(token.Expiry)
		if timeUntilExpiry > 24*time.Hour {
			logger.InfoContext(ctx, "GitHub token still valid, no refresh needed",
				"expires_in", timeUntilExpiry.String())
			output := RefreshGitHubTokenOutput{
				Success:   false,
				ExpiresAt: token.Expiry.Format(time.RFC3339),
				Message:   fmt.Sprintf("Token is still valid for %s. Use force=true to refresh anyway.", timeUntilExpiry.Round(time.Minute)),
			}
			s.responseMiddleware.MeasureResponseSize(ctx, "refresh_github_token", output)
			return nil, output, nil
		}
	}

	// Force a token refresh by calling GetToken again (it will refresh if expired)
	newToken, err := oauthClient.GetToken(ctx)
	if err != nil {
		logger.ErrorContext(ctx, "Failed to refresh GitHub token", "error", err)
		handlerErr = fmt.Errorf("failed to refresh token: %w", err)
		return nil, RefreshGitHubTokenOutput{}, handlerErr
	}

	logger.InfoContext(ctx, "GitHub token refreshed successfully",
		"expires_at", newToken.Expiry)

	output := RefreshGitHubTokenOutput{
		Success:   true,
		ExpiresAt: newToken.Expiry.Format(time.RFC3339),
		Message:   "GitHub token refreshed successfully",
	}

	// Measure response size and record token metrics
	s.responseMiddleware.MeasureResponseSize(ctx, "refresh_github_token", output)

	return nil, output, nil
}

// --- Initialize GitHub Auth Input/Output structures ---

// InitGitHubAuthInput defines input for init_github_auth tool.
type InitGitHubAuthInput struct {
	// No input parameters needed
}

// InitGitHubAuthOutput defines output for init_github_auth tool.
type InitGitHubAuthOutput struct {
	UserCode        string `json:"user_code"        jsonschema:"code to enter on GitHub"`
	VerificationURI string `json:"verification_uri" jsonschema:"URL to visit for authentication"`
	ExpiresIn       int    `json:"expires_in"       jsonschema:"seconds until code expires"`
	Message         string `json:"message"          jsonschema:"instructions for user"`
}

// handleInitGitHubAuth handles the init_github_auth tool.
func (s *MCPServer) handleInitGitHubAuth(ctx context.Context, req *sdk.CallToolRequest, input InitGitHubAuthInput) (*sdk.CallToolResult, InitGitHubAuthOutput, error) {
	startTime := time.Now()
	var handlerErr error
	defer func() {
		s.metrics.RecordToolCall(application.ToolCallMetric{
			ToolName:  "init_github_auth",
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

	// Get GitHub OAuth client
	oauthClient, err := s.getGitHubOAuthClient()
	if err != nil {
		handlerErr = fmt.Errorf("GitHub OAuth client not available: %w", err)
		return nil, InitGitHubAuthOutput{}, handlerErr
	}

	// Initiate device flow
	deviceFlow, err := oauthClient.StartDeviceFlow(ctx)
	if err != nil {
		logger.ErrorContext(ctx, "Failed to initiate GitHub device flow", "error", err)
		handlerErr = fmt.Errorf("failed to initiate device flow: %w", err)
		return nil, InitGitHubAuthOutput{}, handlerErr
	}

	logger.InfoContext(ctx, "GitHub device flow initiated",
		"user_code", deviceFlow.UserCode,
		"verification_uri", deviceFlow.VerificationURI)

	// Store device code in session for polling
	s.mu.Lock()
	if s.deviceCodes == nil {
		s.deviceCodes = make(map[string]string)
	}
	s.deviceCodes[deviceFlow.UserCode] = deviceFlow.DeviceCode
	s.mu.Unlock()

	message := fmt.Sprintf("Please visit %s and enter code: %s (expires in %d seconds)",
		deviceFlow.VerificationURI, deviceFlow.UserCode, deviceFlow.ExpiresIn)

	output := InitGitHubAuthOutput{
		UserCode:        deviceFlow.UserCode,
		VerificationURI: deviceFlow.VerificationURI,
		ExpiresIn:       deviceFlow.ExpiresIn,
		Message:         message,
	}

	// Measure response size and record token metrics
	s.responseMiddleware.MeasureResponseSize(ctx, "init_github_auth", output)

	return nil, output, nil
}

// getGitHubOAuthClient returns the GitHub OAuth client.
func (s *MCPServer) getGitHubOAuthClient() (*infrastructure.GitHubOAuthClient, error) {
	// In a real implementation, this would be initialized during server startup
	// For now, we'll create it on-demand with a default token path
	//nolint:gosec // G101: This is a file path, not a hardcoded credential
	tokenPath := "data/github_token.json"
	client, err := infrastructure.NewGitHubOAuthClient(tokenPath)
	if err != nil {
		return nil, err
	}
	return client, nil
}
