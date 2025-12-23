package hnsw

import (
	"errors"
)

// SearchResult represents a single search result with distance
type SearchResult struct {
	ID       string    // Node ID
	Vector   []float32 // Vector embedding
	Distance float32   // Distance to query
}

// Search performs k-nearest neighbor search
func (g *Graph) Search(query []float32, k int, efSearch int) ([]SearchResult, error) {
	g.mu.RLock()
	defer g.mu.RUnlock()

	if g.entryPoint == nil {
		return nil, errors.New("index is empty")
	}

	if efSearch < k {
		efSearch = k
	}

	ep := g.entryPoint

	// Search from top level to level 0
	for lc := g.maxLevel; lc > 0; lc-- {
		nearest := g.searchLayer(query, ep, 1, lc)
		if len(nearest) > 0 {
			ep = nearest[0].node
		}
	}

	// Search at layer 0 with efSearch
	candidates := g.searchLayer(query, ep, efSearch, 0)

	// Convert to results and take top k
	results := make([]SearchResult, 0, k)
	for i := 0; i < len(candidates) && i < k; i++ {
		results = append(results, SearchResult{
			ID:       candidates[i].node.ID,
			Vector:   candidates[i].node.Vector,
			Distance: candidates[i].distance,
		})
	}

	return results, nil
}

// SearchKNN performs k-nearest neighbor search with default efSearch
func (g *Graph) SearchKNN(query []float32, k int) ([]SearchResult, error) {
	return g.Search(query, k, DefaultEfSearch)
}

// SearchWithCallback performs search and calls callback for each result
// Useful for streaming results or early termination
func (g *Graph) SearchWithCallback(query []float32, k int, efSearch int, callback func(SearchResult) bool) error {
	results, err := g.Search(query, k, efSearch)
	if err != nil {
		return err
	}

	for _, result := range results {
		if !callback(result) {
			break
		}
	}

	return nil
}

// RangeSearch finds all neighbors within a distance threshold
func (g *Graph) RangeSearch(query []float32, maxDistance float32, efSearch int) ([]SearchResult, error) {
	g.mu.RLock()
	defer g.mu.RUnlock()

	if g.entryPoint == nil {
		return nil, errors.New("index is empty")
	}

	ep := g.entryPoint

	// Search from top level to level 0
	for lc := g.maxLevel; lc > 0; lc-- {
		nearest := g.searchLayer(query, ep, 1, lc)
		if len(nearest) > 0 {
			ep = nearest[0].node
		}
	}

	// Search at layer 0
	candidates := g.searchLayer(query, ep, efSearch, 0)

	// Filter by distance threshold
	results := make([]SearchResult, 0)
	for _, candidate := range candidates {
		if candidate.distance <= maxDistance {
			results = append(results, SearchResult{
				ID:       candidate.node.ID,
				Vector:   candidate.node.Vector,
				Distance: candidate.distance,
			})
		}
	}

	return results, nil
}

// BatchSearch performs multiple searches in parallel
func (g *Graph) BatchSearch(queries [][]float32, k int, efSearch int) ([][]SearchResult, error) {
	if len(queries) == 0 {
		return nil, errors.New("no queries provided")
	}

	results := make([][]SearchResult, len(queries))
	errs := make([]error, len(queries))

	// Sequential search for now (can be parallelized later)
	for i, query := range queries {
		results[i], errs[i] = g.Search(query, k, efSearch)
		if errs[i] != nil {
			return nil, errs[i]
		}
	}

	return results, nil
}

// Delete removes a node from the index
func (g *Graph) Delete(id string) error {
	g.mu.Lock()
	defer g.mu.Unlock()

	node, exists := g.nodes[id]
	if !exists {
		return errors.New("node not found")
	}

	// Remove bidirectional links at all levels
	for level := 0; level <= node.Level; level++ {
		neighbors := node.GetNeighbors(level)
		for _, neighbor := range neighbors {
			// Remove this node from neighbor's connections
			neighbor.mu.Lock()
			neighborList := neighbor.Neighbors[level]
			for i, n := range neighborList {
				if n.ID == id {
					neighbor.Neighbors[level] = append(neighborList[:i], neighborList[i+1:]...)
					break
				}
			}
			neighbor.mu.Unlock()
		}
	}

	// Remove from nodes map
	delete(g.nodes, id)

	// Update entry point if needed
	if g.entryPoint != nil && g.entryPoint.ID == id {
		g.entryPoint = g.findNewEntryPoint()
	}

	return nil
}

// findNewEntryPoint finds a new entry point after deletion
func (g *Graph) findNewEntryPoint() *Node {
	var maxNode *Node
	maxLevel := -1

	for _, node := range g.nodes {
		if node.Level > maxLevel {
			maxLevel = node.Level
			maxNode = node
		}
	}

	g.maxLevel = maxLevel
	return maxNode
}

// Clear removes all nodes from the index
func (g *Graph) Clear() {
	g.mu.Lock()
	defer g.mu.Unlock()

	g.nodes = make(map[string]*Node)
	g.entryPoint = nil
	g.maxLevel = 0
}

// GetStatistics returns index statistics
func (g *Graph) GetStatistics() Statistics {
	g.mu.RLock()
	defer g.mu.RUnlock()

	stats := Statistics{
		NodeCount:      len(g.nodes),
		MaxLevel:       g.maxLevel,
		M:              g.m,
		EfConstruction: g.efConstruction,
	}

	if g.entryPoint != nil {
		stats.EntryPointID = g.entryPoint.ID
	}

	// Calculate average connections per level
	connectionCounts := make(map[int]int)
	connectionSums := make(map[int]int)

	for _, node := range g.nodes {
		for level := 0; level <= node.Level; level++ {
			neighbors := node.GetNeighbors(level)
			connectionSums[level] += len(neighbors)
			connectionCounts[level]++
		}
	}

	stats.AvgConnectionsPerLevel = make(map[int]float64)
	for level, sum := range connectionSums {
		if connectionCounts[level] > 0 {
			stats.AvgConnectionsPerLevel[level] = float64(sum) / float64(connectionCounts[level])
		}
	}

	return stats
}

// Statistics holds HNSW index statistics
type Statistics struct {
	NodeCount              int
	MaxLevel               int
	M                      int
	EfConstruction         int
	EntryPointID           string
	AvgConnectionsPerLevel map[int]float64
}

// verifyIntegrity checks graph integrity (for testing/debugging)
func (g *Graph) verifyIntegrity() []string {
	g.mu.RLock()
	defer g.mu.RUnlock()

	errors := make([]string, 0)

	// Check bidirectional links
	for id, node := range g.nodes {
		for level := 0; level <= node.Level; level++ {
			neighbors := node.GetNeighbors(level)
			for _, neighbor := range neighbors {
				// Check if neighbor exists
				if _, exists := g.nodes[neighbor.ID]; !exists {
					errors = append(errors, "node "+id+" has non-existent neighbor "+neighbor.ID)
					continue
				}

				// Check bidirectional link
				found := false
				for _, n := range neighbor.GetNeighbors(level) {
					if n.ID == id {
						found = true
						break
					}
				}
				if !found {
					errors = append(errors, "missing bidirectional link: "+id+" -> "+neighbor.ID)
				}
			}
		}
	}

	// Check entry point
	if g.entryPoint != nil {
		if _, exists := g.nodes[g.entryPoint.ID]; !exists {
			errors = append(errors, "entry point not in nodes map")
		}
	}

	return errors
}
