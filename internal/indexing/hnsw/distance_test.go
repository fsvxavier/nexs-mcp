package hnsw

import (
	"math"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCosineSimilarity(t *testing.T) {
	tests := []struct {
		name     string
		a        []float32
		b        []float32
		expected float32
		delta    float64
	}{
		{
			name:     "identical vectors",
			a:        []float32{1.0, 0.0, 0.0},
			b:        []float32{1.0, 0.0, 0.0},
			expected: 0.0, // distance = 1 - similarity = 1 - 1 = 0
			delta:    0.001,
		},
		{
			name:     "orthogonal vectors",
			a:        []float32{1.0, 0.0, 0.0},
			b:        []float32{0.0, 1.0, 0.0},
			expected: 1.0, // distance = 1 - similarity = 1 - 0 = 1
			delta:    0.001,
		},
		{
			name:     "opposite vectors",
			a:        []float32{1.0, 0.0, 0.0},
			b:        []float32{-1.0, 0.0, 0.0},
			expected: 2.0, // distance = 1 - similarity = 1 - (-1) = 2
			delta:    0.001,
		},
		{
			name:     "similar vectors",
			a:        []float32{1.0, 2.0, 3.0},
			b:        []float32{1.1, 2.1, 3.1},
			expected: 0.001, // very similar, distance near 0
			delta:    0.01,
		},
		{
			name:     "zero vector a",
			a:        []float32{0.0, 0.0, 0.0},
			b:        []float32{1.0, 2.0, 3.0},
			expected: 1.0, // max distance
			delta:    0.001,
		},
		{
			name:     "zero vector b",
			a:        []float32{1.0, 2.0, 3.0},
			b:        []float32{0.0, 0.0, 0.0},
			expected: 1.0, // max distance
			delta:    0.001,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := CosineSimilarity(tt.a, tt.b)
			assert.InDelta(t, tt.expected, result, tt.delta)
		})
	}
}

func TestCosineSimilarityDifferentLengths(t *testing.T) {
	a := []float32{1.0, 2.0, 3.0}
	b := []float32{1.0, 2.0}

	result := CosineSimilarity(a, b)
	assert.Equal(t, float32(math.MaxFloat32), result)
}

func TestEuclideanDistance(t *testing.T) {
	tests := []struct {
		name     string
		a        []float32
		b        []float32
		expected float32
		delta    float64
	}{
		{
			name:     "identical vectors",
			a:        []float32{1.0, 2.0, 3.0},
			b:        []float32{1.0, 2.0, 3.0},
			expected: 0.0,
			delta:    0.001,
		},
		{
			name:     "unit distance",
			a:        []float32{0.0, 0.0, 0.0},
			b:        []float32{1.0, 0.0, 0.0},
			expected: 1.0,
			delta:    0.001,
		},
		{
			name:     "3-4-5 triangle",
			a:        []float32{0.0, 0.0},
			b:        []float32{3.0, 4.0},
			expected: 5.0,
			delta:    0.001,
		},
		{
			name:     "negative coordinates",
			a:        []float32{-1.0, -1.0},
			b:        []float32{1.0, 1.0},
			expected: 2.828, // sqrt(8)
			delta:    0.001,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := EuclideanDistance(tt.a, tt.b)
			assert.InDelta(t, tt.expected, result, tt.delta)
		})
	}
}

func TestEuclideanDistanceDifferentLengths(t *testing.T) {
	a := []float32{1.0, 2.0, 3.0}
	b := []float32{1.0, 2.0}

	result := EuclideanDistance(a, b)
	assert.Equal(t, float32(math.MaxFloat32), result)
}

func TestDotProduct(t *testing.T) {
	tests := []struct {
		name     string
		a        []float32
		b        []float32
		expected float32
		delta    float64
	}{
		{
			name:     "orthogonal vectors",
			a:        []float32{1.0, 0.0, 0.0},
			b:        []float32{0.0, 1.0, 0.0},
			expected: 0.0, // dot product is 0, distance is -0 = 0
			delta:    0.001,
		},
		{
			name:     "parallel vectors",
			a:        []float32{1.0, 2.0, 3.0},
			b:        []float32{1.0, 2.0, 3.0},
			expected: -14.0, // -(1+4+9) = -14
			delta:    0.001,
		},
		{
			name:     "simple case",
			a:        []float32{1.0, 2.0},
			b:        []float32{3.0, 4.0},
			expected: -11.0, // -(3+8) = -11
			delta:    0.001,
		},
		{
			name:     "negative values",
			a:        []float32{-1.0, 2.0},
			b:        []float32{3.0, -4.0},
			expected: 11.0, // -(-3-8) = 11
			delta:    0.001,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := DotProduct(tt.a, tt.b)
			assert.InDelta(t, tt.expected, result, tt.delta)
		})
	}
}

func TestDotProductDifferentLengths(t *testing.T) {
	a := []float32{1.0, 2.0, 3.0}
	b := []float32{1.0, 2.0}

	result := DotProduct(a, b)
	assert.Equal(t, float32(math.MaxFloat32), result)
}

func TestManhattanDistance(t *testing.T) {
	tests := []struct {
		name     string
		a        []float32
		b        []float32
		expected float32
		delta    float64
	}{
		{
			name:     "identical vectors",
			a:        []float32{1.0, 2.0, 3.0},
			b:        []float32{1.0, 2.0, 3.0},
			expected: 0.0,
			delta:    0.001,
		},
		{
			name:     "unit distance x",
			a:        []float32{0.0, 0.0},
			b:        []float32{1.0, 0.0},
			expected: 1.0,
			delta:    0.001,
		},
		{
			name:     "city block distance",
			a:        []float32{0.0, 0.0},
			b:        []float32{3.0, 4.0},
			expected: 7.0, // |3-0| + |4-0| = 7
			delta:    0.001,
		},
		{
			name:     "negative coordinates",
			a:        []float32{-2.0, -3.0},
			b:        []float32{2.0, 3.0},
			expected: 10.0, // |2-(-2)| + |3-(-3)| = 4 + 6 = 10
			delta:    0.001,
		},
		{
			name:     "multi-dimensional",
			a:        []float32{1.0, 2.0, 3.0, 4.0},
			b:        []float32{5.0, 6.0, 7.0, 8.0},
			expected: 16.0, // 4+4+4+4 = 16
			delta:    0.001,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ManhattanDistance(tt.a, tt.b)
			assert.InDelta(t, tt.expected, result, tt.delta)
		})
	}
}

func TestManhattanDistanceDifferentLengths(t *testing.T) {
	a := []float32{1.0, 2.0, 3.0}
	b := []float32{1.0, 2.0}

	result := ManhattanDistance(a, b)
	assert.Equal(t, float32(math.MaxFloat32), result)
}

func BenchmarkCosineSimilarity(b *testing.B) {
	a := generateRandomVector(384)
	vec := generateRandomVector(384)

	b.ResetTimer()
	for range b.N {
		_ = CosineSimilarity(a, vec)
	}
}

func BenchmarkEuclideanDistance(b *testing.B) {
	a := generateRandomVector(384)
	vec := generateRandomVector(384)

	b.ResetTimer()
	for range b.N {
		_ = EuclideanDistance(a, vec)
	}
}

func BenchmarkDotProduct(b *testing.B) {
	a := generateRandomVector(384)
	vec := generateRandomVector(384)

	b.ResetTimer()
	for range b.N {
		_ = DotProduct(a, vec)
	}
}

func BenchmarkManhattanDistance(b *testing.B) {
	a := generateRandomVector(384)
	vec := generateRandomVector(384)

	b.ResetTimer()
	for range b.N {
		_ = ManhattanDistance(a, vec)
	}
}
