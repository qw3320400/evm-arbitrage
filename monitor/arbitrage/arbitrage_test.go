package arbitrage

import (
	"math"
	"math/big"
	"monitor/protocol"
	"os"
	"strconv"
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
	g := NewSwapGraph()
	for _, pair := range pairs {
		g.AddVertices(pair.Token0, pair.Token1)
		g.AddEdges(
			&SwapEdge{
				Key:      pair.Address.Hex() + "+",
				Pair:     pair.Address,
				From:     pair.Token0,
				To:       pair.Token1,
				Distance: pair.Weight0,
			},
			&SwapEdge{
				Key:      pair.Address.Hex() + "-",
				Pair:     pair.Address,
				From:     pair.Token1,
				To:       pair.Token0,
				Distance: pair.Weight1,
			},
		)
	}
	ethAddr := common.HexToAddress("0x4200000000000000000000000000000000000006")
	g.FindCircle(ethAddr)
	t.Log("- find loop time", time.Since(startTime))
}

func Test123(t *testing.T) {
	vertex := 5
	edges := [][3]int64{
		{0, 1, 6},
		{0, 2, 7},
		{1, 2, 8},
		{1, 3, -4},
		{1, 4, 5},
		{2, 3, 9},
		{2, 4, -3},
		{3, 1, 7},
		{4, 0, 2},
		{4, 3, 7},
	}
	dist := make([]int64, vertex)
	prev := make([]int64, vertex)
	dist[0] = 0
	for i := 1; i < len(dist); i++ {
		dist[i] = 1000
	}
	for {
		var change bool
		for _, e := range edges {
			var (
				from, to, distance = e[0], e[1], e[2]
			)
			if dist[from]+distance < dist[to] {
				dist[to] = dist[from] + distance
				prev[to] = from
				change = true
			}
		}
		if !change {
			break
		}
	}
	t.Log(prev, dist)
}

func TestAppend(t *testing.T) {
	a := []string{"1", "2", "3"}
	b := append(a, "4")
	t.Log(a)
	t.Log(b)
}

/*
0.1%		99.6%
0.3%		99.4%
0.5%		99.2%
0.7%		99%
1%			98.71%
2%			97.75%
3%			96.8%
4%			95.875%
5%			94.964%
6%			94.07%
7%			93.20%
8%			92.335%
9%			91.419%
10%			90.661%
*/
func TestAmountOut(t *testing.T) {
	out := protocol.GetAmountOut(7024483748378184, 5075022031094541599, 884916887826466518622968, 30)
	t.Logf("%f", out)
	out = protocol.GetAmountOut(out, 36813941190031336183629, 355002929, 30)
	t.Logf("%f", out)
	out = protocol.GetAmountOut(out, 1238454830614, 785015812149015823715, 30)
	t.Logf("%f", out)
}

func TestFee(t *testing.T) {
	a0i := big.NewInt(0)
	a1i := big.NewInt(0)
	a0o := big.NewInt(0)
	a1o := big.NewInt(0)
	r0 := big.NewInt(0)
	r1 := big.NewInt(0)

	a0i, _ = a0i.SetString("0", 10)
	a1i, _ = a1i.SetString("2000000000000000", 10)
	a0o, _ = a0o.SetString("14472300943115752421", 10)
	a1o, _ = a1o.SetString("0", 10)
	r0, _ = r0.SetString("14863654188967624342618", 10)
	r1, _ = r1.SetString("2051974905703706567", 10)
	t.Log(protocol.CalculatePairFee(a0i, a1i, a0o, a1o, r0, r1))

	a0i, _ = a0i.SetString("2151714554", 10)
	a1i, _ = a1i.SetString("0", 10)
	a0o, _ = a0o.SetString("0", 10)
	a1o, _ = a1o.SetString("2151176174", 10)
	r0, _ = r0.SetString("124614963475", 10)
	r1, _ = r1.SetString("129682145128", 10)
	t.Log(protocol.CalculatePairFee(a0i, a1i, a0o, a1o, r0, r1))

	a0i, _ = a0i.SetString("0", 10)
	a1i, _ = a1i.SetString("1443650043000000000000000000", 10)
	a0o, _ = a0o.SetString("165051070728644039", 10)
	a1o, _ = a1o.SetString("0", 10)
	r0, _ = r0.SetString("990391729848181355", 10)
	r1, _ = r1.SetString("10084639825298959164405905709", 10)
	t.Log(protocol.CalculatePairFee(a0i, a1i, a0o, a1o, r0, r1))
}

func TestAmountsOunt(t *testing.T) {
	wethAddress := common.HexToAddress("0x4200000000000000000000000000000000000006")
	pairs := []string{
		"--------pair 0x41d160033C222E6f3722EC97379867324567d883 0x4200000000000000000000000000000000000006 0xd9aAEc86B65D86f6A7B5B1b0c42FFA531710b6CA 291259013940960551613 463302110785 25",
		"--------pair 0x282f9231E5294E7354744df36461c21e0E68061C 0xB79DD08EA68A908A97220C76d19A6aA9cBDE4376 0xd9aAEc86B65D86f6A7B5B1b0c42FFA531710b6CA 382185083241 381491812370 31",
		"--------pair 0xDDCedA9866C0bcED84561673dDEf77d218b6D90e 0x78a087d713Be963Bf307b18F2Ff8122EF9A63ae9 0xB79DD08EA68A908A97220C76d19A6aA9cBDE4376 115928160956033026 146830 25",
		"--------pair 0x3CD5A4c56C4794d88170b6BdE719656A73026539 0x78a087d713Be963Bf307b18F2Ff8122EF9A63ae9 0x833589fCD6eDb6E08f4c7C32D4f71b54bdA02913 147391892215383617 474986 28",
		"--------pair 0xaEeB835f3Aa21d19ea5E33772DaA9E64f1b6982F 0x4200000000000000000000000000000000000006 0x833589fCD6eDb6E08f4c7C32D4f71b54bdA02913 161529937445973407268 256419444911 31",
	}
	pairPath := []*protocol.UniswapV2Pair{}
	for _, s := range pairs {
		w := strings.Split(s, " ")
		pair := &protocol.UniswapV2Pair{
			Address:  common.HexToAddress(w[1]),
			Token0:   common.HexToAddress(w[2]),
			Token1:   common.HexToAddress(w[3]),
			Reserve0: big.NewInt(0),
			Reserve1: big.NewInt(0),
		}
		pair.Reserve0, _ = pair.Reserve0.SetString(w[4], 10)
		pair.Reserve1, _ = pair.Reserve1.SetString(w[5], 10)
		pair.Fee, _ = strconv.ParseInt(w[6], 10, 64)
		pairPath = append(pairPath, pair)
	}
	t.Log(protocol.GetAmountsOut(wethAddress, float64(44635314010151), pairPath))
}
