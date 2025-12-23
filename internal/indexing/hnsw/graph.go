// Package hnsw implements Hierarchical Navigable Small World (HNSW) algorithm
// for approximate nearest neighbor search.
//
// HNSW is a graph-based index structure that provides:
// - Sub-50ms queries for 10k+ vectors
// - Sub-200ms queries for 100k+ vectors
// - >95% accuracy vs linear search
// - Low memory overhead (~50MB for 10k vectors @ 384 dims)
package hnsw

import (
	"container/heap"
	"math"
	"math/rand"
	"sync"
)

// Parameters for HNSW index
const (
	// M is the number of bi-directional links created for every new element during construction.
	// Reasonable range: 5-48. Higher values = better recall, more memory.
	DefaultM = 16

	// efConstruction controls the index construction time/accuracy trade-off.
	// Higher values = better index quality, slower construction.
	DefaultEfConstruction = 200

	// efSearch controls the search time/accuracy trade-off.
	// Higher values = better recall, slower search.
	DefaultEfSearch = 50
)

// DefaultML is the normalization factor for level assignment.
var DefaultML = 1.0 / math.Log(2.0)

// Node represents a node in the HNSW graph
type Node struct {
	ID        string          // Unique identifier
	Vector    []float32       // Embedding vector
	Level     int             // Maximum level this node appears in
	Neighbors map[int][]*Node // Neighbors at each level: level -> []*Node
	mu        sync.RWMutex    // Protects concurrent access to Neighbors
}

// NewNode creates a new HNSW node
func NewNode(id string, vector []float32, level int) *Node {
	return &Node{
		ID:        id,
		Vector:    vector,
		Level:     level,
		Neighbors: make(map[int][]*Node),
	}
}

// AddNeighbor adds a neighbor at the specified level
func (n *Node) AddNeighbor(level int, neighbor *Node) {
	n.mu.Lock()
	defer n.mu.Unlock()

	// Avoid duplicates
	for _, existing := range n.Neighbors[level] {
		if existing.ID == neighbor.ID {
			return
		}
	}

	n.Neighbors[level] = append(n.Neighbors[level], neighbor)
}

// GetNeighbors returns neighbors at the specified level (thread-safe)
func (n *Node) GetNeighbors(level int) []*Node {
	n.mu.RLock()
	defer n.mu.RUnlock()

	neighbors := n.Neighbors[level]
	result := make([]*Node, len(neighbors))
	copy(result, neighbors)
	return result
}

// Graph represents the HNSW index
type Graph struct {
	nodes          map[string]*Node // All nodes: ID -> Node
	entryPoint     *Node            // Entry point for search
	maxLevel       int              // Current maximum level
	m              int              // Number of connections per node
	efConstruction int              // Size of dynamic candidate list during construction
	ml             float64          // Level generation parameter
	distFunc       DistanceFunc     // Distance function (cosine, euclidean, etc.)
	mu             sync.RWMutex     // Protects concurrent access
}

// DistanceFunc computes distance between two vectors
type DistanceFunc func(a, b []float32) float32

// NewGraph creates a new HNSW graph
func NewGraph(distFunc DistanceFunc) *Graph {
	return &Graph{
		nodes:          make(map[string]*Node),
		m:              DefaultM,
		efConstruction: DefaultEfConstruction,
		ml:             DefaultML,
		distFunc:       distFunc,
	}
}

// SetParameters configures HNSW parameters
func (g *Graph) SetParameters(m, efConstruction int) {
	g.m = m
	g.efConstruction = efConstruction
}

// randomLevel generates a random level for a new node
func (g *Graph) randomLevel() int {
	level := 0
	for rand.Float64() < 1.0/math.Exp(1.0/g.ml) {
		level++
	}
	return level
}

// Insert adds a new vector to the HNSW index
func (g *Graph) Insert(id string, vector []float32) error {
	g.mu.Lock()
	defer g.mu.Unlock()

	// Check if already exists
	if _, exists := g.nodes[id]; exists {
		return nil // Already inserted
	}

	level := g.randomLevel()
	newNode := NewNode(id, vector, level)
	g.nodes[id] = newNode

	// First node - set as entry point
	if g.entryPoint == nil {
		g.entryPoint = newNode
		g.maxLevel = level
		return nil
	}

	// Find nearest neighbors at each level
	ep := g.entryPoint

	// Search from top to target level
	for lc := g.maxLevel; lc > level; lc-- {
		nearest := g.searchLayer(vector, ep, 1, lc)
		if len(nearest) > 0 {
			ep = nearest[0].node
		}
	}

	// Insert at levels 0 to level
	for lc := level; lc >= 0; lc-- {
		candidates := g.searchLayer(vector, ep, g.efConstruction, lc)

		// Select M neighbors
		m := g.m
		if lc == 0 {
			m = g.m * 2 // More connections at layer 0
		}

		neighbors := g.selectNeighbors(candidates, m)

		// Add bidirectional links
		for _, neighbor := range neighbors {
			newNode.AddNeighbor(lc, neighbor.node)
			neighbor.node.AddNeighbor(lc, newNode)

			// Prune neighbors if needed
			g.pruneNeighbors(neighbor.node, lc, m)
		}

		if len(neighbors) > 0 {
			ep = neighbors[0].node
		}
	}

	// Update entry point if this node has higher level
	if level > g.maxLevel {
		g.maxLevel = level
		g.entryPoint = newNode
	}

	return nil
}

// candidateNode represents a candidate node with its distance
type candidateNode struct {
	node     *Node
	distance float32
	index    int // For heap operations
}

// candidateHeap implements heap.Interface for nearest neighbor search
type candidateHeap struct {
	candidates []*candidateNode
	maxHeap    bool // true for max heap (farther first), false for min heap (nearer first)
}

func (h *candidateHeap) Len() int { return len(h.candidates) }

func (h *candidateHeap) Less(i, j int) bool {
	if h.maxHeap {
		return h.candidates[i].distance > h.candidates[j].distance
	}
	return h.candidates[i].distance < h.candidates[j].distance
}

func (h *candidateHeap) Swap(i, j int) {
	h.candidates[i], h.candidates[j] = h.candidates[j], h.candidates[i]
	h.candidates[i].index = i
	h.candidates[j].index = j
}

func (h *candidateHeap) Push(x interface{}) {
	n := len(h.candidates)
	candidate := x.(*candidateNode)
	candidate.index = n
	h.candidates = append(h.candidates, candidate)
}

func (h *candidateHeap) Pop() interface{} {
	old := h.candidates
	n := len(old)
	candidate := old[n-1]
	old[n-1] = nil
	candidate.index = -1
	h.candidates = old[0 : n-1]
	return candidate
}

// searchLayer searches for nearest neighbors at a specific layer
func (g *Graph) searchLayer(query []float32, entryPoint *Node, ef int, layer int) []*candidateNode {
	visited := make(map[string]bool)

	// Min heap for candidates (nearest first)
	candidates := &candidateHeap{maxHeap: false}
	heap.Init(candidates)

	// Max heap for result (farthest first)
	result := &candidateHeap{maxHeap: true}
	heap.Init(result)

	// Start with entry point
	dist := g.distFunc(query, entryPoint.Vector)
	heap.Push(candidates, &candidateNode{node: entryPoint, distance: dist})
	heap.Push(result, &candidateNode{node: entryPoint, distance: dist})
	visited[entryPoint.ID] = true

	for candidates.Len() > 0 {
		current := heap.Pop(candidates).(*candidateNode)

		// Stop if current is farther than farthest in result
		if result.Len() >= ef && current.distance > result.candidates[0].distance {
			break
		}

		// Check neighbors
		for _, neighbor := range current.node.GetNeighbors(layer) {
			if visited[neighbor.ID] {
				continue
			}
			visited[neighbor.ID] = true

			dist := g.distFunc(query, neighbor.Vector)

			if result.Len() < ef || dist < result.candidates[0].distance {
				heap.Push(candidates, &candidateNode{node: neighbor, distance: dist})
				heap.Push(result, &candidateNode{node: neighbor, distance: dist})

				// Keep only ef best
				if result.Len() > ef {
					heap.Pop(result)
				}
			}
		}
	}

	return result.candidates
}

// selectNeighbors selects M best neighbors using heuristic
func (g *Graph) selectNeighbors(candidates []*candidateNode, m int) []*candidateNode {
	if len(candidates) <= m {
		return candidates
	}

	// Simple heuristic: select M nearest
	// Sort by distance
	result := make([]*candidateNode, 0, m)
	for i := 0; i < len(candidates) && i < m; i++ {
		minIdx := i
		for j := i + 1; j < len(candidates); j++ {
			if candidates[j].distance < candidates[minIdx].distance {
				minIdx = j
			}
		}
		candidates[i], candidates[minIdx] = candidates[minIdx], candidates[i]
		result = append(result, candidates[i])
	}

	return result
}

// pruneNeighbors removes excess neighbors from a node
func (g *Graph) pruneNeighbors(node *Node, level int, maxConn int) {
	neighbors := node.GetNeighbors(level)
	if len(neighbors) <= maxConn {
		return
	}

	// Keep only maxConn closest neighbors
	distances := make([]float32, len(neighbors))
	for i, n := range neighbors {
		distances[i] = g.distFunc(node.Vector, n.Vector)
	}

	// Sort by distance
	for i := 0; i < maxConn; i++ {
		minIdx := i
		for j := i + 1; j < len(neighbors); j++ {
			if distances[j] < distances[minIdx] {
				minIdx = j
			}
		}
		neighbors[i], neighbors[minIdx] = neighbors[minIdx], neighbors[i]
		distances[i], distances[minIdx] = distances[minIdx], distances[i]
	}

	// Update neighbors
	node.mu.Lock()
	node.Neighbors[level] = neighbors[:maxConn]
	node.mu.Unlock()
}

// Size returns the number of nodes in the graph
func (g *Graph) Size() int {
	g.mu.RLock()
	defer g.mu.RUnlock()
	return len(g.nodes)
}

// GetNode retrieves a node by ID
func (g *Graph) GetNode(id string) (*Node, bool) {
	g.mu.RLock()
	defer g.mu.RUnlock()
	node, exists := g.nodes[id]
	return node, exists
}
