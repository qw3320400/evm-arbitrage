package arbitrage

import (
	"math"

	"github.com/ethereum/go-ethereum/common"
	"github.com/shopspring/decimal"
)

type SwapGraph struct {
	distances map[common.Address]decimal.Decimal
	// predecessors map[common.Address]common.Address
	preedges map[common.Address]string
	edges    map[string]*SwapEdge
}

type SwapEdge struct {
	Key      string
	Pair     common.Address
	From     common.Address
	To       common.Address
	Distance decimal.Decimal
}

func NewSwapGraph() *SwapGraph {
	return &SwapGraph{
		distances: map[common.Address]decimal.Decimal{},
		// predecessors: map[common.Address]common.Address{},
		preedges: map[common.Address]string{},
		edges:    map[string]*SwapEdge{},
	}
}

func (g *SwapGraph) AddVertices(vertices ...common.Address) {
	for _, v := range vertices {
		g.distances[v] = decimal.NewFromInt(math.MaxInt)
	}
}

func (g *SwapGraph) AddEdges(edges ...*SwapEdge) {
	for _, e := range edges {
		g.edges[e.Key] = e
	}
}

func (g *SwapGraph) BellmanFord(source common.Address) {
	g.distances[source] = decimal.NewFromInt(0)
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
			if newDist := fromDist.Add(e.Distance); newDist.Cmp(toDist) < 0 {
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
		minDistance decimal.Decimal
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
		if newDist := fromDist.Add(e.Distance); newDist.Cmp(toDist) < 0 {
			if minEdge == nil || newDist.Cmp(minDistance) < 0 {
				minEdge = e
				minDistance = newDist
			}
		}
	}
	ret := []common.Address{}
	if minEdge != nil {
		for {
			// fmt.Printf("%s %s %s %s\n", minEdge.Key, minEdge.From, minEdge.To, minEdge.Distance)
			ret = append(ret, minEdge.Pair)

			if minEdge.From == source {
				break
			}
			pair := g.preedges[minEdge.From]
			minEdge = g.edges[pair]
		}
	}
	return ret
}
