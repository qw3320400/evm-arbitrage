package trader

import (
	"context"
	"fmt"
	"math/big"
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

func TestBytes(t *testing.T) {
	bi, _ := big.NewInt(123123).SetString("46922874771987008", 10)
	str := fmt.Sprintf("%020x", bi)
	pairs := []*Route{
		{
			Pair:      common.HexToAddress("0xe12e18f4aa1e923c0be9db1af30f2547ebc31530"),
			Direction: true,
			Fee:       big.NewInt(31),
		},
		{
			Pair:      common.HexToAddress("0x6f58bf1d5d344c01716e8b8b585aee17ad4def86"),
			Direction: false,
			Fee:       big.NewInt(102),
		},
	}
	for _, pair := range pairs {
		var boolToInt int
		if pair.Direction {
			boolToInt = 1
		}
		str += fmt.Sprintf("%040x%02x%04x", pair.Pair, boolToInt, pair.Fee)
	}
	t.Log(str)
	t.Log(common.Bytes2Hex(common.FromHex(str)))
}
