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
	"sync"
	"time"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
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
	go func() {
		for {
			<-time.After(time.Second)
			err := e.subscribeFilter(ctx)
			if err != nil {
				utils.Warnf("subscribe filter fail %s", err)
			}
		}
	}()
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

func (e *EVMMonitor) loopFilter(ctx context.Context) error {
	cli, err := client.GetETHClient(ctx, e.config.Node, e.config.MulticallAddress)
	if err != nil {
		return fmt.Errorf("get eth client fail %s", err)
	}
	blockNumber, err := cli.BlockNumber(ctx)
	if err != nil {
		return fmt.Errorf("get block number fail %s", err)
	}
	filter := ethereum.FilterQuery{
		FromBlock: big.NewInt(int64(blockNumber)),
		Topics: [][]common.Hash{
			{
				protocol.UniswapV2PairEventSyncSign,
				protocol.UniswapV2PairEventSwapSign,
				// protocol.UniswapV2PairEventSyncUint256Sign,
			},
		},
	}
	var filterID hexutil.Uint64
	err = cli.Client.Client().CallContext(ctx, &filterID, "eth_newFilter", filter)
	if err != nil {
		return fmt.Errorf("new filter fail %s", err)
	}
	for {
		<-time.After(time.Second)
		var logs []*types.Log
		err = cli.Client.Client().CallContext(ctx, &logs, "eth_getFilterLogs", filterID)
		if err != nil {
			return fmt.Errorf("get filter logs fail %s", err)
		}
		go e.onNewLogs(ctx, logs)
	}
}

func (e *EVMMonitor) loopLogs(ctx context.Context) error {
	cli, err := client.GetETHClient(ctx, e.config.Node, e.config.MulticallAddress)
	if err != nil {
		return fmt.Errorf("get eth client fail %s", err)
	}
	var (
		lastBlockNumber uint64
		filter          = ethereum.FilterQuery{
			Topics: [][]common.Hash{
				{
					protocol.UniswapV2PairEventSyncSign,
					protocol.UniswapV2PairEventSwapSign,
					// protocol.UniswapV2PairEventSyncUint256Sign,
				},
			},
		}
	)
	for {
		<-time.After(time.Second)

		blockNumber, err := cli.BlockNumber(ctx)
		if err != nil {
			return fmt.Errorf("get block number fail %s", err)
		}
		if blockNumber <= lastBlockNumber {
			continue
		}
		if lastBlockNumber > 0 {
			filter.FromBlock = big.NewInt(int64(lastBlockNumber + 1))
		} else {
			filter.FromBlock = big.NewInt(int64(blockNumber))
		}
		lastBlockNumber = blockNumber
		logs, err := cli.FilterLogs(ctx, filter)
		if err != nil {
			return fmt.Errorf("get filter logs fail %s", err)
		}
		if len(logs) <= 0 {
			continue
		}
		tmp := make([]*types.Log, 0, len(logs))
		for _, log := range logs {
			tmp = append(tmp, &log)
		}
		go e.onNewLogs(ctx, tmp)
	}
}

func (e *EVMMonitor) subscribeWatcher(ctx context.Context) error {
	cli, err := client.GetETHClient(ctx, e.config.Node, e.config.MulticallAddress)
	if err != nil {
		return fmt.Errorf("get eth client fail %s", err)
	}
	blocks := make(chan *types.Header)
	sub, err := cli.SubscribeNewHead(ctx, blocks)
	if err != nil {
		return fmt.Errorf("subscribe new header fail %s", err)
	}
	for {
		select {
		case err = <-sub.Err():
			return fmt.Errorf("subscribe error %s", err)
		case block := <-blocks:
			bloNum := block.Number.Uint64()
			if bloNum > e.latestBlockNumber {
				go e.onNewBlock(ctx, bloNum)
			}
		}
	}
}

func (e *EVMMonitor) subscribeFilter(ctx context.Context) error {
	cli, err := client.GetETHClient(ctx, e.config.Node, e.config.MulticallAddress)
	if err != nil {
		return fmt.Errorf("get eth client fail %s", err)
	}
	logChan := make(chan types.Log, 1000)
	blockNumber, err := cli.BlockNumber(ctx)
	if err != nil {
		return fmt.Errorf("get block number fail %s", err)
	}
	filter := ethereum.FilterQuery{
		FromBlock: big.NewInt(int64(blockNumber)),
		Topics: [][]common.Hash{
			{
				protocol.UniswapV2PairEventSyncSign,
				protocol.UniswapV2PairEventSwapSign,
				// protocol.UniswapV2PairEventSyncUint256Sign,
			},
		},
	}
	sub, err := cli.SubscribeFilterLogs(ctx, filter, logChan)
	if err != nil {
		return fmt.Errorf("subscribe filter fail %s", err)
	}
	var (
		logs     = []*types.Log{}
		logsLock = sync.Mutex{}
		subErr   error
	)
	go func() {
		for log := range logChan {
			if subErr != nil {
				return
			}
			copy := log
			logsLock.Lock()
			logs = append(logs, &copy)
			logsLock.Unlock()
		}
	}()
	for {
		<-time.After(time.Millisecond * 10)
		select {
		case subErr = <-sub.Err():
			if subErr != nil {
				return fmt.Errorf("subscribe error %s", subErr)
			}
		default:
			if len(logs) == 0 {
				continue
			}
			logsLock.Lock()
			tmp := logs
			logs = []*types.Log{}
			logsLock.Unlock()
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
