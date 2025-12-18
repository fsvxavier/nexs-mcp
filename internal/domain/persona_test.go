package domain

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewPersona(t *testing.T) {
	persona := NewPersona("Test Persona", "A test persona", "1.0.0", "Test Author")

	assert.NotEmpty(t, persona.GetID())
	assert.Equal(t, PersonaElement, persona.GetType())
	assert.Equal(t, "Test Persona", persona.GetMetadata().Name)
	assert.Equal(t, "A test persona", persona.GetMetadata().Description)
	assert.Equal(t, "1.0.0", persona.GetMetadata().Version)
	assert.Equal(t, "Test Author", persona.GetMetadata().Author)
	assert.True(t, persona.IsActive())
	assert.Equal(t, PrivacyPublic, persona.PrivacyLevel)
	assert.True(t, persona.HotSwappable)
	assert.Empty(t, persona.BehavioralTraits)
	assert.Empty(t, persona.ExpertiseAreas)
}

func TestPersona_AddBehavioralTrait(t *testing.T) {
	persona := NewPersona("Test", "Test", "1.0.0", "Test")

	tests := []struct {
		name        string
		trait       BehavioralTrait
		expectError bool
		errorMsg    string
	}{
		{
			name: "valid trait",
			trait: BehavioralTrait{
				Name:        "Friendly",
				Description: "Warm and welcoming",
				Intensity:   8,
			},
			expectError: false,
		},
		{
			name: "empty name",
			trait: BehavioralTrait{
				Intensity: 5,
			},
			expectError: true,
			errorMsg:    "name is required",
		},
		{
			name: "intensity too low",
			trait: BehavioralTrait{
				Name:      "Test",
				Intensity: 0,
			},
			expectError: true,
			errorMsg:    "intensity must be between 1 and 10",
		},
		{
			name: "intensity too high",
			trait: BehavioralTrait{
				Name:      "Test",
				Intensity: 11,
			},
			expectError: true,
			errorMsg:    "intensity must be between 1 and 10",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := persona.AddBehavioralTrait(tt.trait)
			if tt.expectError {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errorMsg)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestPersona_AddExpertiseArea(t *testing.T) {
	persona := NewPersona("Test", "Test", "1.0.0", "Test")

	tests := []struct {
		name        string
		area        ExpertiseArea
		expectError bool
		errorMsg    string
	}{
		{
			name: "valid area",
			area: ExpertiseArea{
				Domain:      "Software Engineering",
				Level:       "expert",
				Keywords:    []string{"go", "testing"},
				Description: "Expert in Go",
			},
			expectError: false,
		},
		{
			name: "empty domain",
			area: ExpertiseArea{
				Level: "expert",
			},
			expectError: true,
			errorMsg:    "domain is required",
		},
		{
			name: "invalid level",
			area: ExpertiseArea{
				Domain: "Test",
				Level:  "master",
			},
			expectError: true,
			errorMsg:    "invalid level",
		},
		{
			name: "beginner level",
			area: ExpertiseArea{
				Domain: "Machine Learning",
				Level:  "beginner",
			},
			expectError: false,
		},
		{
			name: "intermediate level",
			area: ExpertiseArea{
				Domain: "Data Science",
				Level:  "intermediate",
			},
			expectError: false,
		},
		{
			name: "advanced level",
			area: ExpertiseArea{
				Domain: "Cloud Architecture",
				Level:  "advanced",
			},
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := persona.AddExpertiseArea(tt.area)
			if tt.expectError {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errorMsg)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestPersona_SetResponseStyle(t *testing.T) {
	persona := NewPersona("Test", "Test", "1.0.0", "Test")

	tests := []struct {
		name        string
		style       ResponseStyle
		expectError bool
		errorMsg    string
	}{
		{
			name: "valid style",
			style: ResponseStyle{
				Tone:        "professional",
				Formality:   "formal",
				Verbosity:   "balanced",
				Perspective: "first-person",
			},
			expectError: false,
		},
		{
			name: "empty tone",
			style: ResponseStyle{
				Formality: "formal",
				Verbosity: "balanced",
			},
			expectError: true,
			errorMsg:    "tone is required",
		},
		{
			name: "invalid formality",
			style: ResponseStyle{
				Tone:      "friendly",
				Formality: "super-formal",
				Verbosity: "balanced",
			},
			expectError: true,
			errorMsg:    "invalid formality",
		},
		{
			name: "invalid verbosity",
			style: ResponseStyle{
				Tone:      "friendly",
				Formality: "casual",
				Verbosity: "very-verbose",
			},
			expectError: true,
			errorMsg:    "invalid verbosity",
		},
		{
			name: "casual formality",
			style: ResponseStyle{
				Tone:      "friendly",
				Formality: "casual",
				Verbosity: "concise",
			},
			expectError: false,
		},
		{
			name: "neutral formality verbose",
			style: ResponseStyle{
				Tone:      "neutral",
				Formality: "neutral",
				Verbosity: "verbose",
			},
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := persona.SetResponseStyle(tt.style)
			if tt.expectError {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errorMsg)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestPersona_SetSystemPrompt(t *testing.T) {
	persona := NewPersona("Test", "Test", "1.0.0", "Test")

	tests := []struct {
		name        string
		prompt      string
		expectError bool
		errorMsg    string
	}{
		{
			name:        "valid prompt",
			prompt:      "You are a helpful assistant specialized in software development.",
			expectError: false,
		},
		{
			name:        "too short",
			prompt:      "Short",
			expectError: true,
			errorMsg:    "at least 10 characters",
		},
		{
			name:        "too long",
			prompt:      string(make([]byte, 2001)),
			expectError: true,
			errorMsg:    "not exceed 2000 characters",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := persona.SetSystemPrompt(tt.prompt)
			if tt.expectError {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errorMsg)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.prompt, persona.SystemPrompt)
			}
		})
	}
}

func TestPersona_SetPrivacyLevel(t *testing.T) {
	persona := NewPersona("Test", "Test", "1.0.0", "Test")

	tests := []struct {
		name        string
		level       PersonaPrivacyLevel
		expectError bool
	}{
		{
			name:        "public",
			level:       PrivacyPublic,
			expectError: false,
		},
		{
			name:        "private",
			level:       PrivacyPrivate,
			expectError: false,
		},
		{
			name:        "shared",
			level:       PrivacyShared,
			expectError: false,
		},
		{
			name:        "invalid",
			level:       PersonaPrivacyLevel("invalid"),
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := persona.SetPrivacyLevel(tt.level)
			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.level, persona.PrivacyLevel)
			}
		})
	}
}

func TestPersona_ShareWith(t *testing.T) {
	persona := NewPersona("Test", "Test", "1.0.0", "Test")

	t.Run("share without setting privacy level", func(t *testing.T) {
		err := persona.ShareWith("user1")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "shared privacy level")
	})

	t.Run("share with valid user", func(t *testing.T) {
		err := persona.SetPrivacyLevel(PrivacyShared)
		require.NoError(t, err)

		err = persona.ShareWith("user1")
		assert.NoError(t, err)
		assert.Contains(t, persona.SharedWith, "user1")
	})

	t.Run("share with duplicate user", func(t *testing.T) {
		err := persona.ShareWith("user1")
		assert.NoError(t, err)
		assert.Len(t, persona.SharedWith, 1)
	})

	t.Run("share with empty user", func(t *testing.T) {
		err := persona.ShareWith("")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "user is required")
	})
}

func TestPersona_UnshareWith(t *testing.T) {
	persona := NewPersona("Test", "Test", "1.0.0", "Test")
	persona.SetPrivacyLevel(PrivacyShared)
	persona.ShareWith("user1")
	persona.ShareWith("user2")

	t.Run("unshare existing user", func(t *testing.T) {
		err := persona.UnshareWith("user1")
		assert.NoError(t, err)
		assert.NotContains(t, persona.SharedWith, "user1")
		assert.Contains(t, persona.SharedWith, "user2")
	})

	t.Run("unshare non-existing user", func(t *testing.T) {
		err := persona.UnshareWith("user3")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "not found")
	})
}

func TestPersona_Validate(t *testing.T) {
	t.Run("valid complete persona", func(t *testing.T) {
		persona := NewPersona("Expert Coder", "A coding expert", "1.0.0", "Test Author")
		persona.AddBehavioralTrait(BehavioralTrait{
			Name:      "Helpful",
			Intensity: 9,
		})
		persona.AddExpertiseArea(ExpertiseArea{
			Domain: "Software Engineering",
			Level:  "expert",
		})
		persona.SetResponseStyle(ResponseStyle{
			Tone:      "professional",
			Formality: "neutral",
			Verbosity: "balanced",
		})
		persona.SetSystemPrompt("You are an expert software engineer.")

		err := persona.Validate()
		assert.NoError(t, err)
	})

	t.Run("missing behavioral traits", func(t *testing.T) {
		persona := NewPersona("Test", "Test", "1.0.0", "Test")
		persona.AddExpertiseArea(ExpertiseArea{Domain: "Test", Level: "beginner"})
		persona.SetResponseStyle(ResponseStyle{Tone: "test", Formality: "casual", Verbosity: "concise"})
		persona.SetSystemPrompt("Test prompt here")

		err := persona.Validate()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "behavioral trait")
	})

	t.Run("missing expertise areas", func(t *testing.T) {
		persona := NewPersona("Test", "Test", "1.0.0", "Test")
		persona.AddBehavioralTrait(BehavioralTrait{Name: "Test", Intensity: 5})
		persona.SetResponseStyle(ResponseStyle{Tone: "test", Formality: "casual", Verbosity: "concise"})
		persona.SetSystemPrompt("Test prompt here")

		err := persona.Validate()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "expertise area")
	})

	t.Run("missing system prompt", func(t *testing.T) {
		persona := NewPersona("Test", "Test", "1.0.0", "Test")
		persona.AddBehavioralTrait(BehavioralTrait{Name: "Test", Intensity: 5})
		persona.AddExpertiseArea(ExpertiseArea{Domain: "Test", Level: "beginner"})
		persona.SetResponseStyle(ResponseStyle{Tone: "test", Formality: "casual", Verbosity: "concise"})

		err := persona.Validate()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "system prompt")
	})

	t.Run("shared persona without shared_with list", func(t *testing.T) {
		persona := NewPersona("Test", "Test", "1.0.0", "Test")
		persona.AddBehavioralTrait(BehavioralTrait{Name: "Test", Intensity: 5})
		persona.AddExpertiseArea(ExpertiseArea{Domain: "Test", Level: "beginner"})
		persona.SetResponseStyle(ResponseStyle{Tone: "test", Formality: "casual", Verbosity: "concise"})
		persona.SetSystemPrompt("Test prompt here")
		persona.SetPrivacyLevel(PrivacyShared)

		err := persona.Validate()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "shared_with")
	})
}

func TestPersona_ActivateDeactivate(t *testing.T) {
	persona := NewPersona("Test", "Test", "1.0.0", "Test")

	assert.True(t, persona.IsActive())

	oldTime := persona.GetMetadata().UpdatedAt
	time.Sleep(time.Millisecond)

	err := persona.Deactivate()
	assert.NoError(t, err)
	assert.False(t, persona.IsActive())
	assert.True(t, persona.GetMetadata().UpdatedAt.After(oldTime))

	oldTime = persona.GetMetadata().UpdatedAt
	time.Sleep(time.Millisecond)

	err = persona.Activate()
	assert.NoError(t, err)
	assert.True(t, persona.IsActive())
	assert.True(t, persona.GetMetadata().UpdatedAt.After(oldTime))
}

func TestPersona_SetMetadata(t *testing.T) {
	persona := NewPersona("Original", "Original", "1.0.0", "Author")

	newMetadata := persona.GetMetadata()
	newMetadata.Name = "Updated"
	newMetadata.Description = "Updated description"

	oldTime := persona.GetMetadata().UpdatedAt
	time.Sleep(time.Millisecond)

	persona.SetMetadata(newMetadata)

	assert.Equal(t, "Updated", persona.GetMetadata().Name)
	assert.Equal(t, "Updated description", persona.GetMetadata().Description)
	assert.True(t, persona.GetMetadata().UpdatedAt.After(oldTime))
}
