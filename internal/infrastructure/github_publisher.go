package infrastructure

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/google/go-github/v57/github"
	"golang.org/x/oauth2"
)

// GitHubPublisher handles GitHub operations for collection publishing.
type GitHubPublisher struct {
	client *github.Client
	token  string
	ctx    context.Context
}

// NewGitHubPublisher creates a new GitHub publisher.
func NewGitHubPublisher(token string) *GitHubPublisher {
	ctx := context.Background()

	var client *github.Client
	if token != "" {
		ts := oauth2.StaticTokenSource(&oauth2.Token{AccessToken: token})
		tc := oauth2.NewClient(ctx, ts)
		client = github.NewClient(tc)
	} else {
		client = github.NewClient(nil)
	}

	return &GitHubPublisher{
		client: client,
		token:  token,
		ctx:    ctx,
	}
}

// ForkRepositoryOptions holds options for forking.
type ForkRepositoryOptions struct {
	Owner        string // Original repo owner
	Repo         string // Original repo name
	Organization string // Optional: fork to organization instead of user
}

// ForkRepository forks a repository.
func (p *GitHubPublisher) ForkRepository(opts *ForkRepositoryOptions) (*github.Repository, error) {
	if p.token == "" {
		return nil, errors.New("GitHub token required for forking")
	}

	// Create fork options
	forkOpts := &github.RepositoryCreateForkOptions{}
	if opts.Organization != "" {
		forkOpts.Organization = opts.Organization
	}

	// Fork repository
	fork, _, err := p.client.Repositories.CreateFork(p.ctx, opts.Owner, opts.Repo, forkOpts)
	if err != nil {
		return nil, fmt.Errorf("failed to fork repository: %w", err)
	}

	// Wait for fork to be ready (GitHub takes a few seconds)
	time.Sleep(5 * time.Second)

	return fork, nil
}

// CloneOptions holds options for cloning.
type CloneOptions struct {
	URL       string // Repository URL
	Directory string // Local directory to clone into
	Branch    string // Optional: specific branch to clone
	Depth     int    // Optional: shallow clone depth (0 = full clone)
}

// CloneRepository clones a repository locally.
func (p *GitHubPublisher) CloneRepository(opts *CloneOptions) error {
	args := []string{"clone"}

	// Add depth if specified (shallow clone)
	if opts.Depth > 0 {
		args = append(args, "--depth", strconv.Itoa(opts.Depth))
	}

	// Add branch if specified
	if opts.Branch != "" {
		args = append(args, "--branch", opts.Branch)
	}

	// Add URL and directory
	args = append(args, opts.URL, opts.Directory)

	// Execute git clone
	cmd := exec.Command("git", args...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("git clone failed: %w (output: %s)", err, string(output))
	}

	return nil
}

// CommitOptions holds options for committing changes.
type CommitOptions struct {
	RepoPath     string   // Local repository path
	Files        []string // Files to add (relative to repo root)
	Message      string   // Commit message
	AuthorName   string   // Author name
	AuthorEmail  string   // Author email
	Branch       string   // Branch to commit to (creates if doesn't exist)
	CreateBranch bool     // Create new branch from current HEAD
}

// CommitChanges commits changes to a local repository.
func (p *GitHubPublisher) CommitChanges(opts *CommitOptions) error {
	// Change to repo directory
	originalDir, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("failed to get current directory: %w", err)
	}
	defer os.Chdir(originalDir)

	if err := os.Chdir(opts.RepoPath); err != nil {
		return fmt.Errorf("failed to change to repo directory: %w", err)
	}

	// Configure git user if specified
	if opts.AuthorName != "" {
		cmd := exec.Command("git", "config", "user.name", opts.AuthorName)
		if output, err := cmd.CombinedOutput(); err != nil {
			return fmt.Errorf("failed to configure git user.name: %s", string(output))
		}
	}

	if opts.AuthorEmail != "" {
		cmd := exec.Command("git", "config", "user.email", opts.AuthorEmail)
		if output, err := cmd.CombinedOutput(); err != nil {
			return fmt.Errorf("failed to configure git user.email: %s", string(output))
		}
	}

	// Create new branch if requested
	if opts.CreateBranch && opts.Branch != "" {
		cmd := exec.Command("git", "checkout", "-b", opts.Branch)
		output, err := cmd.CombinedOutput()
		if err != nil {
			return fmt.Errorf("failed to create branch: %s", string(output))
		}
	} else if opts.Branch != "" {
		// Switch to existing branch
		cmd := exec.Command("git", "checkout", opts.Branch)
		output, err := cmd.CombinedOutput()
		if err != nil {
			return fmt.Errorf("failed to checkout branch: %s", string(output))
		}
	}

	// Add files
	for _, file := range opts.Files {
		cmd := exec.Command("git", "add", file)
		output, err := cmd.CombinedOutput()
		if err != nil {
			return fmt.Errorf("failed to add file %s: %s", file, string(output))
		}
	}

	// Commit
	cmd := exec.Command("git", "commit", "-m", opts.Message)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("git commit failed: %s", string(output))
	}

	return nil
}

// PushOptions holds options for pushing.
type PushOptions struct {
	RepoPath string // Local repository path
	Remote   string // Remote name (default: "origin")
	Branch   string // Branch to push
	Force    bool   // Force push
}

// PushChanges pushes changes to remote repository.
func (p *GitHubPublisher) PushChanges(opts *PushOptions) error {
	// Change to repo directory
	originalDir, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("failed to get current directory: %w", err)
	}
	defer os.Chdir(originalDir)

	if err := os.Chdir(opts.RepoPath); err != nil {
		return fmt.Errorf("failed to change to repo directory: %w", err)
	}

	// Build push command
	remote := opts.Remote
	if remote == "" {
		remote = "origin"
	}

	args := []string{"push"}
	if opts.Force {
		args = append(args, "--force")
	}
	args = append(args, remote, opts.Branch)

	// Execute push
	cmd := exec.Command("git", args...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("git push failed: %w (output: %s)", err, string(output))
	}

	return nil
}

// PullRequestOptions holds options for creating a pull request.
type PullRequestOptions struct {
	Owner       string // Repository owner
	Repo        string // Repository name
	Title       string // PR title
	Body        string // PR description
	Head        string // Head branch (user:branch or org:branch)
	Base        string // Base branch (usually "main" or "master")
	Draft       bool   // Create as draft PR
	Maintainers bool   // Allow maintainers to edit
}

// CreatePullRequest creates a pull request.
func (p *GitHubPublisher) CreatePullRequest(opts *PullRequestOptions) (*github.PullRequest, error) {
	if p.token == "" {
		return nil, errors.New("GitHub token required for creating pull requests")
	}

	// Create PR
	newPR := &github.NewPullRequest{
		Title:               github.String(opts.Title),
		Body:                github.String(opts.Body),
		Head:                github.String(opts.Head),
		Base:                github.String(opts.Base),
		Draft:               github.Bool(opts.Draft),
		MaintainerCanModify: github.Bool(opts.Maintainers),
	}

	pr, _, err := p.client.PullRequests.Create(p.ctx, opts.Owner, opts.Repo, newPR)
	if err != nil {
		return nil, fmt.Errorf("failed to create pull request: %w", err)
	}

	return pr, nil
}

// ReleaseOptions holds options for creating a release.
type ReleaseOptions struct {
	Owner      string   // Repository owner
	Repo       string   // Repository name
	Tag        string   // Release tag
	Name       string   // Release name
	Body       string   // Release description
	Draft      bool     // Create as draft
	Prerelease bool     // Mark as prerelease
	Assets     []string // File paths to upload as assets
}

// CreateRelease creates a GitHub release.
func (p *GitHubPublisher) CreateRelease(opts *ReleaseOptions) (*github.RepositoryRelease, error) {
	if p.token == "" {
		return nil, errors.New("GitHub token required for creating releases")
	}

	// Create release
	newRelease := &github.RepositoryRelease{
		TagName:    github.String(opts.Tag),
		Name:       github.String(opts.Name),
		Body:       github.String(opts.Body),
		Draft:      github.Bool(opts.Draft),
		Prerelease: github.Bool(opts.Prerelease),
	}

	release, _, err := p.client.Repositories.CreateRelease(p.ctx, opts.Owner, opts.Repo, newRelease)
	if err != nil {
		return nil, fmt.Errorf("failed to create release: %w", err)
	}

	// Upload assets
	for _, assetPath := range opts.Assets {
		if err := p.uploadReleaseAsset(opts.Owner, opts.Repo, release.GetID(), assetPath); err != nil {
			return release, fmt.Errorf("failed to upload asset %s: %w", assetPath, err)
		}
	}

	return release, nil
}

// uploadReleaseAsset uploads a file as a release asset.
func (p *GitHubPublisher) uploadReleaseAsset(owner, repo string, releaseID int64, assetPath string) error {
	// Open file
	file, err := os.Open(assetPath)
	if err != nil {
		return fmt.Errorf("failed to open asset file: %w", err)
	}
	defer file.Close()

	// Get file info
	fileInfo, err := file.Stat()
	if err != nil {
		return fmt.Errorf("failed to stat asset file: %w", err)
	}

	// Upload asset
	opts := &github.UploadOptions{
		Name: filepath.Base(assetPath),
	}

	_, _, err = p.client.Repositories.UploadReleaseAsset(p.ctx, owner, repo, releaseID, opts, file)
	if err != nil {
		return fmt.Errorf("failed to upload asset: %w", err)
	}

	_ = fileInfo // Use fileInfo to avoid unused variable warning

	return nil
}

// GetAuthenticatedUser returns the currently authenticated user.
func (p *GitHubPublisher) GetAuthenticatedUser() (*github.User, error) {
	if p.token == "" {
		return nil, errors.New("GitHub token required")
	}

	user, _, err := p.client.Users.Get(p.ctx, "")
	if err != nil {
		return nil, fmt.Errorf("failed to get authenticated user: %w", err)
	}

	return user, nil
}

// GetForkURL returns the fork URL for a repository.
func (p *GitHubPublisher) GetForkURL(owner, repo, username string) string {
	return fmt.Sprintf("https://github.com/%s/%s.git", username, repo)
}

// GetForkHTTPSURL returns the HTTPS clone URL with token embedded.
func (p *GitHubPublisher) GetForkHTTPSURL(owner, repo, username string) string {
	if p.token != "" {
		return fmt.Sprintf("https://%s@github.com/%s/%s.git", p.token, username, repo)
	}
	return p.GetForkURL(owner, repo, username)
}

// BuildPRTemplate builds a PR description from collection metadata.
func BuildPRTemplate(metadata map[string]interface{}) string {
	var sb strings.Builder

	sb.WriteString("## Collection Submission\n\n")

	// Basic info
	if name, ok := metadata["name"].(string); ok {
		sb.WriteString(fmt.Sprintf("**Name:** %s\n", name))
	}
	if version, ok := metadata["version"].(string); ok {
		sb.WriteString(fmt.Sprintf("**Version:** %s\n", version))
	}
	if author, ok := metadata["author"].(string); ok {
		sb.WriteString(fmt.Sprintf("**Author:** %s\n", author))
	}
	if category, ok := metadata["category"].(string); ok {
		sb.WriteString(fmt.Sprintf("**Category:** %s\n", category))
	}

	sb.WriteString("\n### Description\n\n")
	if description, ok := metadata["description"].(string); ok {
		sb.WriteString(description + "\n")
	}

	sb.WriteString("\n### Checklist\n\n")
	sb.WriteString("- [ ] Manifest validated (100+ rules passed)\n")
	sb.WriteString("- [ ] Security scan passed\n")
	sb.WriteString("- [ ] All elements tested\n")
	sb.WriteString("- [ ] Dependencies resolved\n")
	sb.WriteString("- [ ] Documentation complete\n")
	sb.WriteString("- [ ] Examples provided\n")

	// Stats
	sb.WriteString("\n### Statistics\n\n")
	if stats, ok := metadata["stats"].(map[string]interface{}); ok {
		if total, ok := stats["total_elements"].(int); ok {
			sb.WriteString(fmt.Sprintf("- **Total Elements:** %d\n", total))
		}
		if personas, ok := stats["personas"].(int); ok {
			sb.WriteString(fmt.Sprintf("- **Personas:** %d\n", personas))
		}
		if skills, ok := stats["skills"].(int); ok {
			sb.WriteString(fmt.Sprintf("- **Skills:** %d\n", skills))
		}
		if templates, ok := stats["templates"].(int); ok {
			sb.WriteString(fmt.Sprintf("- **Templates:** %d\n", templates))
		}
	}

	// Links
	sb.WriteString("\n### Links\n\n")
	if repo, ok := metadata["repository"].(string); ok && repo != "" {
		sb.WriteString(fmt.Sprintf("- **Repository:** %s\n", repo))
	}
	if docs, ok := metadata["documentation"].(string); ok && docs != "" {
		sb.WriteString(fmt.Sprintf("- **Documentation:** %s\n", docs))
	}
	if homepage, ok := metadata["homepage"].(string); ok && homepage != "" {
		sb.WriteString(fmt.Sprintf("- **Homepage:** %s\n", homepage))
	}

	return sb.String()
}
