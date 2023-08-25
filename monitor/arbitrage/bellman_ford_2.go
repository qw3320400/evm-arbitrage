package arbitrage

import (
	"math"
)

// Graph represents a graph consisting of edges and vertices
type BF2Graph struct {
	edges    []*BF2Edge
	vertices []int
}

// Edge represents a weighted line between two nodes
type BF2Edge struct {
	From, To int
	Weight   float64
}

// NewEdge returns a pointer to a new Edge
func NewEdge(from, to int, weight float64) *BF2Edge {
	return &BF2Edge{From: from, To: to, Weight: weight}
}

// NewGraph returns a graph consisting of given edges and vertices (vertices must count from 0 upwards)
func NewBF2Graph(edges []*BF2Edge, vertices []int) *BF2Graph {
	return &BF2Graph{edges: edges, vertices: vertices}
}

// FindArbitrageLoop returns either an arbitrage loop or a nil map
func (g *BF2Graph) FindArbitrageLoop(source int) []int {
	predecessors, distances := g.BellmanFord(source)
	return g.FindNegativeWeightCycle(predecessors, distances, source)
}

// BellmanFord determines the shortest path and returns the predecessors and distances
func (g *BF2Graph) BellmanFord(source int) ([]int, []float64) {
	size := len(g.vertices)
	distances := make([]float64, size)
	predecessors := make([]int, size)
	for _, v := range g.vertices {
		distances[v] = math.MaxFloat64
	}
	distances[source] = 0

	for i, changes := 0, 0; i < size-1; i, changes = i+1, 0 {
		for _, edge := range g.edges {
			if newDist := distances[edge.From] + edge.Weight; newDist < distances[edge.To] {
				distances[edge.To] = newDist
				predecessors[edge.To] = edge.From
				changes++
			}
		}
		if changes == 0 {
			break
		}
	}
	return predecessors, distances
}

// FindNegativeWeightCycle finds a negative weight cycle from predecessors and a source
func (g *BF2Graph) FindNegativeWeightCycle(predecessors []int, distances []float64, source int) []int {
	for _, edge := range g.edges {
		if distances[edge.From]+edge.Weight < distances[edge.To] {
			return arbitrageLoop(predecessors, source)
		}
	}
	return nil
}

func arbitrageLoop(predecessors []int, source int) []int {
	size := len(predecessors)
	loop := make([]int, size)
	loop[0] = source

	exists := make([]bool, size)
	exists[source] = true

	indices := make([]int, size)

	var index, next int
	for index, next = 1, source; ; index++ {
		next = predecessors[next]
		loop[index] = next
		if exists[next] {
			return loop[indices[next] : index+1]
		}
		indices[next] = index
		exists[next] = true
	}
}
