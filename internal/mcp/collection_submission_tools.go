package mcp

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	sdk "github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/fsvxavier/nexs-mcp/internal/infrastructure"
	"github.com/fsvxavier/nexs-mcp/internal/portfolio"
)

// SubmitElementToCollectionInput represents input for submitting an element to collection
type SubmitElementToCollectionInput struct {
	ElementID          string `json:"element_id"`
	CollectionRepo     string `json:"collection_repo"`
	Title              string `json:"title,omitempty"`
	Description        string `json:"description,omitempty"`
	TargetBranch       string `json:"target_branch,omitempty"`
	SubmissionCategory string `json:"submission_category,omitempty"`
}

// SubmitElementToCollectionOutput represents the output of submission
type SubmitElementToCollectionOutput struct {
	PullRequestNumber int    `json:"pull_request_number"`
	PullRequestURL    string `json:"pull_request_url"`
	ForkURL           string `json:"fork_url"`
	BranchName        string `json:"branch_name"`
	Message           string `json:"message"`
}

// handleSubmitElementToCollection submits an element to the collection via GitHub PR
func (s *MCPServer) handleSubmitElementToCollection(ctx context.Context, req *sdk.CallToolRequest, input SubmitElementToCollectionInput) (*sdk.CallToolResult, SubmitElementToCollectionOutput, error) {
	// Validate input
	if input.ElementID == "" {
		return nil, SubmitElementToCollectionOutput{}, fmt.Errorf("element_id is required")
	}
	if input.CollectionRepo == "" {
		return nil, SubmitElementToCollectionOutput{}, fmt.Errorf("collection_repo is required")
	}

	// Parse collection repository
	collectionOwner, collectionRepo, err := infrastructure.ParseRepoURL(input.CollectionRepo)
	if err != nil {
		return nil, SubmitElementToCollectionOutput{}, fmt.Errorf("invalid collection repository: %w", err)
	}

	// Default target branch
	targetBranch := input.TargetBranch
	if targetBranch == "" {
		targetBranch = "main"
	}

	// Initialize clients
	homeDir, _ := os.UserHomeDir()
	tokenPath := filepath.Join(homeDir, ".nexs-mcp", "github_token.json")
	baseDir := filepath.Join(homeDir, ".nexs-mcp", "elements")

	oauthClient, err := infrastructure.NewGitHubOAuthClient(tokenPath)
	if err != nil {
		return nil, SubmitElementToCollectionOutput{}, fmt.Errorf("failed to initialize OAuth client: %w", err)
	}
	githubClient := infrastructure.NewGitHubClient(oauthClient)

	// Get the element to submit
	element, err := s.repo.GetByID(input.ElementID)
	if err != nil {
		return nil, SubmitElementToCollectionOutput{}, fmt.Errorf("element not found: %w", err)
	}

	// Get username
	username, err := githubClient.GetUser(ctx)
	if err != nil {
		return nil, SubmitElementToCollectionOutput{}, fmt.Errorf("failed to get GitHub username: %w", err)
	}

	// Step 1: Fork the collection repository
	fork, err := githubClient.ForkRepository(ctx, collectionOwner, collectionRepo)
	if err != nil {
		// Check if fork already exists
		if !strings.Contains(err.Error(), "already exists") {
			return nil, SubmitElementToCollectionOutput{}, fmt.Errorf("failed to fork repository: %w", err)
		}
		// Fork exists, continue
	}

	// Wait a bit for fork to be ready
	time.Sleep(3 * time.Second)

	// Step 2: Create a new branch for this submission
	metadata := element.GetMetadata()
	branchName := fmt.Sprintf("submit-%s-%s-%d",
		metadata.Type,
		sanitizeBranchName(metadata.Name),
		time.Now().Unix())

	err = githubClient.CreateBranch(ctx, username, collectionRepo, branchName, targetBranch)
	if err != nil {
		return nil, SubmitElementToCollectionOutput{}, fmt.Errorf("failed to create branch: %w", err)
	}

	// Step 3: Marshal element to YAML
	enhancedRepo, ok := s.repo.(*infrastructure.EnhancedFileElementRepository)
	if !ok {
		return nil, SubmitElementToCollectionOutput{}, fmt.Errorf("enhanced repository required")
	}

	stored := &infrastructure.StoredElement{
		Metadata: metadata,
	}
	yamlContent, err := enhancedRepo.MarshalElement(stored)
	if err != nil {
		return nil, SubmitElementToCollectionOutput{}, fmt.Errorf("failed to marshal element: %w", err)
	}

	// Step 4: Determine file path in collection
	mapper := portfolio.NewGitHubMapper(baseDir)
	githubPath := mapper.ElementToGitHubPath(element)

	// Step 5: Create/update file in the fork
	commitMessage := fmt.Sprintf("Add %s: %s\n\n%s",
		metadata.Type,
		metadata.Name,
		metadata.Description)

	_, err = githubClient.CreateFile(ctx, username, collectionRepo, githubPath, commitMessage, string(yamlContent), branchName)
	if err != nil {
		// Try update if file exists
		existingFile, getErr := githubClient.GetFile(ctx, username, collectionRepo, githubPath, branchName)
		if getErr == nil {
			_, err = githubClient.UpdateFile(ctx, username, collectionRepo, githubPath, commitMessage, string(yamlContent), existingFile.SHA, branchName)
		}
		if err != nil {
			return nil, SubmitElementToCollectionOutput{}, fmt.Errorf("failed to create/update file: %w", err)
		}
	}

	// Step 6: Create pull request
	prTitle := input.Title
	if prTitle == "" {
		prTitle = fmt.Sprintf("Submit %s: %s", metadata.Type, metadata.Name)
	}

	prBody := input.Description
	if prBody == "" {
		prBody = generatePRDescription(element, input.SubmissionCategory)
	}

	headBranch := fmt.Sprintf("%s:%s", username, branchName)
	pr, err := githubClient.CreatePullRequest(ctx, collectionOwner, collectionRepo, prTitle, prBody, headBranch, targetBranch)
	if err != nil {
		return nil, SubmitElementToCollectionOutput{}, fmt.Errorf("failed to create pull request: %w", err)
	}

	output := SubmitElementToCollectionOutput{
		PullRequestNumber: pr.Number,
		PullRequestURL:    pr.URL,
		ForkURL:           fork.URL,
		BranchName:        branchName,
		Message:           fmt.Sprintf("Successfully submitted %s '%s' to collection. PR #%d created.", metadata.Type, metadata.Name, pr.Number),
	}

	return nil, output, nil
}

// sanitizeBranchName sanitizes a string to be used as a branch name
func sanitizeBranchName(name string) string {
	// Replace spaces and special characters with hyphens
	name = strings.ToLower(name)
	name = strings.ReplaceAll(name, " ", "-")
	name = strings.ReplaceAll(name, "_", "-")

	// Remove non-alphanumeric characters except hyphens
	var result strings.Builder
	for _, r := range name {
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') || r == '-' {
			result.WriteRune(r)
		}
	}

	// Trim hyphens from start/end
	sanitized := strings.Trim(result.String(), "-")

	// Limit length
	if len(sanitized) > 50 {
		sanitized = sanitized[:50]
	}

	return sanitized
}

// generatePRDescription generates a PR description for an element submission
func generatePRDescription(element interface{}, category string) string {
	var sb strings.Builder

	metadata := element.(interface{ GetMetadata() interface{} }).GetMetadata()
	meta := metadata.(struct {
		Name        string
		Description string
		Version     string
		Author      string
		Tags        []string
	})

	sb.WriteString("## Element Submission\n\n")
	sb.WriteString(fmt.Sprintf("**Name:** %s\n", meta.Name))
	sb.WriteString(fmt.Sprintf("**Description:** %s\n", meta.Description))
	sb.WriteString(fmt.Sprintf("**Version:** %s\n", meta.Version))
	sb.WriteString(fmt.Sprintf("**Author:** %s\n\n", meta.Author))

	if len(meta.Tags) > 0 {
		sb.WriteString(fmt.Sprintf("**Tags:** %s\n\n", strings.Join(meta.Tags, ", ")))
	}

	if category != "" {
		sb.WriteString(fmt.Sprintf("**Category:** %s\n\n", category))
	}

	sb.WriteString("---\n\n")
	sb.WriteString("This element has been submitted for inclusion in the collection.\n")
	sb.WriteString("Please review the content and provide feedback.\n")

	return sb.String()
}
