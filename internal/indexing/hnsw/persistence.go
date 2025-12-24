package hnsw

import (
	"encoding/json"
	"os"
)

// Save serializes the HNSW index to disk.
func (g *Graph) Save(filepath string) error {
	g.mu.RLock()
	defer g.mu.RUnlock()

	// Create serializable structure
	data := &SerializedGraph{
		M:              g.m,
		EfConstruction: g.efConstruction,
		ML:             g.ml,
		MaxLevel:       g.maxLevel,
		Nodes:          make([]SerializedNode, 0, len(g.nodes)),
	}

	if g.entryPoint != nil {
		data.EntryPointID = g.entryPoint.ID
	}

	// Serialize nodes
	for _, node := range g.nodes {
		sn := SerializedNode{
			ID:        node.ID,
			Vector:    node.Vector,
			Level:     node.Level,
			Neighbors: make(map[int][]string),
		}

		// Serialize neighbors (store only IDs)
		for level, neighbors := range node.Neighbors {
			neighborIDs := make([]string, len(neighbors))
			for i, n := range neighbors {
				neighborIDs[i] = n.ID
			}
			sn.Neighbors[level] = neighborIDs
		}

		data.Nodes = append(data.Nodes, sn)
	}

	// Write to file
	file, err := os.Create(filepath)
	if err != nil {
		return err
	}
	defer func() { _ = file.Close() }()

	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")
	return encoder.Encode(data)
}

// Load deserializes the HNSW index from disk.
func (g *Graph) Load(filepath string) error {
	file, err := os.Open(filepath)
	if err != nil {
		return err
	}
	defer func() { _ = file.Close() }()

	var data SerializedGraph
	decoder := json.NewDecoder(file)
	if err := decoder.Decode(&data); err != nil {
		return err
	}

	g.mu.Lock()
	defer g.mu.Unlock()

	// Restore parameters
	g.m = data.M
	g.efConstruction = data.EfConstruction
	g.ml = data.ML
	g.maxLevel = data.MaxLevel

	// Clear existing data
	g.nodes = make(map[string]*Node)
	g.entryPoint = nil

	// First pass: create all nodes
	for _, sn := range data.Nodes {
		node := NewNode(sn.ID, sn.Vector, sn.Level)
		g.nodes[node.ID] = node
	}

	// Second pass: restore connections
	for _, sn := range data.Nodes {
		node := g.nodes[sn.ID]
		for level, neighborIDs := range sn.Neighbors {
			for _, nid := range neighborIDs {
				if neighbor, exists := g.nodes[nid]; exists {
					node.Neighbors[level] = append(node.Neighbors[level], neighbor)
				}
			}
		}
	}

	// Restore entry point
	if data.EntryPointID != "" {
		if ep, exists := g.nodes[data.EntryPointID]; exists {
			g.entryPoint = ep
		}
	}

	return nil
}

// SerializedGraph is the on-disk representation of the HNSW graph.
type SerializedGraph struct {
	M              int              `json:"m"`
	EfConstruction int              `json:"ef_construction"`
	ML             float64          `json:"ml"`
	MaxLevel       int              `json:"max_level"`
	EntryPointID   string           `json:"entry_point_id"`
	Nodes          []SerializedNode `json:"nodes"`
}

// SerializedNode is the on-disk representation of a node.
type SerializedNode struct {
	ID        string           `json:"id"`
	Vector    []float32        `json:"vector"`
	Level     int              `json:"level"`
	Neighbors map[int][]string `json:"neighbors"` // level -> [neighbor IDs]
}
