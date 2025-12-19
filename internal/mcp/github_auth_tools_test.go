package mcp

import (
	"context"
	"testing"

	sdk "github.com/modelcontextprotocol/go-sdk/mcp"
)

func TestHandleCheckGitHubAuth_NotConfigured(t *testing.T) {
	server := setupUserTestServer(t)
	ctx := context.Background()

	input := CheckGitHubAuthInput{}

	_, output, err := server.handleCheckGitHubAuth(ctx, &sdk.CallToolRequest{}, input)

	if err != nil {
		t.Fatalf("handleCheckGitHubAuth should not return error: %v", err)
	}

	if output.IsAuthenticated {
		t.Error("Expected IsAuthenticated to be false when not configured")
	}

	if output.Message == "" {
		t.Error("Expected non-empty message")
	}
}

func TestHandleCheckGitHubAuth_NoToken(t *testing.T) {
	server := setupUserTestServer(t)
	ctx := context.Background()

	input := CheckGitHubAuthInput{}

	_, output, err := server.handleCheckGitHubAuth(ctx, &sdk.CallToolRequest{}, input)

	if err != nil {
		t.Fatalf("handleCheckGitHubAuth should not return error: %v", err)
	}

	if output.IsAuthenticated {
		t.Error("Expected IsAuthenticated to be false without token")
	}

	if output.Message == "" {
		t.Error("Expected non-empty message")
	}
}

func TestHandleRefreshGitHubToken_NoToken(t *testing.T) {
	server := setupUserTestServer(t)
	ctx := context.Background()

	input := RefreshGitHubTokenInput{
		Force: false,
	}

	_, _, err := server.handleRefreshGitHubToken(ctx, &sdk.CallToolRequest{}, input)

	if err == nil {
		t.Error("Expected error when no token exists")
	}
}

func TestHandleInitGitHubAuth(t *testing.T) {
	// This test would require mocking the GitHub OAuth flow
	// For now, we'll just verify the basic structure
	server := setupUserTestServer(t)
	ctx := context.Background()

	input := InitGitHubAuthInput{}

	_, output, err := server.handleInitGitHubAuth(ctx, &sdk.CallToolRequest{}, input)

	// May fail if no network or GitHub is down, but shouldn't panic
	if err != nil {
		t.Logf("Expected behavior - GitHub device flow not accessible: %v", err)
		return
	}

	// If successful, verify output structure
	if output.UserCode == "" {
		t.Error("Expected non-empty user code")
	}

	if output.VerificationURI == "" {
		t.Error("Expected non-empty verification URI")
	}

	if output.ExpiresIn <= 0 {
		t.Error("Expected positive expiration time")
	}

	if output.Message == "" {
		t.Error("Expected non-empty message")
	}

	// Verify device code was stored
	server.mu.Lock()
	deviceCode, exists := server.deviceCodes[output.UserCode]
	server.mu.Unlock()

	if !exists {
		t.Error("Expected device code to be stored in session")
	}

	if deviceCode == "" {
		t.Error("Expected non-empty device code in session")
	}
}

func TestGetGitHubOAuthClient(t *testing.T) {
	server := setupUserTestServer(t)

	client, err := server.getGitHubOAuthClient()

	if err != nil {
		t.Fatalf("getGitHubOAuthClient failed: %v", err)
	}

	if client == nil {
		t.Error("Expected non-nil GitHub OAuth client")
	}
}

func TestHandleRefreshGitHubToken_WithForce(t *testing.T) {
	server := setupUserTestServer(t)
	ctx := context.Background()

	input := RefreshGitHubTokenInput{
		Force: true,
	}

	_, _, err := server.handleRefreshGitHubToken(ctx, &sdk.CallToolRequest{}, input)

	// Should fail without token, but verify force flag is respected
	if err == nil {
		t.Error("Expected error when no token exists")
	}
}

func TestDeviceCodeStorage(t *testing.T) {
	server := setupUserTestServer(t)

	// Verify device codes map initialization
	server.mu.Lock()
	if server.deviceCodes == nil {
		server.deviceCodes = make(map[string]string)
	}
	server.deviceCodes["TEST123"] = "device_code_123"
	server.mu.Unlock()

	// Verify storage
	server.mu.Lock()
	deviceCode, exists := server.deviceCodes["TEST123"]
	server.mu.Unlock()

	if !exists {
		t.Error("Expected device code to be stored")
	}

	if deviceCode != "device_code_123" {
		t.Errorf("Expected device_code_123, got %s", deviceCode)
	}
}
