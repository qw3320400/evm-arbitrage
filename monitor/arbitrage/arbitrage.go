package arbitrage

import (
	"context"
	"monitor/config"
	"monitor/utils"
)

var arbitrageSingleRoutine = utils.NewSingleRoutine(context.Background())

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

func (a *Arbitrage) OnNewBlockHandler(ctx context.Context, params ...interface{}) error {
	blockNumbers := params[0].([]uint64)
	arbitrageSingleRoutine.Run(ctx, func(ctx context.Context) {
		err := a.doNewBlockHandler(ctx, blockNumbers)
		if err != nil {
			utils.Warnf("arbitrage handle new block fail %+v %s", blockNumbers, err)
		}
	})
	return nil
}

func (a *Arbitrage) doNewBlockHandler(ctx context.Context, blockNumbers []uint64) error {
	utils.Infof("arbitrage handle new block %+v", blockNumbers)
	return nil
}
