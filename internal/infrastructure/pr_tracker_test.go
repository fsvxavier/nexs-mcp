package infrastructure

import (
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/fsvxavier/nexs-mcp/internal/domain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewPRTracker(t *testing.T) {
	tempDir := t.TempDir()

	tracker := NewPRTracker(tempDir)
	assert.NotNil(t, tracker)
	assert.Contains(t, tracker.GetHistoryFile(), tempDir)
}

func TestNewPRTracker_DefaultDir(t *testing.T) {
	tracker := NewPRTracker("")
	assert.NotNil(t, tracker)
	// Should use home directory or current directory
	assert.NotEmpty(t, tracker.GetHistoryFile())
}

func TestPRTracker_LoadHistory_NewFile(t *testing.T) {
	tempDir := t.TempDir()

	tracker := NewPRTracker(tempDir)
	history, err := tracker.LoadHistory()
	require.NoError(t, err)

	assert.NotNil(t, history)
	assert.Equal(t, "1.0.0", history.Version)
	assert.Empty(t, history.Submissions)
	assert.Equal(t, 0, history.Stats.TotalSubmissions)
}

func TestPRTracker_SaveAndLoad(t *testing.T) {
	tempDir := t.TempDir()

	tracker := NewPRTracker(tempDir)

	// Create history with a submission
	history := &PRHistory{
		Version:     "1.0.0",
		Submissions: make(map[string]PRSubmission),
		LastUpdated: time.Now(),
	}

	submission := PRSubmission{
		ID:              "test-1",
		ElementID:       "agent1",
		ElementType:     domain.AgentElement,
		ElementName:     "Test Agent",
		ElementVersion:  "1.0.0",
		RepositoryOwner: "user",
		RepositoryName:  "repo",
		PRNumber:        123,
		PRTitle:         "Add Test Agent",
		PRURL:           "https://github.com/user/repo/pull/123",
		Status:          PRStatusPending,
		SubmittedAt:     time.Now(),
		UpdatedAt:       time.Now(),
		SubmittedBy:     "testuser",
	}

	history.Submissions[submission.ID] = submission

	// Save history
	err := tracker.SaveHistory(history)
	require.NoError(t, err)

	// Load history
	loadedHistory, err := tracker.LoadHistory()
	require.NoError(t, err)

	assert.Equal(t, history.Version, loadedHistory.Version)
	assert.Len(t, loadedHistory.Submissions, 1)
	assert.Contains(t, loadedHistory.Submissions, "test-1")
	assert.Equal(t, 1, loadedHistory.Stats.TotalSubmissions)
	assert.Equal(t, 1, loadedHistory.Stats.Pending)
}

func TestPRTracker_TrackSubmission(t *testing.T) {
	tempDir := t.TempDir()

	tracker := NewPRTracker(tempDir)

	submission := PRSubmission{
		ElementID:       "agent1",
		ElementType:     domain.AgentElement,
		ElementName:     "Test Agent",
		ElementVersion:  "1.0.0",
		RepositoryOwner: "user",
		RepositoryName:  "repo",
		PRNumber:        123,
		PRTitle:         "Add Test Agent",
		PRURL:           "https://github.com/user/repo/pull/123",
		Status:          PRStatusPending,
		SubmittedBy:     "testuser",
	}

	// Track submission
	err := tracker.TrackSubmission(submission)
	require.NoError(t, err)

	// Load and verify
	history, err := tracker.LoadHistory()
	require.NoError(t, err)
	assert.Equal(t, 1, len(history.Submissions))
}

func TestPRTracker_UpdateSubmissionStatus(t *testing.T) {
	tempDir := t.TempDir()

	tracker := NewPRTracker(tempDir)

	// Create initial submission
	submission := PRSubmission{
		ID:              "test-1",
		ElementID:       "agent1",
		ElementType:     domain.AgentElement,
		RepositoryOwner: "user",
		RepositoryName:  "repo",
		PRNumber:        123,
		Status:          PRStatusPending,
		SubmittedBy:     "testuser",
	}

	err := tracker.TrackSubmission(submission)
	require.NoError(t, err)

	// Update status to merged
	err = tracker.UpdateSubmissionStatus("test-1", PRStatusMerged)
	require.NoError(t, err)

	// Verify update
	history, err := tracker.LoadHistory()
	require.NoError(t, err)

	updated := history.Submissions["test-1"]
	assert.Equal(t, PRStatusMerged, updated.Status)
	assert.NotNil(t, updated.MergedAt)
}

func TestPRTracker_UpdateSubmissionStatus_NotFound(t *testing.T) {
	tempDir := t.TempDir()

	tracker := NewPRTracker(tempDir)

	err := tracker.UpdateSubmissionStatus("nonexistent", PRStatusMerged)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
}

func TestPRTracker_GetSubmissionByPRNumber(t *testing.T) {
	tempDir := t.TempDir()

	tracker := NewPRTracker(tempDir)

	submission := PRSubmission{
		ID:              "test-1",
		ElementID:       "agent1",
		RepositoryOwner: "user",
		RepositoryName:  "repo",
		PRNumber:        123,
		Status:          PRStatusPending,
	}

	err := tracker.TrackSubmission(submission)
	require.NoError(t, err)

	// Get by PR number
	found, err := tracker.GetSubmissionByPRNumber("user", "repo", 123)
	require.NoError(t, err)
	assert.NotNil(t, found)
	assert.Equal(t, "agent1", found.ElementID)

	// Not found
	_, err = tracker.GetSubmissionByPRNumber("user", "repo", 999)
	assert.Error(t, err)
}

func TestPRTracker_GetSubmissionsByElement(t *testing.T) {
	tempDir := t.TempDir()

	tracker := NewPRTracker(tempDir)

	// Track multiple submissions for same element
	for i := 1; i <= 3; i++ {
		submission := PRSubmission{
			ID:              fmt.Sprintf("test-%d", i),
			ElementID:       "agent1",
			RepositoryOwner: "user",
			RepositoryName:  "repo",
			PRNumber:        100 + i,
			Status:          PRStatusPending,
		}
		err := tracker.TrackSubmission(submission)
		require.NoError(t, err)
	}

	// Get submissions by element
	submissions, err := tracker.GetSubmissionsByElement("agent1")
	require.NoError(t, err)
	assert.Len(t, submissions, 3)

	// Element not found
	submissions, err = tracker.GetSubmissionsByElement("agent999")
	require.NoError(t, err)
	assert.Empty(t, submissions)
}

func TestPRTracker_GetSubmissionsByStatus(t *testing.T) {
	tempDir := t.TempDir()

	tracker := NewPRTracker(tempDir)

	// Track submissions with different statuses
	statuses := []PRStatus{PRStatusPending, PRStatusMerged, PRStatusPending, PRStatusRejected}
	for i, status := range statuses {
		submission := PRSubmission{
			ID:              fmt.Sprintf("test-%d", i),
			ElementID:       fmt.Sprintf("agent%d", i),
			RepositoryOwner: "user",
			RepositoryName:  "repo",
			PRNumber:        100 + i,
			Status:          status,
		}
		err := tracker.TrackSubmission(submission)
		require.NoError(t, err)
	}

	// Get pending submissions
	pending, err := tracker.GetSubmissionsByStatus(PRStatusPending)
	require.NoError(t, err)
	assert.Len(t, pending, 2)

	// Get merged submissions
	merged, err := tracker.GetSubmissionsByStatus(PRStatusMerged)
	require.NoError(t, err)
	assert.Len(t, merged, 1)
}

func TestPRTracker_GetRecentSubmissions(t *testing.T) {
	tempDir := t.TempDir()

	tracker := NewPRTracker(tempDir)

	// Track multiple submissions with different timestamps
	for i := 1; i <= 5; i++ {
		submission := PRSubmission{
			ID:              fmt.Sprintf("test-%d", i),
			ElementID:       fmt.Sprintf("agent%d", i),
			RepositoryOwner: "user",
			RepositoryName:  "repo",
			PRNumber:        100 + i,
			Status:          PRStatusPending,
			SubmittedAt:     time.Now().Add(-time.Duration(5-i) * time.Hour),
		}
		err := tracker.TrackSubmission(submission)
		require.NoError(t, err)
	}

	// Get recent submissions (limit 3)
	recent, err := tracker.GetRecentSubmissions(3)
	require.NoError(t, err)
	assert.Len(t, recent, 3)

	// Verify order (most recent first)
	assert.Equal(t, "test-5", recent[0].ID)
	assert.Equal(t, "test-4", recent[1].ID)
	assert.Equal(t, "test-3", recent[2].ID)
}

func TestPRTracker_GetStats(t *testing.T) {
	tempDir := t.TempDir()

	tracker := NewPRTracker(tempDir)

	// Track submissions with different statuses
	statuses := []PRStatus{
		PRStatusPending, PRStatusPending,
		PRStatusMerged, PRStatusMerged, PRStatusMerged,
		PRStatusRejected,
		PRStatusDraft,
	}

	for i, status := range statuses {
		submission := PRSubmission{
			ID:              fmt.Sprintf("test-%d", i),
			ElementID:       fmt.Sprintf("agent%d", i),
			RepositoryOwner: "user",
			RepositoryName:  "repo",
			PRNumber:        100 + i,
			Status:          status,
		}
		err := tracker.TrackSubmission(submission)
		require.NoError(t, err)
	}

	// Get stats
	stats, err := tracker.GetStats()
	require.NoError(t, err)
	assert.Equal(t, 7, stats.TotalSubmissions)
	assert.Equal(t, 2, stats.Pending)
	assert.Equal(t, 3, stats.Merged)
	assert.Equal(t, 1, stats.Rejected)
	assert.Equal(t, 1, stats.Draft)
}

func TestPRTracker_AddReviewComment(t *testing.T) {
	tempDir := t.TempDir()

	tracker := NewPRTracker(tempDir)

	submission := PRSubmission{
		ID:              "test-1",
		ElementID:       "agent1",
		RepositoryOwner: "user",
		RepositoryName:  "repo",
		PRNumber:        123,
		Status:          PRStatusPending,
		ReviewComments:  0,
	}

	err := tracker.TrackSubmission(submission)
	require.NoError(t, err)

	// Add review comments
	err = tracker.AddReviewComment("test-1")
	require.NoError(t, err)
	err = tracker.AddReviewComment("test-1")
	require.NoError(t, err)
	// Verify
	history, err := tracker.LoadHistory()
	require.NoError(t, err)
	assert.Equal(t, 2, history.Submissions["test-1"].ReviewComments)
}

func TestPRTracker_UpdateNotes(t *testing.T) {
	tempDir := t.TempDir()

	tracker := NewPRTracker(tempDir)

	submission := PRSubmission{
		ID:              "test-1",
		ElementID:       "agent1",
		RepositoryOwner: "user",
		RepositoryName:  "repo",
		PRNumber:        123,
		Status:          PRStatusPending,
	}

	err := tracker.TrackSubmission(submission)
	require.NoError(t, err)

	// Update notes
	err = tracker.UpdateNotes("test-1", "This is a test note")
	require.NoError(t, err)

	// Verify
	history, err := tracker.LoadHistory()
	require.NoError(t, err)
	assert.Equal(t, "This is a test note", history.Submissions["test-1"].Notes)
}

func TestPRTracker_DeleteSubmission(t *testing.T) {
	tempDir := t.TempDir()

	tracker := NewPRTracker(tempDir)

	submission := PRSubmission{
		ID:              "test-1",
		ElementID:       "agent1",
		RepositoryOwner: "user",
		RepositoryName:  "repo",
		PRNumber:        123,
		Status:          PRStatusPending,
	}

	err := tracker.TrackSubmission(submission)
	require.NoError(t, err)

	// Delete submission
	err = tracker.DeleteSubmission("test-1")
	require.NoError(t, err)

	// Verify deletion
	history, err := tracker.LoadHistory()
	require.NoError(t, err)
	assert.Empty(t, history.Submissions)
}

func TestPRTracker_Clear(t *testing.T) {
	tempDir := t.TempDir()

	tracker := NewPRTracker(tempDir)

	// Track a submission
	submission := PRSubmission{
		ID:              "test-1",
		ElementID:       "agent1",
		RepositoryOwner: "user",
		RepositoryName:  "repo",
		PRNumber:        123,
		Status:          PRStatusPending,
	}

	err := tracker.TrackSubmission(submission)
	require.NoError(t, err)

	// Clear history
	err = tracker.Clear()
	require.NoError(t, err)
	// Verify file is deleted
	_, err = os.Stat(tracker.GetHistoryFile())
	assert.True(t, os.IsNotExist(err))
}
