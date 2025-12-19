package mcp

import (
	"context"
	"testing"

	sdk "github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/fsvxavier/nexs-mcp/internal/infrastructure"
)

func setupUserTestServer(t *testing.T) *MCPServer {
	t.Helper()
	repo := infrastructure.NewInMemoryElementRepository()

	// Clear any existing user session
	GetUserSession().ClearUser()

	return NewMCPServer("nexs-mcp-test", "0.1.0", repo)
}

func TestHandleGetCurrentUser_NoUser(t *testing.T) {
	server := setupUserTestServer(t)
	ctx := context.Background()

	input := GetCurrentUserInput{}

	_, output, err := server.handleGetCurrentUser(ctx, &sdk.CallToolRequest{}, input)

	if err != nil {
		t.Fatalf("handleGetCurrentUser failed: %v", err)
	}

	if output.IsAuthenticated {
		t.Error("Expected IsAuthenticated to be false")
	}

	if output.Username != "" {
		t.Errorf("Expected empty username, got %s", output.Username)
	}

	if output.Message == "" {
		t.Error("Expected non-empty message")
	}
}

func TestHandleSetUserContext(t *testing.T) {
	server := setupUserTestServer(t)
	ctx := context.Background()

	input := SetUserContextInput{
		Username: "alice",
		Metadata: map[string]string{
			"email": "alice@example.com",
			"role":  "admin",
		},
	}

	_, output, err := server.handleSetUserContext(ctx, &sdk.CallToolRequest{}, input)

	if err != nil {
		t.Fatalf("handleSetUserContext failed: %v", err)
	}

	if !output.Success {
		t.Error("Expected Success to be true")
	}

	if output.Username != "alice" {
		t.Errorf("Expected username alice, got %s", output.Username)
	}

	// Verify user was set in session
	session := GetUserSession()
	username, metadata := session.GetUser()

	if username != "alice" {
		t.Errorf("Expected session username alice, got %s", username)
	}

	if metadata["email"] != "alice@example.com" {
		t.Errorf("Expected email alice@example.com, got %s", metadata["email"])
	}

	if metadata["role"] != "admin" {
		t.Errorf("Expected role admin, got %s", metadata["role"])
	}
}

func TestHandleSetUserContext_EmptyUsername(t *testing.T) {
	server := setupUserTestServer(t)
	ctx := context.Background()

	input := SetUserContextInput{
		Username: "",
	}

	_, _, err := server.handleSetUserContext(ctx, &sdk.CallToolRequest{}, input)

	if err == nil {
		t.Error("Expected error for empty username")
	}

	if err.Error() != "username is required" {
		t.Errorf("Expected 'username is required' error, got: %v", err)
	}
}

func TestHandleGetCurrentUser_WithUser(t *testing.T) {
	server := setupUserTestServer(t)
	ctx := context.Background()

	// Set user first
	setInput := SetUserContextInput{
		Username: "bob",
		Metadata: map[string]string{
			"email": "bob@example.com",
		},
	}
	server.handleSetUserContext(ctx, &sdk.CallToolRequest{}, setInput)

	// Get user
	getInput := GetCurrentUserInput{}
	_, output, err := server.handleGetCurrentUser(ctx, &sdk.CallToolRequest{}, getInput)

	if err != nil {
		t.Fatalf("handleGetCurrentUser failed: %v", err)
	}

	if !output.IsAuthenticated {
		t.Error("Expected IsAuthenticated to be true")
	}

	if output.Username != "bob" {
		t.Errorf("Expected username bob, got %s", output.Username)
	}

	if output.Metadata["email"] != "bob@example.com" {
		t.Errorf("Expected email bob@example.com, got %s", output.Metadata["email"])
	}
}

func TestHandleClearUserContext(t *testing.T) {
	server := setupUserTestServer(t)
	ctx := context.Background()

	// Set user first
	setInput := SetUserContextInput{
		Username: "charlie",
		Metadata: map[string]string{
			"email": "charlie@example.com",
		},
	}
	server.handleSetUserContext(ctx, &sdk.CallToolRequest{}, setInput)

	// Clear user
	clearInput := ClearUserContextInput{
		Confirm: true,
	}
	_, output, err := server.handleClearUserContext(ctx, &sdk.CallToolRequest{}, clearInput)

	if err != nil {
		t.Fatalf("handleClearUserContext failed: %v", err)
	}

	if !output.Success {
		t.Error("Expected Success to be true")
	}

	// Verify user was cleared
	session := GetUserSession()
	if session.IsAuthenticated() {
		t.Error("Expected session to be unauthenticated after clear")
	}

	username, _ := session.GetUser()
	if username != "" {
		t.Errorf("Expected empty username after clear, got %s", username)
	}
}

func TestHandleClearUserContext_NoConfirmation(t *testing.T) {
	server := setupUserTestServer(t)
	ctx := context.Background()

	// Set user first
	setInput := SetUserContextInput{
		Username: "dave",
	}
	server.handleSetUserContext(ctx, &sdk.CallToolRequest{}, setInput)

	// Try to clear without confirmation
	clearInput := ClearUserContextInput{
		Confirm: false,
	}
	_, _, err := server.handleClearUserContext(ctx, &sdk.CallToolRequest{}, clearInput)

	if err == nil {
		t.Error("Expected error when confirm is false")
	}

	if err.Error() != "confirmation required: set confirm=true to proceed" {
		t.Errorf("Expected confirmation error, got: %v", err)
	}

	// Verify user was NOT cleared
	session := GetUserSession()
	if !session.IsAuthenticated() {
		t.Error("Expected session to still be authenticated")
	}
}

func TestUserSession_Concurrent(t *testing.T) {
	session := GetUserSession()
	session.ClearUser()

	// Test concurrent access
	done := make(chan bool, 2)

	go func() {
		for i := 0; i < 100; i++ {
			session.SetUser("user1", map[string]string{"id": "1"})
			username, _ := session.GetUser()
			if username != "user1" && username != "user2" {
				t.Errorf("Unexpected username: %s", username)
			}
		}
		done <- true
	}()

	go func() {
		for i := 0; i < 100; i++ {
			session.SetUser("user2", map[string]string{"id": "2"})
			username, _ := session.GetUser()
			if username != "user1" && username != "user2" {
				t.Errorf("Unexpected username: %s", username)
			}
		}
		done <- true
	}()

	<-done
	<-done
}

func TestUserSession_IsAuthenticated(t *testing.T) {
	session := GetUserSession()
	session.ClearUser()

	if session.IsAuthenticated() {
		t.Error("Expected IsAuthenticated to be false initially")
	}

	session.SetUser("testuser", nil)

	if !session.IsAuthenticated() {
		t.Error("Expected IsAuthenticated to be true after SetUser")
	}

	session.ClearUser()

	if session.IsAuthenticated() {
		t.Error("Expected IsAuthenticated to be false after ClearUser")
	}
}

func TestHandleSetUserContext_WithoutMetadata(t *testing.T) {
	server := setupUserTestServer(t)
	ctx := context.Background()

	input := SetUserContextInput{
		Username: "eve",
		Metadata: nil,
	}

	_, output, err := server.handleSetUserContext(ctx, &sdk.CallToolRequest{}, input)

	if err != nil {
		t.Fatalf("handleSetUserContext failed: %v", err)
	}

	if !output.Success {
		t.Error("Expected Success to be true")
	}

	// Verify metadata is empty but not nil
	session := GetUserSession()
	_, metadata := session.GetUser()

	if metadata == nil {
		t.Error("Expected metadata to be non-nil (empty map)")
	}

	if len(metadata) != 0 {
		t.Errorf("Expected empty metadata, got %d entries", len(metadata))
	}
}
