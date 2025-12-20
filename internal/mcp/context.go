package mcp

import (
	"strings"

	"github.com/fsvxavier/nexs-mcp/internal/domain"
)

// GetUserContext creates a UserContext from a username string.
// This is used by MCP handlers to create user context from input fields.
//
// If username is empty or whitespace, returns anonymous user context.
func GetUserContext(username string) *domain.UserContext {
	username = strings.TrimSpace(username)
	return domain.NewUserContext(username)
}

// GetUserContextFromAuthor creates a UserContext using the author field.
// This is a convenience function for handlers that use "author" instead of "user".
func GetUserContextFromAuthor(author string) *domain.UserContext {
	return GetUserContext(author)
}
