package backup

import (
	"archive/tar"
	"compress/gzip"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"

	"github.com/fsvxavier/nexs-mcp/internal/domain"
)

// BackupMetadata contains metadata about a backup.
type BackupMetadata struct {
	Version      string    `json:"version"`
	CreatedAt    time.Time `json:"created_at"`
	ElementCount int       `json:"element_count"`
	TotalSize    int64     `json:"total_size"`
	Checksum     string    `json:"checksum"`
	NexsVersion  string    `json:"nexs_version"`
	BackupType   string    `json:"backup_type"` // full, incremental
	Description  string    `json:"description,omitempty"`
	Author       string    `json:"author,omitempty"`
}

// BackupOptions configures backup behavior.
type BackupOptions struct {
	IncludeBackups bool   // Include previous backups in this backup
	Compression    string // none, fast, best (default: best)
	Description    string // User description of backup
	Author         string // Backup author
}

// BackupService handles portfolio backup operations.
type BackupService struct {
	repository domain.ElementRepository
	baseDir    string // Base directory for elements
}

// NewBackupService creates a new backup service.
func NewBackupService(repo domain.ElementRepository, baseDir string) *BackupService {
	return &BackupService{
		repository: repo,
		baseDir:    baseDir,
	}
}

// Backup creates a complete backup of the portfolio.
func (s *BackupService) Backup(outputPath string, options BackupOptions) (*BackupMetadata, error) {
	// Set default compression
	if options.Compression == "" {
		options.Compression = "best"
	}

	// Create output directory if needed
	outputDir := filepath.Dir(outputPath)
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create output directory: %w", err)
	}

	// Create temporary file for atomic write
	tempFile := outputPath + ".tmp"
	file, err := os.Create(tempFile)
	if err != nil {
		return nil, fmt.Errorf("failed to create backup file: %w", err)
	}
	defer func() {
		file.Close()
		os.Remove(tempFile) // Cleanup on error
	}()

	// Setup gzip compression
	var gzWriter *gzip.Writer
	var writer io.Writer = file

	if options.Compression != "none" {
		level := gzip.BestCompression
		if options.Compression == "fast" {
			level = gzip.BestSpeed
		}
		gzWriter, err = gzip.NewWriterLevel(file, level)
		if err != nil {
			return nil, fmt.Errorf("failed to create gzip writer: %w", err)
		}
		defer gzWriter.Close()
		writer = gzWriter
	}

	// Create tar archive
	tarWriter := tar.NewWriter(writer)
	defer tarWriter.Close()

	// Collect all elements
	elements, err := s.repository.List(domain.ElementFilter{})
	if err != nil {
		return nil, fmt.Errorf("failed to list elements: %w", err)
	}

	var totalSize int64
	checksumHash := sha256.New()

	// PRE-CALCULATE: Serialize all elements and calculate checksum before writing to tar
	type elementData struct {
		fileName string
		data     []byte
	}
	var elementDataList []elementData

	for _, element := range elements {
		// Create a wrapped structure that includes metadata explicitly
		wrappedElement := map[string]interface{}{
			"metadata": element.GetMetadata(),
			"element":  element,
		}

		data, err := json.MarshalIndent(wrappedElement, "", "  ")
		if err != nil {
			return nil, fmt.Errorf("failed to marshal element %s: %w", element.GetID(), err)
		}

		// Create path based on element metadata
		meta := element.GetMetadata()
		fileName := fmt.Sprintf("elements/%s/%s/%s.json",
			meta.Author,
			meta.Type,
			meta.ID,
		)

		elementDataList = append(elementDataList, elementData{
			fileName: fileName,
			data:     data,
		})

		// Add to checksum and size
		checksumHash.Write(data)
		totalSize += int64(len(data))
	}

	// Now create metadata with final checksum
	metadata := &BackupMetadata{
		Version:      "1.0",
		CreatedAt:    time.Now(),
		ElementCount: len(elements),
		TotalSize:    totalSize,
		Checksum:     hex.EncodeToString(checksumHash.Sum(nil)),
		NexsVersion:  "0.4.0-dev", // TODO: Get from version constant
		BackupType:   "full",
		Description:  options.Description,
		Author:       options.Author,
	}

	metadataBytes, err := json.MarshalIndent(metadata, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("failed to marshal metadata: %w", err)
	}

	// Write metadata.json to tar (now with correct checksum)
	if err := s.addFileToTar(tarWriter, "metadata.json", metadataBytes, nil); err != nil {
		return nil, fmt.Errorf("failed to write metadata: %w", err)
	}

	// Write all pre-serialized elements
	for _, ed := range elementDataList {
		if err := s.addFileToTar(tarWriter, ed.fileName, ed.data, nil); err != nil {
			return nil, fmt.Errorf("failed to write element: %w", err)
		}
	}

	// Optionally backup previous backups (checksums already calculated)
	if options.IncludeBackups && s.baseDir != "" {
		backupDir := filepath.Join(s.baseDir, ".backups")
		// Note: We're not updating checksumHash here since it was finalized above
		if err := s.addDirectoryToTar(tarWriter, backupDir, "backups/", nil, &totalSize); err != nil {
			// Non-fatal: just log and continue
			fmt.Fprintf(os.Stderr, "Warning: failed to include previous backups: %v\n", err)
		}
	}

	// Close tar and gzip writers
	if err := tarWriter.Close(); err != nil {
		return nil, fmt.Errorf("failed to close tar writer: %w", err)
	}
	if gzWriter != nil {
		if err := gzWriter.Close(); err != nil {
			return nil, fmt.Errorf("failed to close gzip writer: %w", err)
		}
	}
	if err := file.Close(); err != nil {
		return nil, fmt.Errorf("failed to close file: %w", err)
	}

	// Atomic rename
	if err := os.Rename(tempFile, outputPath); err != nil {
		return nil, fmt.Errorf("failed to finalize backup: %w", err)
	}

	return metadata, nil
}

// addFileToTar adds a single file to the tar archive.
func (s *BackupService) addFileToTar(tw *tar.Writer, name string, data []byte, hash io.Writer) error {
	header := &tar.Header{
		Name:    name,
		Size:    int64(len(data)),
		Mode:    0644,
		ModTime: time.Now(),
	}

	if err := tw.WriteHeader(header); err != nil {
		return err
	}

	if _, err := tw.Write(data); err != nil {
		return err
	}

	// Update checksum
	if hash != nil {
		hash.Write(data)
	}

	return nil
}

// addDirectoryToTar recursively adds a directory to tar.
func (s *BackupService) addDirectoryToTar(tw *tar.Writer, srcDir, destPrefix string, hash io.Writer, totalSize *int64) error {
	return filepath.Walk(srcDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Skip directories themselves
		if info.IsDir() {
			return nil
		}

		// Read file
		data, err := os.ReadFile(path)
		if err != nil {
			return fmt.Errorf("failed to read %s: %w", path, err)
		}

		// Calculate relative path
		relPath, err := filepath.Rel(srcDir, path)
		if err != nil {
			return err
		}
		tarPath := filepath.Join(destPrefix, relPath)

		// Add to tar
		if err := s.addFileToTar(tw, tarPath, data, hash); err != nil {
			return err
		}

		*totalSize += int64(len(data))
		return nil
	})
}

// ValidateBackup checks if a backup file is valid.
func ValidateBackup(backupPath string) (*BackupMetadata, error) {
	file, err := os.Open(backupPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open backup: %w", err)
	}
	defer file.Close()

	// Try to decompress
	gzReader, err := gzip.NewReader(file)
	if err != nil {
		// Maybe not compressed
		file.Seek(0, 0)
		return extractMetadata(file)
	}
	defer gzReader.Close()

	return extractMetadata(gzReader)
}

// extractMetadata extracts metadata from tar archive.
func extractMetadata(reader io.Reader) (*BackupMetadata, error) {
	tarReader := tar.NewReader(reader)

	for {
		header, err := tarReader.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("failed to read tar: %w", err)
		}

		if header.Name == "metadata.json" {
			data, err := io.ReadAll(tarReader)
			if err != nil {
				return nil, fmt.Errorf("failed to read metadata: %w", err)
			}

			var metadata BackupMetadata
			if err := json.Unmarshal(data, &metadata); err != nil {
				return nil, fmt.Errorf("failed to parse metadata: %w", err)
			}

			return &metadata, nil
		}
	}

	return nil, errors.New("metadata.json not found in backup")
}
