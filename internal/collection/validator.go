package collection

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

// Validator provides validation for collection manifests and elements.
type Validator struct {
	basePath string // Base path for resolving relative element paths
}

// NewValidator creates a new manifest validator.
func NewValidator(basePath string) *Validator {
	return &Validator{
		basePath: basePath,
	}
}

// ValidateManifest performs comprehensive validation on a collection manifest.
func (v *Validator) ValidateManifest(manifest *Manifest) error {
	// Basic validation (required fields)
	if err := manifest.Validate(); err != nil {
		return err
	}

	// Validate version format (semver-ish)
	if !isValidVersion(manifest.Version) {
		return fmt.Errorf("invalid version format: %s (expected semver: X.Y.Z)", manifest.Version)
	}

	// Validate name format (lowercase, hyphens, alphanumeric)
	if !isValidCollectionName(manifest.Name) {
		return fmt.Errorf("invalid name format: %s (use lowercase, hyphens, and alphanumeric characters)", manifest.Name)
	}

	// Validate element types if specified
	for i, elem := range manifest.Elements {
		if elem.Type != "" && !isValidElementType(elem.Type) {
			return fmt.Errorf("invalid element type at index %d: %s (must be: persona, skill, template, agent, memory, or ensemble)", i, elem.Type)
		}
	}

	// Validate dependency URIs
	for i, dep := range manifest.Dependencies {
		if !isValidDependencyURI(dep.URI) {
			return fmt.Errorf("invalid dependency URI at index %d: %s (expected format: github://owner/repo[@version], file:///, or https://)", i, dep.URI)
		}
	}

	// Validate hooks
	if manifest.Hooks != nil {
		if err := v.validateHooks(manifest.Hooks); err != nil {
			return fmt.Errorf("invalid hooks: %w", err)
		}
	}

	return nil
}

// ValidateElements checks that all element paths exist and are accessible.
// This is called during installation to ensure collection integrity.
func (v *Validator) ValidateElements(manifest *Manifest) error {
	if v.basePath == "" {
		return fmt.Errorf("base path not set for element validation")
	}

	for i, elem := range manifest.Elements {
		// Resolve path relative to base path
		fullPath := filepath.Join(v.basePath, elem.Path)

		// Check if path contains glob pattern
		if strings.Contains(elem.Path, "*") {
			// Glob pattern - ensure at least one match
			matches, err := filepath.Glob(fullPath)
			if err != nil {
				return fmt.Errorf("invalid glob pattern at element[%d]: %s (%w)", i, elem.Path, err)
			}
			if len(matches) == 0 {
				return fmt.Errorf("no files match glob pattern at element[%d]: %s", i, elem.Path)
			}
		} else {
			// Single file - check existence
			if _, err := os.Stat(fullPath); os.IsNotExist(err) {
				return fmt.Errorf("element file not found at element[%d]: %s", i, elem.Path)
			} else if err != nil {
				return fmt.Errorf("error accessing element file at element[%d]: %s (%w)", i, elem.Path, err)
			}
		}
	}

	return nil
}

// validateHooks validates hook configuration.
func (v *Validator) validateHooks(hooks *Hooks) error {
	allHooks := [][]Hook{
		hooks.PreInstall,
		hooks.PostInstall,
		hooks.PreUpdate,
		hooks.PostUpdate,
		hooks.PreUninstall,
	}

	for _, hookList := range allHooks {
		for i, hook := range hookList {
			if hook.Type == "" {
				return fmt.Errorf("hook[%d] missing type", i)
			}
			if !isValidHookType(hook.Type) {
				return fmt.Errorf("hook[%d] invalid type: %s (must be: command, validate, backup, or confirm)", i, hook.Type)
			}

			// Type-specific validation
			switch hook.Type {
			case "command":
				if hook.Command == "" {
					return fmt.Errorf("hook[%d] type 'command' requires 'command' field", i)
				}
			case "confirm":
				if hook.Message == "" {
					return fmt.Errorf("hook[%d] type 'confirm' requires 'message' field", i)
				}
			case "validate":
				if len(hook.Checks) == 0 {
					return fmt.Errorf("hook[%d] type 'validate' requires 'checks' field", i)
				}
			}
		}
	}

	return nil
}

// isValidVersion checks if version string follows semver format (X.Y.Z).
func isValidVersion(version string) bool {
	// Relaxed semver: major.minor.patch with optional -suffix
	pattern := `^[0-9]+\.[0-9]+\.[0-9]+(-[a-zA-Z0-9.-]+)?$`
	matched, _ := regexp.MatchString(pattern, version)
	return matched
}

// isValidCollectionName checks if collection name is valid.
func isValidCollectionName(name string) bool {
	// Lowercase, hyphens, alphanumeric only
	pattern := `^[a-z0-9][a-z0-9-]*[a-z0-9]$`
	matched, _ := regexp.MatchString(pattern, name)
	return matched
}

// isValidElementType checks if element type is recognized.
func isValidElementType(elementType string) bool {
	validTypes := map[string]bool{
		"persona":  true,
		"skill":    true,
		"template": true,
		"agent":    true,
		"memory":   true,
		"ensemble": true,
	}
	return validTypes[elementType]
}

// isValidDependencyURI checks if dependency URI follows supported formats.
func isValidDependencyURI(uri string) bool {
	// github://owner/repo[@version]
	if strings.HasPrefix(uri, "github://") {
		return true
	}
	// file:///path/to/collection
	if strings.HasPrefix(uri, "file://") {
		return true
	}
	// https://example.com/collection.tar.gz
	if strings.HasPrefix(uri, "https://") || strings.HasPrefix(uri, "http://") {
		return true
	}
	return false
}

// isValidHookType checks if hook type is recognized.
func isValidHookType(hookType string) bool {
	validTypes := map[string]bool{
		"command":  true,
		"validate": true,
		"backup":   true,
		"confirm":  true,
	}
	return validTypes[hookType]
}
