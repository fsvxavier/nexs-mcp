package config

import (
	"flag"
	"os"
	"time"
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

	// LogLevel is the logging level (debug, info, warn, error)
	LogLevel string

	// LogFormat is the log output format (json, text)
	LogFormat string

	// Resources configuration
	Resources ResourcesConfig
}

// ResourcesConfig holds configuration for MCP Resources Protocol
type ResourcesConfig struct {
	// Enabled controls whether resources are exposed to clients
	// Default: false (resources disabled for safety)
	Enabled bool

	// Expose lists which resource URIs to expose
	// Empty means expose all resources when Enabled=true
	// Example: ["capability-index://summary", "capability-index://stats"]
	Expose []string

	// CacheTTL is the duration to cache resource content
	// Default: 5 minutes
	CacheTTL time.Duration
}

// LoadConfig loads configuration from environment variables and command-line flags
func LoadConfig(version string) *Config {
	cfg := &Config{
		ServerName: getEnvOrDefault("NEXS_SERVER_NAME", "nexs-mcp"),
		Version:    version,
		Resources: ResourcesConfig{
			Enabled:  getEnvBool("NEXS_RESOURCES_ENABLED", false),
			Expose:   []string{},
			CacheTTL: getEnvDuration("NEXS_RESOURCES_CACHE_TTL", 5*time.Minute),
		},
	}

	// Define command-line flags
	flag.StringVar(&cfg.StorageType, "storage", getEnvOrDefault("NEXS_STORAGE_TYPE", "file"),
		"Storage type: 'memory' or 'file'")
	flag.StringVar(&cfg.DataDir, "data-dir", getEnvOrDefault("NEXS_DATA_DIR", "data/elements"),
		"Directory for file-based storage")
	flag.StringVar(&cfg.LogLevel, "log-level", getEnvOrDefault("NEXS_LOG_LEVEL", "info"),
		"Log level: 'debug', 'info', 'warn', 'error'")
	flag.StringVar(&cfg.LogFormat, "log-format", getEnvOrDefault("NEXS_LOG_FORMAT", "json"),
		"Log format: 'json' or 'text'")
	flag.BoolVar(&cfg.Resources.Enabled, "resources-enabled", cfg.Resources.Enabled,
		"Enable MCP Resources Protocol (default: false)")
	flag.DurationVar(&cfg.Resources.CacheTTL, "resources-cache-ttl", cfg.Resources.CacheTTL,
		"Cache TTL for resource content")

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

// getEnvBool returns a boolean environment variable value or a default value
func getEnvBool(key string, defaultValue bool) bool {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value == "true" || value == "1" || value == "yes"
}

// getEnvDuration returns a duration environment variable value or a default value
func getEnvDuration(key string, defaultValue time.Duration) time.Duration {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	duration, err := time.ParseDuration(value)
	if err != nil {
		return defaultValue
	}
	return duration
}
