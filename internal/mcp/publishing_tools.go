package mcp

import (
	"archive/tar"
	"compress/gzip"
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	sdk "github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/fsvxavier/nexs-mcp/internal/collection"
	"github.com/fsvxavier/nexs-mcp/internal/collection/security"
	"github.com/fsvxavier/nexs-mcp/internal/infrastructure"
)

// PublishCollectionInput defines input for publish_collection tool
type PublishCollectionInput struct {
	ManifestPath     string `json:"manifest_path" jsonschema:"path to collection.yaml manifest file"`
	Registry         string `json:"registry,omitempty" jsonschema:"target registry repository (format: owner/repo)"`
	GitHubToken      string `json:"github_token" jsonschema:"GitHub personal access token (required for PR creation)"`
	CreateRelease    bool   `json:"create_release,omitempty" jsonschema:"also create a GitHub release in your fork"`
	DryRun           bool   `json:"dry_run,omitempty" jsonschema:"validate and prepare files without creating PR"`
	SkipSecurityScan bool   `json:"skip_security_scan,omitempty" jsonschema:"skip security code scanning (not recommended)"`
}

// PublishCollectionOutput defines output for publish_collection tool
type PublishCollectionOutput struct {
	Status           string                        `json:"status" jsonschema:"success, validation_failed, security_failed, or error"`
	Message          string                        `json:"message" jsonschema:"human-readable progress message"`
	PRURL            string                        `json:"pr_url,omitempty" jsonschema:"pull request URL if successful"`
	PRNumber         int                           `json:"pr_number,omitempty" jsonschema:"pull request number if successful"`
	Tarball          string                        `json:"tarball,omitempty" jsonschema:"tarball path (dry run only)"`
	Checksums        string                        `json:"checksums,omitempty" jsonschema:"checksums file path (dry run only)"`
	ChecksumSHA256   string                        `json:"checksum_sha256,omitempty" jsonschema:"SHA-256 checksum"`
	ValidationErrors []*collection.ValidationError `json:"validation_errors,omitempty" jsonschema:"validation errors if validation failed"`
	SecurityFindings interface{}                   `json:"security_findings,omitempty" jsonschema:"security findings if scan failed"`
	Collection       map[string]interface{}        `json:"collection,omitempty" jsonschema:"collection metadata"`
}

func (s *MCPServer) handlePublishCollection(ctx context.Context, req *sdk.CallToolRequest, input PublishCollectionInput) (*sdk.CallToolResult, PublishCollectionOutput, error) {
	manifestPath := input.ManifestPath
	registry := input.Registry
	githubToken := input.GitHubToken
	createRelease := input.CreateRelease
	dryRun := input.DryRun
	skipSecurityScan := input.SkipSecurityScan

	if registry == "" {
		registry = "fsvxavier/nexs-mcp-collections"
	}

	output := PublishCollectionOutput{}
	var progress strings.Builder
	progress.WriteString("🚀 **Publishing Collection**\n\n")

	// Parse registry owner/repo
	parts := strings.Split(registry, "/")
	if len(parts) != 2 {
		output.Status = "error"
		output.Message = fmt.Sprintf("❌ Invalid registry format: %s (expected: owner/repo)", registry)
		return nil, output, nil
	}
	registryOwner := parts[0]
	registryRepo := parts[1]

	// Step 1: Load and validate manifest
	progress.WriteString("📋 **Step 1/7:** Loading manifest...\n")

	manifestData, err := os.ReadFile(manifestPath)
	if err != nil {
		output.Status = "error"
		output.Message = fmt.Sprintf("❌ Failed to read manifest: %v", err)
		return nil, output, nil
	}

	manifest, err := collection.ParseManifest(manifestData)
	if err != nil {
		output.Status = "error"
		output.Message = fmt.Sprintf("❌ Failed to parse manifest: %v", err)
		return nil, output, nil
	}

	progress.WriteString(fmt.Sprintf("   ✅ Loaded: %s v%s by %s\n\n", manifest.Name, manifest.Version, manifest.Author))

	// Step 2: Comprehensive validation
	progress.WriteString("🔍 **Step 2/7:** Validating manifest (100+ rules)...\n")

	basePath := filepath.Dir(manifestPath)
	validator := collection.NewValidator(basePath)
	validationResult := validator.ValidateComprehensive(manifest)

	if !validationResult.Valid {
		progress.WriteString(fmt.Sprintf("   ❌ Validation failed: %d errors, %d warnings\n\n",
			validationResult.Stats["errors"], validationResult.Stats["warnings"]))

		progress.WriteString("**Errors:**\n")
		for i, verr := range validationResult.Errors {
			if i >= 10 {
				progress.WriteString(fmt.Sprintf("   ... and %d more errors\n", len(validationResult.Errors)-10))
				break
			}
			progress.WriteString(fmt.Sprintf("   - %s: %s\n", verr.Field, verr.Message))
			if verr.Fix != "" {
				progress.WriteString(fmt.Sprintf("     💡 Fix: %s\n", verr.Fix))
			}
		}

		output.Status = "validation_failed"
		output.Message = progress.String()
		output.ValidationErrors = validationResult.Errors
		return nil, output, nil
	}

	progress.WriteString(fmt.Sprintf("   ✅ Passed: %d rules checked, %d warnings\n\n",
		validationResult.Stats["total_rules_checked"], validationResult.Stats["warnings"]))

	// Step 3: Security scan
	var scanResult *security.ScanResult
	if !skipSecurityScan {
		progress.WriteString("🔒 **Step 3/7:** Security scanning...\n")

		scanner := security.NewCodeScanner()
		scanner.SetThreshold(security.SeverityCritical)

		scanResult, err = scanner.Scan(basePath)
		if err != nil {
			progress.WriteString(fmt.Sprintf("   ⚠️  Scan error: %v (continuing...)\n\n", err))
		} else {
			if !scanResult.Clean {
				progress.WriteString(fmt.Sprintf("   ❌ Security issues found: %d critical, %d high\n\n",
					scanResult.Stats["critical"], scanResult.Stats["high"]))

				progress.WriteString("**Critical/High Issues:**\n")
				for i, finding := range scanResult.Findings {
					if finding.Severity != security.SeverityCritical && finding.Severity != security.SeverityHigh {
						continue
					}
					if i >= 10 {
						break
					}
					progress.WriteString(fmt.Sprintf("   - [%s] %s:%d - %s\n",
						finding.Severity, finding.File, finding.Line, finding.Rule.Description))
				}

				output.Status = "security_failed"
				output.Message = progress.String()
				output.SecurityFindings = scanResult.Findings
				return nil, output, nil
			}

			progress.WriteString(fmt.Sprintf("   ✅ Clean: %d files scanned, %d low/medium findings\n\n",
				scanResult.FilesScanned, scanResult.Stats["low"]+scanResult.Stats["medium"]))
		}
	} else {
		progress.WriteString("⚠️  **Step 3/7:** Security scan skipped\n\n")
	}

	// Step 4: Generate checksums and create tarball
	progress.WriteString("📦 **Step 4/7:** Creating tarball with checksums...\n")

	tarballPath := filepath.Join(os.TempDir(), fmt.Sprintf("%s-%s.tar.gz", manifest.Name, manifest.Version))
	checksumsPath := filepath.Join(os.TempDir(), fmt.Sprintf("%s-%s.checksums.txt", manifest.Name, manifest.Version))

	if err := createCollectionTarball(basePath, tarballPath, manifest); err != nil {
		output.Status = "error"
		output.Message = fmt.Sprintf("❌ Failed to create tarball: %v", err)
		return nil, output, nil
	}

	// Generate checksums
	checksumValidator := security.NewChecksumValidator(security.SHA256)
	checksum, err := checksumValidator.Compute(tarballPath)
	if err != nil {
		output.Status = "error"
		output.Message = fmt.Sprintf("❌ Failed to compute checksum: %v", err)
		return nil, output, nil
	}

	// Write checksums file
	checksumContent := fmt.Sprintf("%s  %s\n", checksum, filepath.Base(tarballPath))
	if err := os.WriteFile(checksumsPath, []byte(checksumContent), 0644); err != nil {
		output.Status = "error"
		output.Message = fmt.Sprintf("❌ Failed to write checksums: %v", err)
		return nil, output, nil
	}

	progress.WriteString(fmt.Sprintf("   ✅ Created: %s (checksum: %s)\n\n", filepath.Base(tarballPath), checksum[:16]+"..."))

	// If dry run, stop here
	if dryRun {
		progress.WriteString("🎯 **Dry Run Complete**\n\n")
		progress.WriteString("Files ready for publishing:\n")
		progress.WriteString(fmt.Sprintf("- Tarball: %s\n", tarballPath))
		progress.WriteString(fmt.Sprintf("- Checksums: %s\n", checksumsPath))
		progress.WriteString("\nRun without --dry-run to create PR\n")

		output.Status = "dry_run_success"
		output.Message = progress.String()
		output.Tarball = tarballPath
		output.Checksums = checksumsPath
		output.ChecksumSHA256 = checksum
		return nil, output, nil
	}

	// Step 5: Fork repository
	progress.WriteString("🍴 **Step 5/7:** Forking registry repository...\n")

	publisher := infrastructure.NewGitHubPublisher(githubToken)
	user, err := publisher.GetAuthenticatedUser()
	if err != nil {
		output.Status = "error"
		output.Message = fmt.Sprintf("❌ Failed to authenticate with GitHub: %v", err)
		return nil, output, nil
	}

	fork, err := publisher.ForkRepository(&infrastructure.ForkRepositoryOptions{
		Owner: registryOwner,
		Repo:  registryRepo,
	})
	if err != nil {
		progress.WriteString(fmt.Sprintf("   ℹ️  Fork may already exist: %v\n\n", err))
	} else {
		progress.WriteString(fmt.Sprintf("   ✅ Forked to: %s\n\n", fork.GetFullName()))
	}

	// Step 6: Clone, commit, and push
	progress.WriteString("💾 **Step 6/7:** Cloning fork and committing changes...\n")

	cloneDir := filepath.Join(os.TempDir(), fmt.Sprintf("%s-clone-%d", registryRepo, os.Getpid()))
	defer os.RemoveAll(cloneDir)

	forkURL := publisher.GetForkHTTPSURL(registryOwner, registryRepo, user.GetLogin())
	if err := publisher.CloneRepository(&infrastructure.CloneOptions{
		URL:       forkURL,
		Directory: cloneDir,
		Depth:     1,
	}); err != nil {
		output.Status = "error"
		output.Message = fmt.Sprintf("❌ Failed to clone fork: %v", err)
		return nil, output, nil
	}

	// Copy files to clone
	collectionDir := filepath.Join(cloneDir, "collections", manifest.Author, manifest.Name)
	if err := os.MkdirAll(collectionDir, 0755); err != nil {
		output.Status = "error"
		output.Message = fmt.Sprintf("❌ Failed to create collection directory: %v", err)
		return nil, output, nil
	}

	// Copy manifest, tarball, and checksums
	manifestDest := filepath.Join(collectionDir, "collection.yaml")
	tarballDest := filepath.Join(collectionDir, filepath.Base(tarballPath))
	checksumsDest := filepath.Join(collectionDir, filepath.Base(checksumsPath))

	if err := copyFile(manifestPath, manifestDest); err != nil {
		output.Status = "error"
		output.Message = fmt.Sprintf("❌ Failed to copy manifest: %v", err)
		return nil, output, nil
	}

	if err := copyFile(tarballPath, tarballDest); err != nil {
		output.Status = "error"
		output.Message = fmt.Sprintf("❌ Failed to copy tarball: %v", err)
		return nil, output, nil
	}

	if err := copyFile(checksumsPath, checksumsDest); err != nil {
		output.Status = "error"
		output.Message = fmt.Sprintf("❌ Failed to copy checksums: %v", err)
		return nil, output, nil
	}

	// Commit changes
	branchName := fmt.Sprintf("add-%s-%s", manifest.Name, manifest.Version)
	commitMessage := fmt.Sprintf("Add collection: %s v%s\n\nAuthor: %s\nCategory: %s\nDescription: %s",
		manifest.Name, manifest.Version, manifest.Author, manifest.Category, manifest.Description)

	if err := publisher.CommitChanges(&infrastructure.CommitOptions{
		RepoPath:     cloneDir,
		Files:        []string{filepath.Join("collections", manifest.Author, manifest.Name)},
		Message:      commitMessage,
		AuthorName:   user.GetName(),
		AuthorEmail:  user.GetEmail(),
		Branch:       branchName,
		CreateBranch: true,
	}); err != nil {
		output.Status = "error"
		output.Message = fmt.Sprintf("❌ Failed to commit changes: %v", err)
		return nil, output, nil
	}

	// Push changes
	if err := publisher.PushChanges(&infrastructure.PushOptions{
		RepoPath: cloneDir,
		Remote:   "origin",
		Branch:   branchName,
	}); err != nil {
		output.Status = "error"
		output.Message = fmt.Sprintf("❌ Failed to push changes: %v", err)
		return nil, output, nil
	}

	progress.WriteString(fmt.Sprintf("   ✅ Pushed branch: %s\n\n", branchName))

	// Step 7: Create pull request
	progress.WriteString("🔀 **Step 7/7:** Creating pull request...\n")

	metadata := map[string]interface{}{
		"name":          manifest.Name,
		"version":       manifest.Version,
		"author":        manifest.Author,
		"category":      manifest.Category,
		"description":   manifest.Description,
		"repository":    manifest.Repository,
		"documentation": manifest.Documentation,
		"homepage":      manifest.Homepage,
		"stats": map[string]interface{}{
			"total_elements": len(manifest.Elements),
			"personas":       manifest.Stats.Personas,
			"skills":         manifest.Stats.Skills,
			"templates":      manifest.Stats.Templates,
		},
	}

	prBody := infrastructure.BuildPRTemplate(metadata)

	pr, err := publisher.CreatePullRequest(&infrastructure.PullRequestOptions{
		Owner:       registryOwner,
		Repo:        registryRepo,
		Title:       fmt.Sprintf("Add collection: %s v%s", manifest.Name, manifest.Version),
		Body:        prBody,
		Head:        fmt.Sprintf("%s:%s", user.GetLogin(), branchName),
		Base:        "main",
		Draft:       false,
		Maintainers: true,
	})
	if err != nil {
		output.Status = "error"
		output.Message = fmt.Sprintf("❌ Failed to create pull request: %v", err)
		return nil, output, nil
	}

	progress.WriteString(fmt.Sprintf("   ✅ Pull Request Created!\n"))
	progress.WriteString(fmt.Sprintf("   🔗 URL: %s\n\n", pr.GetHTMLURL()))

	// Optional: Create release
	if createRelease {
		progress.WriteString("🎉 **Bonus:** Creating GitHub release...\n")

		releaseBody := fmt.Sprintf("# %s v%s\n\n%s\n\n## Files\n- Tarball: %s\n- Checksums: %s",
			manifest.Name, manifest.Version, manifest.Description,
			filepath.Base(tarballPath), filepath.Base(checksumsPath))

		release, err := publisher.CreateRelease(&infrastructure.ReleaseOptions{
			Owner:      user.GetLogin(),
			Repo:       registryRepo,
			Tag:        fmt.Sprintf("v%s", manifest.Version),
			Name:       fmt.Sprintf("%s v%s", manifest.Name, manifest.Version),
			Body:       releaseBody,
			Draft:      false,
			Prerelease: false,
			Assets:     []string{tarballPath, checksumsPath},
		})
		if err != nil {
			progress.WriteString(fmt.Sprintf("   ⚠️  Release creation failed: %v\n\n", err))
		} else {
			progress.WriteString(fmt.Sprintf("   ✅ Release: %s\n\n", release.GetHTMLURL()))
		}
	}

	// Success!
	progress.WriteString("✨ **Publication Complete!**\n\n")
	progress.WriteString("**Next Steps:**\n")
	progress.WriteString("1. Review the pull request\n")
	progress.WriteString("2. Address any feedback from maintainers\n")
	progress.WriteString("3. Wait for approval and merge\n\n")
	progress.WriteString(fmt.Sprintf("PR: %s\n", pr.GetHTMLURL()))

	output.Status = "success"
	output.Message = progress.String()
	output.PRURL = pr.GetHTMLURL()
	output.PRNumber = pr.GetNumber()
	output.Collection = map[string]interface{}{
		"name":    manifest.Name,
		"version": manifest.Version,
		"author":  manifest.Author,
	}

	return nil, output, nil
}

// createCollectionTarball creates a tarball of the collection
func createCollectionTarball(basePath, tarballPath string, manifest *collection.Manifest) error {
	file, err := os.Create(tarballPath)
	if err != nil {
		return err
	}
	defer file.Close()

	gzipWriter := gzip.NewWriter(file)
	defer gzipWriter.Close()

	tarWriter := tar.NewWriter(gzipWriter)
	defer tarWriter.Close()

	// Add manifest
	manifestPath := filepath.Join(basePath, "collection.yaml")
	if err := addFileToTar(tarWriter, manifestPath, "collection.yaml"); err != nil {
		return fmt.Errorf("failed to add manifest: %w", err)
	}

	// Add elements
	for _, elem := range manifest.Elements {
		elemPath := filepath.Join(basePath, elem.Path)

		if strings.Contains(elem.Path, "*") {
			matches, err := filepath.Glob(elemPath)
			if err != nil {
				return fmt.Errorf("glob error for %s: %w", elem.Path, err)
			}
			for _, match := range matches {
				relPath, _ := filepath.Rel(basePath, match)
				if err := addFileToTar(tarWriter, match, relPath); err != nil {
					return err
				}
			}
		} else {
			if err := addFileToTar(tarWriter, elemPath, elem.Path); err != nil {
				return err
			}
		}
	}

	return nil
}

// addFileToTar adds a file to a tar archive
func addFileToTar(tarWriter *tar.Writer, filePath, tarPath string) error {
	file, err := os.Open(filePath)
	if err != nil {
		return err
	}
	defer file.Close()

	info, err := file.Stat()
	if err != nil {
		return err
	}

	header, err := tar.FileInfoHeader(info, "")
	if err != nil {
		return err
	}
	header.Name = tarPath

	if err := tarWriter.WriteHeader(header); err != nil {
		return err
	}

	_, err = io.Copy(tarWriter, file)
	return err
}

// copyFile copies a file from src to dst
func copyFile(src, dst string) error {
	srcFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer srcFile.Close()

	dstFile, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer dstFile.Close()

	_, err = io.Copy(dstFile, srcFile)
	return err
}
