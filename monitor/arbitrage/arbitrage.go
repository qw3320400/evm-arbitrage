package arbitrage

import (
	"context"
	"monitor/config"
	"monitor/utils"
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

func (*Arbitrage) Init(context.Context) error {
	return nil
}

func (*Arbitrage) ShutDown(context.Context) {

}
