package onchainmonitor

import (
	"context"
	"fmt"
	"math/big"
	"monitor/action"
	"monitor/client"
	"monitor/config"
	"monitor/protocol"
	"monitor/utils"
	"time"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
)

var (
	_ Monitor      = &EVMMonitor{}
	_ utils.Keeper = &EVMMonitor{}
)

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
	// go e.loopWatcher(ctx)
	// go e.subscribeWatcher(ctx)
	go e.subscribeFilter(ctx)
	return nil
}

func (e *EVMMonitor) ShutDown(ctx context.Context) {

}

func (e *EVMMonitor) loopWatcher(ctx context.Context) {
	for {
		<-time.After(time.Second)

		cli, err := client.GetETHClient(ctx, e.config.Node, e.config.MulticallAddress)
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
	cli, err := client.GetETHClient(ctx, e.config.Node, e.config.MulticallAddress)
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

func (e *EVMMonitor) subscribeFilter(ctx context.Context) {
	cli, err := client.GetETHClient(ctx, e.config.Node, e.config.MulticallAddress)
	if err != nil {
		utils.Errorf("get eth client fail %s", err)
		return
	}
	logChan := make(chan types.Log, 1000)
	blockNumber, err := cli.BlockNumber(ctx)
	if err != nil {
		utils.Errorf("get block number fail %s", err)
		return
	}
	filter := ethereum.FilterQuery{
		FromBlock: big.NewInt(int64(blockNumber)),
		Topics: [][]common.Hash{
			{
				protocol.UniswapV2PairEventSyncSign,
			},
		},
	}
	sub, err := cli.SubscribeFilterLogs(ctx, filter, logChan)
	if err != nil {
		utils.Errorf("subscribe filter fail %s", err)
		return
	}
	logs := []*types.Log{}
	for {
		select {
		case <-sub.Err():
			utils.Errorf("subscribe error %s", err)
		case log := <-logChan:
			logs = append(logs, &log)
		default:
			if len(logs) == 0 {
				continue
			}
			tmp := logs
			logs = []*types.Log{}
			go e.onNewLogs(ctx, tmp)
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
	utils.Infof("on new blocks %d", blockNumbers)
	for _, act := range e.actions {
		tmp := act
		go func() {
			err := tmp.OnNewBlockHandler(ctx, blockNumbers)
			if err != nil {
				utils.Warnf("handle new block fail %d %s", blockNumbers, err)
			}
		}()
	}
}

func (e *EVMMonitor) onNewLogs(ctx context.Context, logs []*types.Log) {
	utils.Infof("on new logs %d", len(logs))
	for _, act := range e.actions {
		tmp := act
		go func() {
			err := tmp.OnNewLogHandler(ctx, logs)
			if err != nil {
				utils.Warnf("handle new logs fail %s", err)
			}
		}()
	}
}
