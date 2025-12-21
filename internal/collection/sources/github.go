package sources

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/google/go-github/v57/github"
	"golang.org/x/oauth2"
	"gopkg.in/yaml.v3"
)

// GitHubSource implements CollectionSource for GitHub repositories.
type GitHubSource struct {
	client      *github.Client
	oauthClient OAuthProvider
	cacheDir    string
}

// OAuthProvider defines the interface for GitHub OAuth operations.
type OAuthProvider interface {
	GetToken(ctx context.Context) (*oauth2.Token, error)
}

// NewGitHubSource creates a new GitHub collection source.
func NewGitHubSource(oauthClient OAuthProvider, cacheDir string) (*GitHubSource, error) {
	if cacheDir == "" {
		homeDir, err := os.UserHomeDir()
		if err != nil {
			return nil, fmt.Errorf("failed to get home directory: %w", err)
		}
		cacheDir = filepath.Join(homeDir, ".nexs", "cache", "collections")
	}

	if err := os.MkdirAll(cacheDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create cache directory: %w", err)
	}

	return &GitHubSource{
		oauthClient: oauthClient,
		cacheDir:    cacheDir,
	}, nil
}

// Name returns the unique name of this source.
func (s *GitHubSource) Name() string {
	return "github"
}

// Supports returns true if this source can handle the given URI.
func (s *GitHubSource) Supports(uri string) bool {
	return strings.HasPrefix(uri, "github://")
}

// Browse discovers collections from GitHub using the Topics API.
func (s *GitHubSource) Browse(ctx context.Context, filter *BrowseFilter) ([]*CollectionMetadata, error) {
	client, err := s.getClient(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get GitHub client: %w", err)
	}

	// Build search query
	query := s.buildSearchQuery(filter)

	// Search repositories
	opts := &github.SearchOptions{
		Sort:  "stars",
		Order: "desc",
		ListOptions: github.ListOptions{
			PerPage: 100,
		},
	}

	if filter != nil {
		if filter.Limit > 0 && filter.Limit < 100 {
			opts.PerPage = filter.Limit
		}
		if filter.Offset > 0 {
			opts.Page = (filter.Offset / opts.PerPage) + 1
		}
	}

	result, _, err := client.Search.Repositories(ctx, query, opts)
	if err != nil {
		return nil, fmt.Errorf("GitHub search failed: %w", err)
	}

	// Convert results to CollectionMetadata
	var collections []*CollectionMetadata
	for _, repo := range result.Repositories {
		metadata, err := s.repoToMetadata(ctx, client, repo)
		if err != nil {
			// Log error but continue processing other results
			continue
		}
		collections = append(collections, metadata)
	}

	return collections, nil
}

// Get retrieves a specific collection by URI.
func (s *GitHubSource) Get(ctx context.Context, uri string) (*Collection, error) {
	// Parse URI: github://owner/repo[@version]
	owner, repo, version, err := s.parseURI(uri)
	if err != nil {
		return nil, fmt.Errorf("invalid URI: %w", err)
	}

	// Clone or update repository
	repoPath, err := s.cloneOrUpdate(ctx, owner, repo, version)
	if err != nil {
		return nil, fmt.Errorf("failed to clone repository: %w", err)
	}

	// Read and parse collection.yaml
	manifestPath := filepath.Join(repoPath, "collection.yaml")
	data, err := os.ReadFile(manifestPath)
	if err != nil {
		return nil, fmt.Errorf("collection.yaml not found: %w", err)
	}

	var manifestMap map[string]interface{}
	if err := yaml.Unmarshal(data, &manifestMap); err != nil {
		return nil, fmt.Errorf("failed to parse manifest: %w", err)
	}

	// Extract required fields
	name, _ := manifestMap["name"].(string)
	manifestVersion, _ := manifestMap["version"].(string)
	author, _ := manifestMap["author"].(string)
	description, _ := manifestMap["description"].(string)

	if name == "" || manifestVersion == "" || author == "" {
		return nil, errors.New("manifest missing required fields (name, version, author)")
	}

	// Extract optional fields
	tags, _ := manifestMap["tags"].([]interface{})
	var tagStrings []string
	for _, t := range tags {
		if s, ok := t.(string); ok {
			tagStrings = append(tagStrings, s)
		}
	}
	category, _ := manifestMap["category"].(string)

	// Get GitHub repo metadata
	client, err := s.getClient(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get GitHub client: %w", err)
	}

	ghRepo, _, err := client.Repositories.Get(ctx, owner, repo)
	if err != nil {
		return nil, fmt.Errorf("failed to get repository info: %w", err)
	}

	metadata := &CollectionMetadata{
		SourceName:  "github",
		URI:         uri,
		Name:        name,
		Version:     manifestVersion,
		Author:      author,
		Description: description,
		Tags:        tagStrings,
		Category:    category,
		Repository:  uri,
		Stars:       ghRepo.GetStargazersCount(),
		Downloads:   0,
	}

	return &Collection{
		Metadata:   metadata,
		Manifest:   manifestMap,
		SourceData: map[string]interface{}{"path": repoPath},
	}, nil
}

// buildSearchQuery builds a GitHub search query from the filter.
func (s *GitHubSource) buildSearchQuery(filter *BrowseFilter) string {
	// Base query: repositories with "nexs-mcp-collection" topic
	query := "topic:nexs-mcp-collection"

	if filter == nil {
		return query
	}

	// Add author filter
	if filter.Author != "" {
		query += " user:" + filter.Author
	}

	// Add text search
	if filter.Query != "" {
		query += fmt.Sprintf(" %s in:name,description,readme", filter.Query)
	}

	// Add category as topic
	if filter.Category != "" {
		query += " topic:" + filter.Category
	}

	// Add additional tags as topics
	var querySb210 strings.Builder
	for _, tag := range filter.Tags {
		querySb210.WriteString(" topic:" + tag)
	}
	query += querySb210.String()

	return query
}

// repoToMetadata converts a GitHub repository to CollectionMetadata.
func (s *GitHubSource) repoToMetadata(ctx context.Context, client *github.Client, repo *github.Repository) (*CollectionMetadata, error) {
	owner := repo.GetOwner().GetLogin()
	name := repo.GetName()

	// Try to get collection.yaml to extract metadata
	content, _, _, err := client.Repositories.GetContents(ctx, owner, name, "collection.yaml", nil)
	if err != nil {
		return nil, errors.New("no collection.yaml found")
	}

	// Decode content
	manifestContent, err := content.GetContent()
	if err != nil {
		return nil, fmt.Errorf("failed to read collection.yaml: %w", err)
	}

	// Parse manifest
	var manifest map[string]interface{}
	if err := yaml.Unmarshal([]byte(manifestContent), &manifest); err != nil {
		return nil, fmt.Errorf("failed to parse manifest: %w", err)
	}

	// Extract fields
	collectionName, _ := manifest["name"].(string)
	author, _ := manifest["author"].(string)
	description, _ := manifest["description"].(string)

	tags, _ := manifest["tags"].([]interface{})
	var tagStrings []string
	for _, t := range tags {
		if s, ok := t.(string); ok {
			tagStrings = append(tagStrings, s)
		}
	}
	category, _ := manifest["category"].(string)

	// Get latest version from tags
	version, err := s.getLatestVersion(ctx, client, owner, name)
	if err != nil {
		// Fallback to manifest version if available
		if v, ok := manifest["version"].(string); ok {
			version = v
		} else {
			version = "0.0.0"
		}
	}

	uriStr := fmt.Sprintf("github://%s/%s", owner, name)

	return &CollectionMetadata{
		SourceName:  "github",
		URI:         uriStr,
		Name:        collectionName,
		Version:     version,
		Author:      author,
		Description: description,
		Tags:        tagStrings,
		Category:    category,
		Repository:  uriStr,
		Stars:       repo.GetStargazersCount(),
		Downloads:   0,
	}, nil
}

// getLatestVersion gets the latest semantic version from git tags.
func (s *GitHubSource) getLatestVersion(ctx context.Context, client *github.Client, owner, repo string) (string, error) {
	// List all tags
	tags, _, err := client.Repositories.ListTags(ctx, owner, repo, &github.ListOptions{PerPage: 100})
	if err != nil {
		return "", err
	}

	// Find the latest semver tag
	var latestVersion string
	semverRegex := regexp.MustCompile(`^v?(\d+\.\d+\.\d+(?:-[a-zA-Z0-9.-]+)?)$`)

	for _, tag := range tags {
		tagName := tag.GetName()
		matches := semverRegex.FindStringSubmatch(tagName)
		if matches != nil {
			version := matches[1]
			if latestVersion == "" || s.compareVersions(version, latestVersion) > 0 {
				latestVersion = version
			}
		}
	}

	if latestVersion == "" {
		return "", errors.New("no valid semver tags found")
	}

	return latestVersion, nil
}

// compareVersions compares two semantic versions (simple implementation)
// Returns: -1 if v1 < v2, 0 if v1 == v2, 1 if v1 > v2.
func (s *GitHubSource) compareVersions(v1, v2 string) int {
	if v1 == v2 {
		return 0
	}
	if v1 > v2 {
		return 1
	}
	return -1
}

// parseURI parses a GitHub URI: github://owner/repo[@version]
func (s *GitHubSource) parseURI(uri string) (owner, repo, version string, err error) {
	if !strings.HasPrefix(uri, "github://") {
		return "", "", "", errors.New("invalid GitHub URI: must start with github://")
	}

	// Remove github:// prefix
	path := strings.TrimPrefix(uri, "github://")

	// Split version if present
	parts := strings.Split(path, "@")
	if len(parts) > 2 {
		return "", "", "", errors.New("invalid URI format: too many @ symbols")
	}

	// Parse owner/repo
	repoPath := parts[0]
	repoParts := strings.Split(repoPath, "/")
	if len(repoParts) != 2 {
		return "", "", "", errors.New("invalid repository format: expected owner/repo")
	}

	owner = repoParts[0]
	repo = repoParts[1]

	// Parse version if present
	if len(parts) == 2 {
		version = parts[1]
	}

	return owner, repo, version, nil
}

// cloneOrUpdate clones or updates a repository to the cache directory.
func (s *GitHubSource) cloneOrUpdate(ctx context.Context, owner, repo, version string) (string, error) {
	repoPath := filepath.Join(s.cacheDir, owner, repo)

	// Check if repository already exists
	if _, err := os.Stat(repoPath); os.IsNotExist(err) {
		// Clone repository
		if err := s.cloneRepo(ctx, owner, repo, repoPath); err != nil {
			return "", err
		}
	} else {
		// Update existing repository
		if err := s.updateRepo(ctx, repoPath); err != nil {
			return "", err
		}
	}

	// Checkout specific version if specified
	if version != "" {
		if err := s.checkoutVersion(ctx, repoPath, version); err != nil {
			return "", err
		}
	}

	return repoPath, nil
}

// cloneRepo clones a GitHub repository.
func (s *GitHubSource) cloneRepo(ctx context.Context, owner, repo, destPath string) error {
	// Create parent directory
	if err := os.MkdirAll(filepath.Dir(destPath), 0755); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	// Build clone URL (using HTTPS for public repos, or with token for private)
	cloneURL := fmt.Sprintf("https://github.com/%s/%s.git", owner, repo)

	// Get token for authentication
	if s.oauthClient != nil {
		token, err := s.oauthClient.GetToken(ctx)
		if err == nil && token != nil {
			// Use token in URL for authentication
			u, err := url.Parse(cloneURL)
			if err == nil {
				u.User = url.UserPassword("oauth2", token.AccessToken)
				cloneURL = u.String()
			}
		}
	}

	// Execute git clone
	cmd := exec.CommandContext(ctx, "git", "clone", cloneURL, destPath)
	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("git clone failed: %w\nOutput: %s", err, output)
	}

	return nil
}

// updateRepo updates an existing repository.
func (s *GitHubSource) updateRepo(ctx context.Context, repoPath string) error {
	// Execute git pull
	cmd := exec.CommandContext(ctx, "git", "-C", repoPath, "pull")
	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("git pull failed: %w\nOutput: %s", err, output)
	}

	return nil
}

// checkoutVersion checks out a specific version (tag or branch).
func (s *GitHubSource) checkoutVersion(ctx context.Context, repoPath, version string) error {
	// Try to checkout as tag first (with v prefix if not present)
	tags := []string{version}
	if !strings.HasPrefix(version, "v") {
		tags = append(tags, "v"+version)
	}

	var lastErr error
	for _, tag := range tags {
		cmd := exec.CommandContext(ctx, "git", "-C", repoPath, "checkout", tag)
		if err := cmd.Run(); err == nil {
			return nil
		} else {
			lastErr = err
		}
	}

	return fmt.Errorf("failed to checkout version %s: %w", version, lastErr)
}

// getClient returns an authenticated GitHub client.
func (s *GitHubSource) getClient(ctx context.Context) (*github.Client, error) {
	if s.client != nil {
		return s.client, nil
	}

	if s.oauthClient == nil {
		// Return unauthenticated client for public repositories
		s.client = github.NewClient(nil)
		return s.client, nil
	}

	token, err := s.oauthClient.GetToken(ctx)
	if err != nil {
		// Fall back to unauthenticated client
		s.client = github.NewClient(nil)
		return s.client, nil
	}

	ts := oauth2.StaticTokenSource(token)
	tc := oauth2.NewClient(ctx, ts)
	s.client = github.NewClient(tc)

	return s.client, nil
}
