package config

import (
	"flag"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLoadConfig(t *testing.T) {
	tests := []struct {
		name     string
		envVars  map[string]string
		args     []string
		expected *Config
	}{
		{
			name: "default configuration",
			envVars: map[string]string{
				"NEXS_SERVER_NAME":  "",
				"NEXS_STORAGE_TYPE": "",
				"NEXS_DATA_DIR":     "",
			},
			args: []string{},
			expected: &Config{
				ServerName:  "nexs-mcp",
				StorageType: "file",
				DataDir:     "data/elements",
				Version:     "test-version",
			},
		},
		{
			name: "environment variables override defaults",
			envVars: map[string]string{
				"NEXS_SERVER_NAME":  "custom-server",
				"NEXS_STORAGE_TYPE": "memory",
				"NEXS_DATA_DIR":     "/custom/path",
			},
			args: []string{},
			expected: &Config{
				ServerName:  "custom-server",
				StorageType: "memory",
				DataDir:     "/custom/path",
				Version:     "test-version",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Reset flags for each subtest
			flag.CommandLine = flag.NewFlagSet(os.Args[0], flag.ContinueOnError)

			// Set environment variables
			for key, value := range tt.envVars {
				if value == "" {
					os.Unsetenv(key)
				} else {
					t.Setenv(key, value)
				}
			}
			defer func() {
				for key := range tt.envVars {
					os.Unsetenv(key)
				}
			}()

			// Set command-line arguments
			oldArgs := os.Args
			os.Args = append([]string{"cmd"}, tt.args...)
			defer func() { os.Args = oldArgs }()

			cfg := LoadConfig("test-version")

			assert.Equal(t, tt.expected.ServerName, cfg.ServerName)
			assert.Equal(t, tt.expected.StorageType, cfg.StorageType)
			assert.Equal(t, tt.expected.DataDir, cfg.DataDir)
			assert.Equal(t, tt.expected.Version, cfg.Version)
		})
	}
}

func TestGetEnvOrDefault(t *testing.T) {
	tests := []struct {
		name         string
		envKey       string
		envValue     string
		defaultValue string
		expected     string
	}{
		{
			name:         "environment variable exists",
			envKey:       "TEST_ENV_VAR",
			envValue:     "test-value",
			defaultValue: "default-value",
			expected:     "test-value",
		},
		{
			name:         "environment variable does not exist",
			envKey:       "NONEXISTENT_VAR",
			envValue:     "",
			defaultValue: "default-value",
			expected:     "default-value",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.envValue != "" {
				t.Setenv(tt.envKey, tt.envValue)
			} else {
				os.Unsetenv(tt.envKey)
			}

			result := getEnvOrDefault(tt.envKey, tt.defaultValue)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestConfig_Struct(t *testing.T) {
	cfg := &Config{
		ServerName:  "test-server",
		StorageType: "memory",
		DataDir:     "/test/path",
		Version:     "1.0.0",
	}

	assert.Equal(t, "test-server", cfg.ServerName)
	assert.Equal(t, "memory", cfg.StorageType)
	assert.Equal(t, "/test/path", cfg.DataDir)
	assert.Equal(t, "1.0.0", cfg.Version)
}
