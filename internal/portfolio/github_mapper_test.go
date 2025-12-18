package portfolio

import (
	"path/filepath"
	"testing"
	"time"

	"github.com/fsvxavier/nexs-mcp/internal/domain"
	"github.com/stretchr/testify/assert"
)

func TestGitHubMapper_LocalToGitHubPath(t *testing.T) {
	baseDir := "/home/user/.nexs-mcp/elements"
	mapper := NewGitHubMapper(baseDir)

	tests := []struct {
		name       string
		localPath  string
		wantGitHub string
		wantErr    bool
	}{
		{
			name:       "Valid local path",
			localPath:  filepath.Join(baseDir, "alice/persona/2025-12-18/persona-123.yaml"),
			wantGitHub: "elements/alice/persona/2025-12-18/persona-123.yaml",
			wantErr:    false,
		},
		{
			name:       "Private user path",
			localPath:  filepath.Join(baseDir, "private-bob/skill/2025-12-18/skill-456.yaml"),
			wantGitHub: "elements/private-bob/skill/2025-12-18/skill-456.yaml",
			wantErr:    false,
		},
		{
			name:       "Path outside base directory",
			localPath:  "/tmp/other/persona-123.yaml",
			wantGitHub: "elements/../../../../tmp/other/persona-123.yaml",
			wantErr:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			githubPath, err := mapper.LocalToGitHubPath(tt.localPath)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				// Convert to forward slashes for comparison
				assert.Equal(t, tt.wantGitHub, filepath.ToSlash(githubPath))
			}
		})
	}
}

func TestGitHubMapper_GitHubToLocalPath(t *testing.T) {
	baseDir := "/home/user/.nexs-mcp/elements"
	mapper := NewGitHubMapper(baseDir)

	tests := []struct {
		name       string
		githubPath string
		wantLocal  string
		wantErr    bool
	}{
		{
			name:       "Valid GitHub path",
			githubPath: "elements/alice/persona/2025-12-18/persona-123.yaml",
			wantLocal:  filepath.Join(baseDir, "alice/persona/2025-12-18/persona-123.yaml"),
			wantErr:    false,
		},
		{
			name:       "Private user path",
			githubPath: "elements/private-bob/skill/2025-12-18/skill-456.yaml",
			wantLocal:  filepath.Join(baseDir, "private-bob/skill/2025-12-18/skill-456.yaml"),
			wantErr:    false,
		},
		{
			name:       "Invalid GitHub path - missing elements prefix",
			githubPath: "alice/persona/2025-12-18/persona-123.yaml",
			wantLocal:  "",
			wantErr:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			localPath, err := mapper.GitHubToLocalPath(tt.githubPath)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.wantLocal, localPath)
			}
		})
	}
}

func TestGitHubMapper_ParseGitHubPath(t *testing.T) {
	mapper := NewGitHubMapper("/tmp")

	tests := []struct {
		name       string
		githubPath string
		wantAuthor string
		wantType   string
		wantDate   string
		wantID     string
		wantErr    bool
	}{
		{
			name:       "Valid GitHub path",
			githubPath: "elements/alice/persona/2025-12-18/persona-123.yaml",
			wantAuthor: "alice",
			wantType:   "persona",
			wantDate:   "2025-12-18",
			wantID:     "persona-123",
			wantErr:    false,
		},
		{
			name:       "Private user path",
			githubPath: "elements/private-bob/skill/2025-12-01/skill-abc.yaml",
			wantAuthor: "private-bob",
			wantType:   "skill",
			wantDate:   "2025-12-01",
			wantID:     "skill-abc",
			wantErr:    false,
		},
		{
			name:       "Invalid - too few parts",
			githubPath: "elements/alice/persona.yaml",
			wantErr:    true,
		},
		{
			name:       "Invalid - wrong prefix",
			githubPath: "files/alice/persona/2025-12-18/persona-123.yaml",
			wantErr:    true,
		},
		{
			name:       "Invalid - bad date format",
			githubPath: "elements/alice/persona/18-12-2025/persona-123.yaml",
			wantErr:    true,
		},
		{
			name:       "Invalid - missing yaml extension",
			githubPath: "elements/alice/persona/2025-12-18/persona-123.txt",
			wantErr:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			info, err := mapper.ParseGitHubPath(tt.githubPath)
			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, info)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, info)
				assert.Equal(t, tt.wantAuthor, info.Author)
				assert.Equal(t, tt.wantType, info.Type)
				assert.Equal(t, tt.wantID, info.ID)

				expectedDate, _ := time.Parse("2006-01-02", tt.wantDate)
				assert.Equal(t, expectedDate.Format("2006-01-02"), info.Date.Format("2006-01-02"))
			}
		})
	}
}

func TestGitHubMapper_ElementToGitHubPath(t *testing.T) {
	mapper := NewGitHubMapper("/tmp")

	createdAt := time.Date(2025, 12, 18, 10, 30, 0, 0, time.UTC)

	persona := domain.NewPersona("Test Persona", "A test", "1.0.0", "alice")
	metadata := persona.GetMetadata()
	metadata.CreatedAt = createdAt
	persona.SetMetadata(metadata)

	githubPath := mapper.ElementToGitHubPath(persona)

	expected := "elements/alice/persona/2025-12-18/" + metadata.ID + ".yaml"
	assert.Equal(t, expected, githubPath)
}

func TestGitHubMapper_FilterElementPaths(t *testing.T) {
	mapper := NewGitHubMapper("/tmp")

	paths := []string{
		"elements/alice/persona/2025-12-18/persona-123.yaml",
		"elements/bob/skill/2025-12-18/skill-456.yaml",
		"README.md",
		"elements/alice/invalid-path.yaml",
		"docs/guide.md",
		"elements/charlie/template/2025-12-18/template-789.yaml",
		"elements/bad/format/persona.yaml",
	}

	validPaths := mapper.FilterElementPaths(paths)

	assert.Len(t, validPaths, 3)
	assert.Contains(t, validPaths, "elements/alice/persona/2025-12-18/persona-123.yaml")
	assert.Contains(t, validPaths, "elements/bob/skill/2025-12-18/skill-456.yaml")
	assert.Contains(t, validPaths, "elements/charlie/template/2025-12-18/template-789.yaml")
}

func TestGitHubMapper_GetAuthorFromPath(t *testing.T) {
	mapper := NewGitHubMapper("/tmp")

	tests := []struct {
		name       string
		githubPath string
		wantAuthor string
		wantErr    bool
	}{
		{
			name:       "Valid path",
			githubPath: "elements/alice/persona/2025-12-18/persona-123.yaml",
			wantAuthor: "alice",
			wantErr:    false,
		},
		{
			name:       "Private user",
			githubPath: "elements/private-bob/skill/2025-12-18/skill-456.yaml",
			wantAuthor: "private-bob",
			wantErr:    false,
		},
		{
			name:       "Invalid path",
			githubPath: "invalid/path.yaml",
			wantErr:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			author, err := mapper.GetAuthorFromPath(tt.githubPath)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.wantAuthor, author)
			}
		})
	}
}

func TestGitHubMapper_GetTypeFromPath(t *testing.T) {
	mapper := NewGitHubMapper("/tmp")

	tests := []struct {
		name       string
		githubPath string
		wantType   string
		wantErr    bool
	}{
		{
			name:       "Persona type",
			githubPath: "elements/alice/persona/2025-12-18/persona-123.yaml",
			wantType:   "persona",
			wantErr:    false,
		},
		{
			name:       "Skill type",
			githubPath: "elements/bob/skill/2025-12-18/skill-456.yaml",
			wantType:   "skill",
			wantErr:    false,
		},
		{
			name:       "Invalid path",
			githubPath: "invalid/path.yaml",
			wantErr:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			elementType, err := mapper.GetTypeFromPath(tt.githubPath)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.wantType, elementType)
			}
		})
	}
}

func TestGitHubMapper_GroupPathsByAuthor(t *testing.T) {
	mapper := NewGitHubMapper("/tmp")

	paths := []string{
		"elements/alice/persona/2025-12-18/persona-123.yaml",
		"elements/alice/skill/2025-12-18/skill-456.yaml",
		"elements/bob/persona/2025-12-18/persona-789.yaml",
		"invalid/path.yaml", // Should be ignored
	}

	groups := mapper.GroupPathsByAuthor(paths)

	assert.Len(t, groups, 2)
	assert.Len(t, groups["alice"], 2)
	assert.Len(t, groups["bob"], 1)
}

func TestGitHubMapper_GroupPathsByType(t *testing.T) {
	mapper := NewGitHubMapper("/tmp")

	paths := []string{
		"elements/alice/persona/2025-12-18/persona-123.yaml",
		"elements/bob/persona/2025-12-18/persona-456.yaml",
		"elements/alice/skill/2025-12-18/skill-789.yaml",
		"invalid/path.yaml", // Should be ignored
	}

	groups := mapper.GroupPathsByType(paths)

	assert.Len(t, groups, 2)
	assert.Len(t, groups["persona"], 2)
	assert.Len(t, groups["skill"], 1)
}

func TestGenerateCommitMessage(t *testing.T) {
	persona := domain.NewPersona("Test Persona", "Description", "1.0.0", "alice")

	message := GenerateCommitMessage("Add", persona)

	metadata := persona.GetMetadata()
	expected := "Add: persona Test Persona (" + metadata.ID + ")"
	assert.Equal(t, expected, message)
}
