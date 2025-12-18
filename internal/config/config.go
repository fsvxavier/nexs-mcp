package config

import (
	"flag"
	"os"
)

// Config holds the application configuration
type Config struct {
	// StorageType can be "memory" or "file"
	StorageType string

	// DataDir is the directory for file-based storage
	DataDir string

	// ServerName is the name of the MCP server
	ServerName string

	// Version is the application version
	Version string
}

// LoadConfig loads configuration from environment variables and command-line flags
func LoadConfig(version string) *Config {
	cfg := &Config{
		ServerName: getEnvOrDefault("NEXS_SERVER_NAME", "nexs-mcp"),
		Version:    version,
	}

	// Define command-line flags
	flag.StringVar(&cfg.StorageType, "storage", getEnvOrDefault("NEXS_STORAGE_TYPE", "file"),
		"Storage type: 'memory' or 'file'")
	flag.StringVar(&cfg.DataDir, "data-dir", getEnvOrDefault("NEXS_DATA_DIR", "data/elements"),
		"Directory for file-based storage")

	flag.Parse()

	return cfg
}

// getEnvOrDefault returns an environment variable value or a default value
func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
