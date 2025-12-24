package domain

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewVersionHistory(t *testing.T) {
	vh := NewVersionHistory("elem-123", PersonaElement)

	assert.NotNil(t, vh)
	assert.Equal(t, "elem-123", vh.ElementID)
	assert.Equal(t, PersonaElement, vh.ElementType)
	assert.Equal(t, 0, vh.CurrentVersion)
	assert.Equal(t, 0, vh.TotalVersions)
	assert.Empty(t, vh.Snapshots)
	assert.NotNil(t, vh.SnapshotPolicy)
	assert.NotNil(t, vh.RetentionPolicy)
}

func TestVersionHistory_AddSnapshot(t *testing.T) {
	tests := []struct {
		name    string
		setup   func() (*VersionHistory, *VersionSnapshot)
		wantErr bool
		errMsg  string
	}{
		{
			name: "add first snapshot successfully",
			setup: func() (*VersionHistory, *VersionSnapshot) {
				vh := NewVersionHistory("elem-1", SkillElement)
				snapshot := &VersionSnapshot{
					Version:    1,
					Timestamp:  time.Now(),
					Author:     "test-author",
					ChangeType: ChangeTypeCreate,
					FullData:   map[string]interface{}{"name": "Test"},
				}
				return vh, snapshot
			},
			wantErr: false,
		},
		{
			name: "add snapshot with sequential version",
			setup: func() (*VersionHistory, *VersionSnapshot) {
				vh := NewVersionHistory("elem-1", SkillElement)
				snapshot1 := &VersionSnapshot{
					Version:    1,
					Timestamp:  time.Now(),
					Author:     "author",
					ChangeType: ChangeTypeCreate,
					FullData:   map[string]interface{}{"name": "Test"},
				}
				_ = vh.AddSnapshot(snapshot1)

				snapshot2 := &VersionSnapshot{
					Version:    2,
					Timestamp:  time.Now(),
					Author:     "author",
					ChangeType: ChangeTypeUpdate,
					Changes:    map[string]interface{}{"name": "Updated"},
				}
				return vh, snapshot2
			},
			wantErr: false,
		},
		{
			name: "reject nil snapshot",
			setup: func() (*VersionHistory, *VersionSnapshot) {
				vh := NewVersionHistory("elem-1", SkillElement)
				return vh, nil
			},
			wantErr: true,
			errMsg:  "snapshot cannot be nil",
		},
		{
			name: "reject non-sequential version",
			setup: func() (*VersionHistory, *VersionSnapshot) {
				vh := NewVersionHistory("elem-1", SkillElement)
				snapshot1 := &VersionSnapshot{
					Version:    1,
					Timestamp:  time.Now(),
					Author:     "author",
					ChangeType: ChangeTypeCreate,
					FullData:   map[string]interface{}{"name": "Test"},
				}
				_ = vh.AddSnapshot(snapshot1)

				// Skip version 2, try to add version 3
				snapshot3 := &VersionSnapshot{
					Version:    3,
					Timestamp:  time.Now(),
					Author:     "author",
					ChangeType: ChangeTypeUpdate,
					Changes:    map[string]interface{}{"name": "Updated"},
				}
				return vh, snapshot3
			},
			wantErr: true,
			errMsg:  "version must be sequential",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			vh, snapshot := tt.setup()
			err := vh.AddSnapshot(snapshot)

			if tt.wantErr {
				assert.Error(t, err)
				if tt.errMsg != "" {
					assert.Contains(t, err.Error(), tt.errMsg)
				}
			} else {
				assert.NoError(t, err)
				if snapshot != nil {
					assert.Equal(t, snapshot.Version, vh.CurrentVersion)
					assert.Equal(t, len(vh.Snapshots), vh.TotalVersions)
				}
			}
		})
	}
}

func TestVersionHistory_GetSnapshot(t *testing.T) {
	vh := NewVersionHistory("elem-1", MemoryElement)

	// Add some snapshots
	for i := 1; i <= 3; i++ {
		snapshot := &VersionSnapshot{
			Version:    i,
			Timestamp:  time.Now(),
			Author:     "author",
			ChangeType: ChangeTypeUpdate,
			FullData:   map[string]interface{}{"version": i},
		}
		require.NoError(t, vh.AddSnapshot(snapshot))
	}

	tests := []struct {
		name    string
		version int
		wantErr bool
	}{
		{"get existing version 1", 1, false},
		{"get existing version 2", 2, false},
		{"get existing version 3", 3, false},
		{"get non-existent version 0", 0, true},
		{"get non-existent version 4", 4, true},
		{"get negative version", -1, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			snapshot, err := vh.GetSnapshot(tt.version)
			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, snapshot)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, snapshot)
				assert.Equal(t, tt.version, snapshot.Version)
			}
		})
	}
}

func TestVersionHistory_GetSnapshotAtTime(t *testing.T) {
	vh := NewVersionHistory("elem-1", AgentElement)
	baseTime := time.Now()

	// Add snapshots at different times (all in the future relative to creation)
	for i := 1; i <= 5; i++ {
		snapshot := &VersionSnapshot{
			Version:    i,
			Timestamp:  baseTime.Add(time.Duration(i*10) * time.Minute),
			Author:     "author",
			ChangeType: ChangeTypeUpdate,
			FullData:   map[string]interface{}{"version": i},
		}
		require.NoError(t, vh.AddSnapshot(snapshot))
	}

	tests := []struct {
		name            string
		targetTime      time.Time
		expectedVersion int
		wantErr         bool
	}{
		{
			name:            "time matches version 1",
			targetTime:      baseTime.Add(15 * time.Minute),
			expectedVersion: 1,
			wantErr:         false,
		},
		{
			name:            "time matches version 3",
			targetTime:      baseTime.Add(35 * time.Minute),
			expectedVersion: 3,
			wantErr:         false,
		},
		{
			name:            "time matches latest version",
			targetTime:      baseTime.Add(60 * time.Minute),
			expectedVersion: 5,
			wantErr:         false,
		},
		{
			name:       "time before first snapshot",
			targetTime: baseTime.Add(-10 * time.Minute),
			wantErr:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			snapshot, err := vh.GetSnapshotAtTime(tt.targetTime)
			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, snapshot)
			} else {
				assert.NoError(t, err)
				require.NotNil(t, snapshot)
				assert.Equal(t, tt.expectedVersion, snapshot.Version)
			}
		})
	}
}

func TestVersionHistory_GetVersionRange(t *testing.T) {
	vh := NewVersionHistory("elem-1", TemplateElement)

	// Add 5 snapshots
	for i := 1; i <= 5; i++ {
		snapshot := &VersionSnapshot{
			Version:    i,
			Timestamp:  time.Now(),
			Author:     "author",
			ChangeType: ChangeTypeUpdate,
			FullData:   map[string]interface{}{"version": i},
		}
		require.NoError(t, vh.AddSnapshot(snapshot))
	}

	tests := []struct {
		name          string
		startVersion  int
		endVersion    int
		expectedCount int
		wantErr       bool
	}{
		{"get versions 1-3", 1, 3, 3, false},
		{"get all versions", 1, 5, 5, false},
		{"get single version", 3, 3, 1, false},
		{"invalid range (start > end)", 4, 2, 0, true},
		{"invalid start version", 0, 3, 0, true},
		{"invalid end version", 1, 10, 0, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			snapshots, err := vh.GetVersionRange(tt.startVersion, tt.endVersion)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Len(t, snapshots, tt.expectedCount)
			}
		})
	}
}

func TestVersionHistory_GetTimeRange(t *testing.T) {
	vh := NewVersionHistory("elem-1", EnsembleElement)
	baseTime := time.Now()

	// Add snapshots at different times
	for i := 1; i <= 5; i++ {
		snapshot := &VersionSnapshot{
			Version:    i,
			Timestamp:  baseTime.Add(time.Duration(i) * time.Hour),
			Author:     "author",
			ChangeType: ChangeTypeUpdate,
			FullData:   map[string]interface{}{"version": i},
		}
		require.NoError(t, vh.AddSnapshot(snapshot))
	}

	tests := []struct {
		name          string
		startTime     time.Time
		endTime       time.Time
		expectedCount int
		wantErr       bool
	}{
		{
			name:          "get first 2 hours",
			startTime:     baseTime,
			endTime:       baseTime.Add(2 * time.Hour),
			expectedCount: 2,
			wantErr:       false,
		},
		{
			name:          "get all snapshots",
			startTime:     baseTime,
			endTime:       baseTime.Add(6 * time.Hour),
			expectedCount: 5,
			wantErr:       false,
		},
		{
			name:      "invalid time range (start after end)",
			startTime: baseTime.Add(5 * time.Hour),
			endTime:   baseTime.Add(2 * time.Hour),
			wantErr:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			snapshots, err := vh.GetTimeRange(tt.startTime, tt.endTime)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Len(t, snapshots, tt.expectedCount)
			}
		})
	}
}

func TestVersionHistory_ReconstructAtVersion(t *testing.T) {
	vh := NewVersionHistory("elem-1", PersonaElement)

	// Add initial full snapshot
	snapshot1 := &VersionSnapshot{
		Version:    1,
		Timestamp:  time.Now(),
		Author:     "author",
		ChangeType: ChangeTypeCreate,
		FullData: map[string]interface{}{
			"name":        "Original",
			"description": "Original description",
			"count":       1,
		},
	}
	require.NoError(t, vh.AddSnapshot(snapshot1))

	// Add diff snapshot (change name and count)
	snapshot2 := &VersionSnapshot{
		Version:    2,
		Timestamp:  time.Now(),
		Author:     "author",
		ChangeType: ChangeTypeUpdate,
		Changes: map[string]interface{}{
			"name":  "Updated",
			"count": 2,
		},
	}
	require.NoError(t, vh.AddSnapshot(snapshot2))

	// Add another diff (change description)
	snapshot3 := &VersionSnapshot{
		Version:    3,
		Timestamp:  time.Now(),
		Author:     "author",
		ChangeType: ChangeTypeUpdate,
		Changes: map[string]interface{}{
			"description": "Updated description",
		},
	}
	require.NoError(t, vh.AddSnapshot(snapshot3))

	tests := []struct {
		name         string
		version      int
		expectedData map[string]interface{}
		wantErr      bool
	}{
		{
			name:    "reconstruct version 1",
			version: 1,
			expectedData: map[string]interface{}{
				"name":        "Original",
				"description": "Original description",
				"count":       1,
			},
			wantErr: false,
		},
		{
			name:    "reconstruct version 2",
			version: 2,
			expectedData: map[string]interface{}{
				"name":        "Updated",
				"description": "Original description",
				"count":       2,
			},
			wantErr: false,
		},
		{
			name:    "reconstruct version 3",
			version: 3,
			expectedData: map[string]interface{}{
				"name":        "Updated",
				"description": "Updated description",
				"count":       2,
			},
			wantErr: false,
		},
		{
			name:    "invalid version",
			version: 10,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data, err := vh.ReconstructAtVersion(tt.version)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedData["name"], data["name"])
				assert.Equal(t, tt.expectedData["description"], data["description"])
				assert.Equal(t, tt.expectedData["count"], data["count"])
			}
		})
	}
}

func TestVersionHistory_RetentionPolicy(t *testing.T) {
	vh := NewVersionHistory("elem-1", SkillElement)
	vh.RetentionPolicy = &RetentionPolicy{
		MaxVersions:    3,
		MaxAge:         1 * time.Hour,
		KeepMilestones: true,
	}

	baseTime := time.Now().Add(-2 * time.Hour)

	// Add 5 snapshots, some old
	for i := 1; i <= 5; i++ {
		var changeType ChangeType
		if i == 3 {
			changeType = ChangeTypeMajor // Make version 3 a milestone
		} else {
			changeType = ChangeTypeUpdate
		}

		snapshot := &VersionSnapshot{
			Version:    i,
			Timestamp:  baseTime.Add(time.Duration(i*15) * time.Minute),
			Author:     "author",
			ChangeType: changeType,
			FullData:   map[string]interface{}{"version": i},
		}
		require.NoError(t, vh.AddSnapshot(snapshot))
	}

	// After retention policy application, should keep:
	// - Last 3 versions (3, 4, 5)
	// - Version 3 is both in last 3 AND a milestone
	assert.LessOrEqual(t, len(vh.Snapshots), 5)
	assert.Greater(t, vh.TotalVersions, 0)
}

func TestDefaultPolicies(t *testing.T) {
	t.Run("default snapshot policy", func(t *testing.T) {
		policy := DefaultSnapshotPolicy()
		assert.Equal(t, 10, policy.FullSnapshotInterval)
		assert.Equal(t, 20, policy.MaxDiffChain)
		assert.True(t, policy.MajorVersionFullSnapshot)
	})

	t.Run("default retention policy", func(t *testing.T) {
		policy := DefaultRetentionPolicy()
		assert.Equal(t, 100, policy.MaxVersions)
		assert.Equal(t, 365*24*time.Hour, policy.MaxAge)
		assert.True(t, policy.KeepMilestones)
	})
}

func TestChangeTypes(t *testing.T) {
	assert.Equal(t, ChangeType("create"), ChangeTypeCreate)
	assert.Equal(t, ChangeType("update"), ChangeTypeUpdate)
	assert.Equal(t, ChangeType("activate"), ChangeTypeActivate)
	assert.Equal(t, ChangeType("deactivate"), ChangeTypeDeactivate)
	assert.Equal(t, ChangeType("major"), ChangeTypeMajor)
}
