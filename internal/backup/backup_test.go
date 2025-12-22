package backup

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/fsvxavier/nexs-mcp/internal/domain"
	"github.com/fsvxavier/nexs-mcp/internal/infrastructure"
)

// Helper function to setup a valid persona for testing.
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
