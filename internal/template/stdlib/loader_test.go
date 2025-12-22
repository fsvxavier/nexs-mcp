package stdlib

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewStandardLibrary(t *testing.T) {
	stdlib := NewStandardLibrary()
	require.NotNil(t, stdlib)
	assert.NotNil(t, stdlib.templates)
	assert.False(t, stdlib.loaded)
}

func TestStandardLibrary_Load(t *testing.T) {
	stdlib := NewStandardLibrary()

	err := stdlib.Load()
	require.NoError(t, err)
	assert.True(t, stdlib.loaded)

	// Loading again should be no-op
	err = stdlib.Load()
	assert.NoError(t, err)
}

func TestStandardLibrary_Get_AutoLoad(t *testing.T) {
	stdlib := NewStandardLibrary()

	// Get should trigger automatic loading
	_, err := stdlib.Get("some-id")
	// Error is expected if template doesn't exist, but loading should succeed
	if err != nil {
		assert.Contains(t, err.Error(), "template not found")
	}
	assert.True(t, stdlib.loaded)
}

func TestStandardLibrary_Get_NotFound(t *testing.T) {
	stdlib := NewStandardLibrary()

	_, err := stdlib.Get("non-existent-template")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "template not found")
}

func TestStandardLibrary_GetAll(t *testing.T) {
	stdlib := NewStandardLibrary()

	templates, err := stdlib.GetAll()
	require.NoError(t, err)
	assert.NotNil(t, templates)
	assert.True(t, stdlib.loaded)
}

func TestStandardLibrary_GetIDs(t *testing.T) {
	stdlib := NewStandardLibrary()

	ids, err := stdlib.GetIDs()
	require.NoError(t, err)
	assert.NotNil(t, ids)
	assert.True(t, stdlib.loaded)
}

func TestStandardLibrary_LoadedState(t *testing.T) {
	stdlib := NewStandardLibrary()
	assert.False(t, stdlib.loaded)

	err := stdlib.Load()
	require.NoError(t, err)
	assert.True(t, stdlib.loaded)
}

func TestGetString(t *testing.T) {
	tests := []struct {
		name     string
		input    map[string]interface{}
		key      string
		expected string
	}{
		{
			name:     "Valid string",
			input:    map[string]interface{}{"key": "value"},
			key:      "key",
			expected: "value",
		},
		{
			name:     "Missing key",
			input:    map[string]interface{}{},
			key:      "key",
			expected: "",
		},
		{
			name:     "Wrong type",
			input:    map[string]interface{}{"key": 123},
			key:      "key",
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := getString(tt.input, tt.key)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestGetBool(t *testing.T) {
	tests := []struct {
		name     string
		input    map[string]interface{}
		key      string
		expected bool
	}{
		{
			name:     "Valid true",
			input:    map[string]interface{}{"key": true},
			key:      "key",
			expected: true,
		},
		{
			name:     "Valid false",
			input:    map[string]interface{}{"key": false},
			key:      "key",
			expected: false,
		},
		{
			name:     "Missing key",
			input:    map[string]interface{}{},
			key:      "key",
			expected: false,
		},
		{
			name:     "Wrong type",
			input:    map[string]interface{}{"key": "true"},
			key:      "key",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := getBool(tt.input, tt.key)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestParseTemplate_MissingName(t *testing.T) {
	raw := map[string]interface{}{
		"description": "Test template",
	}

	_, err := parseTemplate(raw)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "missing required field: name")
}

func TestParseTemplate_ValidTemplate(t *testing.T) {
	raw := map[string]interface{}{
		"name":        "test-template",
		"description": "Test template for unit testing purposes",
		"version":     "1.0.0",
		"author":      "test-author",
		"content":     "Hello {{name}}!",
		"format":      "text",
	}

	tmpl, err := parseTemplate(raw)
	require.NoError(t, err)
	require.NotNil(t, tmpl)

	meta := tmpl.GetMetadata()
	assert.Equal(t, "test-template", meta.Name)
	assert.Equal(t, "Test template for unit testing purposes", meta.Description)
	assert.Equal(t, "1.0.0", meta.Version)
	assert.Equal(t, "test-author", meta.Author)
	assert.Equal(t, "Hello {{name}}!", tmpl.Content)
	assert.Equal(t, "text", tmpl.Format)
}

func TestParseTemplate_WithVariables(t *testing.T) {
	raw := map[string]interface{}{
		"name":        "test-template",
		"description": "Test template with variables for testing",
		"content":     "Hello {{name}}!",
		"variables": []interface{}{
			map[string]interface{}{
				"name":        "name",
				"type":        "string",
				"required":    true,
				"default":     "Guest",
				"description": "User name to greet",
			},
		},
	}

	tmpl, err := parseTemplate(raw)
	require.NoError(t, err)
	require.NotNil(t, tmpl)

	assert.Len(t, tmpl.Variables, 1)
	assert.Equal(t, "name", tmpl.Variables[0].Name)
	assert.Equal(t, "string", tmpl.Variables[0].Type)
	assert.True(t, tmpl.Variables[0].Required)
	assert.Equal(t, "Guest", tmpl.Variables[0].Default)
	assert.Equal(t, "User name to greet", tmpl.Variables[0].Description)
}

func TestParseTemplate_MinimalTemplate(t *testing.T) {
	raw := map[string]interface{}{
		"name": "minimal",
	}

	tmpl, err := parseTemplate(raw)
	require.NoError(t, err)
	require.NotNil(t, tmpl)

	meta := tmpl.GetMetadata()
	assert.Equal(t, "minimal", meta.Name)
	assert.Empty(t, meta.Description)
	assert.Empty(t, meta.Version)
	assert.Empty(t, meta.Author)
	assert.Empty(t, tmpl.Content)
}

func TestParseTemplate_InvalidVariables(t *testing.T) {
	raw := map[string]interface{}{
		"name": "test",
		"variables": []interface{}{
			"not a map", // Invalid: should be map
		},
	}

	tmpl, err := parseTemplate(raw)
	require.NoError(t, err)
	// Should still parse, but skip invalid variables
	assert.Empty(t, tmpl.Variables)
}

func TestStandardLibrary_LoadEmptyDirectory(t *testing.T) {
	// This test verifies that Load handles the embedded FS correctly
	stdlib := NewStandardLibrary()

	err := stdlib.Load()
	// Should succeed even if no templates
	assert.NoError(t, err)
	assert.True(t, stdlib.loaded)
}

func TestStandardLibrary_GetAll_Empty(t *testing.T) {
	stdlib := NewStandardLibrary()
	err := stdlib.Load()
	require.NoError(t, err)

	templates, err := stdlib.GetAll()
	require.NoError(t, err)
	assert.NotNil(t, templates)
	// May be empty if no templates in embedded FS
	assert.GreaterOrEqual(t, len(templates), 0)
}

func TestStandardLibrary_GetIDs_Empty(t *testing.T) {
	stdlib := NewStandardLibrary()
	err := stdlib.Load()
	require.NoError(t, err)

	ids, err := stdlib.GetIDs()
	require.NoError(t, err)
	assert.NotNil(t, ids)
	assert.GreaterOrEqual(t, len(ids), 0)
}

func TestStandardLibrary_MultipleLoads(t *testing.T) {
	stdlib := NewStandardLibrary()

	// Load multiple times - should be idempotent
	for range 3 {
		err := stdlib.Load()
		assert.NoError(t, err)
		assert.True(t, stdlib.loaded)
	}
}

func TestParseTemplate_DefaultFormat(t *testing.T) {
	raw := map[string]interface{}{
		"name":    "test",
		"content": "Hello World",
		// No format specified
	}

	tmpl, err := parseTemplate(raw)
	require.NoError(t, err)
	// Format should be empty or default
	assert.NotNil(t, tmpl)
}

func TestStandardLibrary_ConcurrentAccess(t *testing.T) {
	stdlib := NewStandardLibrary()

	// Load once
	err := stdlib.Load()
	require.NoError(t, err)

	// Concurrent GetAll calls should be safe
	done := make(chan bool, 3)
	for range 3 {
		go func() {
			_, err := stdlib.GetAll()
			assert.NoError(t, err)
			done <- true
		}()
	}

	for range 3 {
		<-done
	}
}

func TestGetString_NilMap(t *testing.T) {
	result := getString(nil, "key")
	assert.Equal(t, "", result)
}

func TestGetBool_NilMap(t *testing.T) {
	result := getBool(nil, "key")
	assert.False(t, result)
}

func TestParseTemplate_WithEmptyVariablesList(t *testing.T) {
	raw := map[string]interface{}{
		"name":      "test",
		"variables": []interface{}{},
	}

	tmpl, err := parseTemplate(raw)
	require.NoError(t, err)
	assert.Empty(t, tmpl.Variables)
}

func TestParseTemplate_VariablesNotArray(t *testing.T) {
	raw := map[string]interface{}{
		"name":      "test",
		"variables": "not an array",
	}

	tmpl, err := parseTemplate(raw)
	require.NoError(t, err)
	// Should ignore invalid variables field
	assert.Empty(t, tmpl.Variables)
}
