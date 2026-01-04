package mcp

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"time"

	sdk "github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/fsvxavier/nexs-mcp/internal/application"
	"github.com/fsvxavier/nexs-mcp/internal/common"
	"github.com/fsvxavier/nexs-mcp/internal/infrastructure"
	"github.com/fsvxavier/nexs-mcp/internal/portfolio"
)

// GitHub authentication state.
type authState struct {
	deviceCode      string
	userCode        string
	verificationURI string
	expiresAt       time.Time
	polling         bool
}

var currentAuthState *authState

// GitHubAuthStartInput represents input for starting GitHub authentication.
type GitHubAuthStartInput struct{}

// GitHubAuthStartOutput represents the output of starting GitHub authentication.
type GitHubAuthStartOutput struct {
	UserCode        string `json:"user_code"`
	VerificationURI string `json:"verification_uri"`
	ExpiresIn       int    `json:"expires_in"`
	Message         string `json:"message"`
}

// handleGitHubAuthStart initiates GitHub OAuth2 device flow.
func (s *MCPServer) handleGitHubAuthStart(ctx context.Context, req *sdk.CallToolRequest, input GitHubAuthStartInput) (*sdk.CallToolResult, GitHubAuthStartOutput, error) {
	startTime := time.Now()
	var handlerErr error
	defer func() {
		s.metrics.RecordToolCall(application.ToolCallMetric{
			ToolName:  "github_auth_start",
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

	// Initialize OAuth client
	homeDir, _ := os.UserHomeDir()
	tokenPath := filepath.Join(homeDir, ".nexs-mcp", "github_token.json")
	oauthClient, err := infrastructure.NewGitHubOAuthClient(tokenPath)
	if err != nil {
		handlerErr = fmt.Errorf("failed to initialize OAuth client: %w", err)
		return nil, GitHubAuthStartOutput{}, handlerErr
	}

	// Start device flow
	response, err := oauthClient.StartDeviceFlow(ctx)
	if err != nil {
		handlerErr = fmt.Errorf("failed to start device flow: %w", err)
		return nil, GitHubAuthStartOutput{}, handlerErr
	}

	// Store auth state for polling
	currentAuthState = &authState{
		deviceCode:      response.DeviceCode,
		userCode:        response.UserCode,
		verificationURI: response.VerificationURI,
		expiresAt:       time.Now().Add(time.Duration(response.ExpiresIn) * time.Second),
		polling:         false,
	}

	// Start background polling
	go func() {
		currentAuthState.polling = true
		token, err := oauthClient.PollForToken(ctx, response.DeviceCode, response.Interval)
		currentAuthState.polling = false

		if err == nil && token != nil {
			_ = oauthClient.SaveToken(token) // Best effort save
		}
	}()

	output := GitHubAuthStartOutput{
		UserCode:        response.UserCode,
		VerificationURI: response.VerificationURI,
		ExpiresIn:       response.ExpiresIn,
		Message: fmt.Sprintf("Visit %s and enter code: %s",
			response.VerificationURI, response.UserCode),
	}

	// Measure response size and record token metrics
	s.responseMiddleware.MeasureResponseSize(ctx, "github_auth_start", output)

	return nil, output, nil
}

// GitHubAuthStatusInput represents input for checking auth status.
type GitHubAuthStatusInput struct{}

// GitHubAuthStatusOutput represents the output of checking auth status.
type GitHubAuthStatusOutput struct {
	Status          string `json:"status"`
	Authenticated   bool   `json:"authenticated"`
	UserCode        string `json:"user_code,omitempty"`
	VerificationURI string `json:"verification_uri,omitempty"`
	ExpiresIn       int    `json:"expires_in,omitempty"`
	Message         string `json:"message"`
}

// handleGitHubAuthStatus checks the status of GitHub authentication.
func (s *MCPServer) handleGitHubAuthStatus(ctx context.Context, req *sdk.CallToolRequest, input GitHubAuthStatusInput) (*sdk.CallToolResult, GitHubAuthStatusOutput, error) {
	startTime := time.Now()
	var handlerErr error
	defer func() {
		s.metrics.RecordToolCall(application.ToolCallMetric{
			ToolName:  "github_auth_status",
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

	homeDir, _ := os.UserHomeDir()
	tokenPath := filepath.Join(homeDir, ".nexs-mcp", "github_token.json")
	oauthClient, err := infrastructure.NewGitHubOAuthClient(tokenPath)
	if err != nil {
		handlerErr = fmt.Errorf("failed to initialize OAuth client: %w", err)
		return nil, GitHubAuthStatusOutput{}, handlerErr
	}

	authenticated := oauthClient.IsAuthenticated(ctx)

	output := GitHubAuthStatusOutput{
		Authenticated: authenticated,
	}

	switch {
	case authenticated:
		output.Status = "authorized"
		output.Message = "GitHub authentication successful"
	case currentAuthState != nil && currentAuthState.polling:
		output.Status = "pending"
		output.UserCode = currentAuthState.userCode
		output.VerificationURI = currentAuthState.verificationURI
		output.ExpiresIn = int(time.Until(currentAuthState.expiresAt).Seconds())
		output.Message = "Waiting for user authorization"
	case currentAuthState != nil && time.Now().Before(currentAuthState.expiresAt):
		output.Status = "pending"
		output.UserCode = currentAuthState.userCode
		output.VerificationURI = currentAuthState.verificationURI
		output.ExpiresIn = int(time.Until(currentAuthState.expiresAt).Seconds())
		output.Message = "Authorization in progress"
	default:
		output.Status = "not_authenticated"
		output.Message = "Not authenticated. Use github_auth_start to begin authentication."
	}

	// Measure response size and record token metrics
	s.responseMiddleware.MeasureResponseSize(ctx, "github_auth_status", output)

	return nil, output, nil
}

// GitHubListReposInput represents input for listing repositories.
type GitHubListReposInput struct{}

// RepositoryInfo represents basic repository information.
type RepositoryInfo struct {
	Owner       string `json:"owner"`
	Name        string `json:"name"`
	FullName    string `json:"full_name"`
	Description string `json:"description"`
	Private     bool   `json:"private"`
	URL         string `json:"url"`
}

// GitHubListReposOutput represents the output of listing repositories.
type GitHubListReposOutput struct {
	Repositories []RepositoryInfo `json:"repositories"`
	Count        int              `json:"count"`
}

// handleGitHubListRepos lists all repositories for the authenticated user.
func (s *MCPServer) handleGitHubListRepos(ctx context.Context, req *sdk.CallToolRequest, input GitHubListReposInput) (*sdk.CallToolResult, GitHubListReposOutput, error) {
	startTime := time.Now()
	var handlerErr error
	defer func() {
		s.metrics.RecordToolCall(application.ToolCallMetric{
			ToolName:  "github_list_repos",
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

	// Initialize clients
	homeDir, _ := os.UserHomeDir()
	tokenPath := filepath.Join(homeDir, ".nexs-mcp", "github_token.json")
	oauthClient, err := infrastructure.NewGitHubOAuthClient(tokenPath)
	if err != nil {
		handlerErr = fmt.Errorf("failed to initialize OAuth client: %w", err)
		return nil, GitHubListReposOutput{}, handlerErr
	}
	githubClient := infrastructure.NewGitHubClient(oauthClient)

	// List repositories
	repos, err := githubClient.ListRepositories(ctx)
	if err != nil {
		handlerErr = fmt.Errorf("failed to list repositories: %w", err)
		return nil, GitHubListReposOutput{}, handlerErr
	}

	// Convert to output format
	repoInfos := make([]RepositoryInfo, len(repos))
	for i, repo := range repos {
		repoInfos[i] = RepositoryInfo{
			Owner:       repo.Owner,
			Name:        repo.Name,
			FullName:    repo.FullName,
			Description: repo.Description,
			Private:     repo.Private,
			URL:         repo.URL,
		}
	}

	output := GitHubListReposOutput{
		Repositories: repoInfos,
		Count:        len(repoInfos),
	}

	// Measure response size and record token metrics
	s.responseMiddleware.MeasureResponseSize(ctx, "github_list_repos", output)

	return nil, output, nil
}

// GitHubSyncPushInput represents input for pushing elements to GitHub.
type GitHubSyncPushInput struct {
	Repository         string `json:"repository"`
	Branch             string `json:"branch,omitempty"`
	ConflictResolution string `json:"conflict_resolution,omitempty"`
}

// GitHubSyncPushOutput represents the output of pushing to GitHub.
type GitHubSyncPushOutput struct {
	Pushed    int      `json:"pushed"`
	Conflicts int      `json:"conflicts"`
	Errors    []string `json:"errors,omitempty"`
	Message   string   `json:"message"`
}

// handleGitHubSyncPush pushes local elements to a GitHub repository.
func (s *MCPServer) handleGitHubSyncPush(ctx context.Context, req *sdk.CallToolRequest, input GitHubSyncPushInput) (*sdk.CallToolResult, GitHubSyncPushOutput, error) {
	startTime := time.Now()
	var handlerErr error
	defer func() {
		s.metrics.RecordToolCall(application.ToolCallMetric{
			ToolName:  "github_sync_push",
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

	if input.Repository == "" {
		handlerErr = errors.New("repository is required")
		return nil, GitHubSyncPushOutput{}, handlerErr
	}

	// Parse repository URL
	owner, repo, err := infrastructure.ParseRepoURL(input.Repository)
	if err != nil {
		handlerErr = fmt.Errorf("invalid repository format: %w", err)
		return nil, GitHubSyncPushOutput{}, handlerErr
	}

	// Default branch
	branch := input.Branch
	if branch == "" {
		branch = common.BranchMain
	}

	// Default conflict resolution
	conflictRes := portfolio.ConflictResolution(input.ConflictResolution)
	if conflictRes == "" {
		conflictRes = portfolio.ConflictNewerWins
	}

	// Initialize clients
	homeDir, _ := os.UserHomeDir()
	tokenPath := filepath.Join(homeDir, ".nexs-mcp", "github_token.json")
	baseDir := filepath.Join(homeDir, ".nexs-mcp", "elements")

	oauthClient, err := infrastructure.NewGitHubOAuthClient(tokenPath)
	if err != nil {
		handlerErr = fmt.Errorf("failed to initialize OAuth client: %w", err)
		return nil, GitHubSyncPushOutput{}, handlerErr
	}
	githubClient := infrastructure.NewGitHubClient(oauthClient)

	// Get enhanced repository from server
	enhancedRepo, ok := s.repo.(*infrastructure.EnhancedFileElementRepository)
	if !ok {
		handlerErr = errors.New("enhanced repository required for GitHub sync")
		return nil, GitHubSyncPushOutput{}, handlerErr
	}

	mapper := portfolio.NewGitHubMapper(baseDir)
	sync := portfolio.NewGitHubSync(githubClient, enhancedRepo, mapper, conflictRes)

	// Push to GitHub
	result, err := sync.Push(ctx, owner, repo, branch)
	if err != nil {
		handlerErr = fmt.Errorf("push failed: %w", err)
		return nil, GitHubSyncPushOutput{}, handlerErr
	}

	output := GitHubSyncPushOutput{
		Pushed:    result.Pushed,
		Conflicts: len(result.Conflicts),
		Errors:    result.Errors,
		Message:   fmt.Sprintf("Pushed %d elements to %s", result.Pushed, input.Repository),
	}

	// Measure response size and record token metrics
	s.responseMiddleware.MeasureResponseSize(ctx, "github_sync_push", output)

	return nil, output, nil
}

// GitHubSyncPullInput represents input for pulling elements from GitHub.
type GitHubSyncPullInput struct {
	Repository         string `json:"repository"`
	Branch             string `json:"branch,omitempty"`
	ConflictResolution string `json:"conflict_resolution,omitempty"`
}

// GitHubSyncPullOutput represents the output of pulling from GitHub.
type GitHubSyncPullOutput struct {
	Pulled    int      `json:"pulled"`
	Conflicts int      `json:"conflicts"`
	Errors    []string `json:"errors,omitempty"`
	Message   string   `json:"message"`
}

// handleGitHubSyncPull pulls elements from a GitHub repository.
func (s *MCPServer) handleGitHubSyncPull(ctx context.Context, req *sdk.CallToolRequest, input GitHubSyncPullInput) (*sdk.CallToolResult, GitHubSyncPullOutput, error) {
	startTime := time.Now()
	var handlerErr error
	defer func() {
		s.metrics.RecordToolCall(application.ToolCallMetric{
			ToolName:  "github_sync_pull",
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

	if input.Repository == "" {
		handlerErr = errors.New("repository is required")
		return nil, GitHubSyncPullOutput{}, handlerErr
	}

	// Parse repository URL
	owner, repo, err := infrastructure.ParseRepoURL(input.Repository)
	if err != nil {
		handlerErr = fmt.Errorf("invalid repository format: %w", err)
		return nil, GitHubSyncPullOutput{}, handlerErr
	}

	// Default branch
	branch := input.Branch
	if branch == "" {
		branch = common.BranchMain
	}

	// Default conflict resolution
	conflictRes := portfolio.ConflictResolution(input.ConflictResolution)
	if conflictRes == "" {
		conflictRes = portfolio.ConflictNewerWins
	}

	// Initialize clients
	homeDir, _ := os.UserHomeDir()
	tokenPath := filepath.Join(homeDir, ".nexs-mcp", "github_token.json")
	baseDir := filepath.Join(homeDir, ".nexs-mcp", "elements")

	oauthClient, err := infrastructure.NewGitHubOAuthClient(tokenPath)
	if err != nil {
		handlerErr = fmt.Errorf("failed to initialize OAuth client: %w", err)
		return nil, GitHubSyncPullOutput{}, handlerErr
	}
	githubClient := infrastructure.NewGitHubClient(oauthClient)

	// Get enhanced repository from server
	enhancedRepo, ok := s.repo.(*infrastructure.EnhancedFileElementRepository)
	if !ok {
		handlerErr = errors.New("enhanced repository required for GitHub sync")
		return nil, GitHubSyncPullOutput{}, handlerErr
	}

	mapper := portfolio.NewGitHubMapper(baseDir)
	sync := portfolio.NewGitHubSync(githubClient, enhancedRepo, mapper, conflictRes)

	// Pull from GitHub
	result, err := sync.Pull(ctx, owner, repo, branch)
	if err != nil {
		handlerErr = fmt.Errorf("pull failed: %w", err)
		return nil, GitHubSyncPullOutput{}, handlerErr
	}

	output := GitHubSyncPullOutput{
		Pulled:    result.Pulled,
		Conflicts: len(result.Conflicts),
		Errors:    result.Errors,
		Message:   fmt.Sprintf("Pulled %d elements from %s", result.Pulled, input.Repository),
	}

	// Measure response size and record token metrics
	s.responseMiddleware.MeasureResponseSize(ctx, "github_sync_pull", output)

	return nil, output, nil
}

// GitHubSyncBidirectionalInput represents input for bidirectional sync.
type GitHubSyncBidirectionalInput struct {
	Repository         string `json:"repository"`
	Branch             string `json:"branch,omitempty"`
	ConflictResolution string `json:"conflict_resolution,omitempty"`
}

// GitHubSyncBidirectionalOutput represents the output of bidirectional sync.
type GitHubSyncBidirectionalOutput struct {
	Pushed    int      `json:"pushed"`
	Pulled    int      `json:"pulled"`
	Conflicts int      `json:"conflicts"`
	Errors    []string `json:"errors,omitempty"`
	Message   string   `json:"message"`
}

// handleGitHubSyncBidirectional performs a full bidirectional sync (pull then push).
func (s *MCPServer) handleGitHubSyncBidirectional(ctx context.Context, req *sdk.CallToolRequest, input GitHubSyncBidirectionalInput) (*sdk.CallToolResult, GitHubSyncBidirectionalOutput, error) {
	startTime := time.Now()
	var handlerErr error
	defer func() {
		s.metrics.RecordToolCall(application.ToolCallMetric{
			ToolName:  "github_sync_bidirectional",
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

	// Parse repository (owner/repo format)
	owner, repo, err := infrastructure.ParseRepoURL(input.Repository)
	if err != nil {
		handlerErr = fmt.Errorf("invalid repository format: %w", err)
		return nil, GitHubSyncBidirectionalOutput{}, handlerErr
	}

	// Default branch
	branch := input.Branch
	if branch == "" {
		branch = common.BranchMain
	}

	// Default conflict resolution
	conflictRes := portfolio.ConflictResolution(input.ConflictResolution)
	if conflictRes == "" {
		conflictRes = portfolio.ConflictNewerWins
	}

	// Initialize clients
	homeDir, _ := os.UserHomeDir()
	tokenPath := filepath.Join(homeDir, ".nexs-mcp", "github_token.json")
	baseDir := filepath.Join(homeDir, ".nexs-mcp", "elements")

	oauthClient, err := infrastructure.NewGitHubOAuthClient(tokenPath)
	if err != nil {
		handlerErr = fmt.Errorf("failed to initialize OAuth client: %w", err)
		return nil, GitHubSyncBidirectionalOutput{}, handlerErr
	}
	githubClient := infrastructure.NewGitHubClient(oauthClient)

	// Get enhanced repository from server
	enhancedRepo, ok := s.repo.(*infrastructure.EnhancedFileElementRepository)
	if !ok {
		handlerErr = errors.New("enhanced repository required for GitHub sync")
		return nil, GitHubSyncBidirectionalOutput{}, handlerErr
	}

	mapper := portfolio.NewGitHubMapper(baseDir)
	sync := portfolio.NewGitHubSync(githubClient, enhancedRepo, mapper, conflictRes)

	// Perform bidirectional sync
	result, err := sync.SyncBidirectional(ctx, owner, repo, branch)
	if err != nil {
		handlerErr = fmt.Errorf("bidirectional sync failed: %w", err)
		return nil, GitHubSyncBidirectionalOutput{}, handlerErr
	}

	output := GitHubSyncBidirectionalOutput{
		Pushed:    result.Pushed,
		Pulled:    result.Pulled,
		Conflicts: len(result.Conflicts),
		Errors:    result.Errors,
		Message:   fmt.Sprintf("Synced with %s: pulled %d, pushed %d elements", input.Repository, result.Pulled, result.Pushed),
	}

	// Measure response size and record token metrics
	s.responseMiddleware.MeasureResponseSize(ctx, "github_sync_bidirectional", output)

	return nil, output, nil
}
