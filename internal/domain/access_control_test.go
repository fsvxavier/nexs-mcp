package domain

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPrivacyLevel_Validation(t *testing.T) {
	tests := []struct {
		name    string
		level   PrivacyLevel
		isValid bool
	}{
		{"Public is valid", PrivacyLevelPublic, true},
		{"Private is valid", PrivacyLevelPrivate, true},
		{"Shared is valid", PrivacyLevelShared, true},
		{"Empty string is invalid", PrivacyLevel(""), false},
		{"Invalid level", PrivacyLevel("invalid"), false},
		{"Mixed case is invalid", PrivacyLevel("Public"), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.level.Validate()
			if tt.isValid {
				assert.NoError(t, err)
			} else {
				assert.Error(t, err)
			}
		})
	}
}

func TestPrivacyLevel_String(t *testing.T) {
	tests := []struct {
		name     string
		level    PrivacyLevel
		expected string
	}{
		{"Public", PrivacyLevelPublic, "public"},
		{"Private", PrivacyLevelPrivate, "private"},
		{"Shared", PrivacyLevelShared, "shared"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, tt.level.String())
		})
	}
}

func TestUserContext_Creation(t *testing.T) {
	ctx := NewUserContext("alice")
	assert.NotNil(t, ctx)
	assert.Equal(t, "alice", ctx.Username)
	assert.NotEmpty(t, ctx.SessionID)
}

func TestUserContext_IsAnonymous(t *testing.T) {
	tests := []struct {
		name      string
		username  string
		anonymous bool
	}{
		{"User alice is not anonymous", "alice", false},
		{"User bob is not anonymous", "bob", false},
		{"Empty username is anonymous", "", true},
		{"Anonymous user", "anonymous", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := NewUserContext(tt.username)
			assert.Equal(t, tt.anonymous, ctx.IsAnonymous())
		})
	}
}

func TestAccessControl_CheckReadPermission(t *testing.T) {
	tests := []struct {
		name          string
		privacyLevel  PrivacyLevel
		owner         string
		currentUser   string
		sharedWith    []string
		shouldAllowed bool
		description   string
	}{
		{
			name:          "Public element - anyone can read",
			privacyLevel:  PrivacyLevelPublic,
			owner:         "alice",
			currentUser:   "bob",
			sharedWith:    nil,
			shouldAllowed: true,
			description:   "Public elements are readable by everyone",
		},
		{
			name:          "Private element - owner can read",
			privacyLevel:  PrivacyLevelPrivate,
			owner:         "alice",
			currentUser:   "alice",
			sharedWith:    nil,
			shouldAllowed: true,
			description:   "Owner can read their private elements",
		},
		{
			name:          "Private element - others cannot read",
			privacyLevel:  PrivacyLevelPrivate,
			owner:         "alice",
			currentUser:   "bob",
			sharedWith:    nil,
			shouldAllowed: false,
			description:   "Non-owners cannot read private elements",
		},
		{
			name:          "Shared element - owner can read",
			privacyLevel:  PrivacyLevelShared,
			owner:         "alice",
			currentUser:   "alice",
			sharedWith:    []string{"bob"},
			shouldAllowed: true,
			description:   "Owner can read their shared elements",
		},
		{
			name:          "Shared element - shared user can read",
			privacyLevel:  PrivacyLevelShared,
			owner:         "alice",
			currentUser:   "bob",
			sharedWith:    []string{"bob", "charlie"},
			shouldAllowed: true,
			description:   "Users in shared list can read",
		},
		{
			name:          "Shared element - non-shared user cannot read",
			privacyLevel:  PrivacyLevelShared,
			owner:         "alice",
			currentUser:   "dave",
			sharedWith:    []string{"bob", "charlie"},
			shouldAllowed: false,
			description:   "Users not in shared list cannot read",
		},
		{
			name:          "Anonymous user cannot read private",
			privacyLevel:  PrivacyLevelPrivate,
			owner:         "alice",
			currentUser:   "",
			sharedWith:    nil,
			shouldAllowed: false,
			description:   "Anonymous users cannot read private elements",
		},
		{
			name:          "Anonymous user can read public",
			privacyLevel:  PrivacyLevelPublic,
			owner:         "alice",
			currentUser:   "",
			sharedWith:    nil,
			shouldAllowed: true,
			description:   "Anonymous users can read public elements",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ac := NewAccessControl()
			ctx := NewUserContext(tt.currentUser)

			allowed := ac.CheckReadPermission(ctx, tt.owner, tt.privacyLevel, tt.sharedWith)
			assert.Equal(t, tt.shouldAllowed, allowed, tt.description)
		})
	}
}

func TestAccessControl_CheckWritePermission(t *testing.T) {
	tests := []struct {
		name          string
		owner         string
		currentUser   string
		shouldAllowed bool
		description   string
	}{
		{
			name:          "Owner can write",
			owner:         "alice",
			currentUser:   "alice",
			shouldAllowed: true,
			description:   "Owner has write permission",
		},
		{
			name:          "Non-owner cannot write",
			owner:         "alice",
			currentUser:   "bob",
			shouldAllowed: false,
			description:   "Non-owner does not have write permission",
		},
		{
			name:          "Anonymous cannot write",
			owner:         "alice",
			currentUser:   "",
			shouldAllowed: false,
			description:   "Anonymous users cannot write",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ac := NewAccessControl()
			ctx := NewUserContext(tt.currentUser)

			allowed := ac.CheckWritePermission(ctx, tt.owner)
			assert.Equal(t, tt.shouldAllowed, allowed, tt.description)
		})
	}
}

func TestAccessControl_CheckDeletePermission(t *testing.T) {
	tests := []struct {
		name          string
		owner         string
		currentUser   string
		shouldAllowed bool
		description   string
	}{
		{
			name:          "Owner can delete",
			owner:         "alice",
			currentUser:   "alice",
			shouldAllowed: true,
			description:   "Owner has delete permission",
		},
		{
			name:          "Non-owner cannot delete",
			owner:         "alice",
			currentUser:   "bob",
			shouldAllowed: false,
			description:   "Non-owner does not have delete permission",
		},
		{
			name:          "Anonymous cannot delete",
			owner:         "alice",
			currentUser:   "",
			shouldAllowed: false,
			description:   "Anonymous users cannot delete",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ac := NewAccessControl()
			ctx := NewUserContext(tt.currentUser)

			allowed := ac.CheckDeletePermission(ctx, tt.owner)
			assert.Equal(t, tt.shouldAllowed, allowed, tt.description)
		})
	}
}

func TestAccessControl_CanShare(t *testing.T) {
	tests := []struct {
		name          string
		owner         string
		currentUser   string
		privacyLevel  PrivacyLevel
		shouldAllowed bool
		description   string
	}{
		{
			name:          "Owner can share shared element",
			owner:         "alice",
			currentUser:   "alice",
			privacyLevel:  PrivacyLevelShared,
			shouldAllowed: true,
			description:   "Owner can modify sharing",
		},
		{
			name:          "Non-owner cannot share",
			owner:         "alice",
			currentUser:   "bob",
			privacyLevel:  PrivacyLevelShared,
			shouldAllowed: false,
			description:   "Non-owner cannot modify sharing",
		},
		{
			name:          "Cannot share public element",
			owner:         "alice",
			currentUser:   "alice",
			privacyLevel:  PrivacyLevelPublic,
			shouldAllowed: false,
			description:   "Public elements don't have sharing lists",
		},
		{
			name:          "Cannot share private element",
			owner:         "alice",
			currentUser:   "alice",
			privacyLevel:  PrivacyLevelPrivate,
			shouldAllowed: false,
			description:   "Private elements don't have sharing lists",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ac := NewAccessControl()
			ctx := NewUserContext(tt.currentUser)

			allowed := ac.CanShare(ctx, tt.owner, tt.privacyLevel)
			assert.Equal(t, tt.shouldAllowed, allowed, tt.description)
		})
	}
}

func TestAccessControl_FilterByPermissions(t *testing.T) {
	// Create test personas with different privacy levels and owners
	persona1 := NewPersona("Public Persona", "desc", "1.0.0", "alice")
	persona1.PrivacyLevel = PrivacyPublic
	persona1.Owner = "alice"

	persona2 := NewPersona("Alice Private", "desc", "1.0.0", "alice")
	persona2.PrivacyLevel = PrivacyPrivate
	persona2.Owner = "alice"

	persona3 := NewPersona("Bob Private", "desc", "1.0.0", "bob")
	persona3.PrivacyLevel = PrivacyPrivate
	persona3.Owner = "bob"

	persona4 := NewPersona("Shared with Bob", "desc", "1.0.0", "alice")
	persona4.PrivacyLevel = PrivacyShared
	persona4.Owner = "alice"
	persona4.SharedWith = []string{"bob"}

	personas := []Element{persona1, persona2, persona3, persona4}

	tests := []struct {
		name          string
		currentUser   string
		expectedCount int
		expectedNames []string
		description   string
	}{
		{
			name:          "Alice sees all her elements + public",
			currentUser:   "alice",
			expectedCount: 3,
			expectedNames: []string{"Public Persona", "Alice Private", "Shared with Bob"},
			description:   "Alice sees public + her private + her shared",
		},
		{
			name:          "Bob sees public + his private + shared with him",
			currentUser:   "bob",
			expectedCount: 3,
			expectedNames: []string{"Public Persona", "Bob Private", "Shared with Bob"},
			description:   "Bob sees public + his elements + elements shared with him",
		},
		{
			name:          "Charlie sees only public",
			currentUser:   "charlie",
			expectedCount: 1,
			expectedNames: []string{"Public Persona"},
			description:   "Charlie only sees public elements",
		},
		{
			name:          "Anonymous sees only public",
			currentUser:   "",
			expectedCount: 1,
			expectedNames: []string{"Public Persona"},
			description:   "Anonymous users only see public elements",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ac := NewAccessControl()
			ctx := NewUserContext(tt.currentUser)

			filtered := ac.FilterByPermissions(ctx, personas)
			assert.Equal(t, tt.expectedCount, len(filtered), tt.description)

			// Verify names match
			names := make([]string, len(filtered))
			for i, elem := range filtered {
				names[i] = elem.GetMetadata().Name
			}
			assert.ElementsMatch(t, tt.expectedNames, names)
		})
	}
}

func TestAccessControl_ValidateOwnership(t *testing.T) {
	tests := []struct {
		name        string
		owner       string
		currentUser string
		shouldError bool
	}{
		{
			name:        "Owner matches current user",
			owner:       "alice",
			currentUser: "alice",
			shouldError: false,
		},
		{
			name:        "Owner does not match current user",
			owner:       "alice",
			currentUser: "bob",
			shouldError: true,
		},
		{
			name:        "Anonymous user cannot own",
			owner:       "alice",
			currentUser: "",
			shouldError: true,
		},
		{
			name:        "Empty owner is invalid",
			owner:       "",
			currentUser: "alice",
			shouldError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ac := NewAccessControl()
			ctx := NewUserContext(tt.currentUser)

			err := ac.ValidateOwnership(ctx, tt.owner)
			if tt.shouldError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
