package domain

import (
	"context"
	"fmt"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// MockCacheService implements CacheService for testing.
type MockCacheService struct {
	data map[string]interface{}
	mu   sync.RWMutex
}

func NewMockCacheService() *MockCacheService {
	return &MockCacheService{
		data: make(map[string]interface{}),
	}
}

func (m *MockCacheService) Get(ctx context.Context, key string) (interface{}, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	val, ok := m.data[key]
	return val, ok
}

func (m *MockCacheService) Set(ctx context.Context, key string, value interface{}, sizeBytes int) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.data[key] = value
	return nil
}

func (m *MockCacheService) Delete(ctx context.Context, key string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	delete(m.data, key)
}

func TestMockCacheService_GetSet(t *testing.T) {
	cache := NewMockCacheService()
	ctx := context.Background()

	// Test Set and Get
	err := cache.Set(ctx, "key1", "value1", 100)
	require.NoError(t, err)

	val, ok := cache.Get(ctx, "key1")
	assert.True(t, ok)
	assert.Equal(t, "value1", val)

	// Test Get non-existent key
	val, ok = cache.Get(ctx, "non-existent")
	assert.False(t, ok)
	assert.Nil(t, val)
}

func TestMockCacheService_Delete(t *testing.T) {
	cache := NewMockCacheService()
	ctx := context.Background()

	// Set a value
	err := cache.Set(ctx, "key1", "value1", 100)
	require.NoError(t, err)

	// Verify it exists
	_, ok := cache.Get(ctx, "key1")
	assert.True(t, ok)

	// Delete it
	cache.Delete(ctx, "key1")

	// Verify it's gone
	_, ok = cache.Get(ctx, "key1")
	assert.False(t, ok)
}

func TestMockCacheService_MultipleKeys(t *testing.T) {
	cache := NewMockCacheService()
	ctx := context.Background()

	// Set multiple values
	for i := range 10 {
		key := fmt.Sprintf("key%d", i)
		value := fmt.Sprintf("value%d", i)
		err := cache.Set(ctx, key, value, 100)
		require.NoError(t, err)
	}

	// Verify all exist
	for i := range 10 {
		key := fmt.Sprintf("key%d", i)
		expectedValue := fmt.Sprintf("value%d", i)
		val, ok := cache.Get(ctx, key)
		assert.True(t, ok)
		assert.Equal(t, expectedValue, val)
	}
}

func TestMockCacheService_Overwrite(t *testing.T) {
	cache := NewMockCacheService()
	ctx := context.Background()

	// Set initial value
	err := cache.Set(ctx, "key1", "value1", 100)
	require.NoError(t, err)

	// Overwrite with new value
	err = cache.Set(ctx, "key1", "value2", 100)
	require.NoError(t, err)

	// Verify new value
	val, ok := cache.Get(ctx, "key1")
	assert.True(t, ok)
	assert.Equal(t, "value2", val)
}

func TestMockCacheService_ConcurrentAccess(t *testing.T) {
	cache := NewMockCacheService()
	ctx := context.Background()
	var wg sync.WaitGroup

	// Concurrent writes
	numGoroutines := 100
	wg.Add(numGoroutines)
	for i := range numGoroutines {
		go func(idx int) {
			defer wg.Done()
			key := fmt.Sprintf("key%d", idx)
			value := fmt.Sprintf("value%d", idx)
			_ = cache.Set(ctx, key, value, 100)
		}(i)
	}
	wg.Wait()

	// Concurrent reads
	wg.Add(numGoroutines)
	for i := range numGoroutines {
		go func(idx int) {
			defer wg.Done()
			key := fmt.Sprintf("key%d", idx)
			_, _ = cache.Get(ctx, key)
		}(i)
	}
	wg.Wait()

	// Verify data integrity
	for i := range numGoroutines {
		key := fmt.Sprintf("key%d", i)
		expectedValue := fmt.Sprintf("value%d", i)
		val, ok := cache.Get(ctx, key)
		assert.True(t, ok, "Key %s should exist", key)
		assert.Equal(t, expectedValue, val, "Value for %s should match", key)
	}
}

func TestMockCacheService_DifferentValueTypes(t *testing.T) {
	cache := NewMockCacheService()
	ctx := context.Background()

	tests := []struct {
		name  string
		key   string
		value interface{}
	}{
		{"string", "key1", "string value"},
		{"int", "key2", 12345},
		{"float", "key3", 123.45},
		{"bool", "key4", true},
		{"slice", "key5", []string{"a", "b", "c"}},
		{"map", "key6", map[string]int{"a": 1, "b": 2}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := cache.Set(ctx, tt.key, tt.value, 100)
			require.NoError(t, err)

			val, ok := cache.Get(ctx, tt.key)
			assert.True(t, ok)
			assert.Equal(t, tt.value, val)
		})
	}
}
