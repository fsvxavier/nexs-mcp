package backup

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/fsvxavier/nexs-mcp/internal/domain"
	"github.com/fsvxavier/nexs-mcp/internal/infrastructure"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

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

func TestRestore_InvalidBackup(t *testing.T) {
	tempDir := t.TempDir()
	elementsDir := filepath.Join(tempDir, "elements")

	repo := infrastructure.NewInMemoryElementRepository()
	backupSvc := NewBackupService(repo, elementsDir)
	restoreSvc := NewRestoreService(repo, backupSvc, elementsDir)

	// Create invalid backup file
	invalidPath := filepath.Join(tempDir, "invalid.tar.gz")
	os.WriteFile(invalidPath, []byte("not a valid tar.gz"), 0644)

	_, err := restoreSvc.Restore(invalidPath, RestoreOptions{
		BackupBefore: false,
	})

	if err == nil {
		t.Error("Expected error when restoring invalid backup")
	}
}

func TestRestore_NonExistentFile(t *testing.T) {
	tempDir := t.TempDir()
	elementsDir := filepath.Join(tempDir, "elements")

	repo := infrastructure.NewInMemoryElementRepository()
	backupSvc := NewBackupService(repo, elementsDir)
	restoreSvc := NewRestoreService(repo, backupSvc, elementsDir)

	_, err := restoreSvc.Restore("/nonexistent/backup.tar.gz", RestoreOptions{
		BackupBefore: false,
	})

	if err == nil {
		t.Error("Expected error when restoring non-existent file")
	}
}

func TestRestore_SkipValidation(t *testing.T) {
	tempDir := t.TempDir()
	elementsDir := filepath.Join(tempDir, "elements")
	backupDir := filepath.Join(tempDir, "backups")

	// Create backup
	originalRepo := infrastructure.NewInMemoryElementRepository()
	backupSvc := NewBackupService(originalRepo, elementsDir)

	persona := setupTestPersona("Skip Validation Test", "Description")
	originalRepo.Create(persona)

	backupPath := filepath.Join(backupDir, "skipval-test.tar.gz")
	backupSvc.Backup(backupPath, BackupOptions{Compression: "best"})

	// Restore with skip validation
	newRepo := infrastructure.NewInMemoryElementRepository()
	restoreSvc := NewRestoreService(newRepo, backupSvc, elementsDir)
	result, err := restoreSvc.Restore(backupPath, RestoreOptions{
		SkipValidation: true,
		BackupBefore:   false,
	})

	if err != nil {
		t.Fatalf("Restore failed: %v", err)
	}

	if !result.Success {
		t.Error("Restore should succeed with skip validation")
	}
}

func TestRollback_NoBackups(t *testing.T) {
	tempDir := t.TempDir()
	elementsDir := filepath.Join(tempDir, "elements")

	// Create .backups directory but leave it empty
	backupDir := filepath.Join(elementsDir, ".backups")
	os.MkdirAll(backupDir, 0755)

	repo := infrastructure.NewInMemoryElementRepository()
	backupSvc := NewBackupService(repo, elementsDir)
	restoreSvc := NewRestoreService(repo, backupSvc, elementsDir)

	_, err := restoreSvc.Rollback()

	if err == nil {
		t.Error("Expected error when rolling back with no backups")
	}
	if err != nil && err.Error() != "no backups found" {
		t.Errorf("Expected 'no backups found' error, got: %v", err)
	}
}

func TestRollback_WithBackup(t *testing.T) {
	tempDir := t.TempDir()
	elementsDir := filepath.Join(tempDir, "elements")
	backupDir := filepath.Join(elementsDir, ".backups")

	// Create backup directory
	os.MkdirAll(backupDir, 0755)

	// Create backup with element
	originalRepo := infrastructure.NewInMemoryElementRepository()
	backupSvc := NewBackupService(originalRepo, elementsDir)

	persona := setupTestPersona("Rollback Test", "Description")
	originalRepo.Create(persona)

	backupPath := filepath.Join(backupDir, "rollback-test.tar.gz")
	backupSvc.Backup(backupPath, BackupOptions{Compression: "best"})

	// Create empty repository and restore service
	newRepo := infrastructure.NewInMemoryElementRepository()
	restoreSvc := NewRestoreService(newRepo, backupSvc, elementsDir)

	// Perform rollback
	result, err := restoreSvc.Rollback()

	if err != nil {
		t.Fatalf("Rollback failed: %v", err)
	}

	if !result.Success {
		t.Error("Rollback should succeed")
	}

	if result.ElementsAdded != 1 {
		t.Errorf("Expected 1 element restored, got %d", result.ElementsAdded)
	}
}

func TestRestoreOptions_Defaults(t *testing.T) {
	opts := RestoreOptions{}

	if opts.Overwrite {
		t.Error("Overwrite should default to false")
	}
	if opts.SkipValidation {
		t.Error("SkipValidation should default to false")
	}
	if opts.MergeStrategy != "" {
		t.Error("MergeStrategy should default to empty string")
	}
	if opts.BackupBefore {
		t.Error("BackupBefore should default to false")
	}
	if opts.DryRun {
		t.Error("DryRun should default to false")
	}
}

func TestRestoreResult_Structure(t *testing.T) {
	result := &RestoreResult{
		Success:         true,
		ElementsAdded:   5,
		ElementsUpdated: 2,
		ElementsSkipped: 1,
		Errors:          []string{"error1"},
		BackupPath:      "/path/to/backup",
		Duration:        time.Second * 10,
		Metadata: &BackupMetadata{
			Version:      "1.0.0",
			ElementCount: 8,
		},
	}

	if !result.Success {
		t.Error("Success should be true")
	}
	if result.ElementsAdded != 5 {
		t.Errorf("Expected 5 added, got %d", result.ElementsAdded)
	}
	if result.ElementsUpdated != 2 {
		t.Errorf("Expected 2 updated, got %d", result.ElementsUpdated)
	}
	if result.ElementsSkipped != 1 {
		t.Errorf("Expected 1 skipped, got %d", result.ElementsSkipped)
	}
	if len(result.Errors) != 1 {
		t.Errorf("Expected 1 error, got %d", len(result.Errors))
	}
	if result.BackupPath == "" {
		t.Error("BackupPath should not be empty")
	}
	if result.Duration == 0 {
		t.Error("Duration should not be zero")
	}
	if result.Metadata == nil {
		t.Error("Metadata should not be nil")
	}
}

func TestNewRestoreService(t *testing.T) {
	tempDir := t.TempDir()
	repo := infrastructure.NewInMemoryElementRepository()
	backupSvc := NewBackupService(repo, tempDir)

	restoreSvc := NewRestoreService(repo, backupSvc, tempDir)

	require.NotNil(t, restoreSvc)
	assert.NotNil(t, restoreSvc.repository)
	assert.NotNil(t, restoreSvc.backupService)
	assert.Equal(t, tempDir, restoreSvc.baseDir)
}
