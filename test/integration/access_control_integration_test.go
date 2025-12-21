package integration_test

import (
	"testing"

	"github.com/fsvxavier/nexs-mcp/internal/domain"
	"github.com/fsvxavier/nexs-mcp/internal/infrastructure"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestAccessControl_FilterPersonas tests filtering personas by permissions.
func TestAccessControl_FilterPersonas(t *testing.T) {
	// Setup repository
	repo := infrastructure.NewInMemoryElementRepository()
	ac := domain.NewAccessControl()

	// Create personas with different privacy levels
	alicePrivate := domain.NewPersona("Alice Private", "Private persona", "1.0.0", "alice")
	alicePrivate.Owner = "alice"
	alicePrivate.PrivacyLevel = domain.PrivacyPrivate
	alicePrivate.SharedWith = []string{}
	require.NoError(t, repo.Create(alicePrivate))

	bobPublic := domain.NewPersona("Bob Public", "Public persona", "1.0.0", "bob")
	bobPublic.Owner = "bob"
	bobPublic.PrivacyLevel = domain.PrivacyPublic
	bobPublic.SharedWith = []string{}
	require.NoError(t, repo.Create(bobPublic))

	charlieShared := domain.NewPersona("Charlie Shared", "Shared persona", "1.0.0", "charlie")
	charlieShared.Owner = "charlie"
	charlieShared.PrivacyLevel = domain.PrivacyShared
	charlieShared.SharedWith = []string{"alice", "bob"}
	require.NoError(t, repo.Create(charlieShared))

	// Get all personas from repo
	allPersonas, err := repo.List(domain.ElementFilter{})
	require.NoError(t, err)
	require.Len(t, allPersonas, 3)

	tests := []struct {
		name          string
		username      string
		expectedCount int
		expectedNames []string
	}{
		{
			name:          "Alice sees her private + Bob's public + Charlie's shared",
			username:      "alice",
			expectedCount: 3,
			expectedNames: []string{"Alice Private", "Bob Public", "Charlie Shared"},
		},
		{
			name:          "Bob sees his public + Charlie's shared",
			username:      "bob",
			expectedCount: 2,
			expectedNames: []string{"Bob Public", "Charlie Shared"},
		},
		{
			name:          "Dave (not in shared list) sees only Bob's public",
			username:      "dave",
			expectedCount: 1,
			expectedNames: []string{"Bob Public"},
		},
		{
			name:          "Anonymous user sees only Bob's public",
			username:      "",
			expectedCount: 1,
			expectedNames: []string{"Bob Public"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			userCtx := domain.NewUserContext(tt.username)
			filtered := ac.FilterByPermissions(userCtx, allPersonas)

			assert.Len(t, filtered, tt.expectedCount)

			// Check that we got the expected personas
			foundNames := make([]string, len(filtered))
			for i, elem := range filtered {
				foundNames[i] = elem.GetMetadata().Name
			}

			for _, expectedName := range tt.expectedNames {
				assert.Contains(t, foundNames, expectedName)
			}
		})
	}
}

// TestAccessControl_ReadPermissions tests read permission checking.
func TestAccessControl_ReadPermissions(t *testing.T) {
	ac := domain.NewAccessControl()

	tests := []struct {
		name         string
		username     string
		owner        string
		privacyLevel domain.PrivacyLevel
		sharedWith   []string
		canRead      bool
	}{
		{
			name:         "Owner can read own private element",
			username:     "alice",
			owner:        "alice",
			privacyLevel: domain.PrivacyLevelPrivate,
			sharedWith:   []string{},
			canRead:      true,
		},
		{
			name:         "Non-owner cannot read private element",
			username:     "bob",
			owner:        "alice",
			privacyLevel: domain.PrivacyLevelPrivate,
			sharedWith:   []string{},
			canRead:      false,
		},
		{
			name:         "Anyone can read public element",
			username:     "bob",
			owner:        "alice",
			privacyLevel: domain.PrivacyLevelPublic,
			sharedWith:   []string{},
			canRead:      true,
		},
		{
			name:         "Anonymous can read public element",
			username:     "",
			owner:        "alice",
			privacyLevel: domain.PrivacyLevelPublic,
			sharedWith:   []string{},
			canRead:      true,
		},
		{
			name:         "Shared user can read shared element",
			username:     "bob",
			owner:        "alice",
			privacyLevel: domain.PrivacyLevelShared,
			sharedWith:   []string{"bob", "charlie"},
			canRead:      true,
		},
		{
			name:         "Non-shared user cannot read shared element",
			username:     "dave",
			owner:        "alice",
			privacyLevel: domain.PrivacyLevelShared,
			sharedWith:   []string{"bob", "charlie"},
			canRead:      false,
		},
		{
			name:         "Anonymous cannot read shared element",
			username:     "",
			owner:        "alice",
			privacyLevel: domain.PrivacyLevelShared,
			sharedWith:   []string{"bob"},
			canRead:      false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			userCtx := domain.NewUserContext(tt.username)
			canRead := ac.CheckReadPermission(userCtx, tt.owner, tt.privacyLevel, tt.sharedWith)
			assert.Equal(t, tt.canRead, canRead)
		})
	}
}

// TestAccessControl_WriteAndDeletePermissions tests write and delete permissions.
func TestAccessControl_WriteAndDeletePermissions(t *testing.T) {
	ac := domain.NewAccessControl()

	tests := []struct {
		name      string
		username  string
		owner     string
		canWrite  bool
		canDelete bool
	}{
		{
			name:      "Owner can write and delete",
			username:  "alice",
			owner:     "alice",
			canWrite:  true,
			canDelete: true,
		},
		{
			name:      "Non-owner cannot write or delete",
			username:  "bob",
			owner:     "alice",
			canWrite:  false,
			canDelete: false,
		},
		{
			name:      "Anonymous cannot write or delete",
			username:  "",
			owner:     "alice",
			canWrite:  false,
			canDelete: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			userCtx := domain.NewUserContext(tt.username)

			canWrite := ac.CheckWritePermission(userCtx, tt.owner)
			assert.Equal(t, tt.canWrite, canWrite, "write permission mismatch")

			canDelete := ac.CheckDeletePermission(userCtx, tt.owner)
			assert.Equal(t, tt.canDelete, canDelete, "delete permission mismatch")
		})
	}
}

// TestAccessControl_SharePermissions tests share permission checking.
func TestAccessControl_SharePermissions(t *testing.T) {
	ac := domain.NewAccessControl()

	tests := []struct {
		name         string
		username     string
		owner        string
		privacyLevel domain.PrivacyLevel
		canShare     bool
	}{
		{
			name:         "Owner can share shared element",
			username:     "alice",
			owner:        "alice",
			privacyLevel: domain.PrivacyLevelShared,
			canShare:     true,
		},
		{
			name:         "Owner cannot share private element",
			username:     "alice",
			owner:        "alice",
			privacyLevel: domain.PrivacyLevelPrivate,
			canShare:     false,
		},
		{
			name:         "Owner cannot share public element",
			username:     "alice",
			owner:        "alice",
			privacyLevel: domain.PrivacyLevelPublic,
			canShare:     false,
		},
		{
			name:         "Non-owner cannot share",
			username:     "bob",
			owner:        "alice",
			privacyLevel: domain.PrivacyLevelShared,
			canShare:     false,
		},
		{
			name:         "Anonymous cannot share",
			username:     "",
			owner:        "alice",
			privacyLevel: domain.PrivacyLevelShared,
			canShare:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			userCtx := domain.NewUserContext(tt.username)
			canShare := ac.CanShare(userCtx, tt.owner, tt.privacyLevel)
			assert.Equal(t, tt.canShare, canShare)
		})
	}
}

// TestAccessControl_OwnershipValidation tests ownership validation.
func TestAccessControl_OwnershipValidation(t *testing.T) {
	ac := domain.NewAccessControl()

	tests := []struct {
		name        string
		username    string
		owner       string
		expectError bool
	}{
		{
			name:        "User is owner",
			username:    "alice",
			owner:       "alice",
			expectError: false,
		},
		{
			name:        "User is not owner",
			username:    "bob",
			owner:       "alice",
			expectError: true,
		},
		{
			name:        "Anonymous is not owner",
			username:    "",
			owner:       "alice",
			expectError: true,
		},
		{
			name:        "Empty owner with empty user is not owner",
			username:    "",
			owner:       "",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			userCtx := domain.NewUserContext(tt.username)
			err := ac.ValidateOwnership(userCtx, tt.owner)
			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
