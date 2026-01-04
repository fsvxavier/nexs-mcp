package application

import (
	"context"
	"sync"
	"time"
)

// AdaptiveCacheService manages cache with adaptive TTL based on access frequency.
type AdaptiveCacheService struct {
	config AdaptiveCacheConfig
	cache  map[string]*CacheEntry
	stats  AdaptiveCacheStats
	mu     sync.RWMutex
}

// AdaptiveCacheConfig configures adaptive caching behavior.
type AdaptiveCacheConfig struct {
	Enabled bool
	MinTTL  time.Duration // Minimum cache TTL (default: 1h)
	MaxTTL  time.Duration // Maximum cache TTL (default: 7 days)
	BaseTTL time.Duration // Baseline cache TTL (default: 24h)
}

// CacheEntry represents a cached item with adaptive TTL.
type CacheEntry struct {
	Key         string
	Value       interface{}
	CreatedAt   time.Time
	LastAccess  time.Time
	AccessCount int64
	TTL         time.Duration
	ExpiresAt   time.Time
	Size        int
}

// AdaptiveCacheStats tracks cache performance metrics.
type AdaptiveCacheStats struct {
	TotalHits      int64
	TotalMisses    int64
	TotalEvictions int64
	TotalEntries   int64
	AvgAccessCount float64
	AvgTTL         time.Duration
	BytesCached    int64
	TTLAdjustments int64
}

// NewAdaptiveCacheService creates a new adaptive cache service.
func NewAdaptiveCacheService(config AdaptiveCacheConfig) *AdaptiveCacheService {
	// Set defaults
	if config.MinTTL == 0 {
		config.MinTTL = 1 * time.Hour
	}
	if config.MaxTTL == 0 {
		config.MaxTTL = 7 * 24 * time.Hour // 7 days
	}
	if config.BaseTTL == 0 {
		config.BaseTTL = 24 * time.Hour
	}

	service := &AdaptiveCacheService{
		config: config,
		cache:  make(map[string]*CacheEntry),
		stats: AdaptiveCacheStats{
			AvgTTL: config.BaseTTL,
		},
	}

	// Start background cleanup goroutine
	if config.Enabled {
		go service.cleanupExpired()
	}

	return service
}

// Get retrieves a value from cache and updates access statistics.
func (s *AdaptiveCacheService) Get(ctx context.Context, key string) (interface{}, bool) {
	if !s.config.Enabled {
		s.mu.Lock()
		s.stats.TotalMisses++
		s.mu.Unlock()
		return nil, false
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	entry, exists := s.cache[key]
	if !exists {
		s.stats.TotalMisses++
		return nil, false
	}

	// Check if expired
	if time.Now().After(entry.ExpiresAt) {
		delete(s.cache, key)
		s.stats.TotalMisses++
		s.stats.TotalEvictions++
		s.stats.TotalEntries--
		return nil, false
	}

	// Update access statistics
	entry.LastAccess = time.Now()
	entry.AccessCount++
	s.stats.TotalHits++

	// Adjust TTL based on access frequency
	s.adjustTTL(entry)

	return entry.Value, true
}

// Set stores a value in cache with adaptive TTL.
func (s *AdaptiveCacheService) Set(ctx context.Context, key string, value interface{}, size int) error {
	if !s.config.Enabled {
		return nil
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	now := time.Now()

	// Check if entry already exists
	if existing, exists := s.cache[key]; exists {
		// Update existing entry
		existing.Value = value
		existing.LastAccess = now
		existing.AccessCount++
		existing.Size = size
		s.adjustTTL(existing)
		return nil
	}

	// Create new entry with base TTL
	entry := &CacheEntry{
		Key:         key,
		Value:       value,
		CreatedAt:   now,
		LastAccess:  now,
		AccessCount: 1,
		TTL:         s.config.BaseTTL,
		ExpiresAt:   now.Add(s.config.BaseTTL),
		Size:        size,
	}

	s.cache[key] = entry
	s.stats.TotalEntries++
	s.stats.BytesCached += int64(size)

	return nil
}

// adjustTTL dynamically adjusts TTL based on access frequency.
// More frequently accessed items get longer TTL (up to MaxTTL).
// Less frequently accessed items get shorter TTL (down to MinTTL).
func (s *AdaptiveCacheService) adjustTTL(entry *CacheEntry) {
	// Calculate access frequency (accesses per hour)
	age := time.Since(entry.CreatedAt)
	if age < time.Hour {
		age = time.Hour // Avoid division by zero
	}
	accessFrequency := float64(entry.AccessCount) / age.Hours()

	// Adjust TTL based on access frequency
	// High frequency (>10 accesses/hour) → MaxTTL
	// Medium frequency (1-10 accesses/hour) → BaseTTL
	// Low frequency (<1 access/hour) → MinTTL
	var newTTL time.Duration
	switch {
	case accessFrequency >= 10:
		newTTL = s.config.MaxTTL
	case accessFrequency >= 1:
		// Linear interpolation between BaseTTL and MaxTTL
		ratio := (accessFrequency - 1.0) / 9.0 // 0.0 to 1.0
		newTTL = s.config.BaseTTL + time.Duration(float64(s.config.MaxTTL-s.config.BaseTTL)*ratio)
	default:
		// Linear interpolation between MinTTL and BaseTTL
		ratio := accessFrequency // 0.0 to 1.0
		newTTL = s.config.MinTTL + time.Duration(float64(s.config.BaseTTL-s.config.MinTTL)*ratio)
	}

	// Only update if TTL changed significantly (>10% difference)
	if oldTTL := entry.TTL; newTTL > oldTTL*11/10 || newTTL < oldTTL*9/10 {
		entry.TTL = newTTL
		entry.ExpiresAt = time.Now().Add(newTTL)
		s.stats.TTLAdjustments++
	}
}

// Delete removes an entry from cache.
func (s *AdaptiveCacheService) Delete(ctx context.Context, key string) {
	if !s.config.Enabled {
		return
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	if entry, exists := s.cache[key]; exists {
		s.stats.BytesCached -= int64(entry.Size)
		s.stats.TotalEntries--
		delete(s.cache, key)
	}
}

// Clear removes all entries from cache.
func (s *AdaptiveCacheService) Clear(ctx context.Context) {
	if !s.config.Enabled {
		return
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	s.cache = make(map[string]*CacheEntry)
	s.stats.TotalEntries = 0
	s.stats.BytesCached = 0
}

// GetStats returns current cache statistics.
func (s *AdaptiveCacheService) GetStats() AdaptiveCacheStats {
	s.mu.RLock()
	defer s.mu.RUnlock()

	// Calculate average access count
	totalAccesses := int64(0)
	totalTTL := time.Duration(0)
	count := int64(0)

	for _, entry := range s.cache {
		totalAccesses += entry.AccessCount
		totalTTL += entry.TTL
		count++
	}

	stats := s.stats
	if count > 0 {
		stats.AvgAccessCount = float64(totalAccesses) / float64(count)
		stats.AvgTTL = totalTTL / time.Duration(count)
	}

	return stats
}

// GetHitRate calculates cache hit rate.
func (s *AdaptiveCacheService) GetHitRate() float64 {
	s.mu.RLock()
	defer s.mu.RUnlock()

	total := s.stats.TotalHits + s.stats.TotalMisses
	if total == 0 {
		return 0.0
	}
	return float64(s.stats.TotalHits) / float64(total)
}

// cleanupExpired periodically removes expired entries.
func (s *AdaptiveCacheService) cleanupExpired() {
	ticker := time.NewTicker(1 * time.Minute)
	defer ticker.Stop()

	for range ticker.C {
		s.mu.Lock()
		now := time.Now()
		for key, entry := range s.cache {
			if now.After(entry.ExpiresAt) {
				s.stats.BytesCached -= int64(entry.Size)
				s.stats.TotalEntries--
				s.stats.TotalEvictions++
				delete(s.cache, key)
			}
		}
		s.mu.Unlock()
	}
}

// Size returns the number of entries in cache.
func (s *AdaptiveCacheService) Size() int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return len(s.cache)
}

// GetEntry retrieves cache entry metadata (for debugging/monitoring).
func (s *AdaptiveCacheService) GetEntry(key string) (*CacheEntry, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	entry, exists := s.cache[key]
	if !exists {
		return nil, false
	}

	// Return a copy to avoid data races
	entryCopy := *entry
	return &entryCopy, true
}
