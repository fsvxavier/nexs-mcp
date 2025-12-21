package infrastructure

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/github"
)

const (
	// GitHub OAuth2 client ID for device flow
	// Note: In production, this should be configured via environment variable.
	DefaultClientID = "Ov23liJUZYB3K5BO6JGP"
)

// DeviceFlowResponse represents the response from GitHub device flow initiation.
type DeviceFlowResponse struct {
	DeviceCode      string `json:"device_code"`
	UserCode        string `json:"user_code"`
	VerificationURI string `json:"verification_uri"`
	ExpiresIn       int    `json:"expires_in"`
	Interval        int    `json:"interval"`
}

// GitHubOAuthClient handles GitHub OAuth2 authentication using device flow.
type GitHubOAuthClient struct {
	clientID     string
	config       *oauth2.Config
	tokenPath    string
	currentToken *oauth2.Token
	encryptor    *TokenEncryptor
}

// NewGitHubOAuthClient creates a new GitHub OAuth client.
func NewGitHubOAuthClient(tokenPath string) (*GitHubOAuthClient, error) {
	clientID := os.Getenv("GITHUB_CLIENT_ID")
	if clientID == "" {
		clientID = DefaultClientID
	}

	config := &oauth2.Config{
		ClientID: clientID,
		Endpoint: github.Endpoint,
		Scopes:   []string{"repo", "user"},
	}

	// Initialize encryptor for secure token storage
	encryptor, err := NewTokenEncryptor()
	if err != nil {
		return nil, fmt.Errorf("failed to initialize token encryptor: %w", err)
	}

	return &GitHubOAuthClient{
		clientID:  clientID,
		config:    config,
		tokenPath: tokenPath,
		encryptor: encryptor,
	}, nil
}

// StartDeviceFlow initiates the GitHub OAuth2 device flow.
func (c *GitHubOAuthClient) StartDeviceFlow(ctx context.Context) (*DeviceFlowResponse, error) {
	deviceAuth, err := c.config.DeviceAuth(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to initiate device flow: %w", err)
	}

	response := &DeviceFlowResponse{
		DeviceCode:      deviceAuth.DeviceCode,
		UserCode:        deviceAuth.UserCode,
		VerificationURI: deviceAuth.VerificationURI,
		ExpiresIn:       int(deviceAuth.Interval * 60), // Convert to seconds
		Interval:        int(deviceAuth.Interval),
	}

	return response, nil
}

// PollForToken polls GitHub for the OAuth2 token after user authorization.
func (c *GitHubOAuthClient) PollForToken(ctx context.Context, deviceCode string, interval int) (*oauth2.Token, error) {
	ticker := time.NewTicker(time.Duration(interval) * time.Second)
	defer ticker.Stop()

	timeout := time.After(10 * time.Minute)

	for {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-timeout:
			return nil, errors.New("device flow timeout: user did not authorize within 10 minutes")
		case <-ticker.C:
			token, err := c.config.DeviceAccessToken(ctx, &oauth2.DeviceAuthResponse{
				DeviceCode: deviceCode,
			})
			if err != nil {
				// Check if it's a "slow down" or "authorization pending" error
				if err.Error() == "authorization_pending" || err.Error() == "slow_down" {
					continue
				}
				return nil, fmt.Errorf("failed to get access token: %w", err)
			}

			c.currentToken = token
			return token, nil
		}
	}
}

// SaveToken saves the OAuth2 token to disk with encryption.
func (c *GitHubOAuthClient) SaveToken(token *oauth2.Token) error {
	// Create directory if it doesn't exist
	dir := filepath.Dir(c.tokenPath)
	if err := os.MkdirAll(dir, 0700); err != nil {
		return fmt.Errorf("failed to create token directory: %w", err)
	}

	// Marshal token to JSON
	data, err := json.Marshal(token)
	if err != nil {
		return fmt.Errorf("failed to marshal token: %w", err)
	}

	// Encrypt token data
	encrypted, err := c.encryptor.Encrypt(data)
	if err != nil {
		return fmt.Errorf("failed to encrypt token: %w", err)
	}

	// Write encrypted data to file with restricted permissions
	if err := os.WriteFile(c.tokenPath, []byte(encrypted), 0600); err != nil {
		return fmt.Errorf("failed to write token file: %w", err)
	}

	c.currentToken = token
	return nil
}

// LoadToken loads and decrypts the OAuth2 token from disk.
func (c *GitHubOAuthClient) LoadToken() (*oauth2.Token, error) {
	encryptedData, err := os.ReadFile(c.tokenPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, errors.New("token file not found: user needs to authenticate")
		}
		return nil, fmt.Errorf("failed to read token file: %w", err)
	}

	// Decrypt token data
	data, err := c.encryptor.Decrypt(string(encryptedData))
	if err != nil {
		return nil, fmt.Errorf("failed to decrypt token: %w", err)
	}

	var token oauth2.Token
	if err := json.Unmarshal(data, &token); err != nil {
		return nil, fmt.Errorf("failed to unmarshal token: %w", err)
	}

	c.currentToken = &token
	return &token, nil
}

// GetToken returns the current token, loading from disk if needed.
func (c *GitHubOAuthClient) GetToken(ctx context.Context) (*oauth2.Token, error) {
	if c.currentToken != nil {
		// Check if token is still valid
		if c.currentToken.Valid() {
			return c.currentToken, nil
		}
	}

	// Try to load from disk
	token, err := c.LoadToken()
	if err != nil {
		return nil, err
	}

	// Refresh token if expired but refreshable
	if !token.Valid() && token.RefreshToken != "" {
		tokenSource := c.config.TokenSource(ctx, token)
		newToken, err := tokenSource.Token()
		if err != nil {
			return nil, fmt.Errorf("failed to refresh token: %w", err)
		}

		if err := c.SaveToken(newToken); err != nil {
			return nil, fmt.Errorf("failed to save refreshed token: %w", err)
		}

		return newToken, nil
	}

	if !token.Valid() {
		return nil, errors.New("token is expired and cannot be refreshed")
	}

	return token, nil
}

// IsAuthenticated checks if there's a valid token available.
func (c *GitHubOAuthClient) IsAuthenticated(ctx context.Context) bool {
	_, err := c.GetToken(ctx)
	return err == nil
}

// ClearToken removes the stored token.
func (c *GitHubOAuthClient) ClearToken() error {
	c.currentToken = nil
	if err := os.Remove(c.tokenPath); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to remove token file: %w", err)
	}
	return nil
}
