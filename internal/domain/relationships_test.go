package domain

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRelationshipMap_Add(t *testing.T) {
	tests := []struct {
		name         string
		initialMap   RelationshipMap
		elementID    string
		relType      RelationshipType
		wantLen      int
		wantContains RelationshipType
	}{
		{
			name:         "add first relationship",
			initialMap:   make(RelationshipMap),
			elementID:    "elem-001",
			relType:      RelationshipRelatedTo,
			wantLen:      1,
			wantContains: RelationshipRelatedTo,
		},
		{
			name: "add second relationship",
			initialMap: RelationshipMap{
				"elem-001": {RelationshipRelatedTo},
			},
			elementID:    "elem-001",
			relType:      RelationshipDependsOn,
			wantLen:      2,
			wantContains: RelationshipDependsOn,
		},
		{
			name: "add duplicate relationship",
			initialMap: RelationshipMap{
				"elem-001": {RelationshipRelatedTo},
			},
			elementID:    "elem-001",
			relType:      RelationshipRelatedTo,
			wantLen:      1,
			wantContains: RelationshipRelatedTo,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rm := tt.initialMap
			rm.Add(tt.elementID, tt.relType)

			rels := rm.Get(tt.elementID)
			assert.Len(t, rels, tt.wantLen)

			found := false
			for _, rel := range rels {
				if rel == tt.wantContains {
					found = true
					break
				}
			}
			assert.True(t, found, "relationship type not found")
		})
	}
}

func TestRelationshipMap_Get(t *testing.T) {
	rm := RelationshipMap{
		"elem-001": {RelationshipRelatedTo, RelationshipDependsOn},
		"elem-002": {RelationshipUses},
	}

	tests := []struct {
		name      string
		elementID string
		wantLen   int
	}{
		{
			name:      "get existing relationships",
			elementID: "elem-001",
			wantLen:   2,
		},
		{
			name:      "get single relationship",
			elementID: "elem-002",
			wantLen:   1,
		},
		{
			name:      "get non-existing",
			elementID: "elem-999",
			wantLen:   0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rels := rm.Get(tt.elementID)
			assert.Len(t, rels, tt.wantLen)
		})
	}
}

func TestRelationshipMap_Has(t *testing.T) {
	rm := RelationshipMap{
		"elem-001": {RelationshipRelatedTo},
		"elem-002": {},
	}

	tests := []struct {
		name      string
		elementID string
		want      bool
	}{
		{
			name:      "has relationships",
			elementID: "elem-001",
			want:      true,
		},
		{
			name:      "empty relationships",
			elementID: "elem-002",
			want:      false,
		},
		{
			name:      "non-existing element",
			elementID: "elem-999",
			want:      false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := rm.Has(tt.elementID)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestRelationshipError(t *testing.T) {
	baseErr := assert.AnError
	relErr := NewRelationshipError("elem-001", RelationshipDependsOn, baseErr)

	assert.Contains(t, relErr.Error(), "elem-001")
	assert.Contains(t, relErr.Error(), string(RelationshipDependsOn))
	assert.ErrorIs(t, relErr, baseErr)
}

func TestRelationshipTypes(t *testing.T) {
	// Validate all relationship type constants
	types := []RelationshipType{
		RelationshipRelatedTo,
		RelationshipDependsOn,
		RelationshipUses,
		RelationshipProduces,
		RelationshipMemberOf,
		RelationshipOwnedBy,
	}

	for _, relType := range types {
		assert.NotEmpty(t, relType, "relationship type should not be empty")
	}
}
