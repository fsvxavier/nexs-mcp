package hnsw

import (
	"math"
)

// CosineSimilarity computes cosine similarity between two vectors
// Returns distance (1 - similarity) so lower values = more similar
func CosineSimilarity(a, b []float32) float32 {
	if len(a) != len(b) {
		return math.MaxFloat32
	}

	var dotProduct, normA, normB float32
	for i := 0; i < len(a); i++ {
		dotProduct += a[i] * b[i]
		normA += a[i] * a[i]
		normB += b[i] * b[i]
	}

	if normA == 0 || normB == 0 {
		return 1.0 // Maximum distance
	}

	similarity := dotProduct / (float32(math.Sqrt(float64(normA))) * float32(math.Sqrt(float64(normB))))

	// Convert similarity to distance: [1, -1] -> [0, 2]
	// Higher similarity = lower distance
	return 1.0 - similarity
}

// EuclideanDistance computes Euclidean distance between two vectors
func EuclideanDistance(a, b []float32) float32 {
	if len(a) != len(b) {
		return math.MaxFloat32
	}

	var sum float32
	for i := 0; i < len(a); i++ {
		diff := a[i] - b[i]
		sum += diff * diff
	}

	return float32(math.Sqrt(float64(sum)))
}

// DotProduct computes negative dot product as distance
// (negative because higher dot product = more similar)
func DotProduct(a, b []float32) float32 {
	if len(a) != len(b) {
		return math.MaxFloat32
	}

	var sum float32
	for i := 0; i < len(a); i++ {
		sum += a[i] * b[i]
	}

	return -sum // Negative so higher similarity = lower distance
}

// ManhattanDistance computes Manhattan (L1) distance
func ManhattanDistance(a, b []float32) float32 {
	if len(a) != len(b) {
		return math.MaxFloat32
	}

	var sum float32
	for i := 0; i < len(a); i++ {
		diff := a[i] - b[i]
		if diff < 0 {
			diff = -diff
		}
		sum += diff
	}

	return sum
}
