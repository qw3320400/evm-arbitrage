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
	duplicate              = cache.New(time.Minute, time.Hour)
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

func (a *Arbitrage) tryTrade(ctx context.Context, path []common.Address, pairs map[interface{}]interface{}) {
	if len(path) == 0 {
		return
	}
	key := AddressList(path).String()
	if _, ok := duplicate.Get(key); ok {
		return
	}
	duplicate.SetDefault(key, struct{}{})

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
		minRecieve    = a.getMinRecieve()
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
		pAmtOut := a.getAmountsOut(amtIn, pairPath)
		if pAmtOut <= amtIn+minRecieve {
			if amtIn < minRecieve {
				// utils.Warnf("------ %f %f %f %f %+v", amtIn, pAmtIn, (pAmtIn-amtIn)/math.Pow10(18), minRecieve/math.Pow10(18), path)
				return
			}
			amtIn *= 0.8
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
		err := a.trader.SwapV2(ctx, amtIn, amtOut, pairPath)
		if err != nil {
			utils.Errorf("SwapV2 fail %s", err)
		}
	}
}

func (a *Arbitrage) getAmountsOut(amountIn float64, pairPath []*protocol.UniswapV2Pair) float64 {
	var (
		pAmtOut  float64 = amountIn
		tokenOut         = a.config.WETHAddress
	)
	for _, pair := range pairPath {
		r0, _ := pair.Reserve0.Float64()
		r1, _ := pair.Reserve1.Float64()
		if pair.Token0 == tokenOut {
			pAmtOut = a.getAmountOut(pAmtOut, r0, r1, float64(pair.Fee))
			tokenOut = pair.Token1
		} else if pair.Token1 == tokenOut {
			pAmtOut = a.getAmountOut(pAmtOut, r1, r0, float64(pair.Fee))
			tokenOut = pair.Token0
		} else {
			return 0
		}
	}
	if tokenOut != a.config.WETHAddress {
		return 0
	}
	return pAmtOut
}

func (a *Arbitrage) getAmountOut(amountIn, reserveIn, reserveOut, fee float64) float64 {
	if amountIn <= 0 || reserveIn <= 0 || reserveOut <= 0 {
		return 0
	}
	amountInWithFee := amountIn * (protocol.FeeBase - fee)
	numerator := amountInWithFee * reserveOut
	denominator := reserveIn*protocol.FeeBase + amountInWithFee
	return numerator / denominator
}

func (a *Arbitrage) getMinRecieve() float64 {
	gasPrice := a.trader.GasPrice()
	eGasPrice := a.trader.ETHGasPrice()
	// TODO base chain
	minRecv := a.config.MinRecieve
	if gasPrice > math.Pow10(9) {
		minRecv *= 5
	}
	if eGasPrice > math.Pow10(9)*20 {
		minRecv *= 5
	} else if eGasPrice > math.Pow10(9)*100 {
		minRecv *= 20
	}
	return minRecv * math.Pow10(18)
}
