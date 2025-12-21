package sources

import (
	"archive/tar"
	"compress/gzip"
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

// LocalSource implements CollectionSource for local filesystem collections.
type LocalSource struct {
	searchPaths []string // Directories to search for collections
}

// NewLocalSource creates a new local filesystem collection source.
func NewLocalSource(searchPaths []string) (*LocalSource, error) {
	if len(searchPaths) == 0 {
		// Use default search paths
		homeDir, err := os.UserHomeDir()
		if err != nil {
			return nil, fmt.Errorf("failed to get home directory: %w", err)
		}
		searchPaths = []string{
			filepath.Join(homeDir, ".nexs", "collections"),
		}
	}

	// Ensure all search paths exist (skip paths we can't create)
	validPaths := make([]string, 0, len(searchPaths))
	for _, path := range searchPaths {
		if err := os.MkdirAll(path, 0755); err != nil {
			// Skip paths we can't create (e.g., system directories)
			continue
		}
		validPaths = append(validPaths, path)
	}

	if len(validPaths) == 0 {
		return nil, errors.New("no valid search paths available")
	}

	return &LocalSource{
		searchPaths: validPaths,
	}, nil
}

// Name returns the unique name of this source.
func (s *LocalSource) Name() string {
	return "local"
}

// Supports returns true if this source can handle the given URI.
func (s *LocalSource) Supports(uri string) bool {
	return strings.HasPrefix(uri, "file://") || filepath.IsAbs(uri)
}

// Browse discovers collections from local filesystem.
func (s *LocalSource) Browse(ctx context.Context, filter *BrowseFilter) ([]*CollectionMetadata, error) {
	var collections []*CollectionMetadata

	for _, searchPath := range s.searchPaths {
		// Walk through directory looking for collection.yaml files
		err := filepath.Walk(searchPath, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return nil // Skip directories with errors
			}

			// Check for context cancellation
			select {
			case <-ctx.Done():
				return ctx.Err()
			default:
			}

			// Look for collection.yaml files
			if info.IsDir() || filepath.Base(path) != "collection.yaml" {
				return nil
			}

			// Parse manifest
			metadata, err := s.parseManifestFile(path)
			if err != nil {
				return nil // Skip invalid manifests
			}

			// Apply filters
			if !s.matchesFilter(metadata, filter) {
				return nil
			}

			// Set URI to the collection directory
			collectionDir := filepath.Dir(path)
			metadata.URI = "file://" + collectionDir
			metadata.SourceName = "local"
			metadata.Repository = collectionDir

			collections = append(collections, metadata)
			return nil
		})

		if err != nil {
			return nil, fmt.Errorf("failed to scan %s: %w", searchPath, err)
		}
	}

	// Apply limit and offset if specified
	if filter != nil {
		if filter.Offset > 0 && filter.Offset < len(collections) {
			collections = collections[filter.Offset:]
		} else if filter.Offset >= len(collections) {
			collections = []*CollectionMetadata{}
		}

		if filter.Limit > 0 && filter.Limit < len(collections) {
			collections = collections[:filter.Limit]
		}
	}

	return collections, nil
}

// Get retrieves a specific collection by URI.
func (s *LocalSource) Get(ctx context.Context, uri string) (*Collection, error) {
	// Parse URI
	path, err := s.parseURI(uri)
	if err != nil {
		return nil, fmt.Errorf("invalid URI: %w", err)
	}

	// Check if path is a tar.gz file
	if strings.HasSuffix(path, ".tar.gz") || strings.HasSuffix(path, ".tgz") {
		return s.getFromArchive(ctx, path, uri)
	}

	// Assume it's a directory
	return s.getFromDirectory(ctx, path, uri)
}

// getFromDirectory loads a collection from a directory.
func (s *LocalSource) getFromDirectory(ctx context.Context, dirPath, uri string) (*Collection, error) {
	// Read collection.yaml
	manifestPath := filepath.Join(dirPath, "collection.yaml")
	data, err := os.ReadFile(manifestPath)
	if err != nil {
		return nil, fmt.Errorf("collection.yaml not found: %w", err)
	}

	var manifestMap map[string]interface{}
	if err := yaml.Unmarshal(data, &manifestMap); err != nil {
		return nil, fmt.Errorf("failed to parse manifest: %w", err)
	}

	// Extract fields
	metadata, err := s.extractMetadata(manifestMap, uri, dirPath)
	if err != nil {
		return nil, err
	}

	return &Collection{
		Metadata:   metadata,
		Manifest:   manifestMap,
		SourceData: map[string]interface{}{"path": dirPath},
	}, nil
}

// getFromArchive loads a collection from a tar.gz archive.
func (s *LocalSource) getFromArchive(ctx context.Context, archivePath, uri string) (*Collection, error) {
	// Create temp directory for extraction
	tempDir, err := os.MkdirTemp("", "nexs-mcp-collection-*")
	if err != nil {
		return nil, fmt.Errorf("failed to create temp directory: %w", err)
	}
	defer func() {
		_ = os.RemoveAll(tempDir) // Cleanup temp directory
	}()

	// Extract archive
	if err := s.extractTarGz(archivePath, tempDir); err != nil {
		return nil, fmt.Errorf("failed to extract archive: %w", err)
	}

	// Look for collection.yaml in extracted files
	manifestPath := filepath.Join(tempDir, "collection.yaml")
	if _, err := os.Stat(manifestPath); os.IsNotExist(err) {
		// Try looking in subdirectories (in case archive has a root folder)
		entries, err := os.ReadDir(tempDir)
		if err != nil {
			return nil, fmt.Errorf("failed to read temp directory: %w", err)
		}

		for _, entry := range entries {
			if entry.IsDir() {
				potentialPath := filepath.Join(tempDir, entry.Name(), "collection.yaml")
				if _, err := os.Stat(potentialPath); err == nil {
					manifestPath = potentialPath
					tempDir = filepath.Join(tempDir, entry.Name())
					break
				}
			}
		}
	}

	// Read and parse manifest
	data, err := os.ReadFile(manifestPath)
	if err != nil {
		return nil, fmt.Errorf("collection.yaml not found in archive: %w", err)
	}

	var manifestMap map[string]interface{}
	if err := yaml.Unmarshal(data, &manifestMap); err != nil {
		return nil, fmt.Errorf("failed to parse manifest: %w", err)
	}

	metadata, err := s.extractMetadata(manifestMap, uri, archivePath)
	if err != nil {
		return nil, err
	}

	return &Collection{
		Metadata: metadata,
		Manifest: manifestMap,
		SourceData: map[string]interface{}{
			"path":         tempDir,
			"archive_path": archivePath,
			"temporary":    true,
		},
	}, nil
}

// parseURI converts a file:// URI to a filesystem path.
func (s *LocalSource) parseURI(uri string) (string, error) {
	if strings.HasPrefix(uri, "file://") {
		return strings.TrimPrefix(uri, "file://"), nil
	}

	if filepath.IsAbs(uri) {
		return uri, nil
	}

	return "", errors.New("invalid local URI: must be file:// or absolute path")
}

// parseManifestFile reads and parses a collection.yaml file.
func (s *LocalSource) parseManifestFile(path string) (*CollectionMetadata, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var manifest map[string]interface{}
	if err := yaml.Unmarshal(data, &manifest); err != nil {
		return nil, err
	}

	return s.extractMetadata(manifest, "", filepath.Dir(path))
}

// extractMetadata extracts CollectionMetadata from a parsed manifest.
func (s *LocalSource) extractMetadata(manifest map[string]interface{}, uri, location string) (*CollectionMetadata, error) {
	name, _ := manifest["name"].(string)
	version, _ := manifest["version"].(string)
	author, _ := manifest["author"].(string)
	description, _ := manifest["description"].(string)

	if name == "" || version == "" || author == "" {
		return nil, errors.New("manifest missing required fields (name, version, author)")
	}

	tags, _ := manifest["tags"].([]interface{})
	var tagStrings []string
	for _, t := range tags {
		if s, ok := t.(string); ok {
			tagStrings = append(tagStrings, s)
		}
	}
	category, _ := manifest["category"].(string)

	return &CollectionMetadata{
		SourceName:  "local",
		URI:         uri,
		Name:        name,
		Version:     version,
		Author:      author,
		Description: description,
		Tags:        tagStrings,
		Category:    category,
		Repository:  location,
	}, nil
}

// matchesFilter checks if metadata matches the browse filter.
func (s *LocalSource) matchesFilter(metadata *CollectionMetadata, filter *BrowseFilter) bool {
	if filter == nil {
		return true
	}

	// Check category
	if filter.Category != "" && metadata.Category != filter.Category {
		return false
	}

	// Check author
	if filter.Author != "" && metadata.Author != filter.Author {
		return false
	}

	// Check tags (must have ALL specified tags)
	if len(filter.Tags) > 0 {
		metadataTags := make(map[string]bool)
		for _, tag := range metadata.Tags {
			metadataTags[tag] = true
		}
		for _, requiredTag := range filter.Tags {
			if !metadataTags[requiredTag] {
				return false
			}
		}
	}

	// Check query (text search in name and description)
	if filter.Query != "" {
		query := strings.ToLower(filter.Query)
		name := strings.ToLower(metadata.Name)
		desc := strings.ToLower(metadata.Description)
		if !strings.Contains(name, query) && !strings.Contains(desc, query) {
			return false
		}
	}

	return true
}

// extractTarGz extracts a tar.gz archive to a destination directory.
func (s *LocalSource) extractTarGz(archivePath, destDir string) error {
	file, err := os.Open(archivePath)
	if err != nil {
		return err
	}
	defer func() {
		_ = file.Close() // Ignore close error on read operation
	}()

	gzReader, err := gzip.NewReader(file)
	if err != nil {
		return fmt.Errorf("failed to create gzip reader: %w", err)
	}
	defer func() {
		_ = gzReader.Close() // Ignore close error on read operation
	}()

	tarReader := tar.NewReader(gzReader)

	for {
		header, err := tarReader.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return fmt.Errorf("failed to read tar entry: %w", err)
		}

		// Construct destination path
		//nolint:gosec // G305: Path traversal is prevented by security check below
		destPath := filepath.Join(destDir, header.Name)

		// Security check: prevent path traversal
		if !strings.HasPrefix(destPath, filepath.Clean(destDir)+string(os.PathSeparator)) {
			return fmt.Errorf("illegal file path in archive: %s", header.Name)
		}

		switch header.Typeflag {
		case tar.TypeDir:
			if err := os.MkdirAll(destPath, 0755); err != nil {
				return fmt.Errorf("failed to create directory: %w", err)
			}

		case tar.TypeReg:
			// Create parent directory if needed
			if err := os.MkdirAll(filepath.Dir(destPath), 0755); err != nil {
				return fmt.Errorf("failed to create parent directory: %w", err)
			}

			// Create file
			outFile, err := os.Create(destPath)
			if err != nil {
				return fmt.Errorf("failed to create file: %w", err)
			}

			//nolint:gosec // G110: Decompression bomb risk mitigated by size checks in caller
			if _, err := io.Copy(outFile, tarReader); err != nil {
				_ = outFile.Close()
				return fmt.Errorf("failed to write file: %w", err)
			}
			if err := outFile.Close(); err != nil {
				return fmt.Errorf("failed to close file: %w", err)
			}

		default:
			// Skip other types (symlinks, devices, etc.)
		}
	}

	return nil
}

// ExportToTarGz exports a collection directory to a tar.gz archive.
func ExportToTarGz(collectionDir, outputPath string) error {
	// Create output file
	outFile, err := os.Create(outputPath)
	if err != nil {
		return fmt.Errorf("failed to create output file: %w", err)
	}
	defer func() {
		if cerr := outFile.Close(); cerr != nil && err == nil {
			err = cerr
		}
	}()

	// Create gzip writer
	gzWriter := gzip.NewWriter(outFile)
	defer func() {
		if cerr := gzWriter.Close(); cerr != nil && err == nil {
			err = cerr
		}
	}()

	// Create tar writer
	tarWriter := tar.NewWriter(gzWriter)
	defer func() {
		if cerr := tarWriter.Close(); cerr != nil && err == nil {
			err = cerr
		}
	}()

	// Walk through collection directory
	return filepath.Walk(collectionDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Get relative path
		relPath, err := filepath.Rel(collectionDir, path)
		if err != nil {
			return err
		}

		// Skip root directory
		if relPath == "." {
			return nil
		}

		// Create tar header
		header, err := tar.FileInfoHeader(info, "")
		if err != nil {
			return err
		}
		header.Name = relPath

		// Write header
		if err := tarWriter.WriteHeader(header); err != nil {
			return err
		}

		// Write file content if it's a regular file
		if !info.IsDir() {
			file, err := os.Open(path)
			if err != nil {
				return err
			}
			defer func() {
				_ = file.Close() // Ignore close error on read operation
			}()

			if _, err := io.Copy(tarWriter, file); err != nil {
				return err
			}
		}

		return nil
	})
}
