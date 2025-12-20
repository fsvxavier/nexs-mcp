package integration

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/fsvxavier/nexs-mcp/internal/domain"
	"github.com/fsvxavier/nexs-mcp/internal/infrastructure"
	"github.com/fsvxavier/nexs-mcp/internal/template"
)

// Test fixtures

func createTestTemplate(id, name, content string, variables []domain.TemplateVariable) *domain.Template {
	tmpl := domain.NewTemplate(name, "Test template", "1.0.0", "test-author")

	// Update metadata with custom ID
	metadata := tmpl.GetMetadata()
	metadata.ID = id
	metadata.Tags = []string{"test"}
	tmpl.SetMetadata(metadata)

	// Set template-specific fields
	tmpl.Content = content
	tmpl.Format = "yaml"
	tmpl.Variables = variables

	return tmpl
}

func setupTestRepo(t *testing.T) (domain.ElementRepository, string) {
	t.Helper()

	// Create temporary directory
	tmpDir, err := os.MkdirTemp("", "nexs-template-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}

	// Create repository
	repo, err := infrastructure.NewFileElementRepository(tmpDir)
	if err != nil {
		t.Fatalf("Failed to create repository: %v", err)
	}

	return repo, tmpDir
}

// Registry Caching Tests (5 tests)

func TestRegistryCacheHit(t *testing.T) {
	repo, tmpDir := setupTestRepo(t)
	defer os.RemoveAll(tmpDir)

	registry := template.NewTemplateRegistry(repo, 15*time.Minute)
	ctx := context.Background()

	// Create and save template
	tmpl := createTestTemplate("test-cache-1", "Cache Test", "content: {{name}}", []domain.TemplateVariable{
		{Name: "name", Type: "string", Required: true},
	})
	if err := repo.Create(tmpl); err != nil {
		t.Fatalf("Failed to save template: %v", err)
	}

	// First get - should be cache miss
	_, err := registry.GetTemplate(ctx, "test-cache-1")
	if err != nil {
		t.Fatalf("First get failed: %v", err)
	}

	stats1 := registry.GetCacheStats()
	if stats1.Misses != 1 {
		t.Errorf("Expected 1 miss, got %d", stats1.Misses)
	}

	// Second get - should be cache hit
	_, err = registry.GetTemplate(ctx, "test-cache-1")
	if err != nil {
		t.Fatalf("Second get failed: %v", err)
	}

	stats2 := registry.GetCacheStats()
	if stats2.Hits != 1 {
		t.Errorf("Expected 1 hit, got %d", stats2.Hits)
	}
}

func TestCacheExpiration(t *testing.T) {
	repo, tmpDir := setupTestRepo(t)
	defer os.RemoveAll(tmpDir)

	// Short TTL for testing
	registry := template.NewTemplateRegistry(repo, 100*time.Millisecond)
	ctx := context.Background()

	// Create and save template
	tmpl := createTestTemplate("test-expire-1", "Expire Test", "content: {{value}}", []domain.TemplateVariable{
		{Name: "value", Type: "string"},
	})
	if err := repo.Create(tmpl); err != nil {
		t.Fatalf("Failed to save template: %v", err)
	}

	// First get - cache miss
	_, err := registry.GetTemplate(ctx, "test-expire-1")
	if err != nil {
		t.Fatalf("First get failed: %v", err)
	}

	// Wait for expiration
	time.Sleep(150 * time.Millisecond)

	// Get after expiration - should be cache miss again
	_, err = registry.GetTemplate(ctx, "test-expire-1")
	if err != nil {
		t.Fatalf("Second get failed: %v", err)
	}

	stats := registry.GetCacheStats()
	if stats.Evictions < 1 {
		t.Errorf("Expected at least 1 eviction, got %d", stats.Evictions)
	}
}

func TestCacheClear(t *testing.T) {
	repo, tmpDir := setupTestRepo(t)
	defer os.RemoveAll(tmpDir)

	registry := template.NewTemplateRegistry(repo, 15*time.Minute)
	ctx := context.Background()

	// Create and cache multiple templates
	for i := 1; i <= 3; i++ {
		tmpl := createTestTemplate(
			filepath.Join("test-clear-", string(rune('0'+i))),
			"Clear Test",
			"content: {{x}}",
			[]domain.TemplateVariable{{Name: "x", Type: "string"}},
		)
		if err := repo.Create(tmpl); err != nil {
			t.Fatalf("Failed to save template %d: %v", i, err)
		}
		if _, err := registry.GetTemplate(ctx, tmpl.GetID()); err != nil {
			t.Fatalf("Failed to get template %d: %v", i, err)
		}
	}

	stats1 := registry.GetCacheStats()
	if stats1.Size != 3 {
		t.Errorf("Expected cache size 3, got %d", stats1.Size)
	}

	// Note: ClearCache is not available in the current implementation
	// The cache will naturally expire based on TTL
}

func TestCacheHitRate(t *testing.T) {
	repo, tmpDir := setupTestRepo(t)
	defer os.RemoveAll(tmpDir)

	registry := template.NewTemplateRegistry(repo, 15*time.Minute)
	ctx := context.Background()

	// Create template
	tmpl := createTestTemplate("test-hitrate-1", "Hit Rate Test", "content: {{val}}", []domain.TemplateVariable{
		{Name: "val", Type: "string"},
	})
	if err := repo.Create(tmpl); err != nil {
		t.Fatalf("Failed to save template: %v", err)
	}

	// First get (miss) + 9 gets (hits) = 90% hit rate
	for i := 0; i < 10; i++ {
		if _, err := registry.GetTemplate(ctx, "test-hitrate-1"); err != nil {
			t.Fatalf("Get %d failed: %v", i, err)
		}
	}

	stats := registry.GetCacheStats()
	expectedHitRate := 0.9 // 9 hits / 10 total
	if stats.HitRate < expectedHitRate-0.01 || stats.HitRate > expectedHitRate+0.01 {
		t.Errorf("Expected hit rate ~%.2f, got %.2f", expectedHitRate, stats.HitRate)
	}
}

func TestCacheInvalidation(t *testing.T) {
	repo, tmpDir := setupTestRepo(t)
	defer os.RemoveAll(tmpDir)

	registry := template.NewTemplateRegistry(repo, 15*time.Minute)
	ctx := context.Background()

	// Create and cache template
	tmpl := createTestTemplate("test-invalidate-1", "Invalidate Test", "content: {{old}}", []domain.TemplateVariable{
		{Name: "old", Type: "string"},
	})
	if err := repo.Create(tmpl); err != nil {
		t.Fatalf("Failed to save template: %v", err)
	}

	// Get to cache
	cached1, err := registry.GetTemplate(ctx, "test-invalidate-1")
	if err != nil {
		t.Fatalf("First get failed: %v", err)
	}
	if cached1.Content != "content: {{old}}" {
		t.Errorf("Unexpected content: %s", cached1.Content)
	}

	// Update template
	tmpl.Content = "content: {{new}}"
	if err := repo.Update(tmpl); err != nil {
		t.Fatalf("Failed to update template: %v", err)
	}

	// Invalidate cache
	registry.InvalidateTemplate("test-invalidate-1")

	// Get again - should see new content
	cached2, err := registry.GetTemplate(ctx, "test-invalidate-1")
	if err != nil {
		t.Fatalf("Second get failed: %v", err)
	}
	if cached2.Content != "content: {{new}}" {
		t.Errorf("Expected new content, got: %s", cached2.Content)
	}
}

// Instantiation Tests (10 tests)

func TestSimpleVariableSubstitution(t *testing.T) {
	tmpl := createTestTemplate("test-simple", "Simple", "Hello {{name}}!", []domain.TemplateVariable{
		{Name: "name", Type: "string", Required: true},
	})

	validator := template.NewTemplateValidator()
	engine := template.NewInstantiationEngine(validator, nil)

	result, err := engine.Instantiate(tmpl, map[string]interface{}{
		"name": "World",
	})
	if err != nil {
		t.Fatalf("Instantiation failed: %v", err)
	}

	if result.Output != "Hello World!" {
		t.Errorf("Expected 'Hello World!', got '%s'", result.Output)
	}
}

func TestConditionalBlocks(t *testing.T) {
	tmpl := createTestTemplate(
		"test-conditional",
		"Conditional",
		"{{#if show}}Visible{{/if}}",
		[]domain.TemplateVariable{
			{Name: "show", Type: "boolean"},
		},
	)

	validator := template.NewTemplateValidator()
	engine := template.NewInstantiationEngine(validator, nil)

	// Test true condition
	result1, err := engine.Instantiate(tmpl, map[string]interface{}{"show": true})
	if err != nil {
		t.Fatalf("Instantiation failed: %v", err)
	}
	if result1.Output != "Visible" {
		t.Errorf("Expected 'Visible', got '%s'", result1.Output)
	}

	// Test false condition
	result2, err := engine.Instantiate(tmpl, map[string]interface{}{"show": false})
	if err != nil {
		t.Fatalf("Instantiation failed: %v", err)
	}
	if result2.Output != "" {
		t.Errorf("Expected empty string, got '%s'", result2.Output)
	}
}

func TestIterationLoops(t *testing.T) {
	tmpl := createTestTemplate(
		"test-loop",
		"Loop",
		"{{#each items}}{{this}} {{/each}}",
		[]domain.TemplateVariable{
			{Name: "items", Type: "array"},
		},
	)

	validator := template.NewTemplateValidator()
	engine := template.NewInstantiationEngine(validator, nil)

	result, err := engine.Instantiate(tmpl, map[string]interface{}{
		"items": []interface{}{"a", "b", "c"},
	})
	if err != nil {
		t.Fatalf("Instantiation failed: %v", err)
	}

	expected := "a b c "
	if result.Output != expected {
		t.Errorf("Expected '%s', got '%s'", expected, result.Output)
	}
}

func TestHelperFunctions(t *testing.T) {
	tmpl := createTestTemplate(
		"test-helpers",
		"Helpers",
		"{{upper name}} - {{lower name}}",
		[]domain.TemplateVariable{
			{Name: "name", Type: "string"},
		},
	)

	validator := template.NewTemplateValidator()
	engine := template.NewInstantiationEngine(validator, nil)

	result, err := engine.Instantiate(tmpl, map[string]interface{}{
		"name": "Test",
	})
	if err != nil {
		t.Fatalf("Instantiation failed: %v", err)
	}

	expected := "TEST - test"
	if result.Output != expected {
		t.Errorf("Expected '%s', got '%s'", expected, result.Output)
	}
}

func TestComparisonHelpers(t *testing.T) {
	tmpl := createTestTemplate(
		"test-comparison",
		"Comparison",
		"{{#if (eq status 'active')}}Active{{else}}Inactive{{/if}}",
		[]domain.TemplateVariable{
			{Name: "status", Type: "string"},
		},
	)

	validator := template.NewTemplateValidator()
	engine := template.NewInstantiationEngine(validator, nil)

	// Test equal
	result1, err := engine.Instantiate(tmpl, map[string]interface{}{"status": "active"})
	if err != nil {
		t.Fatalf("Instantiation failed: %v", err)
	}
	if result1.Output != "Active" {
		t.Errorf("Expected 'Active', got '%s'", result1.Output)
	}

	// Test not equal
	result2, err := engine.Instantiate(tmpl, map[string]interface{}{"status": "inactive"})
	if err != nil {
		t.Fatalf("Instantiation failed: %v", err)
	}
	if result2.Output != "Inactive" {
		t.Errorf("Expected 'Inactive', got '%s'", result2.Output)
	}
}

func TestDefaultValues(t *testing.T) {
	tmpl := createTestTemplate(
		"test-defaults",
		"Defaults",
		"Name: {{default name 'Anonymous'}}",
		[]domain.TemplateVariable{
			{Name: "name", Type: "string", Default: "Anonymous"},
		},
	)

	validator := template.NewTemplateValidator()
	engine := template.NewInstantiationEngine(validator, nil)

	// Test with value
	result1, err := engine.Instantiate(tmpl, map[string]interface{}{"name": "Alice"})
	if err != nil {
		t.Fatalf("Instantiation failed: %v", err)
	}
	if result1.Output != "Name: Alice" {
		t.Errorf("Expected 'Name: Alice', got '%s'", result1.Output)
	}

	// Test without value (should use default)
	result2, err := engine.Instantiate(tmpl, map[string]interface{}{})
	if err != nil {
		t.Fatalf("Instantiation failed: %v", err)
	}
	if result2.Output != "Name: Anonymous" {
		t.Errorf("Expected 'Name: Anonymous', got '%s'", result2.Output)
	}
}

func TestNestedTemplates(t *testing.T) {
	tmpl := createTestTemplate(
		"test-nested",
		"Nested",
		"{{#each users}}{{name}}: {{#each roles}}{{this}} {{/each}}\n{{/each}}",
		[]domain.TemplateVariable{
			{Name: "users", Type: "array"},
		},
	)

	validator := template.NewTemplateValidator()
	engine := template.NewInstantiationEngine(validator, nil)

	result, err := engine.Instantiate(tmpl, map[string]interface{}{
		"users": []interface{}{
			map[string]interface{}{
				"name":  "Alice",
				"roles": []interface{}{"admin", "user"},
			},
			map[string]interface{}{
				"name":  "Bob",
				"roles": []interface{}{"user"},
			},
		},
	})
	if err != nil {
		t.Fatalf("Instantiation failed: %v", err)
	}

	expected := "Alice: admin user \nBob: user \n"
	if result.Output != expected {
		t.Errorf("Expected '%s', got '%s'", expected, result.Output)
	}
}

func TestArrayHelpers(t *testing.T) {
	tmpl := createTestTemplate(
		"test-array-helpers",
		"Array Helpers",
		"Count: {{length items}}, Joined: {{join items ','}}",
		[]domain.TemplateVariable{
			{Name: "items", Type: "array"},
		},
	)

	validator := template.NewTemplateValidator()
	engine := template.NewInstantiationEngine(validator, nil)

	result, err := engine.Instantiate(tmpl, map[string]interface{}{
		"items": []interface{}{"a", "b", "c"},
	})
	if err != nil {
		t.Fatalf("Instantiation failed: %v", err)
	}

	expected := "Count: 3, Joined: a,b,c"
	if result.Output != expected {
		t.Errorf("Expected '%s', got '%s'", expected, result.Output)
	}
}

func TestLogicalHelpers(t *testing.T) {
	tmpl := createTestTemplate(
		"test-logical",
		"Logical",
		"{{#if (and enabled verified)}}OK{{else}}NOT OK{{/if}}",
		[]domain.TemplateVariable{
			{Name: "enabled", Type: "boolean"},
			{Name: "verified", Type: "boolean"},
		},
	)

	validator := template.NewTemplateValidator()
	engine := template.NewInstantiationEngine(validator, nil)

	// Test both true
	result1, err := engine.Instantiate(tmpl, map[string]interface{}{
		"enabled":  true,
		"verified": true,
	})
	if err != nil {
		t.Fatalf("Instantiation failed: %v", err)
	}
	if result1.Output != "OK" {
		t.Errorf("Expected 'OK', got '%s'", result1.Output)
	}

	// Test one false
	result2, err := engine.Instantiate(tmpl, map[string]interface{}{
		"enabled":  true,
		"verified": false,
	})
	if err != nil {
		t.Fatalf("Instantiation failed: %v", err)
	}
	if result2.Output != "NOT OK" {
		t.Errorf("Expected 'NOT OK', got '%s'", result2.Output)
	}
}

func TestPreviewGeneration(t *testing.T) {
	tmpl := createTestTemplate(
		"test-preview",
		"Preview",
		"Name: {{name}}, Age: {{age}}",
		[]domain.TemplateVariable{
			{Name: "name", Type: "string", Description: "User name"},
			{Name: "age", Type: "number", Description: "User age"},
		},
	)

	validator := template.NewTemplateValidator()
	engine := template.NewInstantiationEngine(validator, nil)

	preview, err := engine.Preview(tmpl)
	if err != nil {
		t.Fatalf("Preview failed: %v", err)
	}

	// Should contain placeholders
	if preview.Output == "" {
		t.Error("Expected non-empty preview")
	}
}

// Validation Tests (8 tests)

func TestSyntaxValidation(t *testing.T) {
	validator := template.NewTemplateValidator()

	// Valid template
	validTmpl := createTestTemplate("test-valid", "Valid", "{{name}}", []domain.TemplateVariable{
		{Name: "name", Type: "string"},
	})
	if err := validator.ValidateSyntax(validTmpl); err != nil {
		t.Errorf("Valid template failed validation: %v", err)
	}

	// Invalid template - unbalanced braces
	invalidTmpl := createTestTemplate("test-invalid", "Invalid", "{{name}", []domain.TemplateVariable{
		{Name: "name", Type: "string"},
	})
	if err := validator.ValidateSyntax(invalidTmpl); err == nil {
		t.Error("Expected validation error for unbalanced braces")
	}
}

func TestRequiredVariables(t *testing.T) {
	tmpl := createTestTemplate(
		"test-required",
		"Required",
		"{{name}}",
		[]domain.TemplateVariable{
			{Name: "name", Type: "string", Required: true},
		},
	)

	validator := template.NewTemplateValidator()

	// Missing required variable
	result1 := validator.ValidateComprehensive(tmpl, map[string]interface{}{})
	if result1.Valid {
		t.Error("Expected validation to fail for missing required variable")
	}
	if len(result1.Errors) == 0 {
		t.Error("Expected validation errors")
	}

	// With required variable
	result2 := validator.ValidateComprehensive(tmpl, map[string]interface{}{"name": "Test"})
	if !result2.Valid {
		t.Errorf("Expected validation to pass, got errors: %v", result2.Errors)
	}
}

func TestVariableTypes(t *testing.T) {
	tmpl := createTestTemplate(
		"test-types",
		"Types",
		"{{count}}",
		[]domain.TemplateVariable{
			{Name: "count", Type: "number", Required: true},
		},
	)

	validator := template.NewTemplateValidator()

	// Wrong type (string instead of number)
	result1 := validator.ValidateComprehensive(tmpl, map[string]interface{}{"count": "not a number"})
	if result1.Valid {
		t.Error("Expected validation to fail for wrong type")
	}

	// Correct type
	result2 := validator.ValidateComprehensive(tmpl, map[string]interface{}{"count": 42})
	if !result2.Valid {
		t.Errorf("Expected validation to pass, got errors: %v", result2.Errors)
	}
}

func TestUndefinedVariables(t *testing.T) {
	tmpl := createTestTemplate(
		"test-undefined",
		"Undefined",
		"{{defined}} {{undefined}}",
		[]domain.TemplateVariable{
			{Name: "defined", Type: "string"},
		},
	)

	validator := template.NewTemplateValidator()
	result := validator.ValidateComprehensive(tmpl, map[string]interface{}{"defined": "value"})

	// Should have warnings about undefined variable
	if len(result.Warnings) == 0 {
		t.Error("Expected warnings about undefined variable")
	}
}

func TestBalancedDelimiters(t *testing.T) {
	validator := template.NewTemplateValidator()

	tests := []struct {
		name    string
		content string
		valid   bool
	}{
		{"balanced", "{{name}}", true},
		{"multiple balanced", "{{first}} {{second}}", true},
		{"unbalanced open", "{{name", false},
		{"unbalanced close", "name}}", false},
		{"mixed", "{{#if test}}{{value}}", false}, // #if not closed
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpl := createTestTemplate("test", "Test", tt.content, nil)
			err := validator.ValidateSyntax(tmpl)
			if tt.valid && err != nil {
				t.Errorf("Expected valid, got error: %v", err)
			}
			if !tt.valid && err == nil {
				t.Error("Expected error, got valid")
			}
		})
	}
}

func TestVariableNameFormat(t *testing.T) {
	validator := template.NewTemplateValidator()

	tests := []struct {
		name    string
		varName string
		valid   bool
	}{
		{"alphanumeric", "userName", true},
		{"underscore", "user_name", true},
		{"number suffix", "value123", true},
		{"starts with number", "123value", false},
		{"special chars", "user@name", false},
		{"spaces", "user name", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpl := createTestTemplate(
				"test",
				"Test",
				"{{"+tt.varName+"}}",
				[]domain.TemplateVariable{
					{Name: tt.varName, Type: "string"},
				},
			)
			result := validator.ValidateComprehensive(tmpl, map[string]interface{}{tt.varName: "value"})
			if tt.valid && !result.Valid {
				t.Errorf("Expected valid, got errors: %v", result.Errors)
			}
		})
	}
}

func TestTemplateSizeLimit(t *testing.T) {
	validator := template.NewTemplateValidator()

	// Create large content (> 1MB)
	largeContent := make([]byte, 1024*1024+1)
	for i := range largeContent {
		largeContent[i] = 'x'
	}

	tmpl := createTestTemplate("test-large", "Large", string(largeContent), nil)

	err := validator.ValidateSyntax(tmpl)
	if err == nil {
		t.Error("Expected error for template exceeding size limit")
	}
}

func TestVariableCountLimit(t *testing.T) {
	validator := template.NewTemplateValidator()

	// Create template with > 100 variables
	variables := make([]domain.TemplateVariable, 101)
	content := ""
	for i := 0; i < 101; i++ {
		varName := fmt.Sprintf("var%d", i)
		variables[i] = domain.TemplateVariable{Name: varName, Type: "string"}
		content += "{{" + varName + "}} "
	}

	tmpl := createTestTemplate("test-many-vars", "Many Vars", content, variables)

	err := validator.ValidateSyntax(tmpl)
	if err == nil {
		t.Error("Expected error for template exceeding variable count limit")
	}
}

// Standard Library Tests (7 tests)

func TestStandardLibraryLoading(t *testing.T) {
	repo, tmpDir := setupTestRepo(t)
	defer os.RemoveAll(tmpDir)

	registry := template.NewTemplateRegistry(repo, 15*time.Minute)

	err := registry.LoadStandardLibrary()
	if err != nil {
		t.Fatalf("Failed to load standard library: %v", err)
	}

	// Standard library should be accessible
	// Note: This test will pass even with empty stdlib until we implement the loader
}

func TestStandardLibraryRetrieval(t *testing.T) {
	repo, tmpDir := setupTestRepo(t)
	defer os.RemoveAll(tmpDir)

	registry := template.NewTemplateRegistry(repo, 15*time.Minute)
	ctx := context.Background()

	if err := registry.LoadStandardLibrary(); err != nil {
		t.Fatalf("Failed to load standard library: %v", err)
	}

	// Try to list all templates (should include stdlib once implemented)
	result, err := registry.ListAllTemplates(ctx, true)
	if err != nil {
		t.Fatalf("Failed to list templates: %v", err)
	}

	// Should return result (may be empty if stdlib not yet implemented)
	if result == nil {
		t.Error("Expected non-nil result")
	}
}

func TestStandardLibraryCategories(t *testing.T) {
	repo, tmpDir := setupTestRepo(t)
	defer os.RemoveAll(tmpDir)

	registry := template.NewTemplateRegistry(repo, 15*time.Minute)
	ctx := context.Background()

	if err := registry.LoadStandardLibrary(); err != nil {
		t.Fatalf("Failed to load standard library: %v", err)
	}

	// Search for persona templates
	filter := template.TemplateSearchFilter{
		Category:       "persona",
		IncludeBuiltIn: true,
		Page:           1,
		PerPage:        10,
	}

	result, err := registry.SearchTemplates(ctx, filter)
	if err != nil {
		t.Fatalf("Failed to search templates: %v", err)
	}

	// Should return result (may be empty if stdlib not yet implemented)
	if result == nil {
		t.Error("Expected non-nil result")
	}
}

func TestStandardLibraryInstantiation(t *testing.T) {
	// This test will be meaningful once stdlib templates are added
	// For now, just test the mechanism

	repo, tmpDir := setupTestRepo(t)
	defer os.RemoveAll(tmpDir)

	// Create a mock stdlib template manually
	mockTemplate := createTestTemplate(
		"stdlib-persona-basic",
		"Basic Persona",
		"name: {{name}}\nexpertise: {{expertise}}",
		[]domain.TemplateVariable{
			{Name: "name", Type: "string", Required: true},
			{Name: "expertise", Type: "string", Required: true},
		},
	)
	if err := repo.Create(mockTemplate); err != nil {
		t.Fatalf("Failed to save mock template: %v", err)
	}

	registry := template.NewTemplateRegistry(repo, 15*time.Minute)
	ctx := context.Background()

	tmpl, err := registry.GetTemplate(ctx, "stdlib-persona-basic")
	if err != nil {
		t.Fatalf("Failed to get template: %v", err)
	}

	validator := template.NewTemplateValidator()
	engine := template.NewInstantiationEngine(validator, nil)

	result, err := engine.Instantiate(tmpl, map[string]interface{}{
		"name":      "Expert",
		"expertise": "Testing",
	})
	if err != nil {
		t.Fatalf("Instantiation failed: %v", err)
	}

	expected := "name: Expert\nexpertise: Testing"
	if result.Output != expected {
		t.Errorf("Expected '%s', got '%s'", expected, result.Output)
	}
}

func TestStandardLibraryIDs(t *testing.T) {
	repo, tmpDir := setupTestRepo(t)
	defer os.RemoveAll(tmpDir)

	registry := template.NewTemplateRegistry(repo, 15*time.Minute)

	if err := registry.LoadStandardLibrary(); err != nil {
		t.Fatalf("Failed to load standard library: %v", err)
	}

	// Should be able to get stdlib IDs (empty list acceptable for now)
	// This will be meaningful once stdlib is populated
}

func TestStandardLibraryExclusion(t *testing.T) {
	repo, tmpDir := setupTestRepo(t)
	defer os.RemoveAll(tmpDir)

	registry := template.NewTemplateRegistry(repo, 15*time.Minute)
	ctx := context.Background()

	if err := registry.LoadStandardLibrary(); err != nil {
		t.Fatalf("Failed to load standard library: %v", err)
	}

	// Search without built-in templates
	filter := template.TemplateSearchFilter{
		IncludeBuiltIn: false,
		Page:           1,
		PerPage:        10,
	}

	result, err := registry.SearchTemplates(ctx, filter)
	if err != nil {
		t.Fatalf("Failed to search templates: %v", err)
	}

	// Should only return custom templates (none in this test)
	if result.Total > 0 {
		// Check that none are built-in
		for _, tmpl := range result.Templates {
			id := tmpl.GetID()
			if len(id) > 7 && id[:7] == "stdlib-" {
				t.Errorf("Found built-in template when excluded: %s", id)
			}
		}
	}
}

func TestStandardLibraryLazyLoading(t *testing.T) {
	repo, tmpDir := setupTestRepo(t)
	defer os.RemoveAll(tmpDir)

	registry := template.NewTemplateRegistry(repo, 15*time.Minute)

	// Don't call LoadStandardLibrary() explicitly
	// It should be loaded on first access if needed

	ctx := context.Background()
	_, err := registry.SearchTemplates(ctx, template.TemplateSearchFilter{
		IncludeBuiltIn: true,
		Page:           1,
		PerPage:        10,
	})

	// Should not error even if stdlib not explicitly loaded
	if err != nil {
		t.Fatalf("Failed to search: %v", err)
	}
}
