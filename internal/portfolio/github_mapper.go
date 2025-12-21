package portfolio

import (
	"errors"
	"fmt"
	"path/filepath"
	"strings"
	"time"

	"github.com/fsvxavier/nexs-mcp/internal/domain"
)

// GitHubMapper handles conversion between local file structure and GitHub repository structure.
type GitHubMapper struct {
	baseDir string // Local base directory (e.g., ~/.nexs-mcp/elements)
}

// NewGitHubMapper creates a new GitHub mapper.
func NewGitHubMapper(baseDir string) *GitHubMapper {
	return &GitHubMapper{
		baseDir: baseDir,
	}
}

// LocalToGitHubPath converts local file path to GitHub repository path
// Local: baseDir/author/type/YYYY-MM-DD/id.yaml
// GitHub: elements/author/type/YYYY-MM-DD/id.yaml.
func (m *GitHubMapper) LocalToGitHubPath(localPath string) (string, error) {
	// Remove base directory from path
	relPath, err := filepath.Rel(m.baseDir, localPath)
	if err != nil {
		return "", fmt.Errorf("failed to get relative path: %w", err)
	}

	// Convert OS-specific path separators to forward slashes for GitHub
	githubPath := filepath.ToSlash(relPath)

	// Prepend "elements/" prefix for GitHub structure
	return "elements/" + githubPath, nil
}

// GitHubToLocalPath converts GitHub repository path to local file path
// GitHub: elements/author/type/YYYY-MM-DD/id.yaml
// Local: baseDir/author/type/YYYY-MM-DD/id.yaml.
func (m *GitHubMapper) GitHubToLocalPath(githubPath string) (string, error) {
	// Remove "elements/" prefix
	relPath := strings.TrimPrefix(githubPath, "elements/")
	if relPath == githubPath {
		return "", errors.New("invalid GitHub path: must start with 'elements/'")
	}

	// Convert to OS-specific path
	return filepath.Join(m.baseDir, filepath.FromSlash(relPath)), nil
}

// ElementToGitHubPath generates the GitHub path for an element.
func (m *GitHubMapper) ElementToGitHubPath(element domain.Element) string {
	// Get metadata
	metadata := element.GetMetadata()

	// Format: elements/author/type/YYYY-MM-DD/id.yaml
	dateStr := metadata.CreatedAt.Format("2006-01-02")
	author := metadata.Author
	if author == "" {
		author = "unknown"
	}

	path := fmt.Sprintf("elements/%s/%s/%s/%s.yaml",
		author,
		metadata.Type,
		dateStr,
		metadata.ID,
	)

	return path
}

// ParseGitHubPath extracts information from a GitHub path.
type PathInfo struct {
	Author string
	Type   string
	Date   time.Time
	ID     string
}

// ParseGitHubPath parses a GitHub path into structured information.
func (m *GitHubMapper) ParseGitHubPath(githubPath string) (*PathInfo, error) {
	// Expected format: elements/author/type/YYYY-MM-DD/id.yaml
	parts := strings.Split(githubPath, "/")
	if len(parts) != 5 {
		return nil, fmt.Errorf("invalid GitHub path format: expected 5 parts, got %d", len(parts))
	}

	if parts[0] != "elements" {
		return nil, errors.New("invalid GitHub path: must start with 'elements/'")
	}

	author := parts[1]
	elementType := parts[2]
	dateStr := parts[3]
	filename := parts[4]

	// Parse date
	date, err := time.Parse("2006-01-02", dateStr)
	if err != nil {
		return nil, fmt.Errorf("invalid date format in path: %w", err)
	}

	// Extract ID from filename (remove .yaml extension)
	id := strings.TrimSuffix(filename, ".yaml")
	if id == filename {
		return nil, errors.New("invalid filename: must end with .yaml")
	}

	return &PathInfo{
		Author: author,
		Type:   elementType,
		Date:   date,
		ID:     id,
	}, nil
}

// IsValidGitHubPath checks if a GitHub path follows the expected structure.
func (m *GitHubMapper) IsValidGitHubPath(githubPath string) bool {
	_, err := m.ParseGitHubPath(githubPath)
	return err == nil
}

// FilterElementPaths filters paths to only include valid element files.
func (m *GitHubMapper) FilterElementPaths(paths []string) []string {
	var validPaths []string
	for _, path := range paths {
		if m.IsValidGitHubPath(path) {
			validPaths = append(validPaths, path)
		}
	}
	return validPaths
}

// GetAuthorFromPath extracts the author from a GitHub path.
func (m *GitHubMapper) GetAuthorFromPath(githubPath string) (string, error) {
	info, err := m.ParseGitHubPath(githubPath)
	if err != nil {
		return "", err
	}
	return info.Author, nil
}

// GetTypeFromPath extracts the element type from a GitHub path.
func (m *GitHubMapper) GetTypeFromPath(githubPath string) (string, error) {
	info, err := m.ParseGitHubPath(githubPath)
	if err != nil {
		return "", err
	}
	return info.Type, nil
}

// GenerateCommitMessage generates a commit message for an element operation.
func GenerateCommitMessage(operation string, element domain.Element) string {
	metadata := element.GetMetadata()
	return fmt.Sprintf("%s: %s %s (%s)",
		operation,
		metadata.Type,
		metadata.Name,
		metadata.ID,
	)
}

// GroupPathsByAuthor groups GitHub paths by author.
func (m *GitHubMapper) GroupPathsByAuthor(paths []string) map[string][]string {
	groups := make(map[string][]string)

	for _, path := range paths {
		author, err := m.GetAuthorFromPath(path)
		if err != nil {
			continue
		}
		groups[author] = append(groups[author], path)
	}

	return groups
}

// GroupPathsByType groups GitHub paths by element type.
func (m *GitHubMapper) GroupPathsByType(paths []string) map[string][]string {
	groups := make(map[string][]string)

	for _, path := range paths {
		elementType, err := m.GetTypeFromPath(path)
		if err != nil {
			continue
		}
		groups[elementType] = append(groups[elementType], path)
	}

	return groups
}
