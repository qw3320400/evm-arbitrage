package arbitrage

import (
	"math"
	"monitor/utils"
)

type BF1Edge struct {
	src    int
	dest   int
	weight float64
}
type BF1Graph struct {
	vertices int
	edges    []BF1Edge
}

func InitBF1Graph(vertices int) *BF1Graph {
	return &BF1Graph{
		vertices: vertices,
		edges:    make([]BF1Edge, 0),
	}
}
func (g *BF1Graph) AddEdge(src, dest int, weight float64) {
	g.edges = append(g.edges, BF1Edge{src, dest, weight})
}

func (g *BF1Graph) BellmanFord(source int) ([]float64, []int) {
	dist := make([]float64, g.vertices)
	prev := make([]int, g.vertices)
	for i := 0; i < g.vertices; i++ {
		dist[i] = math.Inf(1)
		prev[i] = -1
	}
	dist[source] = 0

	for i := 1; i < g.vertices; i++ {
		for _, edge := range g.edges {
			u := edge.src
			v := edge.dest
			w := edge.weight
			if dist[u]+w < dist[v] {
				dist[v] = dist[u] + w
				prev[v] = u
			}
		}
	}

	for _, edge := range g.edges {
		u := edge.src
		v := edge.dest
		w := edge.weight
		if dist[u]+w < dist[v] {
			utils.Warnf("Graph contains a negative weight cycle")
			return nil, nil
		}
	}

	return dist, prev
}

func PrintShortestPaths(dist []float64, prev []int, source int) {
	utils.Infof("Shortest Paths from vertex %d", source)
	for i := 0; i < len(dist); i++ {
		if dist[i] == math.Inf(1) {
			utils.Infof("Vertex %d is not reachable", i)
		} else {
			path := []int{}
			j := i
			for j != -1 {
				path = append([]int{j}, path...)
				j = prev[j]
			}
			utils.Infof("Vertex %d: Distance=%f, Path=%v", i, dist[i], path)
		}
	}
}
