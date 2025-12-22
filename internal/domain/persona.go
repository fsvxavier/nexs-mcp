package domain

import (
	"errors"
	"fmt"
	"time"
)

// PersonaPrivacyLevel defines the privacy level of a persona.
type PersonaPrivacyLevel string

const (
	PrivacyPublic  PersonaPrivacyLevel = "public"
	PrivacyPrivate PersonaPrivacyLevel = "private"
	PrivacyShared  PersonaPrivacyLevel = "shared"
)

// BehavioralTrait represents a behavioral characteristic.
type BehavioralTrait struct {
	Name        string `json:"name"                  validate:"required,min=2,max=50" yaml:"name"`
	Description string `json:"description,omitempty" yaml:"description,omitempty"`
	Intensity   int    `json:"intensity"             validate:"min=1,max=10"          yaml:"intensity"` // 1-10 scale
}

// ExpertiseArea represents an area of knowledge or skill.
type ExpertiseArea struct {
	Domain      string   `json:"domain"                validate:"required,min=2,max=100"                               yaml:"domain"`
	Level       string   `json:"level"                 validate:"required,oneof=beginner intermediate advanced expert" yaml:"level"`
	Keywords    []string `json:"keywords,omitempty"    yaml:"keywords,omitempty"`
	Description string   `json:"description,omitempty" yaml:"description,omitempty"`
}

// ResponseStyle defines how the persona communicates.
type ResponseStyle struct {
	Tone            string   `json:"tone"                      validate:"required,min=2,max=50"                   yaml:"tone"`
	Formality       string   `json:"formality"                 validate:"required,oneof=casual neutral formal"    yaml:"formality"`
	Verbosity       string   `json:"verbosity"                 validate:"required,oneof=concise balanced verbose" yaml:"verbosity"`
	Perspective     string   `json:"perspective,omitempty"     yaml:"perspective,omitempty"`
	Characteristics []string `json:"characteristics,omitempty" yaml:"characteristics,omitempty"`
}

// Persona represents a complete persona element.
type Persona struct {
	metadata         ElementMetadata
	BehavioralTraits []BehavioralTrait   `json:"behavioral_traits"     validate:"required,min=1,dive"                  yaml:"behavioral_traits"`
	ExpertiseAreas   []ExpertiseArea     `json:"expertise_areas"       validate:"required,min=1,dive"                  yaml:"expertise_areas"`
	ResponseStyle    ResponseStyle       `json:"response_style"        validate:"required"                             yaml:"response_style"`
	SystemPrompt     string              `json:"system_prompt"         validate:"required,min=10,max=2000"             yaml:"system_prompt"`
	PrivacyLevel     PersonaPrivacyLevel `json:"privacy_level"         validate:"required,oneof=public private shared" yaml:"privacy_level"`
	Owner            string              `json:"owner,omitempty"       yaml:"owner,omitempty"`
	SharedWith       []string            `json:"shared_with,omitempty" yaml:"shared_with,omitempty"`
	HotSwappable     bool                `json:"hot_swappable"         yaml:"hot_swappable"`
	// Sprint 3: Cross-Element Relationships
	RelatedSkills    []string `json:"related_skills,omitempty"    yaml:"related_skills,omitempty"`    // Skill IDs this persona uses
	RelatedTemplates []string `json:"related_templates,omitempty" yaml:"related_templates,omitempty"` // Template IDs this persona uses
	RelatedMemories  []string `json:"related_memories,omitempty"  yaml:"related_memories,omitempty"`  // Memory IDs associated with persona
}

// NewPersona creates a new Persona element.
func NewPersona(name, description, version, author string) *Persona {
	now := time.Now()
	return &Persona{
		metadata: ElementMetadata{
			ID:          GenerateElementID(PersonaElement, name),
			Type:        PersonaElement,
			Name:        name,
			Description: description,
			Version:     version,
			Author:      author,
			Tags:        []string{},
			IsActive:    true,
			CreatedAt:   now,
			UpdatedAt:   now,
		},
		BehavioralTraits: []BehavioralTrait{},
		ExpertiseAreas:   []ExpertiseArea{},
		ResponseStyle:    ResponseStyle{},
		PrivacyLevel:     PrivacyPublic,
		HotSwappable:     true,
		RelatedSkills:    []string{},
		RelatedTemplates: []string{},
		RelatedMemories:  []string{},
	}
}

// GetMetadata returns the element metadata.
func (p *Persona) GetMetadata() ElementMetadata {
	return p.metadata
}

// GetType returns the element type.
func (p *Persona) GetType() ElementType {
	return p.metadata.Type
}

// GetID returns the element ID.
func (p *Persona) GetID() string {
	return p.metadata.ID
}

// IsActive returns whether the element is active.
func (p *Persona) IsActive() bool {
	return p.metadata.IsActive
}

// Activate activates the persona.
func (p *Persona) Activate() error {
	p.metadata.IsActive = true
	p.metadata.UpdatedAt = time.Now()
	return nil
}

// Deactivate deactivates the persona.
func (p *Persona) Deactivate() error {
	p.metadata.IsActive = false
	p.metadata.UpdatedAt = time.Now()
	return nil
}

// Validate validates the persona structure.
func (p *Persona) Validate() error {
	// Validate metadata
	if err := p.metadata.Validate(); err != nil {
		return fmt.Errorf("metadata validation failed: %w", err)
	}

	// Validate behavioral traits
	if len(p.BehavioralTraits) == 0 {
		return errors.New("at least one behavioral trait is required")
	}
	for i, trait := range p.BehavioralTraits {
		if trait.Name == "" {
			return fmt.Errorf("behavioral trait %d: name is required", i)
		}
		if trait.Intensity < 1 || trait.Intensity > 10 {
			return fmt.Errorf("behavioral trait %s: intensity must be between 1 and 10", trait.Name)
		}
	}

	// Validate expertise areas
	if len(p.ExpertiseAreas) == 0 {
		return errors.New("at least one expertise area is required")
	}
	for i, area := range p.ExpertiseAreas {
		if area.Domain == "" {
			return fmt.Errorf("expertise area %d: domain is required", i)
		}
		if area.Level == "" {
			return fmt.Errorf("expertise area %s: level is required", area.Domain)
		}
		validLevels := map[string]bool{"beginner": true, "intermediate": true, "advanced": true, "expert": true}
		if !validLevels[area.Level] {
			return fmt.Errorf("expertise area %s: invalid level %s", area.Domain, area.Level)
		}
	}

	// Validate response style
	if p.ResponseStyle.Tone == "" {
		return errors.New("response style tone is required")
	}
	if p.ResponseStyle.Formality == "" {
		return errors.New("response style formality is required")
	}
	validFormality := map[string]bool{"casual": true, "neutral": true, "formal": true}
	if !validFormality[p.ResponseStyle.Formality] {
		return fmt.Errorf("invalid formality level: %s", p.ResponseStyle.Formality)
	}

	if p.ResponseStyle.Verbosity == "" {
		return errors.New("response style verbosity is required")
	}
	validVerbosity := map[string]bool{"concise": true, "balanced": true, "verbose": true}
	if !validVerbosity[p.ResponseStyle.Verbosity] {
		return fmt.Errorf("invalid verbosity level: %s", p.ResponseStyle.Verbosity)
	}

	// Validate system prompt
	if p.SystemPrompt == "" {
		return errors.New("system prompt is required")
	}
	if len(p.SystemPrompt) < 10 {
		return errors.New("system prompt must be at least 10 characters")
	}
	if len(p.SystemPrompt) > 2000 {
		return errors.New("system prompt must not exceed 2000 characters")
	}

	// Validate privacy level
	validPrivacy := map[PersonaPrivacyLevel]bool{
		PrivacyPublic:  true,
		PrivacyPrivate: true,
		PrivacyShared:  true,
	}
	if !validPrivacy[p.PrivacyLevel] {
		return fmt.Errorf("invalid privacy level: %s", p.PrivacyLevel)
	}

	// Validate shared personas have shared_with list
	if p.PrivacyLevel == PrivacyShared && len(p.SharedWith) == 0 {
		return errors.New("shared personas must have at least one user in shared_with list")
	}

	return nil
}

// SetMetadata updates the persona metadata.
func (p *Persona) SetMetadata(metadata ElementMetadata) {
	p.metadata = metadata
	p.metadata.UpdatedAt = time.Now()
}

// AddBehavioralTrait adds a behavioral trait to the persona.
func (p *Persona) AddBehavioralTrait(trait BehavioralTrait) error {
	if trait.Name == "" {
		return errors.New("trait name is required")
	}
	if trait.Intensity < 1 || trait.Intensity > 10 {
		return errors.New("intensity must be between 1 and 10")
	}
	p.BehavioralTraits = append(p.BehavioralTraits, trait)
	p.metadata.UpdatedAt = time.Now()
	return nil
}

// AddExpertiseArea adds an expertise area to the persona.
func (p *Persona) AddExpertiseArea(area ExpertiseArea) error {
	if area.Domain == "" {
		return errors.New("domain is required")
	}
	validLevels := map[string]bool{"beginner": true, "intermediate": true, "advanced": true, "expert": true}
	if !validLevels[area.Level] {
		return fmt.Errorf("invalid level: %s", area.Level)
	}
	p.ExpertiseAreas = append(p.ExpertiseAreas, area)
	p.metadata.UpdatedAt = time.Now()
	return nil
}

// SetResponseStyle sets the response style.
func (p *Persona) SetResponseStyle(style ResponseStyle) error {
	if style.Tone == "" {
		return errors.New("tone is required")
	}
	validFormality := map[string]bool{"casual": true, "neutral": true, "formal": true}
	if !validFormality[style.Formality] {
		return fmt.Errorf("invalid formality: %s", style.Formality)
	}
	validVerbosity := map[string]bool{"concise": true, "balanced": true, "verbose": true}
	if !validVerbosity[style.Verbosity] {
		return fmt.Errorf("invalid verbosity: %s", style.Verbosity)
	}
	p.ResponseStyle = style
	p.metadata.UpdatedAt = time.Now()
	return nil
}

// SetSystemPrompt sets the system prompt.
func (p *Persona) SetSystemPrompt(prompt string) error {
	if len(prompt) < 10 {
		return errors.New("system prompt must be at least 10 characters")
	}
	if len(prompt) > 2000 {
		return errors.New("system prompt must not exceed 2000 characters")
	}
	p.SystemPrompt = prompt
	p.metadata.UpdatedAt = time.Now()
	return nil
}

// SetPrivacyLevel sets the privacy level.
func (p *Persona) SetPrivacyLevel(level PersonaPrivacyLevel) error {
	validPrivacy := map[PersonaPrivacyLevel]bool{
		PrivacyPublic:  true,
		PrivacyPrivate: true,
		PrivacyShared:  true,
	}
	if !validPrivacy[level] {
		return fmt.Errorf("invalid privacy level: %s", level)
	}
	p.PrivacyLevel = level
	p.metadata.UpdatedAt = time.Now()
	return nil
}

// ShareWith adds a user to the shared_with list.
func (p *Persona) ShareWith(user string) error {
	if user == "" {
		return errors.New("user is required")
	}
	if p.PrivacyLevel != PrivacyShared {
		return errors.New("persona must be set to shared privacy level first")
	}
	// Check if already shared
	for _, u := range p.SharedWith {
		if u == user {
			return nil // Already shared
		}
	}
	p.SharedWith = append(p.SharedWith, user)
	p.metadata.UpdatedAt = time.Now()
	return nil
}

// UnshareWith removes a user from the shared_with list.
func (p *Persona) UnshareWith(user string) error {
	for i, u := range p.SharedWith {
		if u == user {
			p.SharedWith = append(p.SharedWith[:i], p.SharedWith[i+1:]...)
			p.metadata.UpdatedAt = time.Now()
			return nil
		}
	}
	return fmt.Errorf("user %s not found in shared_with list", user)
}

// AddRelatedSkill adds a skill ID to the persona's related skills.
func (p *Persona) AddRelatedSkill(skillID string) {
	if skillID == "" {
		return
	}
	if !containsString(p.RelatedSkills, skillID) {
		p.RelatedSkills = append(p.RelatedSkills, skillID)
		p.metadata.UpdatedAt = time.Now()
	}
}

// RemoveRelatedSkill removes a skill ID from the persona's related skills.
func (p *Persona) RemoveRelatedSkill(skillID string) {
	p.RelatedSkills = removeStringFromSlice(p.RelatedSkills, skillID)
	p.metadata.UpdatedAt = time.Now()
}

// AddRelatedTemplate adds a template ID to the persona's related templates.
func (p *Persona) AddRelatedTemplate(templateID string) {
	if templateID == "" {
		return
	}
	if !containsString(p.RelatedTemplates, templateID) {
		p.RelatedTemplates = append(p.RelatedTemplates, templateID)
		p.metadata.UpdatedAt = time.Now()
	}
}

// RemoveRelatedTemplate removes a template ID from the persona's related templates.
func (p *Persona) RemoveRelatedTemplate(templateID string) {
	p.RelatedTemplates = removeStringFromSlice(p.RelatedTemplates, templateID)
	p.metadata.UpdatedAt = time.Now()
}

// AddRelatedMemory adds a memory ID to the persona's related memories.
func (p *Persona) AddRelatedMemory(memoryID string) {
	if memoryID == "" {
		return
	}
	if !containsString(p.RelatedMemories, memoryID) {
		p.RelatedMemories = append(p.RelatedMemories, memoryID)
		p.metadata.UpdatedAt = time.Now()
	}
}

// RemoveRelatedMemory removes a memory ID from the persona's related memories.
func (p *Persona) RemoveRelatedMemory(memoryID string) {
	p.RelatedMemories = removeStringFromSlice(p.RelatedMemories, memoryID)
	p.metadata.UpdatedAt = time.Now()
}

// GetAllRelatedIDs returns all related element IDs (skills, templates, memories).
func (p *Persona) GetAllRelatedIDs() []string {
	allIDs := make([]string, 0)
	allIDs = append(allIDs, p.RelatedSkills...)
	allIDs = append(allIDs, p.RelatedTemplates...)
	allIDs = append(allIDs, p.RelatedMemories...)
	return allIDs
}
