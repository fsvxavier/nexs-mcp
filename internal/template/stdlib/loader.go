package stdlib

import (
	"embed"
	"fmt"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"

	"github.com/fsvxavier/nexs-mcp/internal/domain"
)

//go:embed templates/*.yaml
var templatesFS embed.FS

// StandardLibrary manages built-in templates
type StandardLibrary struct {
	templates map[string]*domain.Template
	loaded    bool
}

// NewStandardLibrary creates a new standard library instance
func NewStandardLibrary() *StandardLibrary {
	return &StandardLibrary{
		templates: make(map[string]*domain.Template),
		loaded:    false,
	}
}

// Load loads all standard library templates from embedded files
func (sl *StandardLibrary) Load() error {
	if sl.loaded {
		return nil // Already loaded
	}

	entries, err := templatesFS.ReadDir("templates")
	if err != nil {
		return fmt.Errorf("failed to read templates directory: %w", err)
	}

	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".yaml") {
			continue
		}

		path := filepath.Join("templates", entry.Name())
		data, err := templatesFS.ReadFile(path)
		if err != nil {
			return fmt.Errorf("failed to read template %s: %w", entry.Name(), err)
		}

		// Parse YAML
		var rawTemplate map[string]interface{}
		if err := yaml.Unmarshal(data, &rawTemplate); err != nil {
			return fmt.Errorf("failed to parse template %s: %w", entry.Name(), err)
		}

		// Convert to Template
		tmpl, err := parseTemplate(rawTemplate)
		if err != nil {
			return fmt.Errorf("failed to convert template %s: %w", entry.Name(), err)
		}

		sl.templates[tmpl.GetID()] = tmpl
	}

	sl.loaded = true
	return nil
}

// Get retrieves a template by ID
func (sl *StandardLibrary) Get(id string) (*domain.Template, error) {
	if !sl.loaded {
		if err := sl.Load(); err != nil {
			return nil, err
		}
	}

	tmpl, ok := sl.templates[id]
	if !ok {
		return nil, fmt.Errorf("template not found: %s", id)
	}

	return tmpl, nil
}

// GetAll returns all standard library templates
func (sl *StandardLibrary) GetAll() ([]*domain.Template, error) {
	if !sl.loaded {
		if err := sl.Load(); err != nil {
			return nil, err
		}
	}

	templates := make([]*domain.Template, 0, len(sl.templates))
	for _, tmpl := range sl.templates {
		templates = append(templates, tmpl)
	}

	return templates, nil
}

// GetIDs returns all template IDs
func (sl *StandardLibrary) GetIDs() ([]string, error) {
	if !sl.loaded {
		if err := sl.Load(); err != nil {
			return nil, err
		}
	}

	ids := make([]string, 0, len(sl.templates))
	for id := range sl.templates {
		ids = append(ids, id)
	}

	return ids, nil
}

// parseTemplate converts a raw YAML map to a Template
func parseTemplate(raw map[string]interface{}) (*domain.Template, error) {
	// Extract basic fields
	name, _ := raw["name"].(string)
	description, _ := raw["description"].(string)
	version, _ := raw["version"].(string)
	author, _ := raw["author"].(string)
	content, _ := raw["content"].(string)
	format, _ := raw["format"].(string)

	if name == "" {
		return nil, fmt.Errorf("template missing required field: name")
	}

	// Create template using constructor
	tmpl := domain.NewTemplate(name, description, version, author)

	// Set content and format
	tmpl.Content = content
	if format != "" {
		tmpl.Format = format
	}

	// Extract variables
	var variables []domain.TemplateVariable
	if rawVars, ok := raw["variables"].([]interface{}); ok {
		for _, v := range rawVars {
			if varMap, ok := v.(map[string]interface{}); ok {
				variable := domain.TemplateVariable{
					Name:        getString(varMap, "name"),
					Type:        getString(varMap, "type"),
					Required:    getBool(varMap, "required"),
					Default:     getString(varMap, "default"),
					Description: getString(varMap, "description"),
				}
				variables = append(variables, variable)
			}
		}
	}
	tmpl.Variables = variables

	return tmpl, nil
}

// Helper functions

func getString(m map[string]interface{}, key string) string {
	if v, ok := m[key].(string); ok {
		return v
	}
	return ""
}

func getBool(m map[string]interface{}, key string) bool {
	if v, ok := m[key].(bool); ok {
		return v
	}
	return false
}
