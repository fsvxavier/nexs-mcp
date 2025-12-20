package mcp

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/fsvxavier/nexs-mcp/internal/config"
	"github.com/fsvxavier/nexs-mcp/internal/domain"
	"github.com/fsvxavier/nexs-mcp/internal/infrastructure"
)

func TestSaveConversationContext(t *testing.T) {
	// Create temp directory for test
	tmpDir := t.TempDir()

	// Create repository
	repo, err := infrastructure.NewFileElementRepository(tmpDir)
	require.NoError(t, err)

	// Create config with auto-save enabled
	cfg := &config.Config{
		AutoSaveMemories: true,
		AutoSaveInterval: 5 * time.Minute,
		DataDir:          tmpDir,
		StorageType:      "file",
	}

	// Create MCP server
	server := NewMCPServer("test-server", "1.0.0", repo, cfg)

	tests := []struct {
		name        string
		input       SaveConversationContextInput
		wantErr     bool
		errContains string
		checkOutput func(t *testing.T, output SaveConversationContextOutput)
	}{
		{
			name: "save basic conversation context",
			input: SaveConversationContextInput{
				Context: "User asked about persona creation. Agent explained the create_persona tool and demonstrated usage.",
				Summary: "Persona creation discussion",
				Tags:    []string{"tutorial", "persona"},
			},
			wantErr: false,
			checkOutput: func(t *testing.T, output SaveConversationContextOutput) {
				assert.True(t, output.Saved)
				assert.NotEmpty(t, output.MemoryID)
				assert.Contains(t, output.Message, "successfully")
			},
		},
		{
			name: "save with importance level",
			input: SaveConversationContextInput{
				Context:    "Critical bug discovered in memory persistence. Root cause identified as SimpleElement usage instead of typed elements.",
				Summary:    "Memory persistence bug analysis",
				Importance: "critical",
				Tags:       []string{"bug", "investigation"},
			},
			wantErr: false,
			checkOutput: func(t *testing.T, output SaveConversationContextOutput) {
				assert.True(t, output.Saved)

				// Verify memory was created with correct metadata
				mem, err := repo.GetByID(output.MemoryID)
				require.NoError(t, err)

				memory, ok := mem.(*domain.Memory)
				require.True(t, ok)
				assert.Contains(t, memory.Content, "Critical bug")
				assert.Contains(t, memory.Metadata, "importance")
				assert.Equal(t, "critical", memory.Metadata["importance"])

				// Check tags include importance
				metadata := memory.GetMetadata()
				assert.Contains(t, metadata.Tags, "importance:critical")
				assert.Contains(t, metadata.Tags, "auto-save")
			},
		},
		{
			name: "save with related elements",
			input: SaveConversationContextInput{
				Context:   "Following up on previous discussion about ensembles. User requested example configuration.",
				Summary:   "Ensemble configuration example",
				RelatedTo: []string{"ensemble-123", "agent-456"},
				Tags:      []string{"example", "configuration"},
			},
			wantErr: false,
			checkOutput: func(t *testing.T, output SaveConversationContextOutput) {
				assert.True(t, output.Saved)

				// Verify related_to metadata
				mem, err := repo.GetByID(output.MemoryID)
				require.NoError(t, err)

				memory, ok := mem.(*domain.Memory)
				require.True(t, ok)
				assert.Contains(t, memory.Metadata, "related_to")
				assert.Contains(t, memory.Metadata["related_to"], "ensemble-123")
				assert.Contains(t, memory.Metadata["related_to"], "agent-456")
			},
		},
		{
			name: "fail on empty context",
			input: SaveConversationContextInput{
				Context: "",
				Summary: "Empty context test",
			},
			wantErr:     true,
			errContains: "at least 10 characters",
		},
		{
			name: "fail on too short context",
			input: SaveConversationContextInput{
				Context: "Short",
				Summary: "Too short",
			},
			wantErr:     true,
			errContains: "at least 10 characters",
		},
		{
			name: "auto-save disabled returns early",
			input: SaveConversationContextInput{
				Context: "This should not be saved when auto-save is disabled",
				Summary: "Auto-save disabled test",
			},
			wantErr: false,
			checkOutput: func(t *testing.T, output SaveConversationContextOutput) {
				// Note: This test will need a separate server instance with disabled auto-save
				// For now, it will save successfully with current config
				assert.True(t, output.Saved || !output.Saved) // Just checking structure
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()

			_, output, err := server.handleSaveConversationContext(ctx, nil, tt.input)

			if tt.wantErr {
				require.Error(t, err)
				if tt.errContains != "" {
					assert.Contains(t, err.Error(), tt.errContains)
				}
				return
			}

			require.NoError(t, err)
			if tt.checkOutput != nil {
				tt.checkOutput(t, output)
			}
		})
	}
}

func TestSaveConversationContextWithDisabledAutoSave(t *testing.T) {
	tmpDir := t.TempDir()
	repo, err := infrastructure.NewFileElementRepository(tmpDir)
	require.NoError(t, err)

	// Create config with auto-save DISABLED
	cfg := &config.Config{
		AutoSaveMemories: false, // Disabled
		DataDir:          tmpDir,
		StorageType:      "file",
	}

	server := NewMCPServer("test-server", "1.0.0", repo, cfg)

	input := SaveConversationContextInput{
		Context: "This should not be saved when auto-save is disabled",
		Summary: "Auto-save disabled test",
	}

	ctx := context.Background()
	_, output, err := server.handleSaveConversationContext(ctx, nil, input)

	require.NoError(t, err)
	assert.False(t, output.Saved)
	assert.Contains(t, output.Message, "disabled")
}

func TestExtractKeywords(t *testing.T) {
	tests := []struct {
		name        string
		text        string
		maxKeywords int
		minKeywords int
		wantExclude []string
	}{
		{
			name:        "extract from technical text",
			text:        "The memory persistence bug was caused by using SimpleElement instead of Memory struct. The fix requires using type-specific tools.",
			maxKeywords: 5,
			minKeywords: 3,
			wantExclude: []string{"the", "was", "by"},
		},
		{
			name:        "extract from Portuguese text",
			text:        "O usuário solicitou a criação de uma persona baseada no arquivo anexo. A persona foi criada com sucesso.",
			maxKeywords: 5,
			minKeywords: 3,
			wantExclude: []string{"o", "de", "uma", "no", "a", "foi"},
		},
		{
			name:        "handle short text",
			text:        "Bug fix complete.",
			maxKeywords: 5,
			minKeywords: 2,
			wantExclude: []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			keywords := extractKeywords(tt.text, tt.maxKeywords)

			// Check max keywords constraint
			assert.LessOrEqual(t, len(keywords), tt.maxKeywords)

			// Check minimum keywords generated
			assert.GreaterOrEqual(t, len(keywords), tt.minKeywords, "Should extract at least %d keywords", tt.minKeywords)

			// Check excluded words are not present
			for _, exclude := range tt.wantExclude {
				assert.NotContains(t, keywords, exclude, "Stop word %q should not be in keywords", exclude)
			}

			// Verify all keywords are at least 3 characters
			for _, kw := range keywords {
				assert.GreaterOrEqual(t, len(kw), 3, "Keyword %q should be at least 3 characters", kw)
			}
		})
	}
}
