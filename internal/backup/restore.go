package backup

import (
	"archive/tar"
	"compress/gzip"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/fsvxavier/nexs-mcp/internal/domain"
)

// RestoreOptions configures restore behavior
type RestoreOptions struct {
	Overwrite      bool   // Overwrite existing elements (default: false)
	SkipValidation bool   // Skip validation checks (default: false)
	MergeStrategy  string // merge, overwrite, skip (default: skip)
	BackupBefore   bool   // Create backup before restore (default: true)
	DryRun         bool   // Don't actually restore, just validate (default: false)
}

// RestoreResult contains information about the restore operation
type RestoreResult struct {
	Success         bool            `json:"success"`
	ElementsAdded   int             `json:"elements_added"`
	ElementsUpdated int             `json:"elements_updated"`
	ElementsSkipped int             `json:"elements_skipped"`
	Errors          []string        `json:"errors,omitempty"`
	BackupPath      string          `json:"backup_path,omitempty"`
	Duration        time.Duration   `json:"duration"`
	Metadata        *BackupMetadata `json:"metadata"`
}

// RestoreService handles portfolio restoration
type RestoreService struct {
	repository    domain.ElementRepository
	backupService *BackupService
	baseDir       string
}

// NewRestoreService creates a new restore service
func NewRestoreService(repo domain.ElementRepository, backupSvc *BackupService, baseDir string) *RestoreService {
	return &RestoreService{
		repository:    repo,
		backupService: backupSvc,
		baseDir:       baseDir,
	}
}

// Restore restores a portfolio from a backup file
func (s *RestoreService) Restore(backupPath string, options RestoreOptions) (*RestoreResult, error) {
	startTime := time.Now()
	result := &RestoreResult{
		Success: false,
	}

	// Validate backup file
	metadata, err := ValidateBackup(backupPath)
	if err != nil {
		return nil, fmt.Errorf("invalid backup file: %w", err)
	}
	result.Metadata = metadata

	// Create pre-restore backup if requested
	if options.BackupBefore && !options.DryRun {
		preRestoreBackupPath := filepath.Join(s.baseDir, ".backups",
			fmt.Sprintf("pre-restore-%s.tar.gz", time.Now().Format("20060102-150405")))

		if err := os.MkdirAll(filepath.Dir(preRestoreBackupPath), 0755); err != nil {
			return nil, fmt.Errorf("failed to create backup directory: %w", err)
		}

		_, err := s.backupService.Backup(preRestoreBackupPath, BackupOptions{
			Compression: "best",
			Description: "Automatic backup before restore",
		})
		if err != nil {
			return nil, fmt.Errorf("failed to create pre-restore backup: %w", err)
		}
		result.BackupPath = preRestoreBackupPath
	}

	// Open backup file
	file, err := os.Open(backupPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open backup: %w", err)
	}
	defer file.Close()

	// Decompress
	gzReader, err := gzip.NewReader(file)
	if err != nil {
		// Try uncompressed
		file.Seek(0, 0)
		if err := s.restoreFromTar(file, options, result); err != nil {
			return result, err
		}
	} else {
		defer gzReader.Close()
		if err := s.restoreFromTar(gzReader, options, result); err != nil {
			return result, err
		}
	}

	// Verify checksum if not skipping validation
	if !options.SkipValidation && !options.DryRun {
		if err := s.verifyRestoreChecksum(backupPath, metadata.Checksum); err != nil {
			result.Errors = append(result.Errors, fmt.Sprintf("checksum verification failed: %v", err))
			// Don't fail completely, just warn
		}
	}

	result.Success = len(result.Errors) == 0 || result.ElementsAdded > 0
	result.Duration = time.Since(startTime)

	return result, nil
}

// restoreFromTar processes the tar archive and restores elements
func (s *RestoreService) restoreFromTar(reader io.Reader, options RestoreOptions, result *RestoreResult) error {
	tarReader := tar.NewReader(reader)

	for {
		header, err := tarReader.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return fmt.Errorf("failed to read tar: %w", err)
		}

		// Skip metadata and backup directories
		if header.Name == "metadata.json" || strings.HasPrefix(header.Name, "backups/") {
			continue
		}

		// Only process element files
		if !strings.HasPrefix(header.Name, "elements/") || !strings.HasSuffix(header.Name, ".json") {
			continue
		}

		// Read element data
		data, err := io.ReadAll(tarReader)
		if err != nil {
			result.Errors = append(result.Errors, fmt.Sprintf("failed to read %s: %v", header.Name, err))
			continue
		}

		// Parse element (now wrapped with metadata)
		var wrappedElement map[string]interface{}
		if err := json.Unmarshal(data, &wrappedElement); err != nil {
			result.Errors = append(result.Errors, fmt.Sprintf("failed to parse %s: %v", header.Name, err))
			continue
		}

		// Extract metadata
		metadataRaw, ok := wrappedElement["metadata"]
		if !ok {
			result.Errors = append(result.Errors, fmt.Sprintf("missing metadata in %s", header.Name))
			continue
		}

		metadataBytes, err := json.Marshal(metadataRaw)
		if err != nil {
			result.Errors = append(result.Errors, fmt.Sprintf("failed to marshal metadata in %s: %v", header.Name, err))
			continue
		}

		var metadata domain.ElementMetadata
		if err := json.Unmarshal(metadataBytes, &metadata); err != nil {
			result.Errors = append(result.Errors, fmt.Sprintf("failed to parse metadata in %s: %v", header.Name, err))
			continue
		}

		// Extract element data
		elementRaw, ok := wrappedElement["element"]
		if !ok {
			result.Errors = append(result.Errors, fmt.Sprintf("missing element in %s", header.Name))
			continue
		}

		elementBytes, err := json.Marshal(elementRaw)
		if err != nil {
			result.Errors = append(result.Errors, fmt.Sprintf("failed to marshal element in %s: %v", header.Name, err))
			continue
		}

		// Create typed element
		element, err := s.createTypedElement(string(metadata.Type), elementBytes, metadata)
		if err != nil {
			result.Errors = append(result.Errors, fmt.Sprintf("failed to create element from %s: %v", header.Name, err))
			continue
		}

		// Validate element if not skipping
		if !options.SkipValidation {
			if err := element.Validate(); err != nil {
				result.Errors = append(result.Errors, fmt.Sprintf("validation failed for %s: %v", header.Name, err))
				continue
			}
		}

		// Don't actually save in dry run mode
		if options.DryRun {
			result.ElementsAdded++
			continue
		}

		// Check if element already exists
		elementID := element.GetID()
		existing, err := s.repository.GetByID(elementID)

		if err == nil && existing != nil {
			// Element exists - handle based on strategy
			switch options.MergeStrategy {
			case "skip":
				result.ElementsSkipped++
				continue
			case "overwrite", "merge":
				// Update existing element
				if err := s.repository.Update(element); err != nil {
					result.Errors = append(result.Errors, fmt.Sprintf("failed to update %s: %v", elementID, err))
				} else {
					result.ElementsUpdated++
				}
			default:
				if options.Overwrite {
					if err := s.repository.Update(element); err != nil {
						result.Errors = append(result.Errors, fmt.Sprintf("failed to update %s: %v", elementID, err))
					} else {
						result.ElementsUpdated++
					}
				} else {
					result.ElementsSkipped++
				}
			}
		} else {
			// Element doesn't exist - create it
			if err := s.repository.Create(element); err != nil {
				result.Errors = append(result.Errors, fmt.Sprintf("failed to create %s: %v", elementID, err))
			} else {
				result.ElementsAdded++
			}
		}
	}

	return nil
}

// createTypedElement creates a typed element from JSON data
func (s *RestoreService) createTypedElement(elementType string, data []byte, metadata domain.ElementMetadata) (domain.Element, error) {
	switch domain.ElementType(elementType) {
	case domain.PersonaElement:
		var persona domain.Persona
		if err := json.Unmarshal(data, &persona); err != nil {
			return nil, err
		}
		persona.SetMetadata(metadata)
		return &persona, nil

	case domain.SkillElement:
		var skill domain.Skill
		if err := json.Unmarshal(data, &skill); err != nil {
			return nil, err
		}
		skill.SetMetadata(metadata)
		return &skill, nil

	case domain.TemplateElement:
		var template domain.Template
		if err := json.Unmarshal(data, &template); err != nil {
			return nil, err
		}
		template.SetMetadata(metadata)
		return &template, nil

	case domain.AgentElement:
		var agent domain.Agent
		if err := json.Unmarshal(data, &agent); err != nil {
			return nil, err
		}
		agent.SetMetadata(metadata)
		return &agent, nil

	case domain.MemoryElement:
		var memory domain.Memory
		if err := json.Unmarshal(data, &memory); err != nil {
			return nil, err
		}
		memory.SetMetadata(metadata)
		return &memory, nil

	case domain.EnsembleElement:
		var ensemble domain.Ensemble
		if err := json.Unmarshal(data, &ensemble); err != nil {
			return nil, err
		}
		ensemble.SetMetadata(metadata)
		return &ensemble, nil

	default:
		return nil, fmt.Errorf("unknown element type: %s", elementType)
	}
}

// verifyRestoreChecksum verifies the checksum of restored data
func (s *RestoreService) verifyRestoreChecksum(backupPath, expectedChecksum string) error {
	file, err := os.Open(backupPath)
	if err != nil {
		return err
	}
	defer file.Close()

	hash := sha256.New()

	// Decompress if needed
	gzReader, err := gzip.NewReader(file)
	var reader io.Reader
	if err != nil {
		// Not compressed
		file.Seek(0, 0)
		reader = file
	} else {
		defer gzReader.Close()
		reader = gzReader
	}

	// Read tar and hash the file contents (not the tar structure)
	tarReader := tar.NewReader(reader)
	for {
		header, err := tarReader.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return fmt.Errorf("failed to read tar: %w", err)
		}

		// Only hash file contents, not directories or metadata.json
		if !header.FileInfo().IsDir() && !strings.HasSuffix(header.Name, "metadata.json") {
			if _, err := io.Copy(hash, tarReader); err != nil {
				return fmt.Errorf("failed to hash file %s: %w", header.Name, err)
			}
		}
	}

	actualChecksum := hex.EncodeToString(hash.Sum(nil))
	if actualChecksum != expectedChecksum {
		return fmt.Errorf("checksum mismatch: expected %s, got %s", expectedChecksum, actualChecksum)
	}

	return nil
}

// Rollback restores from the most recent backup
func (s *RestoreService) Rollback() (*RestoreResult, error) {
	backupDir := filepath.Join(s.baseDir, ".backups")

	// Find most recent backup
	entries, err := os.ReadDir(backupDir)
	if err != nil {
		return nil, fmt.Errorf("failed to read backup directory: %w", err)
	}

	var mostRecent string
	var mostRecentTime time.Time

	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".tar.gz") {
			continue
		}

		info, err := entry.Info()
		if err != nil {
			continue
		}

		if info.ModTime().After(mostRecentTime) {
			mostRecent = filepath.Join(backupDir, entry.Name())
			mostRecentTime = info.ModTime()
		}
	}

	if mostRecent == "" {
		return nil, fmt.Errorf("no backups found")
	}

	// Restore from most recent backup
	return s.Restore(mostRecent, RestoreOptions{
		Overwrite:     true,
		MergeStrategy: "overwrite",
		BackupBefore:  false, // Don't backup before rollback
	})
}
