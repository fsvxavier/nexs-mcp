package infrastructure

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/fsvxavier/nexs-mcp/internal/domain"
	"gopkg.in/yaml.v3"
)

// LRUCache implements a simple Least Recently Used cache
type LRUCache struct {
	mu       sync.RWMutex
	capacity int
	cache    map[string]*cacheNode
	head     *cacheNode
	tail     *cacheNode
}

type cacheNode struct {
	key   string
	value *StoredElement
	prev  *cacheNode
	next  *cacheNode
}

// NewLRUCache creates a new LRU cache with given capacity
func NewLRUCache(capacity int) *LRUCache {
	lru := &LRUCache{
		capacity: capacity,
		cache:    make(map[string]*cacheNode),
	}
	// Initialize dummy head and tail
	lru.head = &cacheNode{}
	lru.tail = &cacheNode{}
	lru.head.next = lru.tail
	lru.tail.prev = lru.head
	return lru
}

// Get retrieves a value from the cache
func (lru *LRUCache) Get(key string) (*StoredElement, bool) {
	lru.mu.Lock()
	defer lru.mu.Unlock()

	if node, exists := lru.cache[key]; exists {
		lru.moveToFront(node)
		return node.value, true
	}
	return nil, false
}

// Put adds or updates a value in the cache
func (lru *LRUCache) Put(key string, value *StoredElement) {
	lru.mu.Lock()
	defer lru.mu.Unlock()

	if node, exists := lru.cache[key]; exists {
		node.value = value
		lru.moveToFront(node)
		return
	}

	// Create new node
	newNode := &cacheNode{key: key, value: value}
	lru.cache[key] = newNode
	lru.addToFront(newNode)

	// Evict if capacity exceeded
	if len(lru.cache) > lru.capacity {
		lru.evictLRU()
	}
}

// Delete removes a value from the cache
func (lru *LRUCache) Delete(key string) {
	lru.mu.Lock()
	defer lru.mu.Unlock()

	if node, exists := lru.cache[key]; exists {
		lru.removeNode(node)
		delete(lru.cache, key)
	}
}

// Clear removes all entries from the cache
func (lru *LRUCache) Clear() {
	lru.mu.Lock()
	defer lru.mu.Unlock()

	lru.cache = make(map[string]*cacheNode)
	lru.head.next = lru.tail
	lru.tail.prev = lru.head
}

func (lru *LRUCache) moveToFront(node *cacheNode) {
	lru.removeNode(node)
	lru.addToFront(node)
}

func (lru *LRUCache) addToFront(node *cacheNode) {
	node.next = lru.head.next
	node.prev = lru.head
	lru.head.next.prev = node
	lru.head.next = node
}

func (lru *LRUCache) removeNode(node *cacheNode) {
	node.prev.next = node.next
	node.next.prev = node.prev
}

func (lru *LRUCache) evictLRU() {
	toEvict := lru.tail.prev
	if toEvict == lru.head {
		return // No nodes to evict
	}
	lru.removeNode(toEvict)
	delete(lru.cache, toEvict.key)
}

// SearchIndex maintains an inverted index for full-text search
type SearchIndex struct {
	mu    sync.RWMutex
	index map[string][]string // word -> []elementIDs
}

// NewSearchIndex creates a new search index
func NewSearchIndex() *SearchIndex {
	return &SearchIndex{
		index: make(map[string][]string),
	}
}

// Index adds an element to the search index
func (si *SearchIndex) Index(element domain.Element) {
	si.mu.Lock()
	defer si.mu.Unlock()

	metadata := element.GetMetadata()
	words := si.extractWords(metadata)

	for _, word := range words {
		if !si.contains(si.index[word], metadata.ID) {
			si.index[word] = append(si.index[word], metadata.ID)
		}
	}
}

// Remove removes an element from the search index
func (si *SearchIndex) Remove(elementID string) {
	si.mu.Lock()
	defer si.mu.Unlock()

	for word, ids := range si.index {
		si.index[word] = si.removeID(ids, elementID)
	}
}

// Search performs a full-text search and returns matching element IDs
func (si *SearchIndex) Search(query string) []string {
	si.mu.RLock()
	defer si.mu.RUnlock()

	words := si.tokenize(query)
	if len(words) == 0 {
		return []string{}
	}

	// Get IDs for first word
	results := make(map[string]int)
	for _, word := range words {
		if ids, exists := si.index[strings.ToLower(word)]; exists {
			for _, id := range ids {
				results[id]++
			}
		}
	}

	// Convert to slice and sort by relevance (word count)
	var sortedResults []string
	for id := range results {
		sortedResults = append(sortedResults, id)
	}

	return sortedResults
}

func (si *SearchIndex) extractWords(metadata domain.ElementMetadata) []string {
	text := fmt.Sprintf("%s %s %s", metadata.Name, metadata.Description, strings.Join(metadata.Tags, " "))
	return si.tokenize(text)
}

func (si *SearchIndex) tokenize(text string) []string {
	words := strings.Fields(strings.ToLower(text))
	var tokens []string
	for _, word := range words {
		// Remove punctuation
		word = strings.Trim(word, ".,!?;:")
		if len(word) > 2 { // Ignore very short words
			tokens = append(tokens, word)
		}
	}
	return tokens
}

func (si *SearchIndex) contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

func (si *SearchIndex) removeID(slice []string, item string) []string {
	var result []string
	for _, s := range slice {
		if s != item {
			result = append(result, s)
		}
	}
	return result
}

// EnhancedFileElementRepository extends FileElementRepository with advanced features
type EnhancedFileElementRepository struct {
	mu          sync.RWMutex
	baseDir     string
	lruCache    *LRUCache
	searchIndex *SearchIndex
	index       map[string]*StoredElement // Full index for metadata queries
}

// NewEnhancedFileElementRepository creates a new enhanced file-based repository
func NewEnhancedFileElementRepository(baseDir string, cacheSize int) (*EnhancedFileElementRepository, error) {
	if baseDir == "" {
		baseDir = "data/elements"
	}

	if cacheSize <= 0 {
		cacheSize = 100 // Default cache size
	}

	// Create base directory if it doesn't exist
	if err := os.MkdirAll(baseDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create base directory: %w", err)
	}

	repo := &EnhancedFileElementRepository{
		baseDir:     baseDir,
		lruCache:    NewLRUCache(cacheSize),
		searchIndex: NewSearchIndex(),
		index:       make(map[string]*StoredElement),
	}

	// Load existing elements into index
	if err := repo.loadIndex(); err != nil {
		return nil, fmt.Errorf("failed to load index: %w", err)
	}

	return repo, nil
}

// loadIndex loads all existing elements into the index and search index
func (r *EnhancedFileElementRepository) loadIndex() error {
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

		r.index[stored.Metadata.ID] = &stored
		r.searchIndex.Index(&SimpleElement{metadata: stored.Metadata})
		return nil
	})
}

// getFilePath returns the file path for an element with user-specific support
// Structure: baseDir/author/type/YYYY-MM-DD/id.yaml
func (r *EnhancedFileElementRepository) getFilePath(metadata domain.ElementMetadata) string {
	author := metadata.Author
	if author == "" {
		author = "default"
	}

	// Support private user directories
	authorDir := author
	if strings.HasPrefix(author, "private-") {
		authorDir = filepath.Join("private", strings.TrimPrefix(author, "private-"))
	}

	dateDir := metadata.CreatedAt.Format("2006-01-02")
	typeDir := string(metadata.Type)
	filename := fmt.Sprintf("%s.yaml", metadata.ID)
	return filepath.Join(r.baseDir, authorDir, typeDir, dateDir, filename)
}

// Create creates a new element with atomic write
func (r *EnhancedFileElementRepository) Create(element domain.Element) error {
	if element == nil {
		return fmt.Errorf("element cannot be nil")
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	metadata := element.GetMetadata()
	if _, exists := r.index[metadata.ID]; exists {
		return fmt.Errorf("element with ID %s already exists", metadata.ID)
	}

	stored := &StoredElement{Metadata: metadata}

	// Save to file atomically
	if err := r.saveToFileAtomic(stored); err != nil {
		return err
	}

	// Update index
	r.index[metadata.ID] = stored
	r.lruCache.Put(metadata.ID, stored)
	r.searchIndex.Index(element)
	return nil
}

// GetByID retrieves an element by ID using LRU cache
func (r *EnhancedFileElementRepository) GetByID(id string) (domain.Element, error) {
	// Try LRU cache first
	if stored, found := r.lruCache.Get(id); found {
		return r.convertToTypedElement(stored)
	}

	r.mu.RLock()
	stored, exists := r.index[id]
	r.mu.RUnlock()

	if !exists {
		return nil, fmt.Errorf("element with ID %s not found", id)
	}

	// Add to LRU cache
	r.lruCache.Put(id, stored)

	return r.convertToTypedElement(stored)
}

// Update updates an existing element with atomic write
func (r *EnhancedFileElementRepository) Update(element domain.Element) error {
	if element == nil {
		return fmt.Errorf("element cannot be nil")
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	metadata := element.GetMetadata()
	old, exists := r.index[metadata.ID]
	if !exists {
		return fmt.Errorf("element with ID %s not found", metadata.ID)
	}

	// Delete old file if path changed
	oldPath := r.getFilePath(old.Metadata)
	newPath := r.getFilePath(metadata)

	if oldPath != newPath {
		if err := os.Remove(oldPath); err != nil && !os.IsNotExist(err) {
			return fmt.Errorf("failed to remove old file: %w", err)
		}
	}

	stored := &StoredElement{Metadata: metadata}

	// Save to file atomically
	if err := r.saveToFileAtomic(stored); err != nil {
		return err
	}

	// Update index and cache
	r.index[metadata.ID] = stored
	r.lruCache.Put(metadata.ID, stored)
	r.searchIndex.Remove(metadata.ID)
	r.searchIndex.Index(element)
	return nil
}

// Delete removes an element by ID
func (r *EnhancedFileElementRepository) Delete(id string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	stored, exists := r.index[id]
	if !exists {
		return fmt.Errorf("element with ID %s not found", id)
	}

	// Delete file
	path := r.getFilePath(stored.Metadata)
	if err := os.Remove(path); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to remove file: %w", err)
	}

	// Remove from index, cache, and search index
	delete(r.index, id)
	r.lruCache.Delete(id)
	r.searchIndex.Remove(id)
	return nil
}

// List returns all elements matching the filter criteria
func (r *EnhancedFileElementRepository) List(filter domain.ElementFilter) ([]domain.Element, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var results []domain.Element

	for _, stored := range r.index {
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

		elem, err := r.convertToTypedElement(stored)
		if err != nil {
			continue
		}
		results = append(results, elem)
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

// Search performs full-text search on elements
func (r *EnhancedFileElementRepository) Search(query string, filter domain.ElementFilter) ([]domain.Element, error) {
	// Get matching IDs from search index
	matchingIDs := r.searchIndex.Search(query)

	r.mu.RLock()
	defer r.mu.RUnlock()

	var results []domain.Element

	for _, id := range matchingIDs {
		stored, exists := r.index[id]
		if !exists {
			continue
		}

		metadata := stored.Metadata

		// Apply filters
		if filter.Type != nil && metadata.Type != *filter.Type {
			continue
		}

		if filter.IsActive != nil && metadata.IsActive != *filter.IsActive {
			continue
		}

		elem, err := r.convertToTypedElement(stored)
		if err != nil {
			continue
		}
		results = append(results, elem)
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
func (r *EnhancedFileElementRepository) Exists(id string) (bool, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	_, exists := r.index[id]
	return exists, nil
}

// Backup creates a backup of the repository
func (r *EnhancedFileElementRepository) Backup(backupDir string) error {
	r.mu.RLock()
	defer r.mu.RUnlock()

	timestamp := time.Now().Format("20060102-150405")
	backupPath := filepath.Join(backupDir, fmt.Sprintf("backup-%s", timestamp))

	if err := os.MkdirAll(backupPath, 0755); err != nil {
		return fmt.Errorf("failed to create backup directory: %w", err)
	}

	// Copy all files
	return filepath.Walk(r.baseDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() {
			return nil
		}

		relPath, err := filepath.Rel(r.baseDir, path)
		if err != nil {
			return err
		}

		destPath := filepath.Join(backupPath, relPath)
		destDir := filepath.Dir(destPath)

		if err := os.MkdirAll(destDir, 0755); err != nil {
			return err
		}

		data, err := os.ReadFile(path)
		if err != nil {
			return err
		}

		return os.WriteFile(destPath, data, info.Mode())
	})
}

// Restore restores the repository from a backup
func (r *EnhancedFileElementRepository) Restore(backupPath string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	// Clear current data
	if err := os.RemoveAll(r.baseDir); err != nil {
		return fmt.Errorf("failed to clear current data: %w", err)
	}

	if err := os.MkdirAll(r.baseDir, 0755); err != nil {
		return fmt.Errorf("failed to recreate base directory: %w", err)
	}

	// Copy backup files
	if err := filepath.Walk(backupPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() {
			return nil
		}

		relPath, err := filepath.Rel(backupPath, path)
		if err != nil {
			return err
		}

		destPath := filepath.Join(r.baseDir, relPath)
		destDir := filepath.Dir(destPath)

		if err := os.MkdirAll(destDir, 0755); err != nil {
			return err
		}

		data, err := os.ReadFile(path)
		if err != nil {
			return err
		}

		return os.WriteFile(destPath, data, info.Mode())
	}); err != nil {
		return err
	}

	// Reload index
	r.index = make(map[string]*StoredElement)
	r.lruCache.Clear()
	r.searchIndex = NewSearchIndex()

	return r.loadIndex()
}

// saveToFileAtomic saves an element to a YAML file atomically using temp file + rename
func (r *EnhancedFileElementRepository) saveToFileAtomic(stored *StoredElement) error {
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

	// Write to temporary file
	tempPath := path + ".tmp." + generateRandomSuffix()
	if err := os.WriteFile(tempPath, data, 0644); err != nil {
		return fmt.Errorf("failed to write temp file: %w", err)
	}

	// Atomic rename
	if err := os.Rename(tempPath, path); err != nil {
		os.Remove(tempPath) // Clean up temp file on error
		return fmt.Errorf("failed to rename temp file: %w", err)
	}

	return nil
}

// convertToTypedElement converts StoredElement to proper typed element
func (r *EnhancedFileElementRepository) convertToTypedElement(stored *StoredElement) (domain.Element, error) {
	metadata := stored.Metadata

	switch metadata.Type {
	case domain.PersonaElement:
		persona := domain.NewPersona(metadata.Name, metadata.Description, metadata.Version, metadata.Author)
		persona.SetMetadata(metadata)
		return persona, nil

	case domain.SkillElement:
		skill := domain.NewSkill(metadata.Name, metadata.Description, metadata.Version, metadata.Author)
		skill.SetMetadata(metadata)
		return skill, nil

	case domain.TemplateElement:
		template := domain.NewTemplate(metadata.Name, metadata.Description, metadata.Version, metadata.Author)
		template.SetMetadata(metadata)
		return template, nil

	case domain.AgentElement:
		agent := domain.NewAgent(metadata.Name, metadata.Description, metadata.Version, metadata.Author)
		agent.SetMetadata(metadata)
		return agent, nil

	case domain.MemoryElement:
		memory := domain.NewMemory(metadata.Name, metadata.Description, metadata.Version, metadata.Author)
		memory.SetMetadata(metadata)
		return memory, nil

	case domain.EnsembleElement:
		ensemble := domain.NewEnsemble(metadata.Name, metadata.Description, metadata.Version, metadata.Author)
		ensemble.SetMetadata(metadata)
		return ensemble, nil

	default:
		return &SimpleElement{metadata: metadata}, nil
	}
}

// generateRandomSuffix generates a random suffix for temp files
func generateRandomSuffix() string {
	data := []byte(time.Now().String())
	hash := sha256.Sum256(data)
	return hex.EncodeToString(hash[:8])
}
