package backup

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/fsvxavier/nexs-mcp/internal/domain"
	"github.com/fsvxavier/nexs-mcp/internal/infrastructure"
)

// Helper function to setup a valid persona for testing
func setupTestPersona(name, description string) *domain.Persona {
	persona := domain.NewPersona(name, description, "1.0.0", "testuser")
	persona.SetSystemPrompt("Test system prompt")
	persona.AddBehavioralTrait(domain.BehavioralTrait{Name: "professional", Intensity: 8})
	persona.AddExpertiseArea(domain.ExpertiseArea{Domain: "testing", Level: "expert"})
	persona.SetResponseStyle(domain.ResponseStyle{
		Tone:      "professional",
		Formality: "neutral",
		Verbosity: "balanced",
	})
	return persona
}

func TestBackup_Create(t *testing.T) {
	// Setup
	tempDir := t.TempDir()
	elementsDir := filepath.Join(tempDir, "elements")
	backupDir := filepath.Join(tempDir, "backups")

	repo := infrastructure.NewInMemoryElementRepository()
	backupSvc := NewBackupService(repo, elementsDir)

	// Create test elements
	persona := setupTestPersona("Test Persona", "Test Description")

	skill := domain.NewSkill("Test Skill", "Test Description", "1.0.0", "testuser")
	skill.AddTrigger(domain.SkillTrigger{
		Type:     "keyword",
		Keywords: []string{"test"},
	})
	skill.AddProcedure(domain.SkillProcedure{
		Step:   1,
		Action: "Test action",
	})

	if err := repo.Create(persona); err != nil {
		t.Fatalf("Failed to save persona: %v", err)
	}
	if err := repo.Create(skill); err != nil {
		t.Fatalf("Failed to save skill: %v", err)
	}

	// Create backup
	backupPath := filepath.Join(backupDir, "test-backup.tar.gz")
	metadata, err := backupSvc.Backup(backupPath, BackupOptions{
		Compression: "best",
		Description: "Test backup",
		Author:      "testuser",
	})

	if err != nil {
		t.Fatalf("Backup failed: %v", err)
	}

	// Verify metadata
	if metadata.ElementCount != 2 {
		t.Errorf("Expected 2 elements, got %d", metadata.ElementCount)
	}
	if metadata.Checksum == "" {
		t.Error("Expected checksum to be set")
	}
	if metadata.Description != "Test backup" {
		t.Errorf("Expected description 'Test backup', got '%s'", metadata.Description)
	}

	// Verify file exists
	if _, err := os.Stat(backupPath); os.IsNotExist(err) {
		t.Error("Backup file was not created")
	}

	// Verify file size
	info, _ := os.Stat(backupPath)
	if info.Size() == 0 {
		t.Error("Backup file is empty")
	}
}

func TestBackup_ValidateBackup(t *testing.T) {
	// Setup
	tempDir := t.TempDir()
	backupDir := filepath.Join(tempDir, "backups")

	repo := infrastructure.NewInMemoryElementRepository()
	backupSvc := NewBackupService(repo, tempDir)

	// Create a persona
	persona := setupTestPersona("Validate Test", "Description")
	repo.Create(persona)

	// Create backup
	backupPath := filepath.Join(backupDir, "validate-test.tar.gz")
	originalMetadata, err := backupSvc.Backup(backupPath, BackupOptions{
		Compression: "best",
		Author:      "testuser",
	})
	if err != nil {
		t.Fatalf("Backup failed: %v", err)
	}

	// Validate backup
	metadata, err := ValidateBackup(backupPath)
	if err != nil {
		t.Fatalf("Validation failed: %v", err)
	}

	// Compare metadata
	if metadata.ElementCount != originalMetadata.ElementCount {
		t.Errorf("Element count mismatch: %d vs %d", metadata.ElementCount, originalMetadata.ElementCount)
	}
	if metadata.Checksum != originalMetadata.Checksum {
		t.Errorf("Checksum mismatch: original='%s', validated='%s'", originalMetadata.Checksum, metadata.Checksum)
	}
	if metadata.Version != originalMetadata.Version {
		t.Errorf("Version mismatch")
	}
}

func TestBackup_CompressionOptions(t *testing.T) {
	tempDir := t.TempDir()
	repo := infrastructure.NewInMemoryElementRepository()
	backupSvc := NewBackupService(repo, tempDir)

	// Create test element
	persona := setupTestPersona("Compression Test", "Description")
	repo.Create(persona)

	tests := []struct {
		name        string
		compression string
	}{
		{"None", "none"},
		{"Fast", "fast"},
		{"Best", "best"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			backupPath := filepath.Join(tempDir, "backup-"+tt.compression+".tar.gz")

			_, err := backupSvc.Backup(backupPath, BackupOptions{
				Compression: tt.compression,
			})

			if err != nil {
				t.Fatalf("Backup with %s compression failed: %v", tt.compression, err)
			}

			// Verify file exists
			if _, err := os.Stat(backupPath); os.IsNotExist(err) {
				t.Errorf("Backup file not created for %s compression", tt.compression)
			}
		})
	}
}

func TestRestore_BasicRestore(t *testing.T) {
	// Setup
	tempDir := t.TempDir()
	elementsDir := filepath.Join(tempDir, "elements")
	backupDir := filepath.Join(tempDir, "backups")

	// Create original repository and backup
	originalRepo := infrastructure.NewInMemoryElementRepository()
	backupSvc := NewBackupService(originalRepo, elementsDir)

	persona := setupTestPersona("Restore Test", "Description")
	originalRepo.Create(persona)

	skill := domain.NewSkill("Test Skill", "Description", "1.0.0", "testuser")
	skill.AddTrigger(domain.SkillTrigger{
		Type:     "keyword",
		Keywords: []string{"test"},
	})
	skill.AddProcedure(domain.SkillProcedure{
		Step:   1,
		Action: "Test action",
	})
	originalRepo.Create(skill)

	backupPath := filepath.Join(backupDir, "restore-test.tar.gz")
	_, err := backupSvc.Backup(backupPath, BackupOptions{
		Compression: "best",
	})
	if err != nil {
		t.Fatalf("Backup failed: %v", err)
	}

	// Create new empty repository for restore
	newRepo := infrastructure.NewInMemoryElementRepository()
	restoreSvc := NewRestoreService(newRepo, backupSvc, elementsDir)

	// Restore
	result, err := restoreSvc.Restore(backupPath, RestoreOptions{
		Overwrite:    false,
		BackupBefore: false,
	})

	if err != nil {
		t.Fatalf("Restore failed: %v", err)
	}

	// Verify results
	if !result.Success {
		t.Error("Restore was not successful")
	}
	if result.ElementsAdded != 2 {
		t.Errorf("Expected 2 elements added, got %d", result.ElementsAdded)
	}
	if len(result.Errors) > 0 {
		t.Errorf("Unexpected errors: %v", result.Errors)
	}

	// Verify elements were restored
	elements, err := newRepo.List(domain.ElementFilter{})
	if err != nil {
		t.Fatalf("Failed to list elements: %v", err)
	}
	if len(elements) != 2 {
		t.Errorf("Expected 2 elements in repository, got %d", len(elements))
	}
}

func TestRestore_OverwriteExisting(t *testing.T) {
	tempDir := t.TempDir()
	elementsDir := filepath.Join(tempDir, "elements")
	backupDir := filepath.Join(tempDir, "backups")

	// Create backup with original data
	originalRepo := infrastructure.NewInMemoryElementRepository()
	backupSvc := NewBackupService(originalRepo, elementsDir)

	persona := setupTestPersona("Overwrite Test", "Original Description")
	originalRepo.Create(persona)
	originalID := persona.GetID()

	backupPath := filepath.Join(backupDir, "overwrite-test.tar.gz")
	backupSvc.Backup(backupPath, BackupOptions{Compression: "best"})

	// Create repository with modified element using the same ID
	modifiedRepo := infrastructure.NewInMemoryElementRepository()
	modifiedPersona := setupTestPersona("Overwrite Test", "Modified Description")

	// Set the same ID as the original
	meta := modifiedPersona.GetMetadata()
	meta.ID = originalID
	meta.CreatedAt = time.Now()
	meta.UpdatedAt = time.Now()
	modifiedPersona.SetMetadata(meta)

	modifiedRepo.Create(modifiedPersona)

	// Restore with overwrite
	restoreSvc := NewRestoreService(modifiedRepo, backupSvc, elementsDir)
	result, err := restoreSvc.Restore(backupPath, RestoreOptions{
		Overwrite:    true,
		BackupBefore: false,
	})

	if err != nil {
		t.Fatalf("Restore failed: %v", err)
	}

	if result.ElementsUpdated != 1 {
		t.Errorf("Expected 1 element updated, got %d", result.ElementsUpdated)
	}

	// Verify description was restored to original
	restored, _ := modifiedRepo.GetByID(originalID)
	if restored.GetMetadata().Description != "Original Description" {
		t.Error("Element was not properly overwritten")
	}
}

func TestRestore_DryRun(t *testing.T) {
	tempDir := t.TempDir()
	elementsDir := filepath.Join(tempDir, "elements")
	backupDir := filepath.Join(tempDir, "backups")

	// Create backup
	originalRepo := infrastructure.NewInMemoryElementRepository()
	backupSvc := NewBackupService(originalRepo, elementsDir)

	persona := setupTestPersona("DryRun Test", "Description")
	originalRepo.Create(persona)

	backupPath := filepath.Join(backupDir, "dryrun-test.tar.gz")
	backupSvc.Backup(backupPath, BackupOptions{Compression: "best"})

	// Dry run restore
	emptyRepo := infrastructure.NewInMemoryElementRepository()
	restoreSvc := NewRestoreService(emptyRepo, backupSvc, elementsDir)
	result, err := restoreSvc.Restore(backupPath, RestoreOptions{
		DryRun:       true,
		BackupBefore: false,
	})

	if err != nil {
		t.Fatalf("Dry run failed: %v", err)
	}

	// Verify nothing was actually restored
	elements, _ := emptyRepo.List(domain.ElementFilter{})
	if len(elements) != 0 {
		t.Error("Elements were restored during dry run")
	}

	// But result should show what would have been added
	if result.ElementsAdded != 1 {
		t.Errorf("Dry run should have reported 1 element to add, got %d", result.ElementsAdded)
	}
}

func TestRestore_MergeStrategy(t *testing.T) {
	tempDir := t.TempDir()
	elementsDir := filepath.Join(tempDir, "elements")
	backupDir := filepath.Join(tempDir, "backups")

	// Create backup
	originalRepo := infrastructure.NewInMemoryElementRepository()
	backupSvc := NewBackupService(originalRepo, elementsDir)

	persona := setupTestPersona("Merge Test", "Original")
	originalRepo.Create(persona)
	originalID := persona.GetID()

	backupPath := filepath.Join(backupDir, "merge-test.tar.gz")
	backupSvc.Backup(backupPath, BackupOptions{Compression: "best"})

	tests := []struct {
		name          string
		strategy      string
		expectUpdated int
		expectSkipped int
	}{
		{"Skip Strategy", "skip", 0, 1},
		{"Overwrite Strategy", "overwrite", 1, 0},
		{"Merge Strategy", "merge", 1, 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create repo with existing element
			repo := infrastructure.NewInMemoryElementRepository()
			existing := setupTestPersona("Merge Test", "Modified")

			// Set the same ID
			meta := existing.GetMetadata()
			meta.ID = originalID
			meta.CreatedAt = time.Now()
			meta.UpdatedAt = time.Now()
			existing.SetMetadata(meta)

			repo.Create(existing)

			// Restore with specific strategy
			restoreSvc := NewRestoreService(repo, backupSvc, elementsDir)
			result, err := restoreSvc.Restore(backupPath, RestoreOptions{
				MergeStrategy: tt.strategy,
				BackupBefore:  false,
			})

			if err != nil {
				t.Fatalf("Restore failed: %v", err)
			}

			if result.ElementsUpdated != tt.expectUpdated {
				t.Errorf("Expected %d updated, got %d", tt.expectUpdated, result.ElementsUpdated)
			}
			if result.ElementsSkipped != tt.expectSkipped {
				t.Errorf("Expected %d skipped, got %d", tt.expectSkipped, result.ElementsSkipped)
			}
		})
	}
}
