package arbitrage

import (
	"math"
	"math/big"
	"monitor/protocol"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/ethereum/go-ethereum/common"
)

type Vertex struct {
	ID string
}

func TestBellmanFord1(t *testing.T) {
	startTime := time.Now()
	g := InitBF1Graph(5)
	g.AddEdge(0, 1, 6)
	g.AddEdge(0, 2, 7)
	g.AddEdge(1, 2, 8)
	g.AddEdge(1, 3, -4)
	g.AddEdge(1, 4, 5)
	g.AddEdge(2, 3, 9)
	g.AddEdge(2, 4, -3)
	g.AddEdge(3, 1, 7)
	g.AddEdge(4, 0, 2)
	g.AddEdge(4, 3, 7)
	source := 0
	dist, prev := g.BellmanFord(source)
	t.Log("--", dist, prev, time.Since(startTime))
}

func TestBellmanFord2(t *testing.T) {
	startTime := time.Now()
	g := NewBF2Graph(
		[]*BF2Edge{
			NewEdge(0, 1, 6),
			NewEdge(0, 2, 7),
			NewEdge(1, 2, 8),
			NewEdge(1, 3, -4),
			NewEdge(1, 4, 5),
			NewEdge(2, 3, 9),
			NewEdge(2, 4, -3),
			NewEdge(3, 1, 7),
			NewEdge(4, 0, 2),
			NewEdge(4, 3, 7),
		},
		[]int{0, 1, 2, 3, 4},
	)
	source := 0
	prev, dist := g.BellmanFord(source)
	t.Log("--", dist, prev, time.Since(startTime))
}

func TestBellmanFord3(t *testing.T) {
	startTime := time.Now()
	g := NewBF3Graph()
	g.SetOrder(5)
	g.AddEdge(0, 1, 6)
	g.AddEdge(0, 2, 7)
	g.AddEdge(1, 2, 8)
	g.AddEdge(1, 3, -4)
	g.AddEdge(1, 4, 5)
	g.AddEdge(2, 3, 9)
	g.AddEdge(2, 4, -3)
	g.AddEdge(3, 1, 7)
	g.AddEdge(4, 0, 2)
	g.AddEdge(4, 3, 7)
	source := 0
	dist, prev := g.BellmanFord(source)
	t.Log("--", dist, prev, time.Since(startTime))
}

func TestCycle(t *testing.T) {
	startTime := time.Now()

	pairs := map[common.Address]*protocol.UniswapV2Pair{}
	b, err := os.ReadFile("../data/Uniswapv2Pairs")
	if err != nil {
		t.Fatal(err)
	}
	lines := strings.Split(string(b), "\n")
	for _, line := range lines {
		if len(line) == 0 {
			continue
		}
		words := strings.Split(strings.Split(line, "@")[0], ",")
		pair := &protocol.UniswapV2Pair{
			Address: common.HexToAddress(words[0]),
			Token0:  common.HexToAddress(words[1]),
			Token1:  common.HexToAddress(words[2]),
		}
		pair.Reserve0, _ = big.NewInt(0).SetString(words[3], 10)
		pair.Reserve1, _ = big.NewInt(0).SetString(words[4], 10)
		r0, _ := big.NewFloat(0).SetInt(pair.Reserve0).Float64()
		r1, _ := big.NewFloat(0).SetInt(pair.Reserve1).Float64()
		pair.Weight0 = -math.Log(r1 / r0 * 0.997)
		pair.Weight1 = -math.Log(r0 / r1 * 0.997)
		pairs[pair.Address] = pair
	}
	t.Log("- load data time", time.Since(startTime))
	var (
		vertices  = []int{}
		edges     = []*BF2Edge{}
		addresses = []common.Address{}
		addrToIdx = map[common.Address]int{}
	)
	for _, pair := range pairs {
		if pair.Error {
			continue
		}
		var idx0, idx1 int
		if i, ok := addrToIdx[pair.Token0]; !ok {
			idx0 := len(vertices)
			addrToIdx[pair.Token0] = idx0
			vertices = append(vertices, idx0)
			addresses = append(addresses, pair.Token0)
		} else {
			idx0 = i
		}
		if i, ok := addrToIdx[pair.Token1]; !ok {
			idx1 := len(vertices)
			addrToIdx[pair.Token1] = idx1
			vertices = append(vertices, idx1)
			addresses = append(addresses, pair.Token1)
		} else {
			idx1 = i
		}
		edges = append(edges,
			NewEdge(idx0, idx1, pair.Weight0),
			NewEdge(idx1, idx0, pair.Weight1),
		)
	}
	g := NewBF2Graph(edges, vertices)
	ethAddr := common.HexToAddress("0x4200000000000000000000000000000000000006")
	ethIdx := addrToIdx[ethAddr]
	loop := g.FindArbitrageLoop(ethIdx)
	for _, idx := range loop {
		t.Log("-----", addresses[idx])
	}
	t.Log("- find loop time", loop, time.Since(startTime))
}
