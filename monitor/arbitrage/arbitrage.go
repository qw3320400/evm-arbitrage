package arbitrage

import (
	"context"
	"fmt"
	"math"
	"math/big"
	"monitor/config"
	"monitor/protocol"
	"monitor/storage"
	"monitor/utils"
	"time"

	"github.com/ethereum/go-ethereum/common"
)

var (
	_ utils.Keeper = &Arbitrage{}
)

type Arbitrage struct {
	config *config.Config
}

func NewArbitrage(ctx context.Context, conf *config.Config) *Arbitrage {
	return &Arbitrage{
		config: conf,
	}
}

func (a *Arbitrage) Init(ctx context.Context) error {
	go a.loopWatcher(ctx)
	return nil
}

func (*Arbitrage) ShutDown(context.Context) {

}

func (a *Arbitrage) loopWatcher(ctx context.Context) {
	for {
		<-time.After(time.Millisecond * 50)

		err := a.findArbitrage(ctx)
		if err != nil {
			utils.Warnf("find arbitrage fail %s", err)
		}
	}
}

func (a *Arbitrage) findArbitrage(ctx context.Context) error {
	// startTime := time.Now()
	store := storage.GetStorage(storage.StoreKeyUniswapv2Pairs)
	datas := store.LoadAll()
	// utils.Infof("load data finish in %s", time.Since(startTime))
	path := a.bellmanFord(ctx, datas)
	// utils.Infof("bellmanFord finish in %s", time.Since(startTime))
	if len(path) == 0 {
		return nil
	}
	err := a.tryPath(ctx, path, datas)
	if err != nil {
		return fmt.Errorf("try path fail %s", err)
	}
	// utils.Infof("find arbitrage finish in %s", time.Since(startTime))
	return nil
}

func (a *Arbitrage) bellmanFord(ctx context.Context, datas []interface{}) []common.Address {
	var (
		vertices  = []int{}
		edges     = []*BF2Edge{}
		addrToIdx = map[common.Address]int{}
		addresses = []common.Address{}
	)
	for _, data := range datas {
		pair := data.(*protocol.UniswapV2Pair)
		if pair.Error {
			continue
		}
		r0, _ := big.NewFloat(0).SetInt(pair.Reserve0).Float64()
		r1, _ := big.NewFloat(0).SetInt(pair.Reserve1).Float64()
		pair.Weight0 = -math.Log(r1 / r0 * 0.997)
		pair.Weight1 = -math.Log(r0 / r1 * 0.997)

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
	ethAddr := common.HexToAddress(a.config.WETHAddress)
	ethIdx := addrToIdx[ethAddr]
	g := NewBF2Graph(edges, vertices)
	loop := g.FindArbitrageLoop(ethIdx)
	var (
		addressPath   = []common.Address{}
		hasOtherToken bool
	)
	for _, idx := range loop {
		if idx != ethIdx {
			hasOtherToken = true
		}
		if len(addressPath) > 0 && addresses[idx] == addressPath[len(addressPath)-1] {
			continue
		}
		addressPath = append(addressPath, addresses[idx])
	}
	if !hasOtherToken || len(addressPath) == 0 {
		return []common.Address{}
	}
	if addressPath[0] != ethAddr {
		addressPath = append([]common.Address{ethAddr}, addressPath...)
	}
	if addressPath[len(addressPath)-1] != ethAddr {
		addressPath = append(addressPath, ethAddr)
	}
	if len(addressPath) == 3 {
		return []common.Address{}
	}
	return addressPath
}

func (a *Arbitrage) tryPath(ctx context.Context, path []common.Address, datas []interface{}) error {
	pairList := make([]*protocol.UniswapV2Pair, len(path)-1)
	priceList := make([]float64, len(path)-1)
	for i := 1; i < len(path); i++ {
		tokenIn := path[i-1]
		tokenOut := path[i]
		for _, data := range datas {
			pair := data.(*protocol.UniswapV2Pair)
			if (pair.Token0 != tokenIn && pair.Token0 != tokenOut) ||
				(pair.Token1 != tokenIn && pair.Token1 != tokenOut) {
				continue
			}
			r0, _ := big.NewFloat(0).SetInt(pair.Reserve0).Float64()
			r1, _ := big.NewFloat(0).SetInt(pair.Reserve1).Float64()
			var price float64
			if tokenIn == pair.Token0 {
				price = r1 / r0
			} else {
				price = r0 / r1
			}
			if price > priceList[i-1] {
				priceList[i-1] = price
				pairList[i-1] = pair
			}
		}
		if pairList[i-1] == nil {
			utils.Errorf("can not find pair %s %s %+v", tokenIn, tokenOut, path)
			return nil
		}
	}
	var finalPrice float64
	for i, pair := range pairList {
		finalPrice *= priceList[i]
		utils.Warnf("try path pair -- %s %s %s %f", pair.Address, pair.Reserve0, pair.Reserve1, priceList[i])
	}
	utils.Warnf("try path final %f", finalPrice)
	return nil
}
