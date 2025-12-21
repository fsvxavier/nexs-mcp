package infrastructure

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/fsvxavier/nexs-mcp/internal/domain"
)

// PRStatus represents the status of a pull request.
type PRStatus string

const (
	// PRStatusPending indicates PR is awaiting review.
	PRStatusPending PRStatus = "pending"
	// PRStatusMerged indicates PR has been merged.
	PRStatusMerged PRStatus = "merged"
	// PRStatusRejected indicates PR was rejected/closed.
	PRStatusRejected PRStatus = "rejected"
	// PRStatusDraft indicates PR is a draft.
	PRStatusDraft PRStatus = "draft"
)

// PRSubmission represents a single PR submission.
type PRSubmission struct {
	ID              string                 `json:"id"` // Unique submission ID
	ElementID       string                 `json:"element_id"`
	ElementType     domain.ElementType     `json:"element_type"`
	ElementName     string                 `json:"element_name"`
	ElementVersion  string                 `json:"element_version"`
	RepositoryOwner string                 `json:"repository_owner"`
	RepositoryName  string                 `json:"repository_name"`
	PRNumber        int                    `json:"pr_number"`
	PRTitle         string                 `json:"pr_title"`
	PRURL           string                 `json:"pr_url"`
	Status          PRStatus               `json:"status"`
	SubmittedAt     time.Time              `json:"submitted_at"`
	UpdatedAt       time.Time              `json:"updated_at"`
	MergedAt        *time.Time             `json:"merged_at,omitempty"`
	ClosedAt        *time.Time             `json:"closed_at,omitempty"`
	SubmittedBy     string                 `json:"submitted_by"`
	ReviewComments  int                    `json:"review_comments"`
	Notes           string                 `json:"notes,omitempty"`
	Metadata        map[string]interface{} `json:"metadata,omitempty"`
}

// PRHistory contains all PR submissions.
type PRHistory struct {
	Version     string                  `json:"version"`
	Submissions map[string]PRSubmission `json:"submissions"` // key is PR ID
	Stats       PRStats                 `json:"stats"`
	LastUpdated time.Time               `json:"last_updated"`
}

// PRStats contains statistics about PR submissions.
type PRStats struct {
	TotalSubmissions int `json:"total_submissions"`
	Pending          int `json:"pending"`
	Merged           int `json:"merged"`
	Rejected         int `json:"rejected"`
	Draft            int `json:"draft"`
}

// PRTracker tracks pull request submissions.
type PRTracker struct {
	historyFile string
}

// NewPRTracker creates a new PR tracker.
func NewPRTracker(configDir string) *PRTracker {
	if configDir == "" {
		// Use default config directory
		home, err := os.UserHomeDir()
		if err != nil {
			configDir = "."
		} else {
			configDir = filepath.Join(home, ".nexs-mcp")
		}
	}

	historyFile := filepath.Join(configDir, "pr-history.json")
	return &PRTracker{
		historyFile: historyFile,
	}
}

// LoadHistory loads PR history from disk.
func (t *PRTracker) LoadHistory() (*PRHistory, error) {
	// Check if file exists
	if _, err := os.Stat(t.historyFile); os.IsNotExist(err) {
		return &PRHistory{
			Version:     "1.0.0",
			Submissions: make(map[string]PRSubmission),
			Stats:       PRStats{},
			LastUpdated: time.Now(),
		}, nil
	}

	// Read file
	data, err := os.ReadFile(t.historyFile)
	if err != nil {
		return nil, fmt.Errorf("failed to read PR history file: %w", err)
	}

	// Unmarshal JSON
	var history PRHistory
	if err := json.Unmarshal(data, &history); err != nil {
		return nil, fmt.Errorf("failed to unmarshal PR history: %w", err)
	}

	// Initialize maps if nil
	if history.Submissions == nil {
		history.Submissions = make(map[string]PRSubmission)
	}

	return &history, nil
}

// SaveHistory saves PR history to disk.
func (t *PRTracker) SaveHistory(history *PRHistory) error {
	// Create directory if it doesn't exist
	dir := filepath.Dir(t.historyFile)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	// Update stats before saving
	t.updateStats(history)
	history.LastUpdated = time.Now()

	// Marshal to JSON
	data, err := json.MarshalIndent(history, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal PR history: %w", err)
	}

	// Write to file
	if err := os.WriteFile(t.historyFile, data, 0644); err != nil {
		return fmt.Errorf("failed to write PR history file: %w", err)
	}

	return nil
}

// TrackSubmission adds a new PR submission to history.
func (t *PRTracker) TrackSubmission(submission PRSubmission) error {
	history, err := t.LoadHistory()
	if err != nil {
		return fmt.Errorf("failed to load history: %w", err)
	}

	// Generate ID if not set
	if submission.ID == "" {
		submission.ID = t.generateSubmissionID(submission)
	}

	// Set timestamps
	submission.SubmittedAt = time.Now()
	submission.UpdatedAt = submission.SubmittedAt

	// Add to history
	history.Submissions[submission.ID] = submission

	// Save history
	return t.SaveHistory(history)
}

// UpdateSubmissionStatus updates the status of a PR submission.
func (t *PRTracker) UpdateSubmissionStatus(prID string, status PRStatus) error {
	history, err := t.LoadHistory()
	if err != nil {
		return fmt.Errorf("failed to load history: %w", err)
	}

	submission, exists := history.Submissions[prID]
	if !exists {
		return fmt.Errorf("PR submission not found: %s", prID)
	}

	submission.Status = status
	submission.UpdatedAt = time.Now()

	// Update merge/close timestamps
	now := time.Now()
	switch status {
	case PRStatusMerged:
		submission.MergedAt = &now
	case PRStatusRejected:
		submission.ClosedAt = &now
	}

	history.Submissions[prID] = submission

	return t.SaveHistory(history)
}

// GetSubmissionByPRNumber retrieves a submission by PR number.
func (t *PRTracker) GetSubmissionByPRNumber(owner, repo string, prNumber int) (*PRSubmission, error) {
	history, err := t.LoadHistory()
	if err != nil {
		return nil, fmt.Errorf("failed to load history: %w", err)
	}

	for _, submission := range history.Submissions {
		if submission.RepositoryOwner == owner &&
			submission.RepositoryName == repo &&
			submission.PRNumber == prNumber {
			return &submission, nil
		}
	}

	return nil, fmt.Errorf("submission not found for PR #%d in %s/%s", prNumber, owner, repo)
}

// GetSubmissionsByElement retrieves all submissions for a specific element.
func (t *PRTracker) GetSubmissionsByElement(elementID string) ([]PRSubmission, error) {
	history, err := t.LoadHistory()
	if err != nil {
		return nil, fmt.Errorf("failed to load history: %w", err)
	}

	submissions := []PRSubmission{}
	for _, submission := range history.Submissions {
		if submission.ElementID == elementID {
			submissions = append(submissions, submission)
		}
	}

	return submissions, nil
}

// GetSubmissionsByStatus retrieves all submissions with a specific status.
func (t *PRTracker) GetSubmissionsByStatus(status PRStatus) ([]PRSubmission, error) {
	history, err := t.LoadHistory()
	if err != nil {
		return nil, fmt.Errorf("failed to load history: %w", err)
	}

	submissions := []PRSubmission{}
	for _, submission := range history.Submissions {
		if submission.Status == status {
			submissions = append(submissions, submission)
		}
	}

	return submissions, nil
}

// GetRecentSubmissions retrieves the N most recent submissions.
func (t *PRTracker) GetRecentSubmissions(limit int) ([]PRSubmission, error) {
	history, err := t.LoadHistory()
	if err != nil {
		return nil, fmt.Errorf("failed to load history: %w", err)
	}

	// Convert to slice for sorting
	submissions := make([]PRSubmission, 0, len(history.Submissions))
	for _, submission := range history.Submissions {
		submissions = append(submissions, submission)
	}

	// Sort by submission date (descending)
	for i := range len(submissions) - 1 {
		for j := i + 1; j < len(submissions); j++ {
			if submissions[j].SubmittedAt.After(submissions[i].SubmittedAt) {
				submissions[i], submissions[j] = submissions[j], submissions[i]
			}
		}
	}

	// Return limited results
	if limit > 0 && limit < len(submissions) {
		return submissions[:limit], nil
	}

	return submissions, nil
}

// GetStats returns current PR statistics.
func (t *PRTracker) GetStats() (*PRStats, error) {
	history, err := t.LoadHistory()
	if err != nil {
		return nil, fmt.Errorf("failed to load history: %w", err)
	}

	return &history.Stats, nil
}

// AddReviewComment increments the review comment count for a submission.
func (t *PRTracker) AddReviewComment(prID string) error {
	history, err := t.LoadHistory()
	if err != nil {
		return fmt.Errorf("failed to load history: %w", err)
	}

	submission, exists := history.Submissions[prID]
	if !exists {
		return fmt.Errorf("PR submission not found: %s", prID)
	}

	submission.ReviewComments++
	submission.UpdatedAt = time.Now()
	history.Submissions[prID] = submission

	return t.SaveHistory(history)
}

// UpdateNotes updates the notes for a submission.
func (t *PRTracker) UpdateNotes(prID string, notes string) error {
	history, err := t.LoadHistory()
	if err != nil {
		return fmt.Errorf("failed to load history: %w", err)
	}

	submission, exists := history.Submissions[prID]
	if !exists {
		return fmt.Errorf("PR submission not found: %s", prID)
	}

	submission.Notes = notes
	submission.UpdatedAt = time.Now()
	history.Submissions[prID] = submission

	return t.SaveHistory(history)
}

// DeleteSubmission removes a submission from history.
func (t *PRTracker) DeleteSubmission(prID string) error {
	history, err := t.LoadHistory()
	if err != nil {
		return fmt.Errorf("failed to load history: %w", err)
	}

	delete(history.Submissions, prID)

	return t.SaveHistory(history)
}

// Clear clears all PR history.
func (t *PRTracker) Clear() error {
	if _, err := os.Stat(t.historyFile); os.IsNotExist(err) {
		return nil // Nothing to clear
	}

	if err := os.Remove(t.historyFile); err != nil {
		return fmt.Errorf("failed to remove PR history file: %w", err)
	}

	return nil
}

// generateSubmissionID generates a unique ID for a PR submission.
func (t *PRTracker) generateSubmissionID(submission PRSubmission) string {
	return fmt.Sprintf("%s-%s-%d-%d",
		submission.RepositoryOwner,
		submission.RepositoryName,
		submission.PRNumber,
		time.Now().Unix(),
	)
}

// updateStats updates the statistics in the history.
func (t *PRTracker) updateStats(history *PRHistory) {
	stats := PRStats{}
	stats.TotalSubmissions = len(history.Submissions)

	for _, submission := range history.Submissions {
		switch submission.Status {
		case PRStatusPending:
			stats.Pending++
		case PRStatusMerged:
			stats.Merged++
		case PRStatusRejected:
			stats.Rejected++
		case PRStatusDraft:
			stats.Draft++
		}
	}

	history.Stats = stats
}

// GetHistoryFile returns the path to the history file.
func (t *PRTracker) GetHistoryFile() string {
	return t.historyFile
}
