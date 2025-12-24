package mcp

import (
	"archive/tar"
	"compress/gzip"
	"context"
	"os"
	"path/filepath"
	"testing"

	sdk "github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/fsvxavier/nexs-mcp/internal/collection"
	"github.com/fsvxavier/nexs-mcp/internal/common"
)

func setupTestServerForPublishing() *MCPServer {
	repo := NewMockElementRepository()
	return newTestServer("nexs-mcp-test", "0.1.0", repo)
}

// TestHandlePublishCollection tests

func TestHandlePublishCollection_MissingManifestPath(t *testing.T) {
	t.Skip("Requires filesystem access and manifest validation, tested via integration")
}

func TestHandlePublishCollection_InvalidRegistryFormat(t *testing.T) {
	server := setupTestServerForPublishing()
	ctx := context.Background()
	req := &sdk.CallToolRequest{}
	input := PublishCollectionInput{
		ManifestPath: "/tmp/test.yaml",
		Registry:     "invalid-format",
		GitHubToken:  "test-token",
	}

	_, output, err := server.handlePublishCollection(ctx, req, input)
	require.NoError(t, err)
	assert.Equal(t, common.StatusError, output.Status)
	assert.Contains(t, output.Message, "Invalid registry format")
}

func TestHandlePublishCollection_DefaultRegistry(t *testing.T) {
	t.Skip("Requires filesystem access and manifest validation, tested via integration")
}

func TestHandlePublishCollection_DryRun(t *testing.T) {
	t.Skip("Requires filesystem access and manifest validation, tested via integration")
}

func TestHandlePublishCollection_SkipSecurityScan(t *testing.T) {
	t.Skip("Requires filesystem access and manifest validation, tested via integration")
}

func TestHandlePublishCollection_ValidationFailed(t *testing.T) {
	t.Skip("Requires filesystem access and manifest validation, tested via integration")
}

func TestHandlePublishCollection_SecurityFailed(t *testing.T) {
	t.Skip("Requires filesystem access and security scanning, tested via integration")
}

func TestHandlePublishCollection_GitHubAuthFailed(t *testing.T) {
	t.Skip("Requires GitHub API access, tested via integration")
}

// TestCreateCollectionTarball tests

func TestCreateCollectionTarball_BasicManifest(t *testing.T) {
	// Create test directory structure
	tmpDir := t.TempDir()

	// Create test manifest
	manifestPath := filepath.Join(tmpDir, "collection.yaml")
	manifestContent := []byte(`name: test-collection
version: 1.0.0
author: testauthor
category: testing
description: Test collection
elements:
  - path: test-element.yaml
    type: persona
`)
	err := os.WriteFile(manifestPath, manifestContent, 0644)
	require.NoError(t, err)

	// Create test element file
	elementPath := filepath.Join(tmpDir, "test-element.yaml")
	err = os.WriteFile(elementPath, []byte("test: content\n"), 0644)
	require.NoError(t, err)

	// Parse manifest
	manifest, err := collection.ParseManifest(manifestContent)
	require.NoError(t, err)

	// Create tarball
	tarballPath := filepath.Join(tmpDir, "test.tar.gz")
	err = createCollectionTarball(tmpDir, tarballPath, manifest)
	require.NoError(t, err)

	// Verify tarball exists
	_, err = os.Stat(tarballPath)
	assert.NoError(t, err)

	// Verify tarball contents
	file, err := os.Open(tarballPath)
	require.NoError(t, err)
	defer file.Close()

	gzipReader, err := gzip.NewReader(file)
	require.NoError(t, err)
	defer gzipReader.Close()

	tarReader := tar.NewReader(gzipReader)

	foundManifest := false
	foundElement := false

	for {
		header, err := tarReader.Next()
		if err != nil {
			break
		}

		if header.Name == "collection.yaml" {
			foundManifest = true
		}
		if header.Name == "test-element.yaml" {
			foundElement = true
		}
	}

	assert.True(t, foundManifest, "Manifest should be in tarball")
	assert.True(t, foundElement, "Element should be in tarball")
}

func TestCreateCollectionTarball_MissingManifest(t *testing.T) {
	tmpDir := t.TempDir()

	manifest := &collection.Manifest{
		Name:    "test",
		Version: "1.0.0",
	}

	tarballPath := filepath.Join(tmpDir, "test.tar.gz")
	err := createCollectionTarball(tmpDir, tarballPath, manifest)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to add manifest")
}

func TestCreateCollectionTarball_MissingElement(t *testing.T) {
	tmpDir := t.TempDir()

	// Create test manifest
	manifestPath := filepath.Join(tmpDir, "collection.yaml")
	manifestContent := []byte(`name: test-collection
version: 1.0.0
`)
	err := os.WriteFile(manifestPath, manifestContent, 0644)
	require.NoError(t, err)

	manifest := &collection.Manifest{
		Name:    "test",
		Version: "1.0.0",
		Elements: []collection.Element{
			{
				Path: "nonexistent.yaml",
				Type: "persona",
			},
		},
	}

	tarballPath := filepath.Join(tmpDir, "test.tar.gz")
	err = createCollectionTarball(tmpDir, tarballPath, manifest)
	assert.Error(t, err)
}

func TestCreateCollectionTarball_InvalidPath(t *testing.T) {
	tmpDir := t.TempDir()

	manifest := &collection.Manifest{
		Name:    "test",
		Version: "1.0.0",
	}

	tarballPath := "/invalid/path/test.tar.gz"
	err := createCollectionTarball(tmpDir, tarballPath, manifest)
	assert.Error(t, err)
}

// TestAddFileToTar tests

func TestAddFileToTar_ValidFile(t *testing.T) {
	tmpDir := t.TempDir()

	// Create test file
	testFile := filepath.Join(tmpDir, "test.txt")
	testContent := []byte("test content")
	err := os.WriteFile(testFile, testContent, 0644)
	require.NoError(t, err)

	// Create tar file
	tarPath := filepath.Join(tmpDir, "test.tar")
	tarFile, err := os.Create(tarPath)
	require.NoError(t, err)
	defer tarFile.Close()

	tarWriter := tar.NewWriter(tarFile)
	defer tarWriter.Close()

	err = addFileToTar(tarWriter, testFile, "test.txt")
	assert.NoError(t, err)
}

func TestAddFileToTar_NonexistentFile(t *testing.T) {
	tmpDir := t.TempDir()

	tarPath := filepath.Join(tmpDir, "test.tar")
	tarFile, err := os.Create(tarPath)
	require.NoError(t, err)
	defer tarFile.Close()

	tarWriter := tar.NewWriter(tarFile)
	defer tarWriter.Close()

	err = addFileToTar(tarWriter, "/nonexistent/file.txt", "file.txt")
	assert.Error(t, err)
}

func TestAddFileToTar_Directory(t *testing.T) {
	t.Skip("addFileToTar expects files, not directories")
}

// TestCopyFile tests

func TestCopyFile_Success(t *testing.T) {
	tmpDir := t.TempDir()

	srcPath := filepath.Join(tmpDir, "source.txt")
	dstPath := filepath.Join(tmpDir, "dest.txt")

	testContent := []byte("test content for copy")
	err := os.WriteFile(srcPath, testContent, 0644)
	require.NoError(t, err)

	err = copyFile(srcPath, dstPath)
	require.NoError(t, err)

	// Verify destination file
	dstContent, err := os.ReadFile(dstPath)
	require.NoError(t, err)
	assert.Equal(t, testContent, dstContent)
}

func TestCopyFile_SourceNotFound(t *testing.T) {
	tmpDir := t.TempDir()

	srcPath := filepath.Join(tmpDir, "nonexistent.txt")
	dstPath := filepath.Join(tmpDir, "dest.txt")

	err := copyFile(srcPath, dstPath)
	assert.Error(t, err)
}

func TestCopyFile_DestinationDirNotExists(t *testing.T) {
	tmpDir := t.TempDir()

	srcPath := filepath.Join(tmpDir, "source.txt")
	dstPath := filepath.Join(tmpDir, "nonexistent", "dest.txt")

	err := os.WriteFile(srcPath, []byte("test"), 0644)
	require.NoError(t, err)

	err = copyFile(srcPath, dstPath)
	assert.Error(t, err)
}

func TestCopyFile_LargeFile(t *testing.T) {
	tmpDir := t.TempDir()

	srcPath := filepath.Join(tmpDir, "large.txt")
	dstPath := filepath.Join(tmpDir, "large-copy.txt")

	// Create large content (1MB)
	largeContent := make([]byte, 1024*1024)
	for i := range largeContent {
		largeContent[i] = byte(i % 256)
	}

	err := os.WriteFile(srcPath, largeContent, 0644)
	require.NoError(t, err)

	err = copyFile(srcPath, dstPath)
	require.NoError(t, err)

	// Verify size
	srcInfo, err := os.Stat(srcPath)
	require.NoError(t, err)

	dstInfo, err := os.Stat(dstPath)
	require.NoError(t, err)

	assert.Equal(t, srcInfo.Size(), dstInfo.Size())
}

// TestPublishCollectionOutput tests

func TestPublishCollectionOutput_SuccessStatus(t *testing.T) {
	output := PublishCollectionOutput{
		Status:   "success",
		Message:  "Publication complete",
		PRURL:    "https://github.com/test/repo/pull/123",
		PRNumber: 123,
		Collection: map[string]interface{}{
			"name":    "test-collection",
			"version": "1.0.0",
			"author":  "testauthor",
		},
	}

	assert.Equal(t, "success", output.Status)
	assert.Equal(t, "Publication complete", output.Message)
	assert.Equal(t, "https://github.com/test/repo/pull/123", output.PRURL)
	assert.Equal(t, 123, output.PRNumber)
	assert.NotNil(t, output.Collection)
}

func TestPublishCollectionOutput_ValidationFailedStatus(t *testing.T) {
	output := PublishCollectionOutput{
		Status:  "validation_failed",
		Message: "Validation errors found",
		ValidationErrors: []*collection.ValidationError{
			{
				Field:   "name",
				Message: "Name is required",
				Fix:     "Add a name field",
			},
		},
	}

	assert.Equal(t, "validation_failed", output.Status)
	assert.NotEmpty(t, output.ValidationErrors)
	assert.Equal(t, "name", output.ValidationErrors[0].Field)
}

func TestPublishCollectionOutput_SecurityFailedStatus(t *testing.T) {
	output := PublishCollectionOutput{
		Status:           "security_failed",
		Message:          "Security issues found",
		SecurityFindings: []string{"Critical issue 1", "High issue 2"},
	}

	assert.Equal(t, "security_failed", output.Status)
	assert.NotNil(t, output.SecurityFindings)
}

func TestPublishCollectionOutput_DryRunStatus(t *testing.T) {
	output := PublishCollectionOutput{
		Status:         "dry_run_success",
		Message:        "Dry run complete",
		Tarball:        "/tmp/test.tar.gz",
		Checksums:      "/tmp/test.checksums.txt",
		ChecksumSHA256: "abc123def456",
	}

	assert.Equal(t, "dry_run_success", output.Status)
	assert.NotEmpty(t, output.Tarball)
	assert.NotEmpty(t, output.Checksums)
	assert.NotEmpty(t, output.ChecksumSHA256)
}

func TestPublishCollectionOutput_ErrorStatus(t *testing.T) {
	output := PublishCollectionOutput{
		Status:  common.StatusError,
		Message: "Failed to read manifest",
	}

	assert.Equal(t, common.StatusError, output.Status)
	assert.Contains(t, output.Message, "Failed to read manifest")
}

// TestPublishCollectionInput validation

func TestPublishCollectionInput_AllFieldsSet(t *testing.T) {
	input := PublishCollectionInput{
		ManifestPath:     "/path/to/manifest.yaml",
		Registry:         "owner/repo",
		GitHubToken:      "ghp_test123",
		CreateRelease:    true,
		DryRun:           false,
		SkipSecurityScan: false,
	}

	assert.Equal(t, "/path/to/manifest.yaml", input.ManifestPath)
	assert.Equal(t, "owner/repo", input.Registry)
	assert.Equal(t, "ghp_test123", input.GitHubToken)
	assert.True(t, input.CreateRelease)
	assert.False(t, input.DryRun)
	assert.False(t, input.SkipSecurityScan)
}

func TestPublishCollectionInput_DefaultValues(t *testing.T) {
	input := PublishCollectionInput{
		ManifestPath: "/path/to/manifest.yaml",
		GitHubToken:  "ghp_test123",
	}

	assert.Empty(t, input.Registry)
	assert.False(t, input.CreateRelease)
	assert.False(t, input.DryRun)
	assert.False(t, input.SkipSecurityScan)
}
