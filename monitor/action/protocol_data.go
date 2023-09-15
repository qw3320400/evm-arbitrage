package action

import (
	"context"
	"fmt"
	"monitor/client"
	"monitor/config"
	"monitor/protocol"
	"monitor/storage"
	"monitor/utils"

	"github.com/ethereum/go-ethereum/core/types"
)

var _ Action = &ProtocolData{}

type ProtocolData struct {
	config *config.Config
}

func NewProtocolData(ctx context.Context, conf *config.Config) *ProtocolData {
	return &ProtocolData{
		config: conf,
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
			prePair := data.(*protocol.UniswapV2Pair)
			pair.Token0 = prePair.Token0
			pair.Token1 = prePair.Token1
			pair.Fee = prePair.Fee
			pair.Error = prePair.Error
		}
	}
	if len(viewcalls) > 0 {
		cli, err := client.GetETHClient(ctx, p.config.Node, p.config.MulticallAddress)
		if err != nil {
			return fmt.Errorf("get eth client fail %s", err)
		}
		callResult, err := cli.MultiViewCall(ctx, nil, viewcalls)
		if err != nil {
			return fmt.Errorf("multi view call fail %s", err)
		}
		protocol.UniswapV2PairCallResult(pairs, callResult)
	}
	fees, err := protocol.FilterUniswapV2FeeFromLog(ctx, logs)
	if err != nil {
		return fmt.Errorf("filter uniswapv2 fee fail %s", err)
	}
	for addr, fee := range fees {
		utils.Infof("caculated pair swap fee %s %d", addr, fee)
		if pair, ok := pairs[addr]; ok {
			pair.Fee = fee
		}
	}
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
