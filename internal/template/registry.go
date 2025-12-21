package template

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/fsvxavier/nexs-mcp/internal/domain"
	"github.com/fsvxavier/nexs-mcp/internal/template/stdlib"
)

// TemplateRegistry manages template discovery, caching, and indexing.
type TemplateRegistry struct {
	cache  *TemplateCache
	index  *TemplateIndex
	stdlib *stdlib.StandardLibrary
	repo   domain.ElementRepository
	mu     sync.RWMutex
}

// TemplateCache provides fast in-memory template lookup with TTL.
type TemplateCache struct {
	templates map[string]*domain.Template
	expires   map[string]time.Time
	ttl       time.Duration
	hits      uint64
	misses    uint64
	evictions uint64
	mu        sync.RWMutex
}

// TemplateIndex enables rich filtering and search.
type TemplateIndex struct {
	byCategory    map[string][]string // persona, skill, agent, etc.
	byTag         map[string][]string
	byElementType map[string][]string
	byAuthor      map[string][]string
	mu            sync.RWMutex
}

// TemplateSearchFilter defines search criteria.
type TemplateSearchFilter struct {
	Category       string
	Tags           []string
	ElementType    string
	Author         string
	Query          string
	IncludeBuiltIn bool
	Page           int
	PerPage        int
}

// TemplateSearchResult contains search results with metadata.
type TemplateSearchResult struct {
	Templates []*domain.Template
	Total     int
	Page      int
	PerPage   int
	HasMore   bool
}

// CacheStats contains cache performance metrics.
type CacheStats struct {
	Hits      uint64
	Misses    uint64
	Evictions uint64
	Size      int
	HitRate   float64
}

// IndexStats contains index size metrics.
type IndexStats struct {
	Categories     int
	Tags           int
	ElementTypes   int
	Authors        int
	TotalTemplates int
}

// NewTemplateRegistry creates a new template registry.
func NewTemplateRegistry(repo domain.ElementRepository, cacheTTL time.Duration) *TemplateRegistry {
	if cacheTTL == 0 {
		cacheTTL = 15 * time.Minute // Default: 15 minutes
	}

	return &TemplateRegistry{
		cache: &TemplateCache{
			templates: make(map[string]*domain.Template),
			expires:   make(map[string]time.Time),
			ttl:       cacheTTL,
			mu:        sync.RWMutex{},
		},
		index: &TemplateIndex{
			byCategory:    make(map[string][]string),
			byTag:         make(map[string][]string),
			byElementType: make(map[string][]string),
			byAuthor:      make(map[string][]string),
			mu:            sync.RWMutex{},
		},
		stdlib: stdlib.NewStandardLibrary(),
		repo:   repo,
		mu:     sync.RWMutex{},
	}
}

// GetTemplate retrieves a template by ID (checks cache → stdlib → repo).
func (r *TemplateRegistry) GetTemplate(ctx context.Context, id string) (*domain.Template, error) {
	// Check cache first
	if tmpl, found := r.cache.Get(id); found {
		return tmpl, nil
	}

	// Check standard library
	if tmpl, err := r.stdlib.Get(id); err == nil {
		r.cache.Set(id, tmpl)
		return tmpl, nil
	}

	// Load from repository
	element, err := r.repo.GetByID(id)
	if err != nil {
		return nil, fmt.Errorf("template not found: %w", err)
	}

	tmpl, ok := element.(*domain.Template)
	if !ok {
		return nil, fmt.Errorf("element %s is not a template", id)
	}

	// Cache it
	r.cache.Set(id, tmpl)

	// Index it
	r.indexTemplate(tmpl)

	return tmpl, nil
}

// SearchTemplates searches templates with filters.
func (r *TemplateRegistry) SearchTemplates(ctx context.Context, filter TemplateSearchFilter) (*TemplateSearchResult, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	// Collect candidate IDs
	candidateIDs := r.collectCandidates(filter)

	// Load templates
	templates := make([]*domain.Template, 0, len(candidateIDs))
	for _, id := range candidateIDs {
		tmpl, err := r.GetTemplate(ctx, id)
		if err != nil {
			continue // Skip templates that can't be loaded
		}

		// Apply query filter if specified
		if filter.Query != "" && !r.matchesQuery(tmpl, filter.Query) {
			continue
		}

		templates = append(templates, tmpl)
	}

	// Pagination
	total := len(templates)
	page := filter.Page
	if page < 1 {
		page = 1
	}
	perPage := filter.PerPage
	if perPage < 1 {
		perPage = 20
	}

	start := (page - 1) * perPage
	end := start + perPage
	if start >= total {
		return &TemplateSearchResult{
			Templates: []*domain.Template{},
			Total:     total,
			Page:      page,
			PerPage:   perPage,
			HasMore:   false,
		}, nil
	}
	if end > total {
		end = total
	}

	return &TemplateSearchResult{
		Templates: templates[start:end],
		Total:     total,
		Page:      page,
		PerPage:   perPage,
		HasMore:   end < total,
	}, nil
}

// ListAllTemplates returns all templates (repository + stdlib).
func (r *TemplateRegistry) ListAllTemplates(ctx context.Context, includeBuiltIn bool) ([]*domain.Template, error) {
	templates := make([]*domain.Template, 0)

	// Add standard library templates
	if includeBuiltIn {
		if stdlibTemplates, err := r.stdlib.GetAll(); err == nil {
			templates = append(templates, stdlibTemplates...)
		}
	}

	// Add repository templates
	tmplType := domain.TemplateElement
	elements, err := r.repo.List(domain.ElementFilter{Type: &tmplType})
	if err != nil {
		return nil, fmt.Errorf("failed to list templates: %w", err)
	}

	for _, element := range elements {
		if tmpl, ok := element.(*domain.Template); ok {
			templates = append(templates, tmpl)

			// Cache it
			r.cache.Set(tmpl.GetID(), tmpl)

			// Index it
			r.indexTemplate(tmpl)
		}
	}

	return templates, nil
}

// InvalidateCache clears the entire cache.
func (r *TemplateRegistry) InvalidateCache() {
	r.cache.Clear()
}

// InvalidateTemplate removes a specific template from cache.
func (r *TemplateRegistry) InvalidateTemplate(id string) {
	r.cache.Delete(id)
}

// RebuildIndex rebuilds all indices.
func (r *TemplateRegistry) RebuildIndex(ctx context.Context) error {
	r.index.Clear()

	// Index standard library
	if stdlibTemplates, err := r.stdlib.GetAll(); err == nil {
		for _, tmpl := range stdlibTemplates {
			r.indexTemplate(tmpl)
		}
	}

	// Index repository templates
	tmplType := domain.TemplateElement
	elements, err := r.repo.List(domain.ElementFilter{Type: &tmplType})
	if err != nil {
		return fmt.Errorf("failed to rebuild index: %w", err)
	}

	for _, element := range elements {
		if tmpl, ok := element.(*domain.Template); ok {
			r.indexTemplate(tmpl)
		}
	}

	return nil
}

// GetCacheStats returns cache performance metrics.
func (r *TemplateRegistry) GetCacheStats() CacheStats {
	return r.cache.Stats()
}

// GetIndexStats returns index size metrics.
func (r *TemplateRegistry) GetIndexStats() IndexStats {
	return r.index.Stats()
}

// LoadStandardLibrary loads built-in templates.
func (r *TemplateRegistry) LoadStandardLibrary() error {
	return r.stdlib.Load()
}

// indexTemplate adds a template to all relevant indices.
func (r *TemplateRegistry) indexTemplate(tmpl *domain.Template) {
	metadata := tmpl.GetMetadata()
	id := metadata.ID

	r.index.mu.Lock()
	defer r.index.mu.Unlock()

	// Index by category (if available in tags)
	for _, tag := range metadata.Tags {
		if r.isCategory(tag) {
			r.index.byCategory[tag] = r.appendUnique(r.index.byCategory[tag], id)
		}
		r.index.byTag[tag] = r.appendUnique(r.index.byTag[tag], id)
	}

	// Index by element type (inferred from template purpose)
	elementType := r.inferElementType(tmpl)
	if elementType != "" {
		r.index.byElementType[elementType] = r.appendUnique(r.index.byElementType[elementType], id)
	}

	// Index by author
	if metadata.Author != "" {
		r.index.byAuthor[metadata.Author] = r.appendUnique(r.index.byAuthor[metadata.Author], id)
	}
}

// collectCandidates gathers template IDs matching the filter.
func (r *TemplateRegistry) collectCandidates(filter TemplateSearchFilter) []string {
	r.index.mu.RLock()
	defer r.index.mu.RUnlock()

	var candidates []string

	// Filter by category
	switch {
	case filter.Category != "":
		candidates = r.index.byCategory[filter.Category]
	case filter.ElementType != "":
		candidates = r.index.byElementType[filter.ElementType]
	case filter.Author != "":
		candidates = r.index.byAuthor[filter.Author]
	default:
		// No filter, collect all
		seen := make(map[string]bool)
		for _, ids := range r.index.byCategory {
			for _, id := range ids {
				if !seen[id] {
					candidates = append(candidates, id)
					seen[id] = true
				}
			}
		}
		for _, ids := range r.index.byElementType {
			for _, id := range ids {
				if !seen[id] {
					candidates = append(candidates, id)
					seen[id] = true
				}
			}
		}
	}

	// Filter by tags (intersection)
	if len(filter.Tags) > 0 {
		candidates = r.filterByTags(candidates, filter.Tags)
	}

	// Include built-in templates
	if filter.IncludeBuiltIn {
		stdlibIDs, _ := r.stdlib.GetIDs()
		candidates = r.mergeUnique(candidates, stdlibIDs)
	}

	return candidates
}

// filterByTags filters IDs by tags (all tags must match).
func (r *TemplateRegistry) filterByTags(ids []string, tags []string) []string {
	if len(tags) == 0 {
		return ids
	}

	filtered := make([]string, 0)
	for _, id := range ids {
		matchesAll := true
		for _, tag := range tags {
			tagIDs := r.index.byTag[tag]
			if !r.contains(tagIDs, id) {
				matchesAll = false
				break
			}
		}
		if matchesAll {
			filtered = append(filtered, id)
		}
	}

	return filtered
}

// matchesQuery checks if template matches search query.
func (r *TemplateRegistry) matchesQuery(tmpl *domain.Template, query string) bool {
	metadata := tmpl.GetMetadata()

	// Simple substring search (case-insensitive)
	query = toLower(query)

	if contains(toLower(metadata.Name), query) {
		return true
	}
	if contains(toLower(metadata.Description), query) {
		return true
	}
	for _, tag := range metadata.Tags {
		if contains(toLower(tag), query) {
			return true
		}
	}

	return false
}

// isCategory checks if a tag represents a category.
func (r *TemplateRegistry) isCategory(tag string) bool {
	categories := map[string]bool{
		"persona": true, "skill": true, "agent": true,
		"memory": true, "ensemble": true, "template": true,
	}
	return categories[tag]
}

// inferElementType infers the target element type from template.
func (r *TemplateRegistry) inferElementType(tmpl *domain.Template) string {
	metadata := tmpl.GetMetadata()

	// Check tags for element type hints
	for _, tag := range metadata.Tags {
		if r.isCategory(tag) {
			return tag
		}
	}

	// Check name for hints
	name := toLower(metadata.Name)
	if contains(name, "persona") {
		return "persona"
	}
	if contains(name, "skill") {
		return "skill"
	}
	if contains(name, "agent") {
		return "agent"
	}
	if contains(name, "memory") {
		return "memory"
	}
	if contains(name, "ensemble") {
		return "ensemble"
	}

	return "template"
}

// Utility functions

func (r *TemplateRegistry) appendUnique(slice []string, item string) []string {
	if r.contains(slice, item) {
		return slice
	}
	return append(slice, item)
}

func (r *TemplateRegistry) mergeUnique(a, b []string) []string {
	seen := make(map[string]bool)
	result := make([]string, 0, len(a)+len(b))

	for _, item := range a {
		if !seen[item] {
			result = append(result, item)
			seen[item] = true
		}
	}
	for _, item := range b {
		if !seen[item] {
			result = append(result, item)
			seen[item] = true
		}
	}

	return result
}

func (r *TemplateRegistry) contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

// TemplateCache methods

// Get retrieves a template from cache.
func (c *TemplateCache) Get(id string) (*domain.Template, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	tmpl, exists := c.templates[id]
	if !exists {
		c.mu.RUnlock()
		c.mu.Lock()
		c.misses++
		c.mu.Unlock()
		c.mu.RLock()
		return nil, false
	}

	// Check expiration
	if time.Now().After(c.expires[id]) {
		c.mu.RUnlock()
		c.mu.Lock()
		delete(c.templates, id)
		delete(c.expires, id)
		c.evictions++
		c.misses++
		c.mu.Unlock()
		c.mu.RLock()
		return nil, false
	}

	c.mu.RUnlock()
	c.mu.Lock()
	c.hits++
	c.mu.Unlock()
	c.mu.RLock()

	return tmpl, true
}

// Set adds a template to cache.
func (c *TemplateCache) Set(id string, tmpl *domain.Template) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.templates[id] = tmpl
	c.expires[id] = time.Now().Add(c.ttl)
}

// Delete removes a template from cache.
func (c *TemplateCache) Delete(id string) {
	c.mu.Lock()
	defer c.mu.Unlock()

	delete(c.templates, id)
	delete(c.expires, id)
}

// Clear removes all templates from cache.
func (c *TemplateCache) Clear() {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.templates = make(map[string]*domain.Template)
	c.expires = make(map[string]time.Time)
	c.evictions += uint64(len(c.templates))
}

// Stats returns cache statistics.
func (c *TemplateCache) Stats() CacheStats {
	c.mu.RLock()
	defer c.mu.RUnlock()

	total := c.hits + c.misses
	hitRate := 0.0
	if total > 0 {
		hitRate = float64(c.hits) / float64(total)
	}

	return CacheStats{
		Hits:      c.hits,
		Misses:    c.misses,
		Evictions: c.evictions,
		Size:      len(c.templates),
		HitRate:   hitRate,
	}
}

// TemplateIndex methods

// Clear removes all index entries.
func (i *TemplateIndex) Clear() {
	i.mu.Lock()
	defer i.mu.Unlock()

	i.byCategory = make(map[string][]string)
	i.byTag = make(map[string][]string)
	i.byElementType = make(map[string][]string)
	i.byAuthor = make(map[string][]string)
}

// Stats returns index statistics.
func (i *TemplateIndex) Stats() IndexStats {
	i.mu.RLock()
	defer i.mu.RUnlock()

	// Count unique templates
	seen := make(map[string]bool)
	for _, ids := range i.byCategory {
		for _, id := range ids {
			seen[id] = true
		}
	}

	return IndexStats{
		Categories:     len(i.byCategory),
		Tags:           len(i.byTag),
		ElementTypes:   len(i.byElementType),
		Authors:        len(i.byAuthor),
		TotalTemplates: len(seen),
	}
}

// Helper functions

func toLower(s string) string {
	// Simple ASCII lowercase conversion
	result := make([]rune, len(s))
	for i, r := range s {
		if r >= 'A' && r <= 'Z' {
			result[i] = r + 32
		} else {
			result[i] = r
		}
	}
	return string(result)
}

func contains(s, substr string) bool {
	if substr == "" {
		return true
	}
	if s == "" {
		return false
	}

	// Simple substring search
	sLen := len(s)
	subLen := len(substr)

	if subLen > sLen {
		return false
	}

	for i := 0; i <= sLen-subLen; i++ {
		match := true
		for j := range subLen {
			if s[i+j] != substr[j] {
				match = false
				break
			}
		}
		if match {
			return true
		}
	}

	return false
}
