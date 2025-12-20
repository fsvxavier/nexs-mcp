package infrastructure

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/fsvxavier/nexs-mcp/internal/domain"
	"gopkg.in/yaml.v3"
)

// FileElementRepository implements domain.ElementRepository using file-based storage with YAML
type FileElementRepository struct {
	mu      sync.RWMutex
	baseDir string
	cache   map[string]*StoredElement // In-memory cache for faster reads
}

// StoredElement represents an element as stored in YAML files
type StoredElement struct {
	Metadata domain.ElementMetadata `yaml:"metadata"`
	// Type-specific data stored as raw YAML to preserve all fields
	Data map[string]interface{} `yaml:"data,omitempty"`
}

// NewFileElementRepository creates a new file-based repository
func NewFileElementRepository(baseDir string) (*FileElementRepository, error) {
	if baseDir == "" {
		baseDir = "data/elements"
	}

	// Create base directory if it doesn't exist
	if err := os.MkdirAll(baseDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create base directory: %w", err)
	}

	repo := &FileElementRepository{
		baseDir: baseDir,
		cache:   make(map[string]*StoredElement),
	}

	// Load existing elements into cache
	if err := repo.loadCache(); err != nil {
		return nil, fmt.Errorf("failed to load cache: %w", err)
	}

	return repo, nil
}

// loadCache loads all existing elements into memory cache
func (r *FileElementRepository) loadCache() error {
	return filepath.Walk(r.baseDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() || !strings.HasSuffix(path, ".yaml") {
			return nil
		}

		data, err := os.ReadFile(path)
		if err != nil {
			return fmt.Errorf("failed to read file %s: %w", path, err)
		}

		var stored StoredElement
		if err := yaml.Unmarshal(data, &stored); err != nil {
			return fmt.Errorf("failed to unmarshal file %s: %w", path, err)
		}

		r.cache[stored.Metadata.ID] = &stored
		return nil
	})
}

// getFilePath returns the file path for an element
// Structure: baseDir/type/YYYY-MM-DD/id.yaml
func (r *FileElementRepository) getFilePath(metadata domain.ElementMetadata) string {
	typeDir := string(metadata.Type)
	dateDir := metadata.CreatedAt.Format("2006-01-02")
	filename := fmt.Sprintf("%s.yaml", metadata.ID)
	return filepath.Join(r.baseDir, typeDir, dateDir, filename)
}

// Create creates a new element
func (r *FileElementRepository) Create(element domain.Element) error {
	if element == nil {
		return fmt.Errorf("element cannot be nil")
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	metadata := element.GetMetadata()
	if _, exists := r.cache[metadata.ID]; exists {
		return fmt.Errorf("element with ID %s already exists", metadata.ID)
	}

	stored := &StoredElement{
		Metadata: metadata,
		Data:     extractElementData(element),
	}

	// Save to file
	if err := r.saveToFile(stored); err != nil {
		return err
	}

	// Update cache
	r.cache[metadata.ID] = stored
	return nil
}

// GetByID retrieves an element by ID
func (r *FileElementRepository) GetByID(id string) (domain.Element, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	stored, exists := r.cache[id]
	if !exists {
		return nil, fmt.Errorf("element with ID %s not found", id)
	}

	return r.convertToTypedElement(stored)
}

// convertToTypedElement converts a StoredElement to the appropriate typed element
func (r *FileElementRepository) convertToTypedElement(stored *StoredElement) (domain.Element, error) {
	metadata := stored.Metadata

	var element domain.Element

	switch metadata.Type {
	case domain.PersonaElement:
		persona := domain.NewPersona(metadata.Name, metadata.Description, metadata.Version, metadata.Author)
		persona.SetMetadata(metadata)
		restoreElementData(persona, stored.Data)
		element = persona

	case domain.SkillElement:
		skill := domain.NewSkill(metadata.Name, metadata.Description, metadata.Version, metadata.Author)
		skill.SetMetadata(metadata)
		restoreElementData(skill, stored.Data)
		element = skill

	case domain.TemplateElement:
		template := domain.NewTemplate(metadata.Name, metadata.Description, metadata.Version, metadata.Author)
		template.SetMetadata(metadata)
		restoreElementData(template, stored.Data)
		element = template

	case domain.AgentElement:
		agent := domain.NewAgent(metadata.Name, metadata.Description, metadata.Version, metadata.Author)
		agent.SetMetadata(metadata)
		restoreElementData(agent, stored.Data)
		element = agent

	case domain.MemoryElement:
		memory := domain.NewMemory(metadata.Name, metadata.Description, metadata.Version, metadata.Author)
		memory.SetMetadata(metadata)
		restoreElementData(memory, stored.Data)
		element = memory

	case domain.EnsembleElement:
		ensemble := domain.NewEnsemble(metadata.Name, metadata.Description, metadata.Version, metadata.Author)
		ensemble.SetMetadata(metadata)
		restoreElementData(ensemble, stored.Data)
		element = ensemble

	default:
		element = &SimpleElement{metadata: metadata}
	}

	return element, nil
}

// Update updates an existing element
func (r *FileElementRepository) Update(element domain.Element) error {
	if element == nil {
		return fmt.Errorf("element cannot be nil")
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	metadata := element.GetMetadata()
	old, exists := r.cache[metadata.ID]
	if !exists {
		return fmt.Errorf("element with ID %s not found", metadata.ID)
	}

	// Delete old file if date changed
	oldPath := r.getFilePath(old.Metadata)
	newPath := r.getFilePath(metadata)

	if oldPath != newPath {
		if err := os.Remove(oldPath); err != nil && !os.IsNotExist(err) {
			return fmt.Errorf("failed to remove old file: %w", err)
		}
	}

	stored := &StoredElement{
		Metadata: metadata,
		Data:     extractElementData(element),
	}

	// Save to file
	if err := r.saveToFile(stored); err != nil {
		return err
	}

	// Update cache
	r.cache[metadata.ID] = stored
	return nil
}

// Delete removes an element by ID
func (r *FileElementRepository) Delete(id string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	stored, exists := r.cache[id]
	if !exists {
		return fmt.Errorf("element with ID %s not found", id)
	}

	// Delete file
	path := r.getFilePath(stored.Metadata)
	if err := os.Remove(path); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to remove file: %w", err)
	}

	// Remove from cache
	delete(r.cache, id)
	return nil
}

// List returns all elements matching the filter criteria
func (r *FileElementRepository) List(filter domain.ElementFilter) ([]domain.Element, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var results []domain.Element

	for _, stored := range r.cache {
		metadata := stored.Metadata

		// Filter by type
		if filter.Type != nil && metadata.Type != *filter.Type {
			continue
		}

		// Filter by active status
		if filter.IsActive != nil && metadata.IsActive != *filter.IsActive {
			continue
		}

		// Filter by tags
		if len(filter.Tags) > 0 {
			hasAllTags := true
			for _, filterTag := range filter.Tags {
				found := false
				for _, tag := range metadata.Tags {
					if tag == filterTag {
						found = true
						break
					}
				}
				if !found {
					hasAllTags = false
					break
				}
			}
			if !hasAllTags {
				continue
			}
		}

		results = append(results, &SimpleElement{metadata: metadata})
	}

	// Apply pagination
	total := len(results)
	start := filter.Offset
	if start > total {
		start = total
	}

	end := start + filter.Limit
	if filter.Limit == 0 || end > total {
		end = total
	}

	if start < total {
		results = results[start:end]
	} else {
		results = []domain.Element{}
	}

	return results, nil
}

// Exists checks if an element exists
func (r *FileElementRepository) Exists(id string) (bool, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	_, exists := r.cache[id]
	return exists, nil
}

// saveToFile saves an element to a YAML file
func (r *FileElementRepository) saveToFile(stored *StoredElement) error {
	path := r.getFilePath(stored.Metadata)

	// Create directory structure
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	// Marshal to YAML
	data, err := yaml.Marshal(stored)
	if err != nil {
		return fmt.Errorf("failed to marshal element: %w", err)
	}

	// Write to file
	if err := os.WriteFile(path, data, 0644); err != nil {
		return fmt.Errorf("failed to write file: %w", err)
	}

	return nil
}

// SimpleElement is a basic implementation of Element for file operations
type SimpleElement struct {
	metadata domain.ElementMetadata
}

func (s *SimpleElement) GetMetadata() domain.ElementMetadata { return s.metadata }
func (s *SimpleElement) Validate() error                     { return nil }
func (s *SimpleElement) GetType() domain.ElementType         { return s.metadata.Type }
func (s *SimpleElement) GetID() string                       { return s.metadata.ID }
func (s *SimpleElement) IsActive() bool                      { return s.metadata.IsActive }
func (s *SimpleElement) Activate() error {
	s.metadata.IsActive = true
	s.metadata.UpdatedAt = time.Now()
	return nil
}
func (s *SimpleElement) Deactivate() error {
	s.metadata.IsActive = false
	s.metadata.UpdatedAt = time.Now()
	return nil
}
