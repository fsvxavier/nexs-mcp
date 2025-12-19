package mcp

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	sdk "github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/fsvxavier/nexs-mcp/internal/domain"
	"github.com/fsvxavier/nexs-mcp/internal/infrastructure"
)

func setupBackupTestServer(t *testing.T) (*MCPServer, string) {
	t.Helper()
	tempDir := t.TempDir()

	repo := infrastructure.NewInMemoryElementRepository()
	server := NewMCPServer("nexs-mcp-test", "0.1.0", repo)

	return server, tempDir
}

func createTestPersona(name, description string) *domain.Persona {
	persona := domain.NewPersona(name, description, "1.0.0", "testuser")
	persona.SetSystemPrompt("Test prompt")
	persona.AddBehavioralTrait(domain.BehavioralTrait{Name: "professional", Intensity: 8})
	persona.AddExpertiseArea(domain.ExpertiseArea{Domain: "testing", Level: "expert"})
	persona.SetResponseStyle(domain.ResponseStyle{
		Tone:      "professional",
		Formality: "neutral",
		Verbosity: "balanced",
	})
	return persona
}

func TestHandleBackupPortfolio(t *testing.T) {
	server, tempDir := setupBackupTestServer(t)
	ctx := context.Background()

	// Create test elements
	persona := createTestPersona("Test Persona", "Description")
	server.repo.Create(persona)

	skill := domain.NewSkill("Test Skill", "Description", "1.0.0", "testuser")
	skill.AddTrigger(domain.SkillTrigger{Type: "keyword", Keywords: []string{"test"}})
	skill.AddProcedure(domain.SkillProcedure{Step: 1, Action: "Test action"})
	server.repo.Create(skill)

	// Test backup
	backupPath := filepath.Join(tempDir, "test-backup.tar.gz")
	input := BackupPortfolioInput{
		Path:        backupPath,
		Compression: "best",
		Description: "Test backup",
		Author:      "testuser",
	}

	result, output, err := server.handleBackupPortfolio(ctx, &sdk.CallToolRequest{}, input)

	if err != nil {
		t.Fatalf("handleBackupPortfolio failed: %v", err)
	}

	if result != nil {
		t.Errorf("Expected nil result, got: %v", result)
	}

	if output.Path != backupPath {
		t.Errorf("Expected path %s, got %s", backupPath, output.Path)
	}

	if output.ElementCount != 2 {
		t.Errorf("Expected 2 elements, got %d", output.ElementCount)
	}

	if output.Checksum == "" {
		t.Error("Expected non-empty checksum")
	}

	// Verify backup file exists
	if _, err := os.Stat(backupPath); os.IsNotExist(err) {
		t.Error("Backup file was not created")
	}
}

func TestHandleRestorePortfolio(t *testing.T) {
	server, tempDir := setupBackupTestServer(t)
	ctx := context.Background()

	// Create and backup test elements
	persona := createTestPersona("Test Persona", "Description")
	server.repo.Create(persona)

	skill := domain.NewSkill("Test Skill", "Description", "1.0.0", "testuser")
	skill.AddTrigger(domain.SkillTrigger{Type: "keyword", Keywords: []string{"test"}})
	skill.AddProcedure(domain.SkillProcedure{Step: 1, Action: "Test action"})
	server.repo.Create(skill)

	// Create backup
	backupPath := filepath.Join(tempDir, "test-backup.tar.gz")
	backupInput := BackupPortfolioInput{
		Path:        backupPath,
		Compression: "best",
	}
	server.handleBackupPortfolio(ctx, &sdk.CallToolRequest{}, backupInput)

	// Clear repository
	server.repo = infrastructure.NewInMemoryElementRepository()

	// Test restore
	restoreInput := RestorePortfolioInput{
		Path:          backupPath,
		Overwrite:     false,
		BackupBefore:  false,
		MergeStrategy: "skip",
	}

	result, output, err := server.handleRestorePortfolio(ctx, &sdk.CallToolRequest{}, restoreInput)

	if err != nil {
		t.Fatalf("handleRestorePortfolio failed: %v", err)
	}

	if result != nil {
		t.Errorf("Expected nil result, got: %v", result)
	}

	if !output.Success {
		t.Error("Restore should be successful")
	}

	if output.ElementsAdded != 2 {
		t.Errorf("Expected 2 elements added, got %d", output.ElementsAdded)
	}

	if len(output.Errors) > 0 {
		t.Errorf("Unexpected errors: %v", output.Errors)
	}

	// Verify elements were restored
	elements, _ := server.repo.List(domain.ElementFilter{})
	if len(elements) != 2 {
		t.Errorf("Expected 2 elements after restore, got %d", len(elements))
	}
}

func TestHandleActivateElement(t *testing.T) {
	server, _ := setupBackupTestServer(t)
	ctx := context.Background()

	// Create test element and deactivate first
	persona := createTestPersona("Test Persona", "Description")
	persona.Deactivate()
	server.repo.Create(persona)
	elementID := persona.GetID()

	// Test activation
	input := ActivateElementInput{
		ID: elementID,
	}

	result, output, err := server.handleActivateElement(ctx, &sdk.CallToolRequest{}, input)

	if err != nil {
		t.Fatalf("handleActivateElement failed: %v", err)
	}

	if result != nil {
		t.Errorf("Expected nil result, got: %v", result)
	}

	if output.ID != elementID {
		t.Errorf("Expected ID %s, got %s", elementID, output.ID)
	}

	if !output.IsActive {
		t.Error("Element should be active after activation")
	}

	if output.Name != "Test Persona" {
		t.Errorf("Expected name 'Test Persona', got %s", output.Name)
	}

	if output.Type != "persona" {
		t.Errorf("Expected type 'persona', got %s", output.Type)
	}

	// Verify element is active in repository
	element, _ := server.repo.GetByID(elementID)
	if !element.IsActive() {
		t.Error("Element should be active in repository after activation")
	}
}

func TestHandleDeactivateElement(t *testing.T) {
	server, _ := setupBackupTestServer(t)
	ctx := context.Background()

	// Create test element (starts active by default)
	persona := createTestPersona("Test Persona", "Description")
	server.repo.Create(persona)
	elementID := persona.GetID()

	// Verify it's active
	if !persona.IsActive() {
		t.Fatal("Persona should start active")
	}

	// Test deactivation
	input := DeactivateElementInput{
		ID: elementID,
	}

	result, output, err := server.handleDeactivateElement(ctx, &sdk.CallToolRequest{}, input)

	if err != nil {
		t.Fatalf("handleDeactivateElement failed: %v", err)
	}

	if result != nil {
		t.Errorf("Expected nil result, got: %v", result)
	}

	if output.ID != elementID {
		t.Errorf("Expected ID %s, got %s", elementID, output.ID)
	}

	if output.IsActive {
		t.Error("Element should be inactive after deactivation")
	}

	if output.Name != "Test Persona" {
		t.Errorf("Expected name 'Test Persona', got %s", output.Name)
	}

	if output.Type != "persona" {
		t.Errorf("Expected type 'persona', got %s", output.Type)
	}

	// Verify element is inactive in repository
	element, _ := server.repo.GetByID(elementID)
	if element.IsActive() {
		t.Error("Element should be inactive in repository after deactivation")
	}
}

func TestBackupPortfolio_EmptyPath(t *testing.T) {
	server, _ := setupBackupTestServer(t)
	ctx := context.Background()

	// Test with empty path
	input := BackupPortfolioInput{
		Path:        "",
		Compression: "best",
	}

	_, _, err := server.handleBackupPortfolio(ctx, &sdk.CallToolRequest{}, input)

	if err == nil {
		t.Error("Expected error when path is empty")
	}

	if err.Error() != "path is required" {
		t.Errorf("Expected 'path is required' error, got: %v", err)
	}
}

func TestRestorePortfolio_InvalidPath(t *testing.T) {
	server, _ := setupBackupTestServer(t)
	ctx := context.Background()

	// Test with non-existent path
	input := RestorePortfolioInput{
		Path: "/non/existent/path.tar.gz",
	}

	_, _, err := server.handleRestorePortfolio(ctx, &sdk.CallToolRequest{}, input)

	if err == nil {
		t.Error("Expected error when backup file doesn't exist")
	}
}

func TestActivateElement_EmptyID(t *testing.T) {
	server, _ := setupBackupTestServer(t)
	ctx := context.Background()

	// Test with empty ID
	input := ActivateElementInput{
		ID: "",
	}

	_, _, err := server.handleActivateElement(ctx, &sdk.CallToolRequest{}, input)

	if err == nil {
		t.Error("Expected error when ID is empty")
	}

	if err.Error() != "id is required" {
		t.Errorf("Expected 'id is required' error, got: %v", err)
	}
}

func TestActivateElement_NonExistentID(t *testing.T) {
	server, _ := setupBackupTestServer(t)
	ctx := context.Background()

	// Test with non-existent ID
	input := ActivateElementInput{
		ID: "non-existent-id-12345",
	}

	_, _, err := server.handleActivateElement(ctx, &sdk.CallToolRequest{}, input)

	if err == nil {
		t.Error("Expected error when element doesn't exist")
	}
}

func TestDeactivateElement_EmptyID(t *testing.T) {
	server, _ := setupBackupTestServer(t)
	ctx := context.Background()

	// Test with empty ID
	input := DeactivateElementInput{
		ID: "",
	}

	_, _, err := server.handleDeactivateElement(ctx, &sdk.CallToolRequest{}, input)

	if err == nil {
		t.Error("Expected error when ID is empty")
	}

	if err.Error() != "id is required" {
		t.Errorf("Expected 'id is required' error, got: %v", err)
	}
}

func TestBackupPortfolio_CompressionOptions(t *testing.T) {
	server, tempDir := setupBackupTestServer(t)
	ctx := context.Background()

	// Create test element
	persona := createTestPersona("Test Persona", "Description")
	server.repo.Create(persona)

	tests := []struct {
		name        string
		compression string
	}{
		{"None compression", "none"},
		{"Fast compression", "fast"},
		{"Best compression", "best"},
		{"Default compression", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			backupPath := filepath.Join(tempDir, tt.name+".tar.gz")
			input := BackupPortfolioInput{
				Path:        backupPath,
				Compression: tt.compression,
			}

			_, output, err := server.handleBackupPortfolio(ctx, &sdk.CallToolRequest{}, input)

			if err != nil {
				t.Fatalf("Backup failed: %v", err)
			}

			if output.ElementCount != 1 {
				t.Errorf("Expected 1 element, got %d", output.ElementCount)
			}

			// Verify file exists
			if _, err := os.Stat(backupPath); os.IsNotExist(err) {
				t.Errorf("Backup file was not created for %s", tt.name)
			}
		})
	}
}

func TestRestorePortfolio_DryRun(t *testing.T) {
	server, tempDir := setupBackupTestServer(t)
	ctx := context.Background()

	// Create and backup
	persona := createTestPersona("Test Persona", "Description")
	server.repo.Create(persona)

	backupPath := filepath.Join(tempDir, "test-backup.tar.gz")
	backupInput := BackupPortfolioInput{Path: backupPath}
	server.handleBackupPortfolio(ctx, &sdk.CallToolRequest{}, backupInput)

	// Clear repository
	server.repo = infrastructure.NewInMemoryElementRepository()

	// Test dry run
	input := RestorePortfolioInput{
		Path:   backupPath,
		DryRun: true,
	}

	_, output, err := server.handleRestorePortfolio(ctx, &sdk.CallToolRequest{}, input)

	if err != nil {
		t.Fatalf("Dry run failed: %v", err)
	}

	// Repository should still be empty
	elements, _ := server.repo.List(domain.ElementFilter{})
	if len(elements) != 0 {
		t.Errorf("Dry run should not modify repository, got %d elements", len(elements))
	}

	// Output should indicate what would be added
	if output.Success != true {
		t.Error("Dry run should report success")
	}
}
