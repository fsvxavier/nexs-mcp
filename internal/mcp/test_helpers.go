package mcp

import (
	"os"
	"time"

	"github.com/fsvxavier/nexs-mcp/internal/config"
	"github.com/fsvxavier/nexs-mcp/internal/domain"
)

// newTestConfig creates a default config for tests.
func newTestConfig() *config.Config {
	return &config.Config{
		Resources: config.ResourcesConfig{
			Enabled:  false,
			Expose:   []string{},
			CacheTTL: 5 * time.Minute,
		},
	}
}

// newTestServer creates a test MCP server.
func newTestServer(name, version string, repo domain.ElementRepository) *MCPServer {
	// Enable test mode to use mock provider
	os.Setenv("NEXS_TEST_MODE", "1")
	return NewMCPServer(name, version, repo, newTestConfig())
}
