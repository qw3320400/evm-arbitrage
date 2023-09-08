package arbitrage

import (
	"context"
	"math"
	"math/big"
	"monitor/config"
	"monitor/protocol"
	"monitor/storage"
	"monitor/utils"
	"sync"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/shopspring/decimal"
)

var (
	_         utils.Keeper = &Arbitrage{}
	duplicate              = sync.Map{}
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
				Key:      pair.Address.Hex() + "+",
				Pair:     pair.Address,
				From:     pair.Token0,
				To:       pair.Token1,
				Distance: decimal.NewFromFloat(pair.Weight0),
			},
			&SwapEdge{
				Key:      pair.Address.Hex() + "-",
				Pair:     pair.Address,
				From:     pair.Token1,
				To:       pair.Token0,
				Distance: decimal.NewFromFloat(pair.Weight1),
			},
		)
	}
	path := g.FindCircle(common.HexToAddress(a.config.WETHAddress))
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
		ret += (addr.Hex() + ",")
	}
	return ret[:len(ret)-1]
}

func (a *Arbitrage) tryTrade(ctx context.Context, path []common.Address, pairs map[interface{}]interface{}) {
	if len(path) == 0 {
		return
	}
	// key := AddressList(path).String()
	// nowTimestamp := time.Now().Unix()
	// timestamp, loaded := duplicate.LoadOrStore(key, nowTimestamp)
	// if loaded && nowTimestamp < timestamp.(int64)+30 {
	// 	return
	// }
	// duplicate.Store(key, nowTimestamp)

	utils.Infof("tryTrade path :  %+v", path)
	// for {
	// 	var (
	// 		amountIn decimal.Decimal
	// 	)
	// 	pair, ok := pairs[path[0]].(*protocol.UniswapV2Pair)
	// 	if !ok {
	// 		return
	// 	}

	// }
}

func getAmountOut(amountIn, reserveIn, reserveOut decimal.Decimal) decimal.Decimal {
	amountInWithFee := amountIn.Mul(decimal.NewFromInt(997))
	numerator := amountInWithFee.Mul(reserveOut)
	denominator := reserveIn.Mul(decimal.NewFromInt(1000)).Add(amountInWithFee)
	return numerator.Div(denominator)
}
