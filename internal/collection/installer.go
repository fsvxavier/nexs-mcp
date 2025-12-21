package collection

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/fsvxavier/nexs-mcp/internal/collection/sources"
)

// InstallationRecord tracks an installed collection.
type InstallationRecord struct {
	ID              string                 `json:"id"`               // author/name
	Version         string                 `json:"version"`          // Installed version
	URI             string                 `json:"uri"`              // Source URI
	SourceName      string                 `json:"source_name"`      // Source that provided it
	InstalledAt     time.Time              `json:"installed_at"`     // Installation timestamp
	InstallLocation string                 `json:"install_location"` // Where it was installed
	Dependencies    []string               `json:"dependencies"`     // Installed dependencies (IDs)
	Metadata        map[string]interface{} `json:"metadata"`         // Collection metadata
}

// Installer handles collection installation, updates, and removal.
type Installer struct {
	registry       *Registry
	installDir     string
	stateFile      string
	installations  map[string]*InstallationRecord // ID -> Record
	backupDir      string
	skipValidation bool
	skipHooks      bool
}

// NewInstaller creates a new collection installer.
func NewInstaller(registry *Registry, installDir string) (*Installer, error) {
	if installDir == "" {
		homeDir, err := os.UserHomeDir()
		if err != nil {
			return nil, fmt.Errorf("failed to get home directory: %w", err)
		}
		installDir = filepath.Join(homeDir, ".nexs", "collections")
	}

	// Create installation directory
	if err := os.MkdirAll(installDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create install directory: %w", err)
	}

	stateFile := filepath.Join(installDir, ".installed.json")
	backupDir := filepath.Join(installDir, ".backups")

	installer := &Installer{
		registry:      registry,
		installDir:    installDir,
		stateFile:     stateFile,
		installations: make(map[string]*InstallationRecord),
		backupDir:     backupDir,
	}

	// Load existing installations
	if err := installer.loadState(); err != nil {
		// If state file doesn't exist, that's ok
		if !os.IsNotExist(err) {
			return nil, fmt.Errorf("failed to load installation state: %w", err)
		}
	}

	return installer, nil
}

// Install installs a collection and its dependencies.
func (i *Installer) Install(ctx context.Context, uri string, options *InstallOptions) error {
	if options == nil {
		options = &InstallOptions{}
	}

	// Get collection from registry
	collection, err := i.registry.Get(ctx, uri)
	if err != nil {
		return fmt.Errorf("failed to get collection: %w", err)
	}

	// Extract manifest
	manifestMap, ok := collection.Manifest.(map[string]interface{})
	if !ok {
		return errors.New("invalid manifest type")
	}

	// Parse manifest
	manifest, err := parseManifestFromMap(manifestMap)
	if err != nil {
		return fmt.Errorf("failed to parse manifest: %w", err)
	}

	// Check if already installed
	collectionID := manifest.ID()
	if record, exists := i.installations[collectionID]; exists {
		if !options.Force {
			return fmt.Errorf("collection %s@%s already installed", collectionID, record.Version)
		}
		// Uninstall existing version first
		if err := i.Uninstall(ctx, collectionID, &UninstallOptions{Force: true}); err != nil {
			return fmt.Errorf("failed to uninstall existing version: %w", err)
		}
	}

	// Resolve dependencies
	depGraph, err := i.resolveDependencies(ctx, manifest, options)
	if err != nil {
		return fmt.Errorf("dependency resolution failed: %w", err)
	}

	// Install dependencies first (in topological order)
	for _, depURI := range depGraph {
		if depURI == uri {
			continue // Skip self
		}
		if err := i.Install(ctx, depURI, &InstallOptions{SkipDependencies: true}); err != nil {
			return fmt.Errorf("failed to install dependency %s: %w", depURI, err)
		}
	}

	// Perform atomic installation
	if err := i.atomicInstall(ctx, collection, manifest, options); err != nil {
		return fmt.Errorf("installation failed: %w", err)
	}

	return nil
}

// atomicInstall performs atomic installation with rollback capability.
func (i *Installer) atomicInstall(ctx context.Context, collection *sources.Collection, manifest *Manifest, options *InstallOptions) error {
	collectionID := manifest.ID()
	tempDir, err := os.MkdirTemp(i.installDir, ".installing-*")
	if err != nil {
		return fmt.Errorf("failed to create temp directory: %w", err)
	}
	defer func() {
		_ = os.RemoveAll(tempDir) // Cleanup temp directory
	}()

	// Get source path
	sourcePath := ""
	if path, ok := collection.SourceData["path"].(string); ok {
		sourcePath = path
	}
	if sourcePath == "" {
		return errors.New("collection source path not found")
	}

	// Copy collection files to temp directory
	if err := i.copyDirectory(sourcePath, tempDir); err != nil {
		return fmt.Errorf("failed to copy collection files: %w", err)
	}

	// Validate collection
	if !options.SkipValidation && !i.skipValidation {
		validator := NewValidator(tempDir)
		if err := validator.ValidateManifest(manifest); err != nil {
			return fmt.Errorf("validation failed: %w", err)
		}
		if err := validator.ValidateElements(manifest); err != nil {
			return fmt.Errorf("element validation failed: %w", err)
		}
	}

	// Execute pre-install hooks
	if !options.SkipHooks && !i.skipHooks && manifest.Hooks != nil {
		if err := i.executeHooks(ctx, manifest.Hooks.PreInstall, tempDir); err != nil {
			return fmt.Errorf("pre-install hook failed: %w", err)
		}
	}

	// Create backup if overwriting
	finalPath := filepath.Join(i.installDir, manifest.Author, manifest.Name)
	if _, err := os.Stat(finalPath); err == nil {
		if err := i.createBackup(finalPath, collectionID); err != nil {
			return fmt.Errorf("failed to create backup: %w", err)
		}
	}

	// Move to final location (atomic operation)
	if err := os.MkdirAll(filepath.Dir(finalPath), 0755); err != nil {
		return fmt.Errorf("failed to create parent directory: %w", err)
	}

	// Remove existing if present
	_ = os.RemoveAll(finalPath) // Best effort cleanup

	if err := os.Rename(tempDir, finalPath); err != nil {
		// Rollback: restore from backup if it exists
		_ = i.restoreFromBackup(collectionID) // Best effort restore
		return fmt.Errorf("failed to move to final location: %w", err)
	}

	// Execute post-install hooks
	if !options.SkipHooks && !i.skipHooks && manifest.Hooks != nil {
		if err := i.executeHooks(ctx, manifest.Hooks.PostInstall, finalPath); err != nil {
			// Rollback installation
			_ = os.RemoveAll(finalPath)           // Best effort cleanup
			_ = i.restoreFromBackup(collectionID) // Best effort restore
			return fmt.Errorf("post-install hook failed: %w", err)
		}
	}

	// Record installation
	record := &InstallationRecord{
		ID:              collectionID,
		Version:         manifest.Version,
		URI:             collection.Metadata.URI,
		SourceName:      collection.Metadata.SourceName,
		InstalledAt:     time.Now(),
		InstallLocation: finalPath,
		Dependencies:    i.extractDependencyIDsFromManifest(ctx, manifest),
		Metadata: map[string]interface{}{
			"name":        manifest.Name,
			"author":      manifest.Author,
			"description": manifest.Description,
			"category":    manifest.Category,
			"tags":        manifest.Tags,
		},
	}

	i.installations[collectionID] = record
	if err := i.saveState(); err != nil {
		return fmt.Errorf("failed to save installation state: %w", err)
	}

	return nil
}

// Uninstall removes a collection.
func (i *Installer) Uninstall(ctx context.Context, collectionID string, options *UninstallOptions) error {
	if options == nil {
		options = &UninstallOptions{}
	}

	record, exists := i.installations[collectionID]
	if !exists {
		return fmt.Errorf("collection %s not installed", collectionID)
	}

	// Check for dependent collections
	if !options.Force {
		dependents := i.findDependents(collectionID)
		if len(dependents) > 0 {
			return fmt.Errorf("collection is required by: %v (use force to uninstall)", dependents)
		}
	}

	// Remove installation directory
	if err := os.RemoveAll(record.InstallLocation); err != nil {
		return fmt.Errorf("failed to remove installation: %w", err)
	}

	// Remove from state
	delete(i.installations, collectionID)
	if err := i.saveState(); err != nil {
		return fmt.Errorf("failed to save state: %w", err)
	}

	return nil
}

// ListInstalled returns all installed collections.
func (i *Installer) ListInstalled() []*InstallationRecord {
	records := make([]*InstallationRecord, 0, len(i.installations))
	for _, record := range i.installations {
		records = append(records, record)
	}
	return records
}

// GetInstalled returns a specific installation record.
func (i *Installer) GetInstalled(collectionID string) (*InstallationRecord, bool) {
	record, exists := i.installations[collectionID]
	return record, exists
}

// resolveDependencies resolves all dependencies and returns installation order.
func (i *Installer) resolveDependencies(ctx context.Context, manifest *Manifest, options *InstallOptions) ([]string, error) {
	if options.SkipDependencies {
		return []string{}, nil
	}

	visited := make(map[string]bool)
	result := []string{}

	var resolve func(m *Manifest, uri string) error
	resolve = func(m *Manifest, uri string) error {
		id := m.ID()
		if visited[id] {
			return nil
		}
		visited[id] = true

		// Process dependencies
		for _, dep := range m.Dependencies {
			if dep.URI == "" {
				continue
			}

			// Get dependency collection
			depCollection, err := i.registry.Get(ctx, dep.URI)
			if err != nil {
				return fmt.Errorf("failed to get dependency %s: %w", dep.URI, err)
			}

			depManifestMap, ok := depCollection.Manifest.(map[string]interface{})
			if !ok {
				return fmt.Errorf("invalid manifest type for dependency %s", dep.URI)
			}

			depManifest, err := parseManifestFromMap(depManifestMap)
			if err != nil {
				return fmt.Errorf("failed to parse dependency manifest: %w", err)
			}

			depID := depManifest.ID()

			// Check if already installed
			if _, installed := i.installations[depID]; installed {
				continue
			}

			// Recursively resolve
			if err := resolve(depManifest, dep.URI); err != nil {
				return err
			}

			result = append(result, dep.URI)
		}

		return nil
	}

	if err := resolve(manifest, ""); err != nil {
		return nil, err
	}

	return result, nil
}

// executeHooks executes a list of hooks.
func (i *Installer) executeHooks(ctx context.Context, hooks []Hook, workDir string) error {
	for _, hook := range hooks {
		if hook.Type == "command" && hook.Command != "" {
			//nolint:gosec // G204: Hook commands are from trusted collection manifests
			cmd := exec.CommandContext(ctx, "sh", "-c", hook.Command)
			cmd.Dir = workDir
			if output, err := cmd.CombinedOutput(); err != nil {
				return fmt.Errorf("hook command failed: %w\nOutput: %s", err, output)
			}
		}
	}
	return nil
}

// copyDirectory recursively copies a directory.
func (i *Installer) copyDirectory(src, dst string) error {
	return filepath.Walk(src, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Get relative path
		relPath, err := filepath.Rel(src, path)
		if err != nil {
			return err
		}

		// Skip hidden files and directories (except .nexs)
		if strings.HasPrefix(filepath.Base(path), ".") && !strings.HasPrefix(filepath.Base(path), ".nexs") {
			if info.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}

		dstPath := filepath.Join(dst, relPath)

		if info.IsDir() {
			return os.MkdirAll(dstPath, info.Mode())
		}

		// Copy file
		return i.copyFile(path, dstPath, info.Mode())
	})
}

// copyFile copies a single file.
func (i *Installer) copyFile(src, dst string, mode os.FileMode) error {
	srcFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer func() {
		_ = srcFile.Close() // Ignore close error on read operation
	}()

	dstFile, err := os.OpenFile(dst, os.O_RDWR|os.O_CREATE|os.O_TRUNC, mode)
	if err != nil {
		return err
	}
	defer func() {
		if cerr := dstFile.Close(); cerr != nil && err == nil {
			err = cerr
		}
	}()

	_, err = io.Copy(dstFile, srcFile)
	return err
}

// createBackup creates a backup of a collection.
func (i *Installer) createBackup(path, collectionID string) error {
	if err := os.MkdirAll(i.backupDir, 0755); err != nil {
		return err
	}

	backupPath := filepath.Join(i.backupDir, fmt.Sprintf("%s-%d", collectionID, time.Now().Unix()))
	return i.copyDirectory(path, backupPath)
}

// restoreFromBackup restores a collection from backup.
func (i *Installer) restoreFromBackup(collectionID string) error {
	// Find latest backup
	entries, err := os.ReadDir(i.backupDir)
	if err != nil {
		return err
	}

	var latestBackup string
	var latestTime int64

	for _, entry := range entries {
		if strings.HasPrefix(entry.Name(), collectionID+"-") {
			parts := strings.Split(entry.Name(), "-")
			if len(parts) >= 2 {
				var timestamp int64
				if _, err := fmt.Sscanf(parts[len(parts)-1], "%d", &timestamp); err == nil {
					if timestamp > latestTime {
						latestTime = timestamp
						latestBackup = entry.Name()
					}
				}
			}
		}
	}

	if latestBackup == "" {
		return fmt.Errorf("no backup found for %s", collectionID)
	}

	backupPath := filepath.Join(i.backupDir, latestBackup)
	record := i.installations[collectionID]
	if record == nil {
		return errors.New("no installation record found")
	}

	return i.copyDirectory(backupPath, record.InstallLocation)
}

// findDependents finds collections that depend on the given collection.
func (i *Installer) findDependents(collectionID string) []string {
	dependents := []string{}
	for id, record := range i.installations {
		for _, depID := range record.Dependencies {
			if depID == collectionID {
				dependents = append(dependents, id)
				break
			}
		}
	}
	return dependents
}

// extractDependencyIDsFromManifest extracts dependency IDs from manifest by resolving URIs.
func (i *Installer) extractDependencyIDsFromManifest(ctx context.Context, manifest *Manifest) []string {
	ids := make([]string, 0, len(manifest.Dependencies))
	for _, dep := range manifest.Dependencies {
		if dep.URI != "" {
			// Try to get the collection to resolve its ID
			collection, err := i.registry.Get(ctx, dep.URI)
			if err == nil && collection != nil {
				if manifestMap, ok := collection.Manifest.(map[string]interface{}); ok {
					if depManifest, err := parseManifestFromMap(manifestMap); err == nil {
						ids = append(ids, depManifest.ID())
						continue
					}
				}
			}
			// Fallback: store URI if we can't resolve
			ids = append(ids, dep.URI)
		}
	}
	return ids
}

// loadState loads installation state from disk.
func (i *Installer) loadState() error {
	data, err := os.ReadFile(i.stateFile)
	if err != nil {
		return err
	}

	return parseJSONToMap(data, &i.installations)
}

// saveState saves installation state to disk.
func (i *Installer) saveState() error {
	data, err := marshalToJSON(i.installations)
	if err != nil {
		return err
	}

	return os.WriteFile(i.stateFile, data, 0644)
}

// InstallOptions configures installation behavior.
type InstallOptions struct {
	Force            bool // Force reinstallation
	SkipDependencies bool // Skip dependency installation
	SkipValidation   bool // Skip validation
	SkipHooks        bool // Skip hook execution
}

// UninstallOptions configures uninstallation behavior.
type UninstallOptions struct {
	Force bool // Force uninstall even if depended upon
}

// Helper functions for JSON handling (simplified).
func parseManifestFromMap(m map[string]interface{}) (*Manifest, error) {
	manifest := &Manifest{}

	if name, ok := m["name"].(string); ok {
		manifest.Name = name
	}
	if version, ok := m["version"].(string); ok {
		manifest.Version = version
	}
	if author, ok := m["author"].(string); ok {
		manifest.Author = author
	}
	if description, ok := m["description"].(string); ok {
		manifest.Description = description
	}
	if category, ok := m["category"].(string); ok {
		manifest.Category = category
	}

	if tags, ok := m["tags"].([]interface{}); ok {
		manifest.Tags = make([]string, 0, len(tags))
		for _, t := range tags {
			if s, ok := t.(string); ok {
				manifest.Tags = append(manifest.Tags, s)
			}
		}
	}

	// Parse dependencies
	if deps, ok := m["dependencies"].([]interface{}); ok {
		manifest.Dependencies = make([]Dependency, 0, len(deps))
		for _, d := range deps {
			if depMap, ok := d.(map[string]interface{}); ok {
				dep := Dependency{}
				if uri, ok := depMap["uri"].(string); ok {
					dep.URI = uri
				}
				if version, ok := depMap["version"].(string); ok {
					dep.Version = version
				}
				manifest.Dependencies = append(manifest.Dependencies, dep)
			}
		}
	}

	// Parse elements
	if elements, ok := m["elements"].([]interface{}); ok {
		manifest.Elements = make([]Element, 0, len(elements))
		for _, e := range elements {
			if elemMap, ok := e.(map[string]interface{}); ok {
				elem := Element{}
				if elemType, ok := elemMap["type"].(string); ok {
					elem.Type = elemType
				}
				if path, ok := elemMap["path"].(string); ok {
					elem.Path = path
				}
				manifest.Elements = append(manifest.Elements, elem)
			}
		}
	}

	// Parse hooks
	if hooksMap, ok := m["hooks"].(map[string]interface{}); ok {
		manifest.Hooks = &Hooks{}
		if preInstall, ok := hooksMap["pre_install"].([]interface{}); ok {
			manifest.Hooks.PreInstall = parseHooks(preInstall)
		}
		if postInstall, ok := hooksMap["post_install"].([]interface{}); ok {
			manifest.Hooks.PostInstall = parseHooks(postInstall)
		}
	}

	return manifest, nil
}

func parseHooks(hooks []interface{}) []Hook {
	result := make([]Hook, 0, len(hooks))
	for _, h := range hooks {
		if hookMap, ok := h.(map[string]interface{}); ok {
			hook := Hook{}
			if hookType, ok := hookMap["type"].(string); ok {
				hook.Type = hookType
			}
			if command, ok := hookMap["command"].(string); ok {
				hook.Command = command
			}
			result = append(result, hook)
		}
	}
	return result
}

func parseJSONToMap(data []byte, target interface{}) error {
	return json.Unmarshal(data, target)
}

func marshalToJSON(data interface{}) ([]byte, error) {
	return json.MarshalIndent(data, "", "  ")
}
