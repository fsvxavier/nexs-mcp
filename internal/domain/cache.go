package domain

import "context"

// CacheService defines a generic interface for caching services.
// This avoids import cycles between infrastructure and application layers.
type CacheService interface {
	// Get retrieves a value from cache
	Get(ctx context.Context, key string) (interface{}, bool)

	// Set stores a value in cache with an estimated size in bytes
	Set(ctx context.Context, key string, value interface{}, sizeBytes int) error

	// Delete removes a value from cache
	Delete(ctx context.Context, key string)
}
