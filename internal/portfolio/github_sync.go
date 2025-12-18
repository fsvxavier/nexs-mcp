package portfolio

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"time"

	"github.com/fsvxavier/nexs-mcp/internal/domain"
	"github.com/fsvxavier/nexs-mcp/internal/infrastructure"
)

// ConflictResolution defines how to resolve sync conflicts
type ConflictResolution string

const (
	// ConflictLocalWins keeps the local version
	ConflictLocalWins ConflictResolution = "local_wins"
	// ConflictRemoteWins keeps the remote version
	ConflictRemoteWins ConflictResolution = "remote_wins"
	// ConflictNewerWins keeps the version with newer timestamp
	ConflictNewerWins ConflictResolution = "newer_wins"
	// ConflictManual requires manual resolution
	ConflictManual ConflictResolution = "manual"
)

// Conflict represents a sync conflict between local and remote versions
type Conflict struct {
	ElementID       string
	Path            string
	LocalContent    string
	RemoteContent   string
	LocalHash       string
	RemoteHash      string
	LocalUpdatedAt  time.Time
	RemoteUpdatedAt time.Time
}

// SyncResult represents the result of a sync operation
type SyncResult struct {
	Pushed    int
	Pulled    int
	Conflicts []Conflict
	Errors    []string
}

// GitHubSync handles bidirectional synchronization between local and GitHub
type GitHubSync struct {
	githubClient       *infrastructure.GitHubClient
	repository         *infrastructure.EnhancedFileElementRepository
	mapper             *GitHubMapper
	conflictResolution ConflictResolution
}

// NewGitHubSync creates a new GitHub sync manager
func NewGitHubSync(
	githubClient *infrastructure.GitHubClient,
	repository *infrastructure.EnhancedFileElementRepository,
	mapper *GitHubMapper,
	conflictResolution ConflictResolution,
) *GitHubSync {
	return &GitHubSync{
		githubClient:       githubClient,
		repository:         repository,
		mapper:             mapper,
		conflictResolution: conflictResolution,
	}
}

// Push syncs local elements to GitHub repository
func (s *GitHubSync) Push(ctx context.Context, owner, repo, branch string) (*SyncResult, error) {
	result := &SyncResult{}

	// Get all local elements
	filter := domain.ElementFilter{
		Offset: 0,
		Limit:  10000, // Get all elements
	}
	elements, err := s.repository.List(filter)
	if err != nil {
		return nil, fmt.Errorf("failed to list local elements: %w", err)
	}

	for _, element := range elements {
		// Convert element to YAML
		yamlContent, err := s.elementToYAML(element)
		if err != nil {
			result.Errors = append(result.Errors, fmt.Sprintf("failed to marshal %s: %v", element.GetMetadata().ID, err))
			continue
		}

		// Get GitHub path for this element
		githubPath := s.mapper.ElementToGitHubPath(element)

		// Try to get existing file
		existingFile, err := s.githubClient.GetFile(ctx, owner, repo, githubPath, branch)
		if err == nil {
			// File exists, check if update is needed
			conflict := s.detectConflict(element, existingFile.Content, yamlContent)
			if conflict != nil {
				// Handle conflict
				shouldUpdate, err := s.resolveConflict(conflict)
				if err != nil {
					result.Conflicts = append(result.Conflicts, *conflict)
					continue
				}
				if !shouldUpdate {
					continue // Skip this file
				}
			}

			// Update existing file
			message := GenerateCommitMessage("Update", element)
			_, err = s.githubClient.UpdateFile(ctx, owner, repo, githubPath, message, yamlContent, existingFile.SHA, branch)
			if err != nil {
				result.Errors = append(result.Errors, fmt.Sprintf("failed to update %s: %v", githubPath, err))
				continue
			}
		} else {
			// File doesn't exist, create it
			message := GenerateCommitMessage("Add", element)
			_, err = s.githubClient.CreateFile(ctx, owner, repo, githubPath, message, yamlContent, branch)
			if err != nil {
				result.Errors = append(result.Errors, fmt.Sprintf("failed to create %s: %v", githubPath, err))
				continue
			}
		}

		result.Pushed++
	}

	return result, nil
}

// Pull syncs GitHub repository to local elements
func (s *GitHubSync) Pull(ctx context.Context, owner, repo, branch string) (*SyncResult, error) {
	result := &SyncResult{}

	// Get all files from GitHub
	files, err := s.githubClient.ListAllFiles(ctx, owner, repo, branch)
	if err != nil {
		return nil, fmt.Errorf("failed to list GitHub files: %w", err)
	}

	// Filter to only element files
	elementPaths := s.mapper.FilterElementPaths(getFilePaths(files))

	for _, githubPath := range elementPaths {
		// Get file content
		fileContent, err := s.githubClient.GetFile(ctx, owner, repo, githubPath, branch)
		if err != nil {
			result.Errors = append(result.Errors, fmt.Sprintf("failed to get %s: %v", githubPath, err))
			continue
		}

		// Parse YAML to element
		element, err := s.yamlToElement(fileContent.Content)
		if err != nil {
			result.Errors = append(result.Errors, fmt.Sprintf("failed to parse %s: %v", githubPath, err))
			continue
		}

		// Check if element exists locally
		existingElement, err := s.repository.GetByID(element.GetMetadata().ID)
		if err == nil {
			// Element exists, check for conflicts
			existingYAML, _ := s.elementToYAML(existingElement)
			conflict := s.detectConflict(element, existingYAML, fileContent.Content)
			if conflict != nil {
				shouldUpdate, err := s.resolveConflict(conflict)
				if err != nil {
					result.Conflicts = append(result.Conflicts, *conflict)
					continue
				}
				if !shouldUpdate {
					continue
				}
			}

			// Update local element
			if err := s.repository.Update(element); err != nil {
				result.Errors = append(result.Errors, fmt.Sprintf("failed to update %s: %v", element.GetMetadata().ID, err))
				continue
			}
		} else {
			// Element doesn't exist, create it
			if err := s.repository.Create(element); err != nil {
				result.Errors = append(result.Errors, fmt.Sprintf("failed to create %s: %v", element.GetMetadata().ID, err))
				continue
			}
		}

		result.Pulled++
	}

	return result, nil
}

// detectConflict checks if there's a conflict between local and remote versions
func (s *GitHubSync) detectConflict(element domain.Element, localContent, remoteContent string) *Conflict {
	localHash := computeHash(localContent)
	remoteHash := computeHash(remoteContent)

	// No conflict if content is identical
	if localHash == remoteHash {
		return nil
	}

	metadata := element.GetMetadata()

	return &Conflict{
		ElementID:      metadata.ID,
		Path:           s.mapper.ElementToGitHubPath(element),
		LocalContent:   localContent,
		RemoteContent:  remoteContent,
		LocalHash:      localHash,
		RemoteHash:     remoteHash,
		LocalUpdatedAt: metadata.UpdatedAt,
		// RemoteUpdatedAt would need to be parsed from remote YAML
	}
}

// resolveConflict resolves a conflict based on the configured strategy
func (s *GitHubSync) resolveConflict(conflict *Conflict) (bool, error) {
	switch s.conflictResolution {
	case ConflictLocalWins:
		return true, nil // Update remote with local
	case ConflictRemoteWins:
		return false, nil // Keep remote, don't update
	case ConflictNewerWins:
		if conflict.LocalUpdatedAt.After(conflict.RemoteUpdatedAt) {
			return true, nil // Local is newer, update remote
		}
		return false, nil // Remote is newer, keep it
	case ConflictManual:
		return false, fmt.Errorf("manual conflict resolution required")
	default:
		return false, fmt.Errorf("unknown conflict resolution strategy: %s", s.conflictResolution)
	}
}

// elementToYAML converts an element to YAML string
func (s *GitHubSync) elementToYAML(element domain.Element) (string, error) {
	// Use repository's internal method to save element
	// For now, we'll create a simple YAML representation
	metadata := element.GetMetadata()
	stored := &infrastructure.StoredElement{
		Metadata: metadata,
	}

	data, err := s.repository.MarshalElement(stored)
	if err != nil {
		return "", err
	}

	return string(data), nil
}

// yamlToElement converts YAML string to an element
func (s *GitHubSync) yamlToElement(yamlContent string) (domain.Element, error) {
	stored, err := s.repository.UnmarshalElement([]byte(yamlContent))
	if err != nil {
		return nil, err
	}

	// Convert StoredElement to typed element using repository method
	return s.repository.ConvertToTypedElement(stored)
}

// computeHash computes SHA-256 hash of content
func computeHash(content string) string {
	hash := sha256.Sum256([]byte(content))
	return hex.EncodeToString(hash[:])
}

// getFilePaths extracts paths from FileContent slice
func getFilePaths(files []*infrastructure.FileContent) []string {
	paths := make([]string, len(files))
	for i, f := range files {
		paths[i] = f.Path
	}
	return paths
}

// SyncBidirectional performs both push and pull operations
func (s *GitHubSync) SyncBidirectional(ctx context.Context, owner, repo, branch string) (*SyncResult, error) {
	// First pull from GitHub
	pullResult, err := s.Pull(ctx, owner, repo, branch)
	if err != nil {
		return nil, fmt.Errorf("pull failed: %w", err)
	}

	// Then push to GitHub
	pushResult, err := s.Push(ctx, owner, repo, branch)
	if err != nil {
		return nil, fmt.Errorf("push failed: %w", err)
	}

	// Combine results
	result := &SyncResult{
		Pushed:    pushResult.Pushed,
		Pulled:    pullResult.Pulled,
		Conflicts: append(pullResult.Conflicts, pushResult.Conflicts...),
		Errors:    append(pullResult.Errors, pushResult.Errors...),
	}

	return result, nil
}
