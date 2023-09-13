package arbitrage

import (
	"context"
	"fmt"
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
		r0, _ := big.NewFloat(0).SetInt(pair.Reserve0).Float64()
		r1, _ := big.NewFloat(0).SetInt(pair.Reserve1).Float64()
		pair.Weight0 = -math.Log(r1 / r0 * 0.997)
		pair.Weight1 = -math.Log(r0 / r1 * 0.997)

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

	var (
		amtIn, amtOut float64
		canTrade      bool
		fee           = a.getFee(len(path))
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
		var (
			pAmtIn  float64 = amtIn
			tokenIn         = a.config.WETHAddress
		)
		for i := len(path) - 1; i >= 0; i-- {
			pair := pairs[path[i]].(*protocol.UniswapV2Pair)
			if pair == nil {
				return
			}
			r0, _ := pair.Reserve0.Float64()
			r1, _ := pair.Reserve1.Float64()
			if pair.Token0 == tokenIn {
				pAmtIn = a.getAmountOut(pAmtIn, r0, r1)
				tokenIn = pair.Token1
			} else if pair.Token1 == tokenIn {
				pAmtIn = a.getAmountOut(pAmtIn, r1, r0)
				tokenIn = pair.Token0
			} else {
				return
			}
		}
		if tokenIn != a.config.WETHAddress {
			return
		}
		if pAmtIn <= amtIn+3*fee {
			if amtIn < 3*fee {
				return
			}
			amtIn *= 0.7
		} else {
			canTrade = true
			amtOut = pAmtIn
			break
		}
	}
	if canTrade {
		utils.Warnf("tryTrade ok %f %f %+v", amtIn, amtOut, path)
		for i := 0; i < len(path); i++ {
			pair := pairs[path[i]].(*protocol.UniswapV2Pair)
			fmt.Println("--------", pair.Reserve0, pair.Reserve1, pair.Weight0, pair.Weight1)
		}
	}
}

func (a *Arbitrage) getAmountOut(amountIn, reserveIn, reserveOut float64) float64 {
	if amountIn <= 0 || reserveIn <= 0 || reserveOut <= 0 {
		return 0
	}
	amountInWithFee := amountIn * 9970 * 0.99999
	numerator := amountInWithFee * reserveOut
	denominator := reserveIn*10000 + amountInWithFee
	return numerator / denominator
}

func (a *Arbitrage) getFee(pathLength int) float64 {
	gasUse := float64(100000 + 40000*pathLength)
	gasPrice := a.trader.GasPrice()
	return gasUse * gasPrice
}
