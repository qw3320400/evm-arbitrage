package onchainmonitor

import (
	"context"
	"fmt"
	"monitor/action"
	"monitor/client"
	"monitor/config"
	"monitor/utils"
	"time"

	"github.com/ethereum/go-ethereum/core/types"
)

var _ Monitor = &EVMMonitor{}

type EVMMonitor struct {
	config  *config.Config
	actions []action.Action

	latestBlockNumber uint64
}

func NewEVMMonitor(ctx context.Context, conf *config.Config, actions []action.Action) *EVMMonitor {
	return &EVMMonitor{
		config:  conf,
		actions: actions,
	}
}

func (e *EVMMonitor) Init(ctx context.Context) error {
	for _, act := range e.actions {
		err := act.Init(ctx)
		if err != nil {
			return fmt.Errorf("init action fail %s", err)
		}
	}
	if e.config.Loop {
		go e.loopWatcher(ctx)
	} else {
		go e.subscribeWatcher(ctx)
	}
	return nil
}

func (e *EVMMonitor) loopWatcher(ctx context.Context) {
	if !e.config.Loop || e.config.LoopTime <= 0 {
		utils.Errorf("param error %+v", e)
		return
	}
	for {
		<-time.After(e.config.LoopTime)

		cli, err := client.GetETHClient(ctx, e.config.Node)
		if err != nil {
			utils.Errorf("get eth client fail %s", err)
			continue
		}
		bloNum, err := cli.BlockNumber(ctx)
		if err != nil {
			utils.Errorf("get block number fail %s", err)
			continue
		}
		if bloNum > e.latestBlockNumber {
			go e.onNewBlock(ctx, bloNum)
		}
	}
}

func (e *EVMMonitor) subscribeWatcher(ctx context.Context) {
	cli, err := client.GetETHClient(ctx, e.config.Node)
	if err != nil {
		utils.Errorf("get eth client fail %s", err)
		return
	}
	blocks := make(chan *types.Header)
	sub, err := cli.SubscribeNewHead(ctx, blocks)
	if err != nil {
		utils.Errorf("subscribe new header fail %s", err)
		return
	}
	for {
		select {
		case <-sub.Err():
			utils.Errorf("subscribe error %s", err)
		case block := <-blocks:
			bloNum := block.Number.Uint64()
			if bloNum > e.latestBlockNumber {
				go e.onNewBlock(ctx, bloNum)
			}
		}
	}
}

func (e *EVMMonitor) onNewBlock(ctx context.Context, blockNumber uint64) {
	if blockNumber <= e.latestBlockNumber {
		return
	}
	blockNumbers := []uint64{}
	if e.latestBlockNumber == 0 {
		blockNumbers = append(blockNumbers, blockNumber)
	} else {
		for i := e.latestBlockNumber + 1; i <= blockNumber; i++ {
			blockNumbers = append(blockNumbers, i)
		}
	}
	e.latestBlockNumber = blockNumber
	utils.Infof("on new block %d", blockNumbers)
	for _, act := range e.actions {
		err := act.OnNewBlockHandler(ctx, blockNumbers)
		if err != nil {
			utils.Errorf("handle new block fail %s", err)
			continue
		}
	}
}
