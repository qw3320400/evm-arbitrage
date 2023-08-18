package action

import (
	"context"
	"fmt"
	"math/big"
	"monitor/client"
	"monitor/config"
	"monitor/utils"
	"time"

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
	for _, blockNumber := range blockNumbers {
		p.pool <- struct{}{}
		go func(num uint64) {
			utils.Retry(ctx, func(ctx context.Context) error {
				return p.fetchBlockProtocolData(ctx, num)
			}, time.Second, -1)
			<-p.pool
		}(blockNumber)
	}
	return nil
}

func (p *ProtocolData) fetchBlockProtocolData(ctx context.Context, blockNumber uint64) error {
	blockData, err := p.getBlockData(ctx, blockNumber)
	if err != nil {
		return fmt.Errorf("get block data fail %s", err)
	}
	return nil
}

type BlockData struct {
	Receipts []*types.Receipt
}

func (p *ProtocolData) getBlockData(ctx context.Context, blockNumber uint64) (*BlockData, error) {
	startTime := time.Now()

	cli, err := client.GetETHClient(ctx, p.config.Node, p.config.MulticallAddress)
	if err != nil {
		return nil, fmt.Errorf("get eth client fail %s", err)
	}
	blockReceipt := []*types.Receipt{}
	var receiptErr error
	utils.Retry(ctx, func(ctx context.Context) error {
		err = cli.Client.Client().CallContext(ctx, &blockReceipt, "eth_getBlockReceipts", big.NewInt(int64(blockNumber)))
		if err != nil {
			receiptErr = err
			return nil
		}
		if len(blockReceipt) <= 0 {
			return fmt.Errorf("can't get block %d receipt, length is 0", blockNumber)
		}
		return nil
	}, time.Millisecond*200, -1)
	if receiptErr != nil {
		return nil, fmt.Errorf("get block receipt fail %s", err)
	}
	utils.Infof("get %d block and receipt done in %s", blockNumber, time.Since(startTime))
	return &BlockData{
		Receipts: blockReceipt,
	}, nil
}
