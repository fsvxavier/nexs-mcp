package collection

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseManifest(t *testing.T) {
	tests := []struct {
		name     string
		yaml     string
		wantErr  bool
		validate func(*testing.T, *Manifest)
	}{
		{
			name: "valid minimal manifest",
			yaml: `
name: "test-collection"
version: "1.0.0"
author: "testuser"
description: "A test collection"
elements:
  - path: "personas/expert.yaml"
    type: "persona"
`,
			wantErr: false,
			validate: func(t *testing.T, m *Manifest) {
				assert.Equal(t, "test-collection", m.Name)
				assert.Equal(t, "1.0.0", m.Version)
				assert.Equal(t, "testuser", m.Author)
				assert.Equal(t, "A test collection", m.Description)
				assert.Len(t, m.Elements, 1)
				assert.Equal(t, "personas/expert.yaml", m.Elements[0].Path)
			},
		},
		{
			name: "complete manifest with all fields",
			yaml: `
name: "complete-collection"
version: "2.1.0"
author: "fsvxavier"
description: "Complete collection example"
tags: ["devops", "kubernetes"]
category: "professional-toolkit"
license: "MIT"
min_nexs_version: "0.4.0"
homepage: "https://github.com/fsvxavier/nexs-mcp-collection"
repository: "https://github.com/fsvxavier/nexs-mcp-collection"
maintainers:
  - name: "Francisco Xavier"
    email: "fx@example.com"
    github: "fsvxavier"
dependencies:
  - uri: "github://fsvxavier/base-skills@^1.0.0"
    description: "Core skills"
    optional: false
elements:
  - path: "personas/*.yaml"
    type: "persona"
    description: "Expert personas"
  - path: "skills/deploy/*.yaml"
    type: "skill"
config:
  default_persona: "personas/devops.yaml"
  auto_activate_skills:
    - "skills/deploy/kubectl.yaml"
  default_privacy_level: "public"
  custom:
    tool_version: "1.0.0"
keywords: ["automation", "deployment"]
`,
			wantErr: false,
			validate: func(t *testing.T, m *Manifest) {
				assert.Equal(t, "complete-collection", m.Name)
				assert.Equal(t, "2.1.0", m.Version)
				assert.Len(t, m.Tags, 2)
				assert.Equal(t, "MIT", m.License)
				assert.Len(t, m.Maintainers, 1)
				assert.Equal(t, "fsvxavier", m.Maintainers[0].GitHub)
				assert.Len(t, m.Dependencies, 1)
				assert.Equal(t, "github://fsvxavier/base-skills@^1.0.0", m.Dependencies[0].URI)
				assert.Len(t, m.Elements, 2)
				assert.NotNil(t, m.Config)
				assert.Equal(t, "personas/devops.yaml", m.Config.DefaultPersona)
				assert.Len(t, m.Keywords, 2)
			},
		},
		{
			name: "invalid YAML",
			yaml: `
name: test
version: 1.0.0
invalid: [unclosed
`,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			manifest, err := ParseManifest([]byte(tt.yaml))

			if tt.wantErr {
				assert.Error(t, err)
				return
			}

			require.NoError(t, err)
			require.NotNil(t, manifest)

			if tt.validate != nil {
				tt.validate(t, manifest)
			}
		})
	}
}

func TestManifest_Validate(t *testing.T) {
	tests := []struct {
		name     string
		manifest *Manifest
		wantErr  bool
		errMsg   string
	}{
		{
			name: "valid manifest",
			manifest: &Manifest{
				Name:        "test",
				Version:     "1.0.0",
				Author:      "testuser",
				Description: "Test description",
				Elements: []Element{
					{Path: "test.yaml"},
				},
			},
			wantErr: false,
		},
		{
			name: "missing name",
			manifest: &Manifest{
				Version:     "1.0.0",
				Author:      "testuser",
				Description: "Test",
				Elements:    []Element{{Path: "test.yaml"}},
			},
			wantErr: true,
			errMsg:  "'name' is required",
		},
		{
			name: "missing version",
			manifest: &Manifest{
				Name:        "test",
				Author:      "testuser",
				Description: "Test",
				Elements:    []Element{{Path: "test.yaml"}},
			},
			wantErr: true,
			errMsg:  "'version' is required",
		},
		{
			name: "missing author",
			manifest: &Manifest{
				Name:        "test",
				Version:     "1.0.0",
				Description: "Test",
				Elements:    []Element{{Path: "test.yaml"}},
			},
			wantErr: true,
			errMsg:  "'author' is required",
		},
		{
			name: "missing description",
			manifest: &Manifest{
				Name:     "test",
				Version:  "1.0.0",
				Author:   "testuser",
				Elements: []Element{{Path: "test.yaml"}},
			},
			wantErr: true,
			errMsg:  "'description' is required",
		},
		{
			name: "empty elements",
			manifest: &Manifest{
				Name:        "test",
				Version:     "1.0.0",
				Author:      "testuser",
				Description: "Test",
				Elements:    []Element{},
			},
			wantErr: true,
			errMsg:  "'elements' must contain at least one element",
		},
		{
			name: "element with empty path",
			manifest: &Manifest{
				Name:        "test",
				Version:     "1.0.0",
				Author:      "testuser",
				Description: "Test",
				Elements: []Element{
					{Path: ""},
				},
			},
			wantErr: true,
			errMsg:  "element[0].path is required",
		},
		{
			name: "dependency with empty URI",
			manifest: &Manifest{
				Name:        "test",
				Version:     "1.0.0",
				Author:      "testuser",
				Description: "Test",
				Elements:    []Element{{Path: "test.yaml"}},
				Dependencies: []Dependency{
					{URI: ""},
				},
			},
			wantErr: true,
			errMsg:  "dependency[0].uri is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.manifest.Validate()

			if tt.wantErr {
				assert.Error(t, err)
				if tt.errMsg != "" {
					assert.Contains(t, err.Error(), tt.errMsg)
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestManifest_ID(t *testing.T) {
	manifest := &Manifest{
		Name:   "test-collection",
		Author: "fsvxavier",
	}

	assert.Equal(t, "fsvxavier/test-collection", manifest.ID())
}

func TestManifest_FullID(t *testing.T) {
	manifest := &Manifest{
		Name:    "test-collection",
		Author:  "fsvxavier",
		Version: "1.2.3",
	}

	assert.Equal(t, "fsvxavier/test-collection@1.2.3", manifest.FullID())
}

func TestManifest_ToYAML(t *testing.T) {
	manifest := &Manifest{
		Name:        "test",
		Version:     "1.0.0",
		Author:      "testuser",
		Description: "Test collection",
		Elements: []Element{
			{Path: "test.yaml", Type: "persona"},
		},
	}

	data, err := manifest.ToYAML()
	require.NoError(t, err)
	assert.NotEmpty(t, data)

	// Parse it back
	parsed, err := ParseManifest(data)
	require.NoError(t, err)
	assert.Equal(t, manifest.Name, parsed.Name)
	assert.Equal(t, manifest.Version, parsed.Version)
	assert.Equal(t, manifest.Author, parsed.Author)
}
