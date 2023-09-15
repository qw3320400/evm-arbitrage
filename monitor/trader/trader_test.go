package trader

import (
	"context"
	"monitor/config"
	"testing"

	"github.com/ethereum/go-ethereum/common"
)

func TestGasPrice(t *testing.T) {
	ctx := context.Background()
	conf := &config.Config{
		Node:             "wss://distinguished-long-frog.base-mainnet.discover.quiknode.pro/9733b4ce6e9bbd6556771ea11f7a910d7ba0c50a/",
		MulticallAddress: common.HexToAddress("0xcA11bde05977b3631167028862bE2a173976CA11"),
		WETHAddress:      common.HexToAddress("0x4200000000000000000000000000000000000006"),
	}
	trader := NewTrader(ctx, conf)
	err := trader.fetchGasPrice(ctx)
	if err != nil {
		t.Fatal(err)
	}
	t.Log(trader.GasPrice() * 200000 / 1000000000000000000)
}
