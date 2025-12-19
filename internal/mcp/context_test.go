package mcp

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetUserContext(t *testing.T) {
	tests := []struct {
		name         string
		username     string
		expectedUser string
		expectedAnon bool
	}{
		{
			name:         "empty username returns anonymous",
			username:     "",
			expectedUser: "",
			expectedAnon: true,
		},
		{
			name:         "whitespace username returns anonymous",
			username:     "   ",
			expectedUser: "",
			expectedAnon: true,
		},
		{
			name:         "valid username alice",
			username:     "alice",
			expectedUser: "alice",
			expectedAnon: false,
		},
		{
			name:         "valid username bob",
			username:     "bob",
			expectedUser: "bob",
			expectedAnon: false,
		},
		{
			name:         "username with whitespace is trimmed",
			username:     "  charlie  ",
			expectedUser: "charlie",
			expectedAnon: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := GetUserContext(tt.username)
			assert.Equal(t, tt.expectedUser, ctx.Username)
			assert.Equal(t, tt.expectedAnon, ctx.IsAnonymous())
			assert.NotEmpty(t, ctx.SessionID)
		})
	}
}

func TestGetUserContextFromAuthor(t *testing.T) {
	tests := []struct {
		name         string
		author       string
		expectedUser string
		expectedAnon bool
	}{
		{
			name:         "empty author returns anonymous",
			author:       "",
			expectedUser: "",
			expectedAnon: true,
		},
		{
			name:         "valid author",
			author:       "alice",
			expectedUser: "alice",
			expectedAnon: false,
		},
		{
			name:         "author with whitespace is trimmed",
			author:       "  bob  ",
			expectedUser: "bob",
			expectedAnon: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := GetUserContextFromAuthor(tt.author)
			assert.Equal(t, tt.expectedUser, ctx.Username)
			assert.Equal(t, tt.expectedAnon, ctx.IsAnonymous())
			assert.NotEmpty(t, ctx.SessionID)
		})
	}
}
