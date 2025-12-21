package domain

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"time"
)

// Memory represents persistent context storage.
type Memory struct {
	metadata    ElementMetadata
	Content     string            `json:"content"                validate:"required"           yaml:"content"`
	DateCreated string            `json:"date_created"           yaml:"date_created"` // YYYY-MM-DD
	ContentHash string            `json:"content_hash"           yaml:"content_hash"`
	SearchIndex []string          `json:"search_index,omitempty" yaml:"search_index,omitempty"`
	Metadata    map[string]string `json:"metadata,omitempty"     yaml:"metadata,omitempty"`
}

// NewMemory creates a new Memory element.
func NewMemory(name, description, version, author string) *Memory {
	now := time.Now()
	return &Memory{
		metadata: ElementMetadata{
			ID:          GenerateElementID(MemoryElement, name),
			Type:        MemoryElement,
			Name:        name,
			Description: description,
			Version:     version,
			Author:      author,
			Tags:        []string{},
			IsActive:    true,
			CreatedAt:   now,
			UpdatedAt:   now,
		},
		DateCreated: now.Format("2006-01-02"),
		Metadata:    make(map[string]string),
		SearchIndex: []string{},
	}
}

func (m *Memory) GetMetadata() ElementMetadata { return m.metadata }
func (m *Memory) GetType() ElementType         { return m.metadata.Type }
func (m *Memory) GetID() string                { return m.metadata.ID }
func (m *Memory) IsActive() bool               { return m.metadata.IsActive }

func (m *Memory) Activate() error {
	m.metadata.IsActive = true
	m.metadata.UpdatedAt = time.Now()
	return nil
}

func (m *Memory) Deactivate() error {
	m.metadata.IsActive = false
	m.metadata.UpdatedAt = time.Now()
	return nil
}

func (m *Memory) Validate() error {
	if err := m.metadata.Validate(); err != nil {
		return fmt.Errorf("metadata validation failed: %w", err)
	}
	if m.Content == "" {
		return errors.New("content is required")
	}
	return nil
}

func (m *Memory) SetMetadata(metadata ElementMetadata) {
	m.metadata = metadata
	m.metadata.UpdatedAt = time.Now()
}

// ComputeHash computes SHA-256 hash of content for deduplication.
func (m *Memory) ComputeHash() {
	hash := sha256.Sum256([]byte(m.Content))
	m.ContentHash = hex.EncodeToString(hash[:])
}
