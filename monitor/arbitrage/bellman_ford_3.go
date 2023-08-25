package arbitrage

import (
	"fmt"
	"math"
)

type BF3Graph struct {
	V        int         // number of vertices (order of the graph)
	E        int         // number of edges (size of the graph)
	Nmat     [][]int     // neighbour matrix (adjacency matrix)
	Emat     [][]float32 // edge matrix (edge weights)
	directed bool        // is the graph a directed graph?
}

func NewBF3Graph() *BF3Graph {
	g := new(BF3Graph)
	g.V = 0
	g.E = 0
	g.directed = false
	return g
}

// set total number of vertices
func (g *BF3Graph) SetOrder(i int) {
	g.V = i
	a := make([][]int, g.V)
	for i := range a {
		a[i] = make([]int, g.V)
	}
	g.Nmat = a
	b := make([][]float32, g.V)
	for i := range b {
		b[i] = make([]float32, g.V)
	}
	g.Emat = b
}

// add a new edge between two vertices,
// but only if the vertices are within the order of G
// Each edge requires a value (edge length)
// Setting up parallel edges is avoided.
func (g *BF3Graph) AddEdge(v1, v2 int, l float64) {
	var e []int
	if v1 < g.V && v2 < g.V {
		if g.directed {
			e = append(e, v1, v2)
		} else { // if the graph is not directed, sort the edge
			e = edge(v1, v2)
		}
		if !g.CheckEdges(e) {
			g.Nmat[e[0]][e[1]] = 1
			g.Emat[e[0]][e[1]] = float32(l)
			if !g.directed { // for undirected graphs the matrices are symmetric
				g.Nmat[e[1]][e[0]] = 1
				g.Emat[e[1]][e[0]] = float32(l)
			}
			g.E++ //update size of G
		}
		// } else {
		// err := fmt.Errorf("Vertex does not exist")
		// return err
	}
}

// run a check on a graph: is a given edge part of the graph?
func (g *BF3Graph) CheckEdges(edge []int) bool {
	v1 := edge[0]
	v2 := edge[1]
	return g.Nmat[v1][v2] == 1
}

// delete an edge between two vertices (if it is present) from the graph
func (g *BF3Graph) DelEdge(v1, v2 int) {
	g.Nmat[v1][v2] = 0
	g.Emat[v1][v2] = 0
	if !g.directed {
		g.Nmat[v2][v1] = 0
		g.Emat[v2][v1] = 0
	}
}

// quickly convert a pair of two vertices into
// an edge (in the correct order)
func edge(v1, v2 int) []int {
	var e []int
	if v1 <= v2 {
		e = append(e, v1, v2)
	} else {
		e = append(e, v2, v1)
	}
	return e
}

// Get the weight (length) of an edge
func (g *BF3Graph) GetWeight(v1, v2 int) float32 {
	return g.Emat[v1][v2]
}

// Get the degree of a vertex (i.e, the number of its connected neighbours)
// it is equal to the sum of row v of the adjacency matrix
func (g *BF3Graph) Degree(v int) int {
	deg := 0
	for _, k := range g.Nmat[v] {
		if k == 1 {
			deg++
		}
	}
	return deg
}

// Get a list of the connected neighbours of a vertex
// also outputs the degree of the vertex
func (g *BF3Graph) Neighbours(v int) ([]int, int) {
	var nei []int
	deg := 0
	/* this should work for directed and undirected graphs
	since G.Nmat[v] is the part of the adjacency matrix
	belonging to a single vertex v */
	for i, k := range g.Nmat[v] {
		if k == 1 {
			nei = append(nei, i)
			deg++
		}
	}
	return nei, deg
}

// Nlist is a function that returns a list of all neighbours for each vertex
func (g *BF3Graph) Nlist() [][]int {
	neigh := make([][]int, g.V)
	for i := 0; i < g.V; i++ {
		neigh[i], _ = g.Neighbours(i) // get the neighbour lists for easy access
	}
	return neigh
}

// disconnect a vertex from the graph
// (i.e., remove all its edges)
func (g *BF3Graph) DisconnectVert(v int) {
	nei, _ := g.Neighbours(v)
	for _, k := range nei {
		g.DelEdge(v, k)
	}
}

func (g *BF3Graph) BellmanFord(start int) ([]float64, []int) {
	// Initialize the distances.
	// I.e., this is the total distance from the source to any given point
	dist := make([]float64, g.V)
	// Initialize the predecessors.
	// I.e., this is the predecessor for any given point in the path from the source
	prev := make([]int, g.V)
	// Initialize explicit list of edges (obtained from adjacency matrix).
	var edges [][]int
	// Initialize data
	for i := 0; i < g.V; i++ {
		dist[i] = math.Inf(0)      // set distance to vertex i to "infinity"
		prev[i] = -1               // set predecessor of vertex i to "undefined"
		for j := 0; j < g.V; j++ { // it has to be all edge combinations, i.e., both (u,v) and (v,u)
			if g.Nmat[i][j] == 1 {
				edges = append(edges, []int{i, j})
			}
		}
	}
	dist[start] = 0.0   // set distance of the source vertex to 0
	prev[start] = start // set the predecessor of the source to itself

	// repeated relaxation of edges (i => n-1 times)
	for i := 1; i < g.V; i++ {
		for _, e := range edges {
			u := e[0]
			v := e[1]
			newdist := dist[u] + float64(g.Emat[u][v])
			if newdist < dist[v] {
				dist[v] = newdist
				prev[v] = u
			}
		}
	}

	// check for negative cycles (would be the n-th iteration of the for-loop above)
	for _, e := range edges {
		u := e[0]
		v := e[1]
		newdist := dist[u] + float64(g.Emat[u][v])
		if newdist < dist[v] {
			err := fmt.Errorf("warning: the graph contains a negativ-weight cycle")
			fmt.Println(err)
		}
	}

	return dist, prev
}
