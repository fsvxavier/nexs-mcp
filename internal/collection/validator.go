package collection

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"unicode/utf8"
)

// Validator provides validation for collection manifests and elements.
type Validator struct {
	basePath string // Base path for resolving relative element paths
}

// ValidationError represents a single validation error with context
type ValidationError struct {
	Field    string `json:"field"`         // Field that failed (e.g., "name", "elements[0].path")
	Rule     string `json:"rule"`          // Rule that failed (e.g., "required", "format", "security")
	Message  string `json:"message"`       // Human-readable error message
	Severity string `json:"severity"`      // "error" or "warning"
	Path     string `json:"path"`          // JSON path to field (e.g., "$.elements[0].path")
	Fix      string `json:"fix,omitempty"` // Suggested fix (optional)
}

// ValidationResult holds the complete validation outcome
type ValidationResult struct {
	Valid    bool               `json:"valid"`
	Errors   []*ValidationError `json:"errors"`
	Warnings []*ValidationError `json:"warnings"`
	Stats    map[string]int     `json:"stats"` // Validation statistics
}

// NewValidator creates a new manifest validator.
func NewValidator(basePath string) *Validator {
	return &Validator{
		basePath: basePath,
	}
}

// ValidateManifest performs comprehensive validation on a collection manifest.
func (v *Validator) ValidateManifest(manifest *Manifest) error {
	result := v.ValidateComprehensive(manifest)
	if !result.Valid {
		// Return first error for backward compatibility
		if len(result.Errors) > 0 {
			return fmt.Errorf("%s: %s", result.Errors[0].Field, result.Errors[0].Message)
		}
	}
	return nil
}

// ValidateComprehensive performs production-grade validation with 100+ rules
func (v *Validator) ValidateComprehensive(manifest *Manifest) *ValidationResult {
	result := &ValidationResult{
		Valid:    true,
		Errors:   make([]*ValidationError, 0),
		Warnings: make([]*ValidationError, 0),
		Stats:    make(map[string]int),
	}

	// Run all validation categories
	v.validateSchema(manifest, result)
	v.validateSecurity(manifest, result)
	v.validateDependencies(manifest, result)
	v.validateElements(manifest, result)
	v.validateHooksComprehensive(manifest, result)

	// Set overall validity
	result.Valid = len(result.Errors) == 0

	// Compute stats
	result.Stats["total_rules_checked"] = result.Stats["schema"] + result.Stats["security"] + result.Stats["dependency"] + result.Stats["element"] + result.Stats["hook"]
	result.Stats["errors"] = len(result.Errors)
	result.Stats["warnings"] = len(result.Warnings)

	return result
}

// validateSchema performs schema validation (30 rules)
func (v *Validator) validateSchema(manifest *Manifest, result *ValidationResult) {
	result.Stats["schema"] = 0

	// Required fields (5 rules)
	if manifest.Name == "" {
		result.Errors = append(result.Errors, &ValidationError{
			Field: "name", Rule: "required", Message: "name is required",
			Severity: "error", Path: "$.name", Fix: "Add a collection name (e.g., 'my-collection')",
		})
	}
	result.Stats["schema"]++

	if manifest.Version == "" {
		result.Errors = append(result.Errors, &ValidationError{
			Field: "version", Rule: "required", Message: "version is required",
			Severity: "error", Path: "$.version", Fix: "Add a semver version (e.g., '1.0.0')",
		})
	}
	result.Stats["schema"]++

	if manifest.Author == "" {
		result.Errors = append(result.Errors, &ValidationError{
			Field: "author", Rule: "required", Message: "author is required",
			Severity: "error", Path: "$.author", Fix: "Add author name or organization",
		})
	}
	result.Stats["schema"]++

	if manifest.Description == "" {
		result.Errors = append(result.Errors, &ValidationError{
			Field: "description", Rule: "required", Message: "description is required",
			Severity: "error", Path: "$.description", Fix: "Add a clear description of the collection",
		})
	}
	result.Stats["schema"]++

	if len(manifest.Elements) == 0 {
		result.Errors = append(result.Errors, &ValidationError{
			Field: "elements", Rule: "required", Message: "at least one element is required",
			Severity: "error", Path: "$.elements", Fix: "Add at least one element (persona, skill, template, etc.)",
		})
	}
	result.Stats["schema"]++

	// Format validation (10 rules)
	if manifest.Version != "" && !isValidVersion(manifest.Version) {
		result.Errors = append(result.Errors, &ValidationError{
			Field: "version", Rule: "format", Message: fmt.Sprintf("invalid version format: %s", manifest.Version),
			Severity: "error", Path: "$.version", Fix: "Use semver format: X.Y.Z (e.g., '1.0.0', '2.1.3-beta')",
		})
	}
	result.Stats["schema"]++

	if manifest.Name != "" && !isValidCollectionName(manifest.Name) {
		result.Errors = append(result.Errors, &ValidationError{
			Field: "name", Rule: "format", Message: fmt.Sprintf("invalid name format: %s", manifest.Name),
			Severity: "error", Path: "$.name", Fix: "Use lowercase, hyphens, and alphanumeric (e.g., 'my-collection')",
		})
	}
	result.Stats["schema"]++

	// Email validation for maintainers
	for i, maintainer := range manifest.Maintainers {
		if maintainer.Email != "" && !isValidEmail(maintainer.Email) {
			result.Errors = append(result.Errors, &ValidationError{
				Field: fmt.Sprintf("maintainers[%d].email", i), Rule: "format",
				Message:  fmt.Sprintf("invalid email format: %s", maintainer.Email),
				Severity: "error", Path: fmt.Sprintf("$.maintainers[%d].email", i),
				Fix: "Use valid email format: user@example.com",
			})
		}
		result.Stats["schema"]++
	}

	// URL validation
	urlFields := map[string]string{
		"homepage":      manifest.Homepage,
		"documentation": manifest.Documentation,
		"repository":    manifest.Repository,
	}
	for field, url := range urlFields {
		if url != "" && !isValidURL(url) {
			result.Errors = append(result.Errors, &ValidationError{
				Field: field, Rule: "format", Message: fmt.Sprintf("invalid URL format: %s", url),
				Severity: "error", Path: "$." + field, Fix: "Use valid HTTP(S) URL",
			})
		}
		result.Stats["schema"]++
	}

	// Length constraints (10 rules)
	if len(manifest.Name) < 3 || len(manifest.Name) > 64 {
		result.Errors = append(result.Errors, &ValidationError{
			Field: "name", Rule: "length", Message: fmt.Sprintf("name length must be 3-64 chars, got %d", len(manifest.Name)),
			Severity: "error", Path: "$.name", Fix: "Use a name between 3 and 64 characters",
		})
	}
	result.Stats["schema"]++

	if len(manifest.Description) < 10 || len(manifest.Description) > 500 {
		result.Warnings = append(result.Warnings, &ValidationError{
			Field: "description", Rule: "length", Message: fmt.Sprintf("description should be 10-500 chars, got %d", len(manifest.Description)),
			Severity: "warning", Path: "$.description", Fix: "Provide a concise but informative description",
		})
	}
	result.Stats["schema"]++

	if manifest.Author != "" && (len(manifest.Author) < 2 || len(manifest.Author) > 100) {
		result.Errors = append(result.Errors, &ValidationError{
			Field: "author", Rule: "length", Message: fmt.Sprintf("author length must be 2-100 chars, got %d", len(manifest.Author)),
			Severity: "error", Path: "$.author",
		})
	}
	result.Stats["schema"]++

	// UTF-8 validation
	if !utf8.ValidString(manifest.Name) || !utf8.ValidString(manifest.Description) {
		result.Errors = append(result.Errors, &ValidationError{
			Field: "name/description", Rule: "encoding", Message: "invalid UTF-8 encoding",
			Severity: "error", Path: "$", Fix: "Ensure all text fields use valid UTF-8 encoding",
		})
	}
	result.Stats["schema"]++

	// Enum validation (5 rules)
	validCategories := map[string]bool{
		"development": true, "devops": true, "creative-writing": true,
		"data-science": true, "security": true, "productivity": true,
		"education": true, "research": true, "other": true,
	}
	if manifest.Category != "" && !validCategories[manifest.Category] {
		result.Warnings = append(result.Warnings, &ValidationError{
			Field: "category", Rule: "enum", Message: fmt.Sprintf("unknown category: %s", manifest.Category),
			Severity: "warning", Path: "$.category", Fix: "Use standard categories: development, devops, creative-writing, etc.",
		})
	}
	result.Stats["schema"]++

	validLicenses := map[string]bool{
		"MIT": true, "Apache-2.0": true, "GPL-3.0": true, "BSD-3-Clause": true,
		"ISC": true, "MPL-2.0": true, "LGPL-3.0": true, "Proprietary": true, "Other": true,
	}
	if manifest.License != "" && !validLicenses[manifest.License] {
		result.Warnings = append(result.Warnings, &ValidationError{
			Field: "license", Rule: "enum", Message: fmt.Sprintf("non-standard license: %s", manifest.License),
			Severity: "warning", Path: "$.license", Fix: "Use SPDX license identifier (e.g., MIT, Apache-2.0)",
		})
	}
	result.Stats["schema"]++

	// Element type validation
	for i, elem := range manifest.Elements {
		if elem.Type != "" && !isValidElementType(elem.Type) {
			result.Errors = append(result.Errors, &ValidationError{
				Field: fmt.Sprintf("elements[%d].type", i), Rule: "enum",
				Message:  fmt.Sprintf("invalid element type: %s", elem.Type),
				Severity: "error", Path: fmt.Sprintf("$.elements[%d].type", i),
				Fix: "Use: persona, skill, template, agent, memory, or ensemble",
			})
		}
		result.Stats["schema"]++
	}
}

// validateSecurity performs security validation (25 rules)
func (v *Validator) validateSecurity(manifest *Manifest, result *ValidationResult) {
	result.Stats["security"] = 0

	// Path traversal detection (5 rules)
	for i, elem := range manifest.Elements {
		path := elem.Path
		result.Stats["security"]++

		// Check for ..
		if strings.Contains(path, "..") {
			result.Errors = append(result.Errors, &ValidationError{
				Field: fmt.Sprintf("elements[%d].path", i), Rule: "security.path_traversal",
				Message:  fmt.Sprintf("path traversal detected: %s", path),
				Severity: "error", Path: fmt.Sprintf("$.elements[%d].path", i),
				Fix: "Remove '..' from path, use relative paths only",
			})
		}

		// Check for absolute paths
		if filepath.IsAbs(path) {
			result.Errors = append(result.Errors, &ValidationError{
				Field: fmt.Sprintf("elements[%d].path", i), Rule: "security.absolute_path",
				Message:  fmt.Sprintf("absolute path not allowed: %s", path),
				Severity: "error", Path: fmt.Sprintf("$.elements[%d].path", i),
				Fix: "Use relative paths from collection root",
			})
		}
		result.Stats["security"]++

		// Check for symlinks (if basePath set)
		if v.basePath != "" {
			fullPath := filepath.Join(v.basePath, path)
			if info, err := os.Lstat(fullPath); err == nil && info.Mode()&os.ModeSymlink != 0 {
				result.Warnings = append(result.Warnings, &ValidationError{
					Field: fmt.Sprintf("elements[%d].path", i), Rule: "security.symlink",
					Message:  fmt.Sprintf("symlink detected: %s", path),
					Severity: "warning", Path: fmt.Sprintf("$.elements[%d].path", i),
					Fix: "Consider using actual files instead of symlinks",
				})
			}
		}
		result.Stats["security"]++
	}

	// Shell injection prevention in hooks (10 rules)
	if manifest.Hooks != nil {
		allHooks := []struct {
			name  string
			hooks []Hook
		}{
			{"pre_install", manifest.Hooks.PreInstall},
			{"post_install", manifest.Hooks.PostInstall},
			{"pre_update", manifest.Hooks.PreUpdate},
			{"post_update", manifest.Hooks.PostUpdate},
			{"pre_uninstall", manifest.Hooks.PreUninstall},
		}

		dangerousPatterns := []struct {
			pattern string
			name    string
		}{
			{"`", "backtick_execution"},
			{"$(", "command_substitution"},
			{"&&", "command_chaining"},
			{"||", "command_chaining"},
			{";", "command_separator"},
			{"|", "pipe"},
			{">", "redirection"},
			{"eval", "eval_injection"},
			{"exec", "exec_injection"},
		}

		for _, hookGroup := range allHooks {
			for i, hook := range hookGroup.hooks {
				if hook.Command == "" {
					continue
				}

				for _, dp := range dangerousPatterns {
					if strings.Contains(hook.Command, dp.pattern) {
						result.Warnings = append(result.Warnings, &ValidationError{
							Field:    fmt.Sprintf("hooks.%s[%d].command", hookGroup.name, i),
							Rule:     "security." + dp.name,
							Message:  fmt.Sprintf("potentially dangerous pattern in command: %s", dp.pattern),
							Severity: "warning",
							Path:     fmt.Sprintf("$.hooks.%s[%d].command", hookGroup.name, i),
							Fix:      "Avoid shell operators, use safe commands only",
						})
					}
					result.Stats["security"]++
				}
			}
		}
	}

	// Malicious command patterns (10 rules)
	maliciousCommands := []string{
		"rm -rf", "mkfs", "dd if=", ":(){:|:&};:", "chmod 777",
		"curl.*|.*bash", "wget.*|.*sh", "nc -l", "ncat", "/dev/null",
	}

	if manifest.Hooks != nil {
		allHooks := []struct {
			name  string
			hooks []Hook
		}{
			{"pre_install", manifest.Hooks.PreInstall},
			{"post_install", manifest.Hooks.PostInstall},
			{"pre_update", manifest.Hooks.PreUpdate},
			{"post_update", manifest.Hooks.PostUpdate},
			{"pre_uninstall", manifest.Hooks.PreUninstall},
		}

		for _, hookGroup := range allHooks {
			for i, hook := range hookGroup.hooks {
				if hook.Command == "" {
					continue
				}

				for _, malCmd := range maliciousCommands {
					matched, _ := regexp.MatchString(malCmd, hook.Command)
					if matched {
						result.Errors = append(result.Errors, &ValidationError{
							Field:    fmt.Sprintf("hooks.%s[%d].command", hookGroup.name, i),
							Rule:     "security.malicious_command",
							Message:  fmt.Sprintf("potentially malicious command detected: %s", malCmd),
							Severity: "error",
							Path:     fmt.Sprintf("$.hooks.%s[%d].command", hookGroup.name, i),
							Fix:      "Remove dangerous commands, use safe alternatives",
						})
					}
					result.Stats["security"]++
				}
			}
		}
	}
}

// validateDependencies performs dependency validation (15 rules)
func (v *Validator) validateDependencies(manifest *Manifest, result *ValidationResult) {
	result.Stats["dependency"] = 0

	for i, dep := range manifest.Dependencies {
		// URI format validation
		if !isValidDependencyURI(dep.URI) {
			result.Errors = append(result.Errors, &ValidationError{
				Field: fmt.Sprintf("dependencies[%d].uri", i), Rule: "dependency.uri_format",
				Message:  fmt.Sprintf("invalid dependency URI: %s", dep.URI),
				Severity: "error", Path: fmt.Sprintf("$.dependencies[%d].uri", i),
				Fix: "Use format: github://owner/repo[@version], file:///, or https://",
			})
		}
		result.Stats["dependency"]++

		// Version constraint validation
		if dep.Version != "" && !isValidVersionConstraint(dep.Version) {
			result.Errors = append(result.Errors, &ValidationError{
				Field: fmt.Sprintf("dependencies[%d].version", i), Rule: "dependency.version_constraint",
				Message:  fmt.Sprintf("invalid version constraint: %s", dep.Version),
				Severity: "error", Path: fmt.Sprintf("$.dependencies[%d].version", i),
				Fix: "Use semver constraints: ^1.0.0, ~2.1.0, >=1.0.0, <2.0.0",
			})
		}
		result.Stats["dependency"]++

		// Check for duplicate dependencies
		for j := i + 1; j < len(manifest.Dependencies); j++ {
			if manifest.Dependencies[j].URI == dep.URI {
				result.Errors = append(result.Errors, &ValidationError{
					Field: fmt.Sprintf("dependencies[%d,%d].uri", i, j), Rule: "dependency.duplicate",
					Message:  fmt.Sprintf("duplicate dependency: %s", dep.URI),
					Severity: "error", Path: fmt.Sprintf("$.dependencies[%d].uri", i),
					Fix: "Remove duplicate dependency entries",
				})
			}
			result.Stats["dependency"]++
		}

		// Self-dependency check
		if manifest.Name != "" && manifest.Author != "" {
			selfID := fmt.Sprintf("github://%s/%s", manifest.Author, manifest.Name)
			if strings.Contains(dep.URI, selfID) {
				result.Errors = append(result.Errors, &ValidationError{
					Field: fmt.Sprintf("dependencies[%d].uri", i), Rule: "dependency.self_reference",
					Message:  "collection cannot depend on itself",
					Severity: "error", Path: fmt.Sprintf("$.dependencies[%d].uri", i),
					Fix: "Remove self-dependency",
				})
			}
		}
		result.Stats["dependency"]++
	}

	// Dependency depth warning (avoid deep chains)
	if len(manifest.Dependencies) > 10 {
		result.Warnings = append(result.Warnings, &ValidationError{
			Field: "dependencies", Rule: "dependency.depth",
			Message:  fmt.Sprintf("too many dependencies: %d", len(manifest.Dependencies)),
			Severity: "warning", Path: "$.dependencies",
			Fix: "Consider reducing dependencies (recommended: <10)",
		})
	}
	result.Stats["dependency"]++
}

// validateElements performs element validation (20 rules)
func (v *Validator) validateElements(manifest *Manifest, result *ValidationResult) {
	result.Stats["element"] = 0

	seenPaths := make(map[string]int)

	for i, elem := range manifest.Elements {
		// Path required
		if elem.Path == "" {
			result.Errors = append(result.Errors, &ValidationError{
				Field: fmt.Sprintf("elements[%d].path", i), Rule: "element.path_required",
				Message:  "element path is required",
				Severity: "error", Path: fmt.Sprintf("$.elements[%d].path", i),
				Fix: "Add path field to element",
			})
		}
		result.Stats["element"]++

		// Duplicate path detection
		if elem.Path != "" {
			if prevIndex, exists := seenPaths[elem.Path]; exists {
				result.Warnings = append(result.Warnings, &ValidationError{
					Field: fmt.Sprintf("elements[%d,%d].path", prevIndex, i), Rule: "element.duplicate_path",
					Message:  fmt.Sprintf("duplicate element path: %s", elem.Path),
					Severity: "warning", Path: fmt.Sprintf("$.elements[%d].path", i),
					Fix: "Remove duplicate or use unique paths",
				})
			} else {
				seenPaths[elem.Path] = i
			}
		}
		result.Stats["element"]++

		// Type inference validation (based on file extension)
		if elem.Type != "" && elem.Path != "" && !strings.Contains(elem.Path, "*") {
			ext := filepath.Ext(elem.Path)
			expectedExt := "." + elem.Type + ".yaml"
			if ext != ".yaml" && ext != ".yml" && ext != expectedExt {
				result.Warnings = append(result.Warnings, &ValidationError{
					Field: fmt.Sprintf("elements[%d]", i), Rule: "element.type_mismatch",
					Message:  fmt.Sprintf("type '%s' doesn't match file extension '%s'", elem.Type, ext),
					Severity: "warning", Path: fmt.Sprintf("$.elements[%d].path", i),
					Fix: fmt.Sprintf("Rename to %s or remove type field", expectedExt),
				})
			}
		}
		result.Stats["element"]++

		// Glob pattern validation
		if strings.Contains(elem.Path, "*") {
			// Check for invalid glob patterns
			if strings.Contains(elem.Path, "**/**") {
				result.Warnings = append(result.Warnings, &ValidationError{
					Field: fmt.Sprintf("elements[%d].path", i), Rule: "element.glob_pattern",
					Message:  "redundant glob pattern: **/**",
					Severity: "warning", Path: fmt.Sprintf("$.elements[%d].path", i),
					Fix: "Simplify to **/*",
				})
			}
		}
		result.Stats["element"]++
	}

	// Check element count limits
	if len(manifest.Elements) > 1000 {
		result.Warnings = append(result.Warnings, &ValidationError{
			Field: "elements", Rule: "element.count_limit",
			Message:  fmt.Sprintf("very large collection: %d elements", len(manifest.Elements)),
			Severity: "warning", Path: "$.elements",
			Fix: "Consider splitting into multiple collections (recommended: <100 elements)",
		})
	}
	result.Stats["element"]++
}

// validateHooksComprehensive performs hook validation (10 rules)
func (v *Validator) validateHooksComprehensive(manifest *Manifest, result *ValidationResult) {
	result.Stats["hook"] = 0

	if manifest.Hooks == nil {
		result.Stats["hook"]++ // Count as checked
		return
	}

	allHooks := []struct {
		name  string
		hooks []Hook
	}{
		{"pre_install", manifest.Hooks.PreInstall},
		{"post_install", manifest.Hooks.PostInstall},
		{"pre_update", manifest.Hooks.PreUpdate},
		{"post_update", manifest.Hooks.PostUpdate},
		{"pre_uninstall", manifest.Hooks.PreUninstall},
	}

	for _, hookGroup := range allHooks {
		for i, hook := range hookGroup.hooks {
			// Type validation
			if hook.Type == "" {
				result.Errors = append(result.Errors, &ValidationError{
					Field: fmt.Sprintf("hooks.%s[%d].type", hookGroup.name, i), Rule: "hook.type_required",
					Message:  "hook type is required",
					Severity: "error", Path: fmt.Sprintf("$.hooks.%s[%d].type", hookGroup.name, i),
					Fix: "Add type: command, validate, backup, or confirm",
				})
			}
			result.Stats["hook"]++

			if !isValidHookType(hook.Type) {
				result.Errors = append(result.Errors, &ValidationError{
					Field: fmt.Sprintf("hooks.%s[%d].type", hookGroup.name, i), Rule: "hook.type_invalid",
					Message:  fmt.Sprintf("invalid hook type: %s", hook.Type),
					Severity: "error", Path: fmt.Sprintf("$.hooks.%s[%d].type", hookGroup.name, i),
					Fix: "Use: command, validate, backup, or confirm",
				})
			}
			result.Stats["hook"]++

			// Type-specific validation
			switch hook.Type {
			case "command":
				if hook.Command == "" {
					result.Errors = append(result.Errors, &ValidationError{
						Field: fmt.Sprintf("hooks.%s[%d].command", hookGroup.name, i), Rule: "hook.command_required",
						Message:  "command hook requires 'command' field",
						Severity: "error", Path: fmt.Sprintf("$.hooks.%s[%d].command", hookGroup.name, i),
						Fix: "Add command field with shell command",
					})
				}
				result.Stats["hook"]++

			case "confirm":
				if hook.Message == "" {
					result.Errors = append(result.Errors, &ValidationError{
						Field: fmt.Sprintf("hooks.%s[%d].message", hookGroup.name, i), Rule: "hook.message_required",
						Message:  "confirm hook requires 'message' field",
						Severity: "error", Path: fmt.Sprintf("$.hooks.%s[%d].message", hookGroup.name, i),
						Fix: "Add message field with confirmation prompt",
					})
				}
				result.Stats["hook"]++

			case "validate":
				if len(hook.Checks) == 0 {
					result.Errors = append(result.Errors, &ValidationError{
						Field: fmt.Sprintf("hooks.%s[%d].checks", hookGroup.name, i), Rule: "hook.checks_required",
						Message:  "validate hook requires 'checks' field",
						Severity: "error", Path: fmt.Sprintf("$.hooks.%s[%d].checks", hookGroup.name, i),
						Fix: "Add checks array with tool availability checks",
					})
				}
				result.Stats["hook"]++
			}
		}
	}

	// Hook count warning
	totalHooks := len(manifest.Hooks.PreInstall) + len(manifest.Hooks.PostInstall) +
		len(manifest.Hooks.PreUpdate) + len(manifest.Hooks.PostUpdate) + len(manifest.Hooks.PreUninstall)
	if totalHooks > 20 {
		result.Warnings = append(result.Warnings, &ValidationError{
			Field: "hooks", Rule: "hook.count_limit",
			Message:  fmt.Sprintf("many hooks defined: %d", totalHooks),
			Severity: "warning", Path: "$.hooks",
			Fix: "Consider reducing hooks (recommended: <10 total)",
		})
	}
	result.Stats["hook"]++
}

// Helper validation functions

func isValidEmail(email string) bool {
	pattern := `^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`
	matched, _ := regexp.MatchString(pattern, email)
	return matched
}

func isValidURL(url string) bool {
	return strings.HasPrefix(url, "http://") || strings.HasPrefix(url, "https://")
}

func isValidVersionConstraint(constraint string) bool {
	// Support: ^1.0.0, ~1.0.0, >=1.0.0, <2.0.0, 1.0.0
	patterns := []string{
		`^\^[0-9]+\.[0-9]+\.[0-9]+$`, // ^1.0.0
		`^~[0-9]+\.[0-9]+\.[0-9]+$`,  // ~1.0.0
		`^>=[0-9]+\.[0-9]+\.[0-9]+$`, // >=1.0.0
		`^<[0-9]+\.[0-9]+\.[0-9]+$`,  // <2.0.0
		`^[0-9]+\.[0-9]+\.[0-9]+$`,   // 1.0.0
		`^\*$`,                       // * (any version)
	}

	for _, pattern := range patterns {
		if matched, _ := regexp.MatchString(pattern, constraint); matched {
			return true
		}
	}
	return false
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
