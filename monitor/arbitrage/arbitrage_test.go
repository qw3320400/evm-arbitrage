package arbitrage

import (
	"testing"

	"github.com/dominikbraun/graph"
)

type Vertex struct {
	ID string
}

func TestGraph(t *testing.T) {
	hash := func(v *Vertex) string {
		return v.ID
	}
	g := graph.New[string, *Vertex](hash)
	g.AddVertex(&Vertex{"london"})
	g.AddVertex(&Vertex{"munich"})
	g.AddVertex(&Vertex{"paris"})
	g.AddVertex(&Vertex{"madrid"})
	g.AddEdge("london", "munich", graph.EdgeWeight(3))
	g.AddEdge("london", "paris", graph.EdgeWeight(2))
	g.AddEdge("london", "madrid", graph.EdgeWeight(5))
	g.AddEdge("munich", "madrid", graph.EdgeWeight(6))
	g.AddEdge("munich", "paris", graph.EdgeWeight(2))
	g.AddEdge("paris", "madrid", graph.EdgeWeight(4))
	graph.ShortestPath[]()
}
