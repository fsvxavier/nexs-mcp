package infrastructure

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewGitHubPublisher(t *testing.T) {
	tests := []struct {
		name  string
		token string
	}{
		{
			name:  "With Token",
			token: "ghp_test_token",
		},
		{
			name:  "Without Token",
			token: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			publisher := NewGitHubPublisher(tt.token)
			assert.NotNil(t, publisher)
			assert.NotNil(t, publisher.client)
			assert.NotNil(t, publisher.ctx)
			assert.Equal(t, tt.token, publisher.token)
		})
	}
}

func TestGetForkURL(t *testing.T) {
	publisher := NewGitHubPublisher("")

	tests := []struct {
		name     string
		owner    string
		repo     string
		username string
		expected string
	}{
		{
			name:     "Standard Repository",
			owner:    "original-owner",
			repo:     "test-repo",
			username: "forked-user",
			expected: "https://github.com/forked-user/test-repo.git",
		},
		{
			name:     "Organization Repository",
			owner:    "org",
			repo:     "collection-repo",
			username: "user",
			expected: "https://github.com/user/collection-repo.git",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			url := publisher.GetForkURL(tt.owner, tt.repo, tt.username)
			assert.Equal(t, tt.expected, url)
		})
	}
}

func TestGetForkHTTPSURL(t *testing.T) {
	tests := []struct {
		name     string
		token    string
		owner    string
		repo     string
		username string
		expected string
	}{
		{
			name:     "With Token",
			token:    "test_token",
			owner:    "original-owner",
			repo:     "test-repo",
			username: "forked-user",
			expected: "https://test_token@github.com/forked-user/test-repo.git",
		},
		{
			name:     "Without Token",
			token:    "",
			owner:    "original-owner",
			repo:     "test-repo",
			username: "forked-user",
			expected: "https://github.com/forked-user/test-repo.git",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			publisher := NewGitHubPublisher(tt.token)
			url := publisher.GetForkHTTPSURL(tt.owner, tt.repo, tt.username)
			assert.Equal(t, tt.expected, url)
		})
	}
}

func TestBuildPRTemplate_Complete(t *testing.T) {
	metadata := map[string]interface{}{
		"name":          "Test Collection",
		"version":       "1.0.0",
		"author":        "Test Author",
		"category":      "Testing",
		"description":   "This is a test collection for unit testing",
		"repository":    "https://github.com/test/repo",
		"documentation": "https://docs.example.com",
		"homepage":      "https://example.com",
		"stats": map[string]interface{}{
			"total_elements": 10,
			"personas":       3,
			"skills":         5,
			"templates":      2,
		},
	}

	template := BuildPRTemplate(metadata)

	// Verify basic structure
	assert.Contains(t, template, "## Collection Submission")
	assert.Contains(t, template, "**Name:** Test Collection")
	assert.Contains(t, template, "**Version:** 1.0.0")
	assert.Contains(t, template, "**Author:** Test Author")
	assert.Contains(t, template, "**Category:** Testing")

	// Verify description
	assert.Contains(t, template, "### Description")
	assert.Contains(t, template, "This is a test collection for unit testing")

	// Verify checklist
	assert.Contains(t, template, "### Checklist")
	assert.Contains(t, template, "- [ ] Manifest validated (100+ rules passed)")
	assert.Contains(t, template, "- [ ] Security scan passed")
	assert.Contains(t, template, "- [ ] All elements tested")
	assert.Contains(t, template, "- [ ] Dependencies resolved")
	assert.Contains(t, template, "- [ ] Documentation complete")
	assert.Contains(t, template, "- [ ] Examples provided")

	// Verify statistics
	assert.Contains(t, template, "### Statistics")
	assert.Contains(t, template, "- **Total Elements:** 10")
	assert.Contains(t, template, "- **Personas:** 3")
	assert.Contains(t, template, "- **Skills:** 5")
	assert.Contains(t, template, "- **Templates:** 2")

	// Verify links
	assert.Contains(t, template, "### Links")
	assert.Contains(t, template, "- **Repository:** https://github.com/test/repo")
	assert.Contains(t, template, "- **Documentation:** https://docs.example.com")
	assert.Contains(t, template, "- **Homepage:** https://example.com")
}

func TestBuildPRTemplate_MinimalMetadata(t *testing.T) {
	metadata := map[string]interface{}{
		"name":    "Minimal Collection",
		"version": "0.1.0",
	}

	template := BuildPRTemplate(metadata)

	// Should still have basic structure
	assert.Contains(t, template, "## Collection Submission")
	assert.Contains(t, template, "**Name:** Minimal Collection")
	assert.Contains(t, template, "**Version:** 0.1.0")
	assert.Contains(t, template, "### Checklist")

	// Should not have optional fields
	assert.NotContains(t, template, "**Author:**")
	assert.NotContains(t, template, "**Category:**")
}

func TestBuildPRTemplate_EmptyMetadata(t *testing.T) {
	metadata := map[string]interface{}{}

	template := BuildPRTemplate(metadata)

	// Should have minimal structure
	assert.Contains(t, template, "## Collection Submission")
	assert.Contains(t, template, "### Description")
	assert.Contains(t, template, "### Checklist")
	assert.Contains(t, template, "### Statistics")
	assert.Contains(t, template, "### Links")
}

func TestBuildPRTemplate_WithStats(t *testing.T) {
	metadata := map[string]interface{}{
		"name": "Stats Collection",
		"stats": map[string]interface{}{
			"total_elements": 20,
			"personas":       5,
			"skills":         10,
			"templates":      5,
		},
	}

	template := BuildPRTemplate(metadata)

	assert.Contains(t, template, "- **Total Elements:** 20")
	assert.Contains(t, template, "- **Personas:** 5")
	assert.Contains(t, template, "- **Skills:** 10")
	assert.Contains(t, template, "- **Templates:** 5")
}

func TestBuildPRTemplate_WithPartialStats(t *testing.T) {
	metadata := map[string]interface{}{
		"name": "Partial Stats",
		"stats": map[string]interface{}{
			"total_elements": 15,
			"personas":       3,
			// Missing skills and templates
		},
	}

	template := BuildPRTemplate(metadata)

	assert.Contains(t, template, "- **Total Elements:** 15")
	assert.Contains(t, template, "- **Personas:** 3")
	// Should not have skills and templates if not provided
	assert.NotContains(t, template, "- **Skills:**")
	assert.NotContains(t, template, "- **Templates:**")
}

func TestBuildPRTemplate_WithLinks(t *testing.T) {
	metadata := map[string]interface{}{
		"name":          "Links Collection",
		"repository":    "https://github.com/user/repo",
		"documentation": "https://docs.example.com",
		"homepage":      "https://example.com",
	}

	template := BuildPRTemplate(metadata)

	assert.Contains(t, template, "### Links")
	assert.Contains(t, template, "- **Repository:** https://github.com/user/repo")
	assert.Contains(t, template, "- **Documentation:** https://docs.example.com")
	assert.Contains(t, template, "- **Homepage:** https://example.com")
}

func TestBuildPRTemplate_EmptyLinks(t *testing.T) {
	metadata := map[string]interface{}{
		"name":          "No Links",
		"repository":    "",
		"documentation": "",
		"homepage":      "",
	}

	template := BuildPRTemplate(metadata)

	assert.Contains(t, template, "### Links")
	// Empty links should not appear
	linksIdx := strings.Index(template, "### Links")
	if linksIdx == -1 {
		t.Fatal("Expected ### Links section")
	}
	linkSection := template[linksIdx:]
	assert.NotContains(t, linkSection, "- **Repository:**")
	assert.NotContains(t, linkSection, "- **Documentation:**")
	assert.NotContains(t, linkSection, "- **Homepage:**")
}

func TestForkRepositoryOptions_Structure(t *testing.T) {
	opts := ForkRepositoryOptions{
		Owner:        "original-owner",
		Repo:         "test-repo",
		Organization: "my-org",
	}

	assert.Equal(t, "original-owner", opts.Owner)
	assert.Equal(t, "test-repo", opts.Repo)
	assert.Equal(t, "my-org", opts.Organization)
}

func TestCloneOptions_Structure(t *testing.T) {
	opts := CloneOptions{
		URL:       "https://github.com/user/repo.git",
		Directory: "/tmp/repo",
		Branch:    "main",
		Depth:     1,
	}

	assert.Equal(t, "https://github.com/user/repo.git", opts.URL)
	assert.Equal(t, "/tmp/repo", opts.Directory)
	assert.Equal(t, "main", opts.Branch)
	assert.Equal(t, 1, opts.Depth)
}

func TestCommitOptions_Structure(t *testing.T) {
	opts := CommitOptions{
		RepoPath:     "/tmp/repo",
		Files:        []string{"file1.txt", "file2.txt"},
		Message:      "Test commit",
		AuthorName:   "Test Author",
		AuthorEmail:  "test@example.com",
		Branch:       "feature-branch",
		CreateBranch: true,
	}

	assert.Equal(t, "/tmp/repo", opts.RepoPath)
	assert.Equal(t, []string{"file1.txt", "file2.txt"}, opts.Files)
	assert.Equal(t, "Test commit", opts.Message)
	assert.Equal(t, "Test Author", opts.AuthorName)
	assert.Equal(t, "test@example.com", opts.AuthorEmail)
	assert.Equal(t, "feature-branch", opts.Branch)
	assert.True(t, opts.CreateBranch)
}

func TestPushOptions_Structure(t *testing.T) {
	opts := PushOptions{
		RepoPath: "/tmp/repo",
		Remote:   "origin",
		Branch:   "main",
		Force:    false,
	}

	assert.Equal(t, "/tmp/repo", opts.RepoPath)
	assert.Equal(t, "origin", opts.Remote)
	assert.Equal(t, "main", opts.Branch)
	assert.False(t, opts.Force)
}

func TestPullRequestOptions_Structure(t *testing.T) {
	opts := PullRequestOptions{
		Owner:       "owner",
		Repo:        "repo",
		Title:       "Test PR",
		Body:        "Test description",
		Head:        "user:feature",
		Base:        "main",
		Draft:       false,
		Maintainers: true,
	}

	assert.Equal(t, "owner", opts.Owner)
	assert.Equal(t, "repo", opts.Repo)
	assert.Equal(t, "Test PR", opts.Title)
	assert.Equal(t, "Test description", opts.Body)
	assert.Equal(t, "user:feature", opts.Head)
	assert.Equal(t, "main", opts.Base)
	assert.False(t, opts.Draft)
	assert.True(t, opts.Maintainers)
}

func TestReleaseOptions_Structure(t *testing.T) {
	opts := ReleaseOptions{
		Owner:      "owner",
		Repo:       "repo",
		Tag:        "v1.0.0",
		Name:       "Release 1.0.0",
		Body:       "Release notes",
		Draft:      false,
		Prerelease: false,
		Assets:     []string{"asset1.zip", "asset2.tar.gz"},
	}

	assert.Equal(t, "owner", opts.Owner)
	assert.Equal(t, "repo", opts.Repo)
	assert.Equal(t, "v1.0.0", opts.Tag)
	assert.Equal(t, "Release 1.0.0", opts.Name)
	assert.Equal(t, "Release notes", opts.Body)
	assert.False(t, opts.Draft)
	assert.False(t, opts.Prerelease)
	assert.Equal(t, []string{"asset1.zip", "asset2.tar.gz"}, opts.Assets)
}

func TestBuildPRTemplate_LongDescription(t *testing.T) {
	longDesc := strings.Repeat("This is a very long description. ", 20)
	metadata := map[string]interface{}{
		"name":        "Long Desc Collection",
		"description": longDesc,
	}

	template := BuildPRTemplate(metadata)

	assert.Contains(t, template, longDesc)
	assert.Contains(t, template, "### Description")
}

func TestBuildPRTemplate_SpecialCharacters(t *testing.T) {
	metadata := map[string]interface{}{
		"name":        "Special & Characters < > \"",
		"description": "Description with special chars: & < > \" '",
		"author":      "Author <test@example.com>",
	}

	template := BuildPRTemplate(metadata)

	// Special characters should be preserved
	assert.Contains(t, template, "Special & Characters < > \"")
	assert.Contains(t, template, "Description with special chars: & < > \" '")
	assert.Contains(t, template, "Author <test@example.com>")
}
