package domain

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewMemory(t *testing.T) {
	memory := NewMemory("test-memory", "Test Memory", "1.0", "author")

	assert.NotNil(t, memory)
	assert.Equal(t, "test-memory", memory.metadata.Name)
	assert.Equal(t, MemoryElement, memory.metadata.Type)
	assert.NotEmpty(t, memory.DateCreated)
	assert.True(t, memory.metadata.IsActive)
}

func TestMemory_Validate(t *testing.T) {
	tests := []struct {
		name    string
		setup   func() *Memory
		wantErr bool
	}{
		{
			name: "valid memory",
			setup: func() *Memory {
				mem := NewMemory("valid", "Valid Memory", "1.0", "author")
				mem.Content = "Some content"
				return mem
			},
			wantErr: false,
		},
		{
			name: "empty content",
			setup: func() *Memory {
				mem := NewMemory("invalid", "Invalid Memory", "1.0", "author")
				mem.Content = ""
				return mem
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mem := tt.setup()
			err := mem.Validate()
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestMemory_ComputeHash(t *testing.T) {
	mem := NewMemory("test", "Test", "1.0", "author")
	mem.Content = "Hello World"

	mem.ComputeHash()
	assert.NotEmpty(t, mem.ContentHash)
	assert.Len(t, mem.ContentHash, 64) // SHA-256 hex string length

	// Same content should produce same hash
	mem2 := NewMemory("test2", "Test2", "1.0", "author")
	mem2.Content = "Hello World"
	mem2.ComputeHash()
	assert.Equal(t, mem.ContentHash, mem2.ContentHash)

	// Different content should produce different hash
	mem3 := NewMemory("test3", "Test3", "1.0", "author")
	mem3.Content = "Different Content"
	mem3.ComputeHash()
	assert.NotEqual(t, mem.ContentHash, mem3.ContentHash)
}

func TestMemory_ActivateDeactivate(t *testing.T) {
	mem := NewMemory("test", "Test", "1.0", "author")

	assert.True(t, mem.IsActive())

	err := mem.Deactivate()
	assert.NoError(t, err)
	assert.False(t, mem.IsActive())

	err = mem.Activate()
	assert.NoError(t, err)
	assert.True(t, mem.IsActive())
}
