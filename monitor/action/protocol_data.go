package action

import (
	"context"
	"fmt"
	"monitor/client"
	"monitor/config"
	"monitor/protocol"
	"monitor/storage"

	"github.com/ethereum/go-ethereum/core/types"
)

var _ Action = &ProtocolData{}

type ProtocolData struct {
	config *config.Config
	pool   chan struct{}
}

func NewProtocolData(ctx context.Context, conf *config.Config) *ProtocolData {
	if conf.MaxConcurrency <= 0 {
		conf.MaxConcurrency = 1
	}
	return &ProtocolData{
		config: conf,
		pool:   make(chan struct{}, conf.MaxConcurrency),
	}
}

func (*ProtocolData) Init(context.Context) error {
	return nil
}

func (p *ProtocolData) OnNewBlockHandler(ctx context.Context, params ...interface{}) error {
	blockNumbers := params[0].([]uint64)
	return p.doNewBlockHandler(ctx, blockNumbers)
}

func (p *ProtocolData) doNewBlockHandler(ctx context.Context, blockNumbers []uint64) error {
	if len(blockNumbers) == 0 {
		return nil
	}
	return nil
}

func (p *ProtocolData) OnNewLogHandler(ctx context.Context, params ...interface{}) error {
	logs := params[0].([]*types.Log)
	return p.doNewLogHandler(ctx, logs)
}

func (p *ProtocolData) doNewLogHandler(ctx context.Context, logs []*types.Log) error {
	if len(logs) == 0 {
		return nil
	}
	err := p.doNewLogHandlerUniswapV2(ctx, logs)
	if err != nil {
		return fmt.Errorf("handle uniswapv2 data fail %s", err)
	}
	return nil
}

func (p *ProtocolData) doNewLogHandlerUniswapV2(ctx context.Context, logs []*types.Log) error {
	pairs, err := protocol.FilterUniswapV2PairFromLog(ctx, logs)
	if err != nil {
		return fmt.Errorf("filter uniswapv2 pair fail %s", err)
	}
	pairStore := storage.GetStorage(storage.StoreKeyUniswapv2Pairs)
	viewcalls := []*client.ViewCall{}
	for _, pair := range pairs {
		if data := pairStore.Load(pair.Address); data == nil {
			viewcalls = append(viewcalls, protocol.NewUniswapV2PairInfoCalls(pair)...)
		} else {
			pair.Token0 = data.(*protocol.UniswapV2Pair).Token0
			pair.Token1 = data.(*protocol.UniswapV2Pair).Token1
			pair.Error = data.(*protocol.UniswapV2Pair).Error
		}
	}
	cli, err := client.GetETHClient(ctx, p.config.Node, p.config.MulticallAddress)
	if err != nil {
		return fmt.Errorf("get eth client fail %s", err)
	}
	callResult, err := cli.MultiViewCall(ctx, nil, viewcalls)
	if err != nil {
		return fmt.Errorf("multi view call fail %s", err)
	}
	protocol.UniswapV2PairCallResult(pairs, callResult)
	var (
		storeKeys  = []interface{}{}
		storeDatas = []interface{}{}
	)
	for key, pair := range pairs {
		storeKeys = append(storeKeys, key)
		storeDatas = append(storeDatas, pair)
	}
	pairStore.Store(storeKeys, storeDatas)
	return nil
}
