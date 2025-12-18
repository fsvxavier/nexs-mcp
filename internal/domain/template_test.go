package domain

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewTemplate(t *testing.T) {
	template := NewTemplate("test-template", "Test Template", "1.0", "author")

	assert.NotNil(t, template)
	assert.Equal(t, "test-template", template.metadata.Name)
	assert.Equal(t, TemplateElement, template.metadata.Type)
	assert.Equal(t, "markdown", template.Format)
	assert.True(t, template.metadata.IsActive)
}

func TestTemplate_Validate(t *testing.T) {
	tests := []struct {
		name    string
		setup   func() *Template
		wantErr bool
	}{
		{
			name: "valid template",
			setup: func() *Template {
				tmpl := NewTemplate("valid", "Valid Template", "1.0", "author")
				tmpl.Content = "Hello {{name}}"
				return tmpl
			},
			wantErr: false,
		},
		{
			name: "empty content",
			setup: func() *Template {
				tmpl := NewTemplate("invalid", "Invalid Template", "1.0", "author")
				tmpl.Content = ""
				return tmpl
			},
			wantErr: true,
		},
		{
			name: "invalid format",
			setup: func() *Template {
				tmpl := NewTemplate("invalid", "Invalid Template", "1.0", "author")
				tmpl.Content = "test"
				tmpl.Format = "invalid"
				return tmpl
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpl := tt.setup()
			err := tmpl.Validate()
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestTemplate_Render(t *testing.T) {
	tmpl := NewTemplate("test", "Test", "1.0", "author")
	tmpl.Content = "Hello {{name}}, you are {{age}} years old"
	tmpl.Variables = []TemplateVariable{
		{Name: "name", Type: "string", Required: true},
		{Name: "age", Type: "number", Required: false, Default: "18"},
	}

	t.Run("all values provided", func(t *testing.T) {
		result, err := tmpl.Render(map[string]string{
			"name": "John",
			"age":  "25",
		})
		require.NoError(t, err)
		assert.Equal(t, "Hello John, you are 25 years old", result)
	})

	t.Run("missing required variable", func(t *testing.T) {
		_, err := tmpl.Render(map[string]string{
			"age": "25",
		})
		assert.Error(t, err)
	})

	t.Run("using default value", func(t *testing.T) {
		result, err := tmpl.Render(map[string]string{
			"name": "Jane",
		})
		require.NoError(t, err)
		assert.Equal(t, "Hello Jane, you are 18 years old", result)
	})
}

func TestTemplate_ActivateDeactivate(t *testing.T) {
	tmpl := NewTemplate("test", "Test", "1.0", "author")

	assert.True(t, tmpl.IsActive())

	err := tmpl.Deactivate()
	assert.NoError(t, err)
	assert.False(t, tmpl.IsActive())

	err = tmpl.Activate()
	assert.NoError(t, err)
	assert.True(t, tmpl.IsActive())
}
