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

func (a *Arbitrage) OnNewBlockHandler(ctx context.Context, blockNumbers []uint64) error {
	// cli, err := client.GetETHClient(ctx, a.config.Node)
	// if err != nil {
	// 	return fmt.Errorf("get eth client fail %s", err)
	// }
	// cli.Client().Call()
	return nil
}
