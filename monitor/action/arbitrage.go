package action

import (
	"context"
	"monitor/config"
)

var _ Action = &Arbitrage{}

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

func (a *Arbitrage) OnNewBlockHandler(ctx context.Context, blockNumber uint64) error {
	return nil
}
