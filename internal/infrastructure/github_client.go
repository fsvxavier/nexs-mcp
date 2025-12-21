package infrastructure

import (
	"context"
	"fmt"
	"strings"

	"github.com/google/go-github/v57/github"
	"golang.org/x/oauth2"
)

// GitHubClientInterface defines the interface for GitHub operations.
type GitHubClientInterface interface {
	ListRepositories(ctx context.Context) ([]*Repository, error)
	GetFile(ctx context.Context, owner, repo, path, branch string) (*FileContent, error)
	CreateFile(ctx context.Context, owner, repo, path, message, content, branch string) (*CommitInfo, error)
	UpdateFile(ctx context.Context, owner, repo, path, message, content, sha, branch string) (*CommitInfo, error)
	DeleteFile(ctx context.Context, owner, repo, path, message, sha, branch string) error
	ListFilesInDirectory(ctx context.Context, owner, repo, path, branch string) ([]string, error)
	ListAllFiles(ctx context.Context, owner, repo, branch string) ([]string, error)
	GetUser(ctx context.Context) (string, error)
	CreateRepository(ctx context.Context, name, description string, private bool) (*Repository, error)
}

// GitHubClient wraps the GitHub API client with high-level operations.
type GitHubClient struct {
	client      *github.Client
	oauthClient *GitHubOAuthClient
}

// NewGitHubClient creates a new GitHub API client.
func NewGitHubClient(oauthClient *GitHubOAuthClient) *GitHubClient {
	return &GitHubClient{
		oauthClient: oauthClient,
	}
}

// ensureAuthenticated ensures the client is authenticated and returns the authenticated client.
func (c *GitHubClient) ensureAuthenticated(ctx context.Context) (*github.Client, error) {
	if c.client != nil {
		return c.client, nil
	}

	token, err := c.oauthClient.GetToken(ctx)
	if err != nil {
		return nil, fmt.Errorf("not authenticated: %w", err)
	}

	ts := oauth2.StaticTokenSource(token)
	tc := oauth2.NewClient(ctx, ts)
	c.client = github.NewClient(tc)

	return c.client, nil
}

// Repository represents a GitHub repository.
type Repository struct {
	Owner         string
	Name          string
	FullName      string
	Description   string
	Private       bool
	URL           string
	DefaultBranch string
}

// FileContent represents a file in a GitHub repository.
type FileContent struct {
	Path    string
	Content string
	SHA     string
	Size    int
}

// CommitInfo represents information about a commit.
type CommitInfo struct {
	SHA     string
	Message string
	Author  string
	Date    string
}

// ListRepositories lists all repositories for the authenticated user.
func (c *GitHubClient) ListRepositories(ctx context.Context) ([]*Repository, error) {
	client, err := c.ensureAuthenticated(ctx)
	if err != nil {
		return nil, err
	}

	var allRepos []*Repository
	opt := &github.RepositoryListByAuthenticatedUserOptions{
		ListOptions: github.ListOptions{PerPage: 100},
	}

	for {
		repos, resp, err := client.Repositories.ListByAuthenticatedUser(ctx, opt)
		if err != nil {
			return nil, fmt.Errorf("failed to list repositories: %w", err)
		}

		for _, repo := range repos {
			allRepos = append(allRepos, &Repository{
				Owner:         repo.GetOwner().GetLogin(),
				Name:          repo.GetName(),
				FullName:      repo.GetFullName(),
				Description:   repo.GetDescription(),
				Private:       repo.GetPrivate(),
				URL:           repo.GetHTMLURL(),
				DefaultBranch: repo.GetDefaultBranch(),
			})
		}

		if resp.NextPage == 0 {
			break
		}
		opt.Page = resp.NextPage
	}

	return allRepos, nil
}

// GetFile retrieves a file from a GitHub repository.
func (c *GitHubClient) GetFile(ctx context.Context, owner, repo, path, branch string) (*FileContent, error) {
	client, err := c.ensureAuthenticated(ctx)
	if err != nil {
		return nil, err
	}

	opts := &github.RepositoryContentGetOptions{
		Ref: branch,
	}

	fileContent, _, _, err := client.Repositories.GetContents(ctx, owner, repo, path, opts)
	if err != nil {
		return nil, fmt.Errorf("failed to get file %s: %w", path, err)
	}

	if fileContent == nil {
		return nil, fmt.Errorf("path %s is a directory, not a file", path)
	}

	content, err := fileContent.GetContent()
	if err != nil {
		return nil, fmt.Errorf("failed to decode file content: %w", err)
	}

	return &FileContent{
		Path:    fileContent.GetPath(),
		Content: content,
		SHA:     fileContent.GetSHA(),
		Size:    fileContent.GetSize(),
	}, nil
}

// CreateFile creates a new file in a GitHub repository.
func (c *GitHubClient) CreateFile(ctx context.Context, owner, repo, path, message, content, branch string) (*CommitInfo, error) {
	client, err := c.ensureAuthenticated(ctx)
	if err != nil {
		return nil, err
	}

	opts := &github.RepositoryContentFileOptions{
		Message: github.String(message),
		Content: []byte(content),
		Branch:  github.String(branch),
	}

	result, _, err := client.Repositories.CreateFile(ctx, owner, repo, path, opts)
	if err != nil {
		return nil, fmt.Errorf("failed to create file %s: %w", path, err)
	}

	return &CommitInfo{
		SHA:     result.GetSHA(),
		Message: message,
	}, nil
}

// UpdateFile updates an existing file in a GitHub repository.
func (c *GitHubClient) UpdateFile(ctx context.Context, owner, repo, path, message, content, sha, branch string) (*CommitInfo, error) {
	client, err := c.ensureAuthenticated(ctx)
	if err != nil {
		return nil, err
	}

	opts := &github.RepositoryContentFileOptions{
		Message: github.String(message),
		Content: []byte(content),
		SHA:     github.String(sha),
		Branch:  github.String(branch),
	}

	result, _, err := client.Repositories.UpdateFile(ctx, owner, repo, path, opts)
	if err != nil {
		return nil, fmt.Errorf("failed to update file %s: %w", path, err)
	}

	return &CommitInfo{
		SHA:     result.GetSHA(),
		Message: message,
	}, nil
}

// DeleteFile deletes a file from a GitHub repository.
func (c *GitHubClient) DeleteFile(ctx context.Context, owner, repo, path, message, sha, branch string) error {
	client, err := c.ensureAuthenticated(ctx)
	if err != nil {
		return err
	}

	opts := &github.RepositoryContentFileOptions{
		Message: github.String(message),
		SHA:     github.String(sha),
		Branch:  github.String(branch),
	}

	_, _, err = client.Repositories.DeleteFile(ctx, owner, repo, path, opts)
	if err != nil {
		return fmt.Errorf("failed to delete file %s: %w", path, err)
	}

	return nil
}

// ListFilesInDirectory lists all files in a directory (non-recursive).
func (c *GitHubClient) ListFilesInDirectory(ctx context.Context, owner, repo, path, branch string) ([]string, error) {
	client, err := c.ensureAuthenticated(ctx)
	if err != nil {
		return nil, err
	}

	opts := &github.RepositoryContentGetOptions{
		Ref: branch,
	}

	_, dirContents, _, err := client.Repositories.GetContents(ctx, owner, repo, path, opts)
	if err != nil {
		return nil, fmt.Errorf("failed to list directory %s: %w", path, err)
	}

	var files []string
	for _, item := range dirContents {
		if item.GetType() == "file" {
			files = append(files, item.GetPath())
		}
	}

	return files, nil
}

// ListAllFiles recursively lists all files in a repository tree.
func (c *GitHubClient) ListAllFiles(ctx context.Context, owner, repo, branch string) ([]string, error) {
	client, err := c.ensureAuthenticated(ctx)
	if err != nil {
		return nil, err
	}

	// Get the tree recursively
	tree, _, err := client.Git.GetTree(ctx, owner, repo, branch, true)
	if err != nil {
		return nil, fmt.Errorf("failed to get repository tree: %w", err)
	}

	var files []string
	for _, entry := range tree.Entries {
		if entry.GetType() == "blob" {
			files = append(files, entry.GetPath())
		}
	}

	return files, nil
}

// GetUser returns the authenticated user's information.
func (c *GitHubClient) GetUser(ctx context.Context) (string, error) {
	client, err := c.ensureAuthenticated(ctx)
	if err != nil {
		return "", err
	}

	user, _, err := client.Users.Get(ctx, "")
	if err != nil {
		return "", fmt.Errorf("failed to get user info: %w", err)
	}

	return user.GetLogin(), nil
}

// CreateRepository creates a new repository for the authenticated user.
func (c *GitHubClient) CreateRepository(ctx context.Context, name, description string, private bool) (*Repository, error) {
	client, err := c.ensureAuthenticated(ctx)
	if err != nil {
		return nil, err
	}

	repo := &github.Repository{
		Name:        github.String(name),
		Description: github.String(description),
		Private:     github.Bool(private),
		AutoInit:    github.Bool(true), // Initialize with README
	}

	created, _, err := client.Repositories.Create(ctx, "", repo)
	if err != nil {
		return nil, fmt.Errorf("failed to create repository: %w", err)
	}

	return &Repository{
		Owner:         created.GetOwner().GetLogin(),
		Name:          created.GetName(),
		FullName:      created.GetFullName(),
		Description:   created.GetDescription(),
		Private:       created.GetPrivate(),
		URL:           created.GetHTMLURL(),
		DefaultBranch: created.GetDefaultBranch(),
	}, nil
}

// ParseRepoURL parses a GitHub repository URL into owner and repo name.
func ParseRepoURL(url string) (owner, repo string, err error) {
	// Support formats:
	// - https://github.com/owner/repo
	// - owner/repo
	// - github.com/owner/repo

	url = strings.TrimPrefix(url, "https://")
	url = strings.TrimPrefix(url, "http://")
	url = strings.TrimPrefix(url, "github.com/")
	url = strings.TrimSuffix(url, ".git")

	parts := strings.Split(url, "/")
	if len(parts) < 2 {
		return "", "", fmt.Errorf("invalid repository URL format: %s", url)
	}

	owner = parts[0]
	repo = parts[1]

	return owner, repo, nil
}

// ForkRepository forks a repository to the authenticated user's account.
func (c *GitHubClient) ForkRepository(ctx context.Context, owner, repo string) (*Repository, error) {
	client, err := c.ensureAuthenticated(ctx)
	if err != nil {
		return nil, fmt.Errorf("authentication failed: %w", err)
	}

	// Create fork
	repoObj, _, err := client.Repositories.CreateFork(ctx, owner, repo, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to fork repository: %w", err)
	}

	return &Repository{
		Owner:       repoObj.GetOwner().GetLogin(),
		Name:        repoObj.GetName(),
		FullName:    repoObj.GetFullName(),
		Description: repoObj.GetDescription(),
		Private:     repoObj.GetPrivate(),
		URL:         repoObj.GetHTMLURL(),
	}, nil
}

// CreateBranch creates a new branch from a base branch.
func (c *GitHubClient) CreateBranch(ctx context.Context, owner, repo, newBranch, baseBranch string) error {
	client, err := c.ensureAuthenticated(ctx)
	if err != nil {
		return fmt.Errorf("authentication failed: %w", err)
	}

	// Get base branch reference
	baseRef, _, err := client.Git.GetRef(ctx, owner, repo, "refs/heads/"+baseBranch)
	if err != nil {
		return fmt.Errorf("failed to get base branch: %w", err)
	}

	// Create new branch reference
	newRef := &github.Reference{
		Ref: github.String("refs/heads/" + newBranch),
		Object: &github.GitObject{
			SHA: baseRef.Object.SHA,
		},
	}

	_, _, err = client.Git.CreateRef(ctx, owner, repo, newRef)
	if err != nil {
		return fmt.Errorf("failed to create branch: %w", err)
	}

	return nil
}

// CreatePullRequest creates a pull request.
func (c *GitHubClient) CreatePullRequest(ctx context.Context, owner, repo, title, body, head, base string) (*PullRequest, error) {
	client, err := c.ensureAuthenticated(ctx)
	if err != nil {
		return nil, fmt.Errorf("authentication failed: %w", err)
	}

	newPR := &github.NewPullRequest{
		Title: github.String(title),
		Body:  github.String(body),
		Head:  github.String(head),
		Base:  github.String(base),
	}

	pr, _, err := client.PullRequests.Create(ctx, owner, repo, newPR)
	if err != nil {
		return nil, fmt.Errorf("failed to create pull request: %w", err)
	}

	return &PullRequest{
		Number: pr.GetNumber(),
		Title:  pr.GetTitle(),
		Body:   pr.GetBody(),
		State:  pr.GetState(),
		URL:    pr.GetHTMLURL(),
		Head:   pr.GetHead().GetRef(),
		Base:   pr.GetBase().GetRef(),
	}, nil
}

// PullRequest represents a GitHub pull request.
type PullRequest struct {
	Number int
	Title  string
	Body   string
	State  string
	URL    string
	Head   string
	Base   string
}
