package arbitrage

import (
	"context"
	"math"
	"math/big"
	"monitor/config"
	"monitor/protocol"
	"monitor/storage"
	"monitor/trader"
	"monitor/utils"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/patrickmn/go-cache"
)

var (
	_         utils.Keeper = &Arbitrage{}
	duplicate              = cache.New(time.Minute, 12*time.Hour)
	failPair               = cache.New(time.Second*30, time.Minute*10)
)

type Arbitrage struct {
	config *config.Config
	trader *trader.Trader
}

func NewArbitrage(ctx context.Context, conf *config.Config, trader *trader.Trader) *Arbitrage {
	if conf.MinRecieve <= 0 {
		conf.MinRecieve = 0.0001
	}
	return &Arbitrage{
		config: conf,
		trader: trader,
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
		<-time.After(time.Millisecond * 100)

		err := a.findArbitrage(ctx)
		if err != nil {
			utils.Warnf("find arbitrage fail %s", err)
		}
	}
}

func (a *Arbitrage) findArbitrage(ctx context.Context) error {
	// startTime := time.Now()
	store := storage.GetStorage(storage.StoreKeyUniswapv2Pairs)
	pairs := store.LoadAll()
	// utils.Infof("load data finish in %s", time.Since(startTime))
	g := NewSwapGraph()
	for _, p := range pairs {
		pair := p.(*protocol.UniswapV2Pair)
		if pair.Error || pair.Fee < 0 {
			continue
		}
		r0, _ := big.NewFloat(0).SetInt(pair.Reserve0).Float64()
		r1, _ := big.NewFloat(0).SetInt(pair.Reserve1).Float64()
		pair.Weight0 = -math.Log10(r1 / r0 * (protocol.FeeBase - float64(pair.Fee)) / protocol.FeeBase)
		pair.Weight1 = -math.Log10(r0 / r1 * (protocol.FeeBase - float64(pair.Fee)) / protocol.FeeBase)

		g.AddVertices(pair.Token0, pair.Token1)
		g.AddEdges(
			&SwapEdge{
				Key:      string(pair.Address.Bytes()) + "+",
				Pair:     pair.Address,
				From:     pair.Token0,
				To:       pair.Token1,
				Distance: pair.Weight0,
			},
			&SwapEdge{
				Key:      string(pair.Address.Bytes()) + "-",
				Pair:     pair.Address,
				From:     pair.Token1,
				To:       pair.Token0,
				Distance: pair.Weight1,
			},
		)
	}
	path := g.FindCircle(a.config.WETHAddress)
	if len(path) > 0 {
		go a.tryTrade(ctx, path, pairs)
		go a.failPair(ctx, path)
	}
	// utils.Infof("find arbitrage finish in %s", time.Since(startTime))
	return nil
}

type AddressList []common.Address

func (l AddressList) String() string {
	ret := ""
	if len(l) == 0 {
		return ret
	}
	for _, addr := range l {
		ret += string(addr.Bytes())
	}
	return ret
}

func (a *Arbitrage) failPair(ctx context.Context, path []common.Address) {
	var (
		maxPair  common.Address
		maxCount int64
	)
	for _, pair := range path {
		pairStr := pair.String()
		var count int64 = 1
		if c, ok := failPair.Get(pairStr); ok {
			count = c.(int64) + 1
		}
		failPair.SetDefault(pairStr, count)
		if count > 10000 && count > maxCount {
			maxCount = count
			maxPair = pair
		}
	}
	if maxCount > 0 {
		utils.Errorf("found potential fail pair %d %s", maxCount, maxPair)
	}
}

func (a *Arbitrage) tryTrade(ctx context.Context, path []common.Address, pairs map[interface{}]interface{}) {
	if len(path) == 0 {
		return
	}
	key := AddressList(path).String()
	dupCount, ok := duplicate.Get(key)
	if ok {
		return
	}
	duplicate.SetDefault(key, int(0))

	pairPath := make([]*protocol.UniswapV2Pair, 0, len(path))
	for i := len(path) - 1; i >= 0; i-- {
		pair, ok := pairs[path[i]].(*protocol.UniswapV2Pair)
		if !ok {
			return
		}
		pairPath = append(pairPath, pair)
	}
	var (
		amtIn, amtOut float64
		canTrade      bool
		minRecieve    = a.trader.EstimateFee(len(pairPath)) * 1.5
	)
	pair0 := pairs[path[len(path)-1]].(*protocol.UniswapV2Pair)
	if pair0 == nil {
		return
	}
	if pair0.Token0 == a.config.WETHAddress {
		amtIn, _ = pair0.Reserve0.Float64()
		amtIn *= 0.1
	} else if pair0.Token1 == a.config.WETHAddress {
		amtIn, _ = pair0.Reserve1.Float64()
		amtIn *= 0.1
	} else {
		return
	}
	for {
		pAmtOut := protocol.GetAmountsOut(a.config.WETHAddress, amtIn, pairPath)
		if pAmtOut <= amtIn+minRecieve {
			if amtIn < minRecieve {
				// utils.Warnf("------ %f %f %f %f %+v", amtIn, pAmtIn, (pAmtIn-amtIn)/math.Pow10(18), minRecieve/math.Pow10(18), path)
				return
			}
			amtIn *= 0.9
		} else {
			canTrade = true
			amtOut = pAmtOut
			break
		}
	}
	if canTrade {
		utils.Warnf("tryTrade ok %f %f %f %f", amtIn, amtOut, (amtOut-amtIn)/math.Pow10(18), minRecieve/math.Pow10(18))
		for i := len(path) - 1; i >= 0; i-- {
			pair := pairs[path[i]].(*protocol.UniswapV2Pair)
			utils.Warnf("--------pair %s %s %s %s %s %d", pair.Address, pair.Token0, pair.Token1, pair.Reserve0, pair.Reserve1, pair.Fee)
		}
		err := a.trader.SwapV2(ctx, amtIn, pairPath)
		if err != nil {
			failCount := 0
			if dupCount != nil {
				failCount = dupCount.(int)
			}
			failCount++
			duplicate.Set(key, failCount, time.Minute*time.Duration(10*failCount))
			utils.Errorf("SwapV2 fail %s", err)
		}
	}
}
