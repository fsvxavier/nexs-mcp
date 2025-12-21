package template

import (
	"strings"
	"testing"

	"github.com/fsvxavier/nexs-mcp/internal/domain"
)

// TestNewInstantiationEngine tests engine creation.
func TestNewInstantiationEngine(t *testing.T) {
	validator := NewTemplateValidator()
	engine := NewInstantiationEngine(validator, nil)

	if engine == nil {
		t.Fatal("Expected non-nil engine")
	}
	if engine.validator == nil {
		t.Error("Expected validator to be set")
	}
	if engine.options == nil {
		t.Error("Expected default options to be set")
	}
	if engine.options.MaxDepth != 10 {
		t.Errorf("Expected MaxDepth 10, got %d", engine.options.MaxDepth)
	}
	if engine.options.MaxIterations != 1000 {
		t.Errorf("Expected MaxIterations 1000, got %d", engine.options.MaxIterations)
	}
}

// TestNewInstantiationEngine_WithOptions tests engine with custom options.
func TestNewInstantiationEngine_WithOptions(t *testing.T) {
	validator := NewTemplateValidator()
	options := &EngineOptions{
		MaxDepth:           5,
		MaxIterations:      500,
		StrictMode:         true,
		AllowUnsafeHelpers: false,
	}
	engine := NewInstantiationEngine(validator, options)

	if engine.options.MaxDepth != 5 {
		t.Errorf("Expected MaxDepth 5, got %d", engine.options.MaxDepth)
	}
	if engine.options.MaxIterations != 500 {
		t.Errorf("Expected MaxIterations 500, got %d", engine.options.MaxIterations)
	}
	if !engine.options.StrictMode {
		t.Error("Expected StrictMode to be true")
	}
}

// TestInstantiate_SimpleTemplate tests basic template rendering.
func TestInstantiate_SimpleTemplate(t *testing.T) {
	validator := NewTemplateValidator()
	engine := NewInstantiationEngine(validator, nil)

	tmpl := domain.NewTemplate("simple", "Simple template", "1.0.0", "tester")
	tmpl.Content = "Hello {{name}}!"
	tmpl.Format = "text"

	variables := map[string]interface{}{
		"name": "World",
	}

	result, err := engine.Instantiate(tmpl, variables)
	if err != nil {
		t.Fatalf("Instantiation failed: %v", err)
	}

	expected := "Hello World!"
	if result.Output != expected {
		t.Errorf("Expected output %q, got %q", expected, result.Output)
	}
}

// TestInstantiate_WithDefaultVariable tests template with default values.
func TestInstantiate_WithDefaultVariable(t *testing.T) {
	validator := NewTemplateValidator()
	engine := NewInstantiationEngine(validator, nil)

	tmpl := domain.NewTemplate("default", "Template with defaults", "1.0.0", "tester")
	tmpl.Content = "Hello {{name}}!"
	tmpl.Format = "text"
	tmpl.Variables = []domain.TemplateVariable{
		{Name: "name", Type: "string", Required: false, Default: "Guest"},
	}

	variables := map[string]interface{}{}

	result, err := engine.Instantiate(tmpl, variables)
	if err != nil {
		t.Fatalf("Instantiation failed: %v", err)
	}

	expected := "Hello Guest!"
	if result.Output != expected {
		t.Errorf("Expected output %q, got %q", expected, result.Output)
	}
}

// TestInstantiate_MissingRequiredVariable tests error on missing required variable.
func TestInstantiate_MissingRequiredVariable(t *testing.T) {
	validator := NewTemplateValidator()
	engine := NewInstantiationEngine(validator, nil)

	tmpl := domain.NewTemplate("required", "Template with required var", "1.0.0", "tester")
	tmpl.Content = "Hello {{name}}!"
	tmpl.Format = "text"
	tmpl.Variables = []domain.TemplateVariable{
		{Name: "name", Type: "string", Required: true},
	}

	variables := map[string]interface{}{}

	_, err := engine.Instantiate(tmpl, variables)
	if err == nil {
		t.Error("Expected error for missing required variable")
	}
	if !strings.Contains(err.Error(), "required variable") {
		t.Errorf("Expected 'required variable' in error, got: %v", err)
	}
}

// TestInstantiate_WithHelpers tests template with built-in helpers.
func TestInstantiate_WithHelpers(t *testing.T) {
	validator := NewTemplateValidator()
	engine := NewInstantiationEngine(validator, nil)

	tmpl := domain.NewTemplate("helpers", "Template with helpers", "1.0.0", "tester")
	// Use built-in 'with' helper instead of custom helper
	tmpl.Content = "{{#with person}}Name: {{name}}{{/with}}"
	tmpl.Format = "text"

	variables := map[string]interface{}{
		"person": map[string]interface{}{
			"name": "Alice",
		},
	}

	result, err := engine.Instantiate(tmpl, variables)
	if err != nil {
		t.Fatalf("Instantiation failed: %v", err)
	}

	expected := "Name: Alice"
	if result.Output != expected {
		t.Errorf("Expected output %q, got %q", expected, result.Output)
	}
}

// TestInstantiate_ConditionalTemplate tests if/else logic.
func TestInstantiate_ConditionalTemplate(t *testing.T) {
	validator := NewTemplateValidator()
	engine := NewInstantiationEngine(validator, nil)

	tmpl := domain.NewTemplate("conditional", "Conditional template", "1.0.0", "tester")
	tmpl.Content = "{{#if admin}}Admin{{else}}User{{/if}}"
	tmpl.Format = "text"

	// Test with admin=true
	result, err := engine.Instantiate(tmpl, map[string]interface{}{"admin": true})
	if err != nil {
		t.Fatalf("Instantiation failed: %v", err)
	}
	if result.Output != "Admin" {
		t.Errorf("Expected 'Admin', got %q", result.Output)
	}

	// Test with admin=false
	result, err = engine.Instantiate(tmpl, map[string]interface{}{"admin": false})
	if err != nil {
		t.Fatalf("Instantiation failed: %v", err)
	}
	if result.Output != "User" {
		t.Errorf("Expected 'User', got %q", result.Output)
	}
}

// TestInstantiate_LoopTemplate tests loop iteration.
func TestInstantiate_LoopTemplate(t *testing.T) {
	validator := NewTemplateValidator()
	engine := NewInstantiationEngine(validator, nil)

	tmpl := domain.NewTemplate("loop", "Loop template", "1.0.0", "tester")
	tmpl.Content = "{{#each items}}{{this}} {{/each}}"
	tmpl.Format = "text"

	variables := map[string]interface{}{
		"items": []string{"apple", "banana", "cherry"},
	}

	result, err := engine.Instantiate(tmpl, variables)
	if err != nil {
		t.Fatalf("Instantiation failed: %v", err)
	}

	expected := "apple banana cherry "
	if result.Output != expected {
		t.Errorf("Expected output %q, got %q", expected, result.Output)
	}
}

// TestInstantiate_EmptyTemplate tests empty template.
func TestInstantiate_EmptyTemplate(t *testing.T) {
	validator := NewTemplateValidator()
	engine := NewInstantiationEngine(validator, nil)

	tmpl := domain.NewTemplate("empty", "Empty template", "1.0.0", "tester")
	tmpl.Content = ""
	tmpl.Format = "text"

	result, err := engine.Instantiate(tmpl, map[string]interface{}{})
	if err != nil {
		t.Fatalf("Instantiation failed: %v", err)
	}

	if result.Output != "" {
		t.Errorf("Expected empty output, got %q", result.Output)
	}
}

// TestInstantiate_ComplexTemplate tests more complex template.
func TestInstantiate_ComplexTemplate(t *testing.T) {
	validator := NewTemplateValidator()
	engine := NewInstantiationEngine(validator, nil)

	tmpl := domain.NewTemplate("complex", "Complex template", "1.0.0", "tester")
	tmpl.Content = `Hello {{name}}!
{{#if premium}}
  You are a premium user.
{{else}}
  You are a regular user.
{{/if}}
Your items: {{#each items}}{{this}}, {{/each}}`
	tmpl.Format = "text"

	variables := map[string]interface{}{
		"name":    "Alice",
		"premium": true,
		"items":   []string{"item1", "item2", "item3"},
	}

	result, err := engine.Instantiate(tmpl, variables)
	if err != nil {
		t.Fatalf("Instantiation failed: %v", err)
	}

	if !strings.Contains(result.Output, "Hello Alice!") {
		t.Error("Expected output to contain 'Hello Alice!'")
	}
	if !strings.Contains(result.Output, "premium user") {
		t.Error("Expected output to contain 'premium user'")
	}
	if !strings.Contains(result.Output, "item1") {
		t.Error("Expected output to contain 'item1'")
	}
}

// TestEngineOptions tests EngineOptions structure.
func TestEngineOptions(t *testing.T) {
	options := &EngineOptions{
		MaxDepth:           15,
		MaxIterations:      2000,
		StrictMode:         true,
		AllowUnsafeHelpers: true,
	}

	if options.MaxDepth != 15 {
		t.Errorf("Expected MaxDepth 15, got %d", options.MaxDepth)
	}
	if options.MaxIterations != 2000 {
		t.Errorf("Expected MaxIterations 2000, got %d", options.MaxIterations)
	}
	if !options.StrictMode {
		t.Error("Expected StrictMode to be true")
	}
	if !options.AllowUnsafeHelpers {
		t.Error("Expected AllowUnsafeHelpers to be true")
	}
}

// TestInstantiationResult tests result structure.
func TestInstantiationResult(t *testing.T) {
	result := &InstantiationResult{
		Output: "test output",
		Variables: map[string]interface{}{
			"key": "value",
		},
		UsedHelpers: []string{"helper1", "helper2"},
		Warnings:    []string{"warning1"},
	}

	if result.Output != "test output" {
		t.Errorf("Expected output 'test output', got %q", result.Output)
	}
	if len(result.UsedHelpers) != 2 {
		t.Errorf("Expected 2 used helpers, got %d", len(result.UsedHelpers))
	}
	if len(result.Warnings) != 1 {
		t.Errorf("Expected 1 warning, got %d", len(result.Warnings))
	}
}
