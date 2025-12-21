package domain

import (
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
)

// PrivacyLevel defines the visibility level of an element.
type PrivacyLevel string

const (
	// PrivacyLevelPublic means anyone can read the element.
	PrivacyLevelPublic PrivacyLevel = "public"

	// PrivacyLevelPrivate means only the owner can read the element.
	PrivacyLevelPrivate PrivacyLevel = "private"

	// PrivacyLevelShared means owner and specific users can read the element.
	PrivacyLevelShared PrivacyLevel = "shared"
)

// Validate checks if the privacy level is valid.
func (p PrivacyLevel) Validate() error {
	switch p {
	case PrivacyLevelPublic, PrivacyLevelPrivate, PrivacyLevelShared:
		return nil
	default:
		return fmt.Errorf("invalid privacy level: %s (must be public, private, or shared)", p)
	}
}

// String returns the string representation of the privacy level.
func (p PrivacyLevel) String() string {
	return string(p)
}

// UserContext represents the current user making a request.
type UserContext struct {
	// Username is the identifier of the current user
	Username string

	// SessionID is a unique identifier for this session
	SessionID string

	// AuthenticatedAt is when the user authenticated
	AuthenticatedAt time.Time
}

// NewUserContext creates a new user context.
func NewUserContext(username string) *UserContext {
	return &UserContext{
		Username:        username,
		SessionID:       uuid.New().String(),
		AuthenticatedAt: time.Now(),
	}
}

// IsAnonymous returns true if the user is not authenticated.
func (u *UserContext) IsAnonymous() bool {
	return u.Username == "" || u.Username == "anonymous"
}

// AccessControl provides methods for checking permissions.
type AccessControl struct {
	// Can be extended with additional fields for more complex permission logic
}

// NewAccessControl creates a new AccessControl instance.
func NewAccessControl() *AccessControl {
	return &AccessControl{}
}

// CheckReadPermission checks if the current user can read an element.
func (ac *AccessControl) CheckReadPermission(
	ctx *UserContext,
	owner string,
	privacyLevel PrivacyLevel,
	sharedWith []string,
) bool {
	// Public elements are readable by everyone
	if privacyLevel == PrivacyLevelPublic {
		return true
	}

	// Anonymous users can only read public elements
	if ctx.IsAnonymous() {
		return false
	}

	// Owner can always read their own elements
	if ctx.Username == owner {
		return true
	}

	// For shared elements, check if user is in the shared list
	if privacyLevel == PrivacyLevelShared {
		for _, sharedUser := range sharedWith {
			if sharedUser == ctx.Username {
				return true
			}
		}
	}

	// Private elements are only readable by owner
	return false
}

// CheckWritePermission checks if the current user can update an element.
func (ac *AccessControl) CheckWritePermission(
	ctx *UserContext,
	owner string,
) bool {
	// Only owner can write (update) their elements
	if ctx.IsAnonymous() {
		return false
	}
	return ctx.Username == owner
}

// CheckDeletePermission checks if the current user can delete an element.
func (ac *AccessControl) CheckDeletePermission(
	ctx *UserContext,
	owner string,
) bool {
	// Only owner can delete their elements
	if ctx.IsAnonymous() {
		return false
	}
	return ctx.Username == owner
}

// CanShare checks if the current user can modify the shared list of an element.
func (ac *AccessControl) CanShare(
	ctx *UserContext,
	owner string,
	privacyLevel PrivacyLevel,
) bool {
	// Only owner can modify sharing
	if ctx.IsAnonymous() || ctx.Username != owner {
		return false
	}

	// Only shared elements have a sharing list
	return privacyLevel == PrivacyLevelShared
}

// FilterByPermissions filters a list of elements based on read permissions.
func (ac *AccessControl) FilterByPermissions(
	ctx *UserContext,
	elements []Element,
) []Element {
	filtered := make([]Element, 0, len(elements))

	for _, elem := range elements {
		// Get metadata to access author
		meta := elem.GetMetadata()

		// Determine privacy level, owner, and sharedWith based on element type
		// Currently only Persona has privacy controls
		var privacyLevel = PrivacyLevelPublic
		var owner = meta.Author
		var sharedWith []string

		// Type assertion for Persona (the only type with privacy controls for now)
		if persona, ok := elem.(*Persona); ok {
			privacyLevel = PrivacyLevel(persona.PrivacyLevel)
			if persona.Owner != "" {
				owner = persona.Owner
			}
			sharedWith = persona.SharedWith
		}

		if ac.CheckReadPermission(ctx, owner, privacyLevel, sharedWith) {
			filtered = append(filtered, elem)
		}
	}

	return filtered
}

// ValidateOwnership checks if the current user is the owner or if owner is valid.
func (ac *AccessControl) ValidateOwnership(
	ctx *UserContext,
	owner string,
) error {
	if owner == "" {
		return errors.New("owner cannot be empty")
	}

	if ctx.IsAnonymous() {
		return errors.New("anonymous users cannot own elements")
	}

	if ctx.Username != owner {
		return fmt.Errorf("user %s is not the owner (owner is %s)", ctx.Username, owner)
	}

	return nil
}

// GetDefaultPrivacyLevel returns the default privacy level for new elements.
func GetDefaultPrivacyLevel() PrivacyLevel {
	return PrivacyLevelPrivate
}

// PermissionError represents an access control error.
type PermissionError struct {
	Operation string
	Username  string
	ElementID string
	Reason    string
}

func (e *PermissionError) Error() string {
	return fmt.Sprintf(
		"permission denied: user %s cannot %s element %s: %s",
		e.Username,
		e.Operation,
		e.ElementID,
		e.Reason,
	)
}

// NewPermissionError creates a new permission error.
func NewPermissionError(operation, username, elementID, reason string) *PermissionError {
	return &PermissionError{
		Operation: operation,
		Username:  username,
		ElementID: elementID,
		Reason:    reason,
	}
}
