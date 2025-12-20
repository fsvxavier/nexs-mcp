package collection

import (
	"archive/tar"
	"compress/gzip"
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

// Manager handles collection update, export, and publishing operations
type Manager struct {
	installer *Installer
	registry  *Registry
}

// NewManager creates a new collection manager
func NewManager(installer *Installer, registry *Registry) *Manager {
	return &Manager{
		installer: installer,
		registry:  registry,
	}
}

// UpdateResult contains information about a collection update
type UpdateResult struct {
	CollectionID    string `json:"collection_id"`
	OldVersion      string `json:"old_version"`
	NewVersion      string `json:"new_version"`
	Updated         bool   `json:"updated"`
	Message         string `json:"message"`
	UpdateAvailable bool   `json:"update_available"`
}

// CheckUpdates checks all installed collections for available updates
func (m *Manager) CheckUpdates(ctx context.Context) ([]*UpdateResult, error) {
	results := make([]*UpdateResult, 0)
	installed := m.installer.ListInstalled()

	for _, record := range installed {
		result, err := m.CheckUpdate(ctx, record.ID)
		if err != nil {
			// Continue checking other collections even if one fails
			results = append(results, &UpdateResult{
				CollectionID: record.ID,
				OldVersion:   record.Version,
				Updated:      false,
				Message:      fmt.Sprintf("Error checking update: %v", err),
			})
			continue
		}
		results = append(results, result)
	}

	return results, nil
}

// CheckUpdate checks if an update is available for a specific collection
func (m *Manager) CheckUpdate(ctx context.Context, collectionID string) (*UpdateResult, error) {
	record, exists := m.installer.GetInstalled(collectionID)
	if !exists {
		return nil, fmt.Errorf("collection %s not installed", collectionID)
	}

	// Get latest version from source
	collection, err := m.registry.Get(ctx, record.URI)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch collection from source: %w", err)
	}

	// Parse manifest
	manifestMap, ok := collection.Manifest.(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("invalid manifest type")
	}

	manifest, err := parseManifestFromMap(manifestMap)
	if err != nil {
		return nil, fmt.Errorf("failed to parse manifest: %w", err)
	}

	result := &UpdateResult{
		CollectionID: collectionID,
		OldVersion:   record.Version,
		NewVersion:   manifest.Version,
	}

	// Compare versions
	if manifest.Version != record.Version {
		result.UpdateAvailable = true
		result.Message = fmt.Sprintf("Update available: %s → %s", record.Version, manifest.Version)
	} else {
		result.UpdateAvailable = false
		result.Message = "Up to date"
	}

	return result, nil
}

// Update updates a specific collection to the latest version
func (m *Manager) Update(ctx context.Context, collectionID string, options *UpdateOptions) (*UpdateResult, error) {
	if options == nil {
		options = &UpdateOptions{}
	}

	// Check if update is available
	result, err := m.CheckUpdate(ctx, collectionID)
	if err != nil {
		return nil, err
	}

	if !result.UpdateAvailable {
		return result, nil
	}

	record, _ := m.installer.GetInstalled(collectionID)

	// Get collection manifest for hooks
	collection, err := m.registry.Get(ctx, record.URI)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch collection: %w", err)
	}

	manifestMap, ok := collection.Manifest.(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("invalid manifest type")
	}

	manifest, err := parseManifestFromMap(manifestMap)
	if err != nil {
		return nil, fmt.Errorf("failed to parse manifest: %w", err)
	}

	// Execute pre-update hooks
	if !options.SkipHooks && manifest.Hooks != nil && len(manifest.Hooks.PreUpdate) > 0 {
		if err := m.executeHooks(ctx, manifest.Hooks.PreUpdate, record.InstallLocation); err != nil {
			return nil, fmt.Errorf("pre-update hook failed: %w", err)
		}
	}

	// Uninstall old version (with force to handle dependencies)
	if err := m.installer.Uninstall(ctx, collectionID, &UninstallOptions{Force: true}); err != nil {
		return nil, fmt.Errorf("failed to uninstall old version: %w", err)
	}

	// Install new version
	installOpts := &InstallOptions{
		Force:            true,
		SkipDependencies: options.SkipDependencies,
		SkipValidation:   options.SkipValidation,
		SkipHooks:        options.SkipHooks,
	}

	if err := m.installer.Install(ctx, record.URI, installOpts); err != nil {
		return nil, fmt.Errorf("failed to install new version: %w", err)
	}

	// Execute post-update hooks
	if !options.SkipHooks && manifest.Hooks != nil && len(manifest.Hooks.PostUpdate) > 0 {
		if err := m.executeHooks(ctx, manifest.Hooks.PostUpdate, record.InstallLocation); err != nil {
			// Log error but don't fail the update
			result.Message += fmt.Sprintf("; Warning: post-update hook failed: %v", err)
		}
	}

	result.Updated = true
	result.Message = fmt.Sprintf("Successfully updated from %s to %s", result.OldVersion, result.NewVersion)

	return result, nil
}

// UpdateAll updates all collections that have available updates
func (m *Manager) UpdateAll(ctx context.Context, options *UpdateOptions) ([]*UpdateResult, error) {
	if options == nil {
		options = &UpdateOptions{}
	}

	results := make([]*UpdateResult, 0)
	installed := m.installer.ListInstalled()

	for _, record := range installed {
		result, err := m.Update(ctx, record.ID, options)
		if err != nil {
			results = append(results, &UpdateResult{
				CollectionID: record.ID,
				OldVersion:   record.Version,
				Updated:      false,
				Message:      fmt.Sprintf("Update failed: %v", err),
			})
			continue
		}
		results = append(results, result)
	}

	return results, nil
}

// ExportOptions configures collection export behavior
type ExportOptions struct {
	IncludeBackups  bool     // Include backup files
	Compression     string   // Compression level: none, fast, best (default: best)
	ExcludePatterns []string // File patterns to exclude
}

// Export exports a collection to a tar.gz archive
func (m *Manager) Export(ctx context.Context, collectionID, outputPath string, options *ExportOptions) error {
	if options == nil {
		options = &ExportOptions{
			Compression: "best",
		}
	}

	// Get installation record
	record, exists := m.installer.GetInstalled(collectionID)
	if !exists {
		return fmt.Errorf("collection %s not installed", collectionID)
	}

	// Verify installation directory exists
	if _, err := os.Stat(record.InstallLocation); os.IsNotExist(err) {
		return fmt.Errorf("collection directory not found: %s", record.InstallLocation)
	}

	// Create output file
	outFile, err := os.Create(outputPath)
	if err != nil {
		return fmt.Errorf("failed to create output file: %w", err)
	}
	defer outFile.Close()

	// Create gzip writer with appropriate compression level
	var gzipLevel int
	switch options.Compression {
	case "none":
		gzipLevel = gzip.NoCompression
	case "fast":
		gzipLevel = gzip.BestSpeed
	default: // "best" or unspecified
		gzipLevel = gzip.BestCompression
	}

	gzipWriter, err := gzip.NewWriterLevel(outFile, gzipLevel)
	if err != nil {
		return fmt.Errorf("failed to create gzip writer: %w", err)
	}
	defer gzipWriter.Close()

	// Create tar writer
	tarWriter := tar.NewWriter(gzipWriter)
	defer tarWriter.Close()

	// Walk the collection directory and add files to archive
	baseDir := record.InstallLocation
	return filepath.Walk(baseDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Get relative path
		relPath, err := filepath.Rel(baseDir, path)
		if err != nil {
			return err
		}

		// Skip root directory
		if relPath == "." {
			return nil
		}

		// Check exclude patterns
		if m.shouldExclude(relPath, info, options) {
			if info.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}

		// Create tar header
		header, err := tar.FileInfoHeader(info, "")
		if err != nil {
			return fmt.Errorf("failed to create tar header: %w", err)
		}

		header.Name = relPath

		// Write header
		if err := tarWriter.WriteHeader(header); err != nil {
			return fmt.Errorf("failed to write tar header: %w", err)
		}

		// Write file content (if not a directory)
		if !info.IsDir() {
			file, err := os.Open(path)
			if err != nil {
				return err
			}
			defer file.Close()

			if _, err := io.Copy(tarWriter, file); err != nil {
				return fmt.Errorf("failed to write file content: %w", err)
			}
		}

		return nil
	})
}

// PublishOptions configures collection publishing behavior
type PublishOptions struct {
	GitHubRepo     string // GitHub repository (owner/repo)
	Branch         string // Target branch (default: main)
	CommitMessage  string // Commit message
	CreateRelease  bool   // Create a GitHub release
	ReleaseTag     string // Release tag (defaults to version)
	ReleaseNotes   string // Release notes
	Force          bool   // Force push
	SkipValidation bool   // Skip manifest validation
}

// Publish publishes a collection to GitHub
func (m *Manager) Publish(ctx context.Context, collectionID string, options *PublishOptions) error {
	if options == nil {
		options = &PublishOptions{
			Branch: "main",
		}
	}

	// Get installation record
	record, exists := m.installer.GetInstalled(collectionID)
	if !exists {
		return fmt.Errorf("collection %s not installed", collectionID)
	}

	// Load manifest for validation
	manifestPath := filepath.Join(record.InstallLocation, "collection.yaml")
	manifestData, err := os.ReadFile(manifestPath)
	if err != nil {
		return fmt.Errorf("failed to read manifest: %w", err)
	}

	var manifest Manifest
	if err := yaml.Unmarshal(manifestData, &manifest); err != nil {
		return fmt.Errorf("failed to parse manifest: %w", err)
	}

	// Validate manifest
	if !options.SkipValidation {
		validator := NewValidator(record.InstallLocation)
		if err := validator.ValidateManifest(&manifest); err != nil {
			return fmt.Errorf("manifest validation failed: %w", err)
		}
	}

	// Determine GitHub repository
	repoPath := options.GitHubRepo
	if repoPath == "" {
		// Try to extract from manifest repository field
		if manifest.Repository != "" {
			// Parse GitHub URL
			repoPath = m.extractGitHubRepo(manifest.Repository)
		}
	}

	if repoPath == "" {
		return fmt.Errorf("GitHub repository not specified and not found in manifest")
	}

	// Check if directory is a git repository
	gitDir := filepath.Join(record.InstallLocation, ".git")
	isGitRepo := false
	if _, err := os.Stat(gitDir); err == nil {
		isGitRepo = true
	}

	// Initialize git repository if needed
	if !isGitRepo {
		if err := m.initGitRepo(ctx, record.InstallLocation, repoPath, options.Branch); err != nil {
			return fmt.Errorf("failed to initialize git repository: %w", err)
		}
	}

	// Stage all files
	if err := m.gitCommand(ctx, record.InstallLocation, "add", "."); err != nil {
		return fmt.Errorf("failed to stage files: %w", err)
	}

	// Commit changes
	commitMsg := options.CommitMessage
	if commitMsg == "" {
		commitMsg = fmt.Sprintf("Release version %s", manifest.Version)
	}

	if err := m.gitCommand(ctx, record.InstallLocation, "commit", "-m", commitMsg); err != nil {
		// Check if there are no changes to commit
		if !strings.Contains(err.Error(), "nothing to commit") {
			return fmt.Errorf("failed to commit changes: %w", err)
		}
	}

	// Push to GitHub
	pushArgs := []string{"push", "origin", options.Branch}
	if options.Force {
		pushArgs = append(pushArgs, "--force")
	}

	if err := m.gitCommand(ctx, record.InstallLocation, pushArgs...); err != nil {
		return fmt.Errorf("failed to push to GitHub: %w", err)
	}

	// Create release if requested
	if options.CreateRelease {
		releaseTag := options.ReleaseTag
		if releaseTag == "" {
			releaseTag = "v" + manifest.Version
		}

		if err := m.createGitHubRelease(ctx, record.InstallLocation, releaseTag, options.ReleaseNotes); err != nil {
			return fmt.Errorf("failed to create release: %w", err)
		}
	}

	return nil
}

// Helper functions

func (m *Manager) executeHooks(ctx context.Context, hooks []Hook, workDir string) error {
	for _, hook := range hooks {
		if hook.Type == "command" && hook.Command != "" {
			cmd := exec.CommandContext(ctx, "sh", "-c", hook.Command)
			cmd.Dir = workDir
			if output, err := cmd.CombinedOutput(); err != nil {
				return fmt.Errorf("hook command failed: %w\nOutput: %s", err, output)
			}
		}
	}
	return nil
}

func (m *Manager) shouldExclude(relPath string, info os.FileInfo, options *ExportOptions) bool {
	// Always exclude backups unless explicitly included
	if !options.IncludeBackups && strings.Contains(relPath, ".backup") {
		return true
	}

	// Exclude .git directory
	if strings.HasPrefix(relPath, ".git") {
		return true
	}

	// Check custom exclude patterns
	for _, pattern := range options.ExcludePatterns {
		if matched, _ := filepath.Match(pattern, filepath.Base(relPath)); matched {
			return true
		}
	}

	return false
}

func (m *Manager) extractGitHubRepo(repoURL string) string {
	// Extract owner/repo from various GitHub URL formats
	// e.g., https://github.com/owner/repo, git@github.com:owner/repo.git
	repoURL = strings.TrimSpace(repoURL)
	repoURL = strings.TrimSuffix(repoURL, ".git")

	if strings.Contains(repoURL, "github.com/") {
		parts := strings.Split(repoURL, "github.com/")
		if len(parts) == 2 {
			return parts[1]
		}
	}

	if strings.Contains(repoURL, "github.com:") {
		parts := strings.Split(repoURL, "github.com:")
		if len(parts) == 2 {
			return parts[1]
		}
	}

	return ""
}

func (m *Manager) initGitRepo(ctx context.Context, workDir, remote, branch string) error {
	// Initialize git repository
	if err := m.gitCommand(ctx, workDir, "init"); err != nil {
		return err
	}

	// Set default branch
	if err := m.gitCommand(ctx, workDir, "checkout", "-b", branch); err != nil {
		return err
	}

	// Add remote
	remoteURL := fmt.Sprintf("https://github.com/%s.git", remote)
	if err := m.gitCommand(ctx, workDir, "remote", "add", "origin", remoteURL); err != nil {
		return err
	}

	return nil
}

func (m *Manager) gitCommand(ctx context.Context, workDir string, args ...string) error {
	cmd := exec.CommandContext(ctx, "git", args...)
	cmd.Dir = workDir
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("git command failed: %w\nOutput: %s", err, output)
	}
	return nil
}

func (m *Manager) createGitHubRelease(ctx context.Context, workDir, tag, notes string) error {
	// Create tag
	if err := m.gitCommand(ctx, workDir, "tag", "-a", tag, "-m", notes); err != nil {
		return err
	}

	// Push tag
	if err := m.gitCommand(ctx, workDir, "push", "origin", tag); err != nil {
		return err
	}

	// Note: Creating the actual GitHub release would require GitHub API integration
	// This is a simplified version that just creates and pushes the tag
	// Full implementation would use the GitHub API to create a release with notes

	return nil
}

// UpdateOptions configures update behavior
type UpdateOptions struct {
	SkipDependencies bool
	SkipValidation   bool
	SkipHooks        bool
}
