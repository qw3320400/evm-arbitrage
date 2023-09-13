package trader

import (
	"context"
	"fmt"
	"math/big"
	"monitor/client"
	"monitor/config"
	"monitor/utils"
	"time"
)

var (
	_ utils.Keeper = &Trader{}
)

type Trader struct {
	config   *config.Config
	gasPrice *big.Int
}

func NewTrader(ctx context.Context, conf *config.Config) *Trader {
	return &Trader{
		config:   conf,
		gasPrice: big.NewInt(15000000),
	}
}

func (t *Trader) Init(ctx context.Context) error {
	go t.loopWatcher(ctx)
	return nil
}

func (*Trader) ShutDown(context.Context) {

}

func (t *Trader) loopWatcher(ctx context.Context) {
	for {
		err := t.fetchGasPrice(ctx)
		if err != nil {
			utils.Warnf("fetch gas price fail %s", err)
		}

		<-time.After(time.Second * 5)
	}
}

func (t *Trader) fetchGasPrice(ctx context.Context) error {
	cli, err := client.GetETHClient(ctx, t.config.Node, t.config.MulticallAddress)
	if err != nil {
		return fmt.Errorf("get eth client fail %s", err)
	}
	t.gasPrice, err = cli.SuggestGasPrice(ctx)
	if err != nil {
		return fmt.Errorf("get suggest gas price fail %s", err)
	}
	utils.Infof("current suggest gas price is %f gwei", t.GasPrice()/1000000000)
	return nil
}

func (t *Trader) GasPrice() float64 {
	gp, _ := t.gasPrice.Float64()
	return gp
}
