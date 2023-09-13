package arbitrage

import (
	"github.com/ethereum/go-ethereum/common"
)

type SwapGraph struct {
	distances map[common.Address]float64
	// predecessors map[common.Address]common.Address
	preedges map[common.Address]string
	edges    map[string]*SwapEdge
}

type SwapEdge struct {
	Key      string
	Pair     common.Address
	From     common.Address
	To       common.Address
	Distance float64
}

func NewSwapGraph() *SwapGraph {
	return &SwapGraph{
		distances: map[common.Address]float64{},
		// predecessors: map[common.Address]common.Address{},
		preedges: map[common.Address]string{},
		edges:    map[string]*SwapEdge{},
	}
}

func (g *SwapGraph) AddVertices(vertices ...common.Address) {
	for _, v := range vertices {
		g.distances[v] = 1000000000000000000
	}
}

func (g *SwapGraph) AddEdges(edges ...*SwapEdge) {
	for _, e := range edges {
		g.edges[e.Key] = e
	}
}

func (g *SwapGraph) BellmanFord(source common.Address) {
	g.distances[source] = 0
	for i := 0; i < 5; i++ {
		var change bool
		for _, e := range g.edges {
			fromDist, ok := g.distances[e.From]
			if !ok {
				continue
			}
			toDist, ok := g.distances[e.To]
			if !ok {
				continue
			}
			if newDist := fromDist + e.Distance; newDist < toDist {
				change = true
				g.distances[e.To] = newDist
				// g.predecessors[e.To] = e.From
				g.preedges[e.To] = e.Key
			}
		}
		if !change {
			break
		}
	}
}

func (g *SwapGraph) FindCircle(source common.Address) []common.Address {
	g.BellmanFord(source)
	var (
		minEdge     *SwapEdge
		minDistance float64
	)
	for _, e := range g.edges {
		if e.To != source {
			continue
		}
		fromDist, ok := g.distances[e.From]
		if !ok {
			continue
		}
		toDist, ok := g.distances[e.To]
		if !ok {
			continue
		}
		if newDist := fromDist + e.Distance; newDist < toDist {
			if minEdge == nil || newDist < minDistance {
				minEdge = e
				minDistance = newDist
			}
		}
	}
	ret := []common.Address{}
	if minEdge != nil {
		var loop = true
		for i := 0; i < 10; i++ {
			// fmt.Printf("%s %s %s %s\n", minEdge.Key, minEdge.From, minEdge.To, minEdge.Distance)
			ret = append(ret, minEdge.Pair)

			if minEdge.From == source {
				loop = false
				break
			}
			pair := g.preedges[minEdge.From]
			minEdge = g.edges[pair]
		}
		if loop {
			return []common.Address{}
		}
	}
	return ret
}
