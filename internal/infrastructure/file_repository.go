package infrastructure

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"sync"
	"time"
	"unicode"

	"github.com/fsvxavier/nexs-mcp/internal/domain"
	"golang.org/x/text/runes"
	"golang.org/x/text/transform"
	"golang.org/x/text/unicode/norm"
	"gopkg.in/yaml.v3"
)

// FileElementRepository implements domain.ElementRepository using file-based storage with YAML.
type FileElementRepository struct {
	mu            sync.RWMutex
	baseDir       string
	cache         map[string]*StoredElement // In-memory cache for faster reads
	adaptiveCache domain.CacheService       // Adaptive cache for GetByID operations
}

// StoredElement represents an element as stored in YAML files.
type StoredElement struct {
	Metadata domain.ElementMetadata `yaml:"metadata"`
	// Type-specific data stored as raw YAML to preserve all fields
	Data map[string]interface{} `yaml:"data,omitempty"`
}

// NewFileElementRepository creates a new file-based repository.
func NewFileElementRepository(baseDir string) (*FileElementRepository, error) {
	if baseDir == "" {
		baseDir = "./nexs-mcp/elements"
	}

	// Create base directory if it doesn't exist
	if err := os.MkdirAll(baseDir, 0o755); err != nil {
		return nil, fmt.Errorf("failed to create base directory: %w", err)
	}

	// Create element type subdirectories upfront for organized storage
	elementTypes := []domain.ElementType{
		domain.PersonaElement,
		domain.SkillElement,
		domain.TemplateElement,
		domain.AgentElement,
		domain.MemoryElement,
		domain.EnsembleElement,
	}
	for _, elemType := range elementTypes {
		typeDir := filepath.Join(baseDir, string(elemType))
		if err := os.MkdirAll(typeDir, 0o755); err != nil {
			return nil, fmt.Errorf("failed to create %s directory: %w", elemType, err)
		}
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

// loadCache loads all existing elements into memory cache and migrates IDs/filenames to normalized form.
func (r *FileElementRepository) loadCache() error {
	// temp storage from old files
	temp := make(map[string]*StoredElement)
	origPaths := make(map[string]string)

	// Read all files
	err := filepath.Walk(r.baseDir, func(path string, info os.FileInfo, err error) error {
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

		// Fallback: if metadata fields are empty, try raw map (handle different YAML producers)
		if stored.Metadata.ID == "" || stored.Metadata.Name == "" {
			var raw map[string]interface{}
			_ = yaml.Unmarshal(data, &raw)
			if rawMeta, ok := raw["metadata"].(map[string]interface{}); ok {
				if stored.Metadata.ID == "" {
					if v, ok := rawMeta["id"].(string); ok {
						stored.Metadata.ID = v
					}
				}
				if stored.Metadata.Name == "" {
					if v, ok := rawMeta["name"].(string); ok {
						stored.Metadata.Name = v
					}
				}
				if stored.Metadata.CreatedAt.IsZero() {
					if v, ok := rawMeta["created_at"].(string); ok {
						if t, err := time.Parse(time.RFC3339, v); err == nil {
							stored.Metadata.CreatedAt = t
						}
					}
				}
				if stored.Metadata.UpdatedAt.IsZero() {
					if v, ok := rawMeta["updated_at"].(string); ok {
						if t, err := time.Parse(time.RFC3339, v); err == nil {
							stored.Metadata.UpdatedAt = t
						}
					}
				}
			}

			if rawMeta, ok := raw["Metadata"].(map[string]interface{}); ok {
				if stored.Metadata.ID == "" {
					if v, ok := rawMeta["id"].(string); ok {
						stored.Metadata.ID = v
					}
				}
				if stored.Metadata.Name == "" {
					if v, ok := rawMeta["name"].(string); ok {
						stored.Metadata.Name = v
					}
				}
				if stored.Metadata.CreatedAt.IsZero() {
					if v, ok := rawMeta["created_at"].(string); ok {
						if t, err := time.Parse(time.RFC3339, v); err == nil {
							stored.Metadata.CreatedAt = t
						}
					}
				}
				if stored.Metadata.UpdatedAt.IsZero() {
					if v, ok := rawMeta["updated_at"].(string); ok {
						if t, err := time.Parse(time.RFC3339, v); err == nil {
							stored.Metadata.UpdatedAt = t
						}
					}
				}
			}
		}

		temp[stored.Metadata.ID] = &stored
		origPaths[stored.Metadata.ID] = path
		return nil
	})
	if err != nil {
		return err
	}

	// Compute old->new ID mapping
	oldToNew := make(map[string]string)
	for oldID, stored := range temp {
		meta := stored.Metadata
		// Try to extract timestamp (last '_' part)
		parts := strings.Split(oldID, "_")
		timestamp := ""
		if len(parts) > 1 {
			timestamp = parts[len(parts)-1]
		}
		// Ensure we have a valid element type; fall back to parsing from old ID if missing
		elemType := meta.Type
		if !domain.ValidateElementType(elemType) {
			if len(parts) > 0 {
				elemType = domain.ElementType(parts[0])
				if !domain.ValidateElementType(elemType) {
					elemType = meta.Type // keep original (may be zero)
				}
			}
		}
		stored.Metadata.Type = elemType

		nameFragment := sanitizeFileName(meta.Name)
		if nameFragment == "" && len(parts) > 1 {
			nameFragment = sanitizeFileName(parts[1])
		}
		newID := string(elemType) + "_" + nameFragment
		if timestamp != "" {
			newID = newID + "_" + timestamp
		}
		if nameFragment == "" {
			// Fall back to old ID if we couldn't infer a name
			newID = oldID
		}
		oldToNew[oldID] = newID
		stored.Metadata.ID = newID
	}

	// Update internal references (persona.related_skills, skill.metadata.custom.related_personas)
	for _, stored := range temp {
		// Update persona related skills if present in Data
		if stored.Metadata.Type == domain.PersonaElement {
			if rs, ok := stored.Data["related_skills"].([]interface{}); ok {
				newSlice := []string{}
				for _, v := range rs {
					if s, ok := v.(string); ok {
						if newID, found := oldToNew[s]; found {
							s = newID
						}
						newSlice = append(newSlice, s)
					}
				}
				stored.Data["related_skills"] = newSlice
			}
		}

		// Update skill related_personas in metadata.Custom
		if stored.Metadata.Type == domain.SkillElement {
			if stored.Metadata.Custom != nil {
				if rp, ok := stored.Metadata.Custom["related_personas"].([]interface{}); ok {
					newSlice := []string{}
					for _, v := range rp {
						if s, ok := v.(string); ok {
							if newID, found := oldToNew[s]; found {
								s = newID
							}
							newSlice = append(newSlice, s)
						}
					}
					stored.Metadata.Custom["related_personas"] = newSlice
				}
			}
		}
	}

	// Write migrated files and populate cache
	for oldID, stored := range temp {
		expectedPath := r.getFilePath(stored.Metadata)
		dir := filepath.Dir(expectedPath)
		if err := os.MkdirAll(dir, 0o755); err != nil {
			return fmt.Errorf("failed to create dir %s: %w", dir, err)
		}

		data, err := yaml.Marshal(stored)
		if err != nil {
			return fmt.Errorf("failed to marshal element %s: %w", stored.Metadata.ID, err)
		}

		if err := os.WriteFile(expectedPath, data, 0o644); err != nil {
			return fmt.Errorf("failed to write file %s: %w", expectedPath, err)
		}

		// remove old file if different
		if origPath, ok := origPaths[oldID]; ok && origPath != expectedPath {
			_ = os.Remove(origPath)
		}

		r.cache[stored.Metadata.ID] = stored
	}

	return nil
}

// sanitizeFileName returns a snake_case ASCII-safe filename without extension.
func sanitizeFileName(s string) string {
	// Normalize unicode to remove accents
	t := transform.Chain(norm.NFD, runes.Remove(runes.In(unicode.Mn)), norm.NFC)
	res, _, _ := transform.String(t, s)

	// Lowercase
	res = strings.ToLower(res)

	// Replace any non-alphanumeric char with underscore
	re := regexp.MustCompile("[^a-z0-9]+")
	res = re.ReplaceAllString(res, "_")

	// Trim underscores
	res = strings.Trim(res, "_")

	return res
}

// getFilePath returns the file path for an element
// Structure: baseDir/type/YYYY-MM-DD/sanitized-id.yaml.
func (r *FileElementRepository) getFilePath(metadata domain.ElementMetadata) string {
	typeDir := string(metadata.Type)
	dateDir := metadata.CreatedAt.Format("2006-01-02")
	filename := sanitizeFileName(metadata.ID) + ".yaml"
	return filepath.Join(r.baseDir, typeDir, dateDir, filename)
}

// Create creates a new element.
func (r *FileElementRepository) Create(element domain.Element) error {
	if element == nil {
		return errors.New("element cannot be nil")
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

// SetAdaptiveCache sets the adaptive cache for this repository.
func (r *FileElementRepository) SetAdaptiveCache(cache domain.CacheService) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.adaptiveCache = cache
}

// GetByID retrieves an element by ID.
func (r *FileElementRepository) GetByID(id string) (domain.Element, error) {
	r.mu.RLock()
	cache := r.adaptiveCache
	r.mu.RUnlock()

	// Try adaptive cache first (caches converted elements)
	if cache != nil {
		cacheKey := "element:" + id
		if cached, found := cache.Get(context.Background(), cacheKey); found {
			return cached.(domain.Element), nil
		}
	}

	r.mu.RLock()
	stored, exists := r.cache[id]
	r.mu.RUnlock()

	if !exists {
		return nil, fmt.Errorf("element with ID %s not found", id)
	}

	// Convert stored element to typed element (expensive operation)
	element, err := r.convertToTypedElement(stored)
	if err != nil {
		return nil, err
	}

	// Cache the converted element
	if cache != nil {
		cacheKey := "element:" + id
		// Estimate ~2KB per element for cache size
		_ = cache.Set(context.Background(), cacheKey, element, 2048)
	}

	return element, nil
}

// convertToTypedElement converts a StoredElement to the appropriate typed element.
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

// Update updates an existing element.
func (r *FileElementRepository) Update(element domain.Element) error {
	if element == nil {
		return errors.New("element cannot be nil")
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

// Delete removes an element by ID.
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

// List returns all elements matching the filter criteria.
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

// Exists checks if an element exists.
func (r *FileElementRepository) Exists(id string) (bool, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	_, exists := r.cache[id]
	return exists, nil
}

// saveToFile saves an element to a YAML file.
func (r *FileElementRepository) saveToFile(stored *StoredElement) error {
	path := r.getFilePath(stored.Metadata)

	// Create directory structure
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	// Marshal to YAML
	data, err := yaml.Marshal(stored)
	if err != nil {
		return fmt.Errorf("failed to marshal element: %w", err)
	}

	// Write to file
	if err := os.WriteFile(path, data, 0o644); err != nil {
		return fmt.Errorf("failed to write file: %w", err)
	}

	return nil
}

// SimpleElement is a basic implementation of Element for file operations.
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
