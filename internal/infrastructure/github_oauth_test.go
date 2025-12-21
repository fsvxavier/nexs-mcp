package infrastructure

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/oauth2"
)

func TestGitHubOAuthClient_TokenPersistence(t *testing.T) {
	// Create temporary token file
	tmpDir := t.TempDir()
	_ = tmpDir // Using t.TempDir() for automatic cleanup

	tokenPath := filepath.Join(tmpDir, "test_token.json")
	client, err := NewGitHubOAuthClient(tokenPath)
	require.NoError(t, err)

	// Create a test token
	testToken := &oauth2.Token{
		AccessToken:  "test-access-token",
		TokenType:    "Bearer",
		RefreshToken: "test-refresh-token",
		Expiry:       time.Now().Add(1 * time.Hour),
	}

	// Save token
	err = client.SaveToken(testToken)
	assert.NoError(t, err)

	// Verify file was created
	_, err = os.Stat(tokenPath)
	assert.NoError(t, err)

	// Load token
	loadedToken, err := client.LoadToken()
	assert.NoError(t, err)
	assert.NotNil(t, loadedToken)
	assert.Equal(t, testToken.AccessToken, loadedToken.AccessToken)
	assert.Equal(t, testToken.TokenType, loadedToken.TokenType)
	assert.Equal(t, testToken.RefreshToken, loadedToken.RefreshToken)
}

func TestGitHubOAuthClient_LoadToken_FileNotExists(t *testing.T) {
	tmpDir := t.TempDir()
	_ = tmpDir // Using t.TempDir() for automatic cleanup

	tokenPath := filepath.Join(tmpDir, "nonexistent_token.json")
	client, err := NewGitHubOAuthClient(tokenPath)
	require.NoError(t, err)

	// Load token from non-existent file
	token, err := client.LoadToken()
	assert.Error(t, err)
	assert.Nil(t, token)
}

func TestGitHubOAuthClient_IsAuthenticated_WithValidToken(t *testing.T) {
	tmpDir := t.TempDir()
	_ = tmpDir // Using t.TempDir() for automatic cleanup

	tokenPath := filepath.Join(tmpDir, "test_token.json")
	client, err := NewGitHubOAuthClient(tokenPath)
	require.NoError(t, err)

	// Save a valid token
	testToken := &oauth2.Token{
		AccessToken: "valid-token",
		TokenType:   "Bearer",
		Expiry:      time.Now().Add(1 * time.Hour),
	}
	err = client.SaveToken(testToken)
	require.NoError(t, err)

	// Check authentication
	ctx := context.Background()
	authenticated := client.IsAuthenticated(ctx)
	assert.True(t, authenticated)
}

func TestGitHubOAuthClient_IsAuthenticated_WithExpiredToken(t *testing.T) {
	tmpDir := t.TempDir()
	_ = tmpDir // Using t.TempDir() for automatic cleanup

	tokenPath := filepath.Join(tmpDir, "test_token.json")
	client, err := NewGitHubOAuthClient(tokenPath)
	require.NoError(t, err)

	// Save an expired token
	expiredToken := &oauth2.Token{
		AccessToken: "expired-token",
		TokenType:   "Bearer",
		Expiry:      time.Now().Add(-1 * time.Hour),
	}
	err = client.SaveToken(expiredToken)
	require.NoError(t, err)

	// Check authentication (should be false for expired token)
	ctx := context.Background()
	authenticated := client.IsAuthenticated(ctx)
	assert.False(t, authenticated)
}

func TestGitHubOAuthClient_IsAuthenticated_NoToken(t *testing.T) {
	tmpDir := t.TempDir()
	_ = tmpDir // Using t.TempDir() for automatic cleanup

	tokenPath := filepath.Join(tmpDir, "nonexistent_token.json")
	client, err := NewGitHubOAuthClient(tokenPath)
	require.NoError(t, err)

	// Check authentication without token
	ctx := context.Background()
	authenticated := client.IsAuthenticated(ctx)
	assert.False(t, authenticated)
}

func TestGitHubOAuthClient_SaveToken_CreatesDirectory(t *testing.T) {
	tmpDir := t.TempDir()
	_ = tmpDir // Using t.TempDir() for automatic cleanup

	// Use a path with non-existent subdirectory
	tokenPath := filepath.Join(tmpDir, "subdir", "token.json")
	client, err := NewGitHubOAuthClient(tokenPath)
	require.NoError(t, err)

	testToken := &oauth2.Token{
		AccessToken: "test-token",
		TokenType:   "Bearer",
		Expiry:      time.Now().Add(1 * time.Hour),
	}

	// Save token (should create directory)
	err = client.SaveToken(testToken)
	assert.NoError(t, err)

	// Verify directory was created
	dir := filepath.Dir(tokenPath)
	info, err := os.Stat(dir)
	assert.NoError(t, err)
	assert.True(t, info.IsDir())
}

func TestGitHubOAuthClient_TokenFilePermissions(t *testing.T) {
	tmpDir := t.TempDir()
	_ = tmpDir // Using t.TempDir() for automatic cleanup

	tokenPath := filepath.Join(tmpDir, "token.json")
	client, err := NewGitHubOAuthClient(tokenPath)
	require.NoError(t, err)

	testToken := &oauth2.Token{
		AccessToken: "secret-token",
		TokenType:   "Bearer",
		Expiry:      time.Now().Add(1 * time.Hour),
	}

	// Save token
	err = client.SaveToken(testToken)
	require.NoError(t, err)

	// Check file permissions (should be 0600 for security)
	info, err := os.Stat(tokenPath)
	require.NoError(t, err)

	// On Unix systems, check that file is only readable/writable by owner
	mode := info.Mode()
	assert.Equal(t, os.FileMode(0600), mode.Perm(), "Token file should have 0600 permissions")
}

func TestGitHubOAuthClient_GetToken_ValidToken(t *testing.T) {
	tmpDir := t.TempDir()
	_ = tmpDir // Using t.TempDir() for automatic cleanup

	tokenPath := filepath.Join(tmpDir, "token.json")
	client, err := NewGitHubOAuthClient(tokenPath)
	require.NoError(t, err)

	// Save a valid token
	testToken := &oauth2.Token{
		AccessToken: "valid-access-token",
		TokenType:   "Bearer",
		Expiry:      time.Now().Add(1 * time.Hour),
	}
	err = client.SaveToken(testToken)
	require.NoError(t, err)

	// Get token
	ctx := context.Background()
	token, err := client.GetToken(ctx)
	assert.NoError(t, err)
	assert.NotNil(t, token)
	assert.Equal(t, testToken.AccessToken, token.AccessToken)
	assert.True(t, token.Valid())
}

func TestGitHubOAuthClient_GetToken_NoToken(t *testing.T) {
	tmpDir := t.TempDir()
	_ = tmpDir // Using t.TempDir() for automatic cleanup

	tokenPath := filepath.Join(tmpDir, "nonexistent.json")
	client, err := NewGitHubOAuthClient(tokenPath)
	require.NoError(t, err)

	// Get token when no token exists
	ctx := context.Background()
	token, err := client.GetToken(ctx)
	assert.Error(t, err)
	assert.Nil(t, token)
}

func TestGitHubOAuthClient_NewGitHubOAuthClient(t *testing.T) {
	tokenPath := "/tmp/test_token.json"
	client, err := NewGitHubOAuthClient(tokenPath)
	require.NoError(t, err)

	assert.NotNil(t, client)
	assert.Equal(t, tokenPath, client.tokenPath)
	assert.NotNil(t, client.config)
	assert.Equal(t, "https://github.com/login/device/code", client.config.Endpoint.DeviceAuthURL)
	assert.Equal(t, "https://github.com/login/oauth/access_token", client.config.Endpoint.TokenURL)
}
