package mcp

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestGitHubAuthStartOutput_Structure(t *testing.T) {
	output := GitHubAuthStartOutput{
		UserCode:        "ABCD-1234",
		VerificationURI: "https://github.com/login/device",
		ExpiresIn:       900,
		Message:         "Visit https://github.com/login/device and enter code: ABCD-1234",
	}

	assert.Equal(t, "ABCD-1234", output.UserCode)
	assert.Equal(t, "https://github.com/login/device", output.VerificationURI)
	assert.Equal(t, 900, output.ExpiresIn)
	assert.Contains(t, output.Message, "ABCD-1234")
}

func TestGitHubAuthStatusOutput_Structure(t *testing.T) {
	output := GitHubAuthStatusOutput{
		Status:          "authorized",
		Authenticated:   true,
		UserCode:        "",
		VerificationURI: "",
		ExpiresIn:       0,
		Message:         "GitHub authentication successful",
	}

	assert.Equal(t, "authorized", output.Status)
	assert.True(t, output.Authenticated)
	assert.Equal(t, "GitHub authentication successful", output.Message)
}

func TestGitHubAuthStatusOutput_PendingStatus(t *testing.T) {
	output := GitHubAuthStatusOutput{
		Status:          "pending",
		Authenticated:   false,
		UserCode:        "TEST-CODE",
		VerificationURI: "https://github.com/login/device",
		ExpiresIn:       300,
		Message:         "Waiting for user authorization",
	}

	assert.Equal(t, "pending", output.Status)
	assert.False(t, output.Authenticated)
	assert.Equal(t, "TEST-CODE", output.UserCode)
	assert.Greater(t, output.ExpiresIn, 0)
}

func TestRepositoryInfo_Structure(t *testing.T) {
	repoInfo := RepositoryInfo{
		Owner:       "fsvxavier",
		Name:        "nexs-mcp",
		FullName:    "fsvxavier/nexs-mcp",
		Description: "NEXS MCP Server",
		Private:     false,
		URL:         "https://github.com/fsvxavier/nexs-mcp",
	}

	assert.Equal(t, "fsvxavier", repoInfo.Owner)
	assert.Equal(t, "nexs-mcp", repoInfo.Name)
	assert.Equal(t, "fsvxavier/nexs-mcp", repoInfo.FullName)
	assert.False(t, repoInfo.Private)
}

func TestGitHubListReposOutput_Structure(t *testing.T) {
	repos := []RepositoryInfo{
		{
			Owner:    "user1",
			Name:     "repo1",
			FullName: "user1/repo1",
			Private:  false,
		},
		{
			Owner:    "user1",
			Name:     "repo2",
			FullName: "user1/repo2",
			Private:  true,
		},
	}

	output := GitHubListReposOutput{
		Repositories: repos,
		Count:        len(repos),
	}

	assert.Equal(t, 2, output.Count)
	assert.Len(t, output.Repositories, 2)
	assert.Equal(t, "repo1", output.Repositories[0].Name)
	assert.Equal(t, "repo2", output.Repositories[1].Name)
}

func TestGitHubSyncPushInput_Structure(t *testing.T) {
	input := GitHubSyncPushInput{
		Repository:         "owner/repo",
		Branch:             "main",
		ConflictResolution: "local_wins",
	}

	assert.Equal(t, "owner/repo", input.Repository)
	assert.Equal(t, "main", input.Branch)
	assert.Equal(t, "local_wins", input.ConflictResolution)
}

func TestGitHubSyncPushOutput_Structure(t *testing.T) {
	output := GitHubSyncPushOutput{
		Pushed:    5,
		Conflicts: 2,
		Errors:    []string{"error1", "error2"},
		Message:   "Pushed 5 elements to owner/repo",
	}

	assert.Equal(t, 5, output.Pushed)
	assert.Equal(t, 2, output.Conflicts)
	assert.Len(t, output.Errors, 2)
	assert.Contains(t, output.Message, "5 elements")
}

func TestGitHubSyncPullInput_Structure(t *testing.T) {
	input := GitHubSyncPullInput{
		Repository:         "fsvxavier/nexs-mcp",
		Branch:             "develop",
		ConflictResolution: "newer_wins",
	}

	assert.Equal(t, "fsvxavier/nexs-mcp", input.Repository)
	assert.Equal(t, "develop", input.Branch)
	assert.Equal(t, "newer_wins", input.ConflictResolution)
}

func TestGitHubSyncPullOutput_Structure(t *testing.T) {
	output := GitHubSyncPullOutput{
		Pulled:    10,
		Conflicts: 1,
		Errors:    []string{},
		Message:   "Pulled 10 elements from owner/repo",
	}

	assert.Equal(t, 10, output.Pulled)
	assert.Equal(t, 1, output.Conflicts)
	assert.Empty(t, output.Errors)
	assert.Contains(t, output.Message, "10 elements")
}

func TestAuthState_Structure(t *testing.T) {
	now := time.Now()
	state := &authState{
		deviceCode:      "device-code-123",
		userCode:        "USER-CODE",
		verificationURI: "https://github.com/login/device",
		expiresAt:       now.Add(15 * time.Minute),
		polling:         true,
	}

	assert.Equal(t, "device-code-123", state.deviceCode)
	assert.Equal(t, "USER-CODE", state.userCode)
	assert.Equal(t, "https://github.com/login/device", state.verificationURI)
	assert.True(t, state.polling)
	assert.True(t, state.expiresAt.After(now))
}

func TestAuthState_ExpirationCheck(t *testing.T) {
	pastTime := time.Now().Add(-10 * time.Minute)
	state := &authState{
		deviceCode:      "expired-code",
		userCode:        "EXPIRED",
		verificationURI: "https://github.com/login/device",
		expiresAt:       pastTime,
		polling:         false,
	}

	// Check if expired
	assert.True(t, time.Now().After(state.expiresAt), "Auth state should be expired")
	assert.False(t, state.polling, "Should not be polling for expired state")
}

func TestGitHubSyncInput_DefaultBranch(t *testing.T) {
	// Test that empty branch can be detected and defaulted to "main"
	input := GitHubSyncPushInput{
		Repository: "owner/repo",
		Branch:     "",
	}

	branch := input.Branch
	if branch == "" {
		branch = "main"
	}

	assert.Equal(t, "main", branch)
}

func TestGitHubSyncInput_DefaultConflictResolution(t *testing.T) {
	// Test that empty conflict resolution can be detected and defaulted
	input := GitHubSyncPullInput{
		Repository:         "owner/repo",
		ConflictResolution: "",
	}

	conflictRes := input.ConflictResolution
	if conflictRes == "" {
		conflictRes = "newer_wins"
	}

	assert.Equal(t, "newer_wins", conflictRes)
}

// Note: Full integration tests for MCP handlers would require mocking the SDK
// and GitHub clients. These tests focus on data structure validation and basic
// logic that can be tested without external dependencies.
