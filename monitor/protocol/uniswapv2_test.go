package protocol

import (
	"context"
	"math/big"
	"monitor/client"
	"monitor/utils"
	"testing"

	"github.com/ethereum/go-ethereum/common"
)

func TestPairInfo(t *testing.T) {
	ctx := context.Background()
	addr := "0x41d160033c222e6f3722ec97379867324567d883"
	pair := &UniswapV2Pair{
		Address: common.HexToAddress(addr),
	}
	viewcalls := []*client.ViewCall{}
	viewcalls = append(viewcalls, NewUniswapV2PairInfoCalls(pair)...)
	viewcalls = append(viewcalls, NewUniswapV2PairStateCalls(pair)...)
	cli, err := client.GetETHClient(ctx,
		"wss://distinguished-long-frog.base-mainnet.discover.quiknode.pro/9733b4ce6e9bbd6556771ea11f7a910d7ba0c50a/",
		"0xcA11bde05977b3631167028862bE2a173976CA11")
	if err != nil {
		t.Fatal(err)
	}
	callResult, err := cli.MultiViewCall(ctx, nil, viewcalls)
	if err != nil {
		t.Fatal(err)
	}
	pairs := map[common.Address]*UniswapV2Pair{
		common.HexToAddress(addr): pair,
	}
	UniswapV2PairCallResult(pairs, callResult)
	for _, pair := range pairs {
		utils.Infof("---- pair %s %+v", pair)
	}
}

func TestFileData(t *testing.T) {
	old := &UniswapV2Pair{
		Address:  common.HexToAddress("0x41d160033C222E6f3722EC97379867324567d883"),
		Token0:   common.HexToAddress("0x4200000000000000000000000000000000000006"),
		Token1:   common.HexToAddress("0xd9aAEc86B65D86f6A7B5B1b0c42FFA531710b6CA"),
		Reserve0: big.NewInt(0),
		Reserve1: big.NewInt(0),
		StateFromLogUpdate: &StateFromLogUpdate{
			BlockNumber: 1111,
			TxIndex:     22,
			LogIndex:    33,
		},
	}
	_, ok := old.Reserve0.SetString("3471420952218753871639", 10)
	if !ok {
		t.Fatal()
	}
	_, ok = old.Reserve1.SetString("5817201169869", 10)
	if !ok {
		t.Fatal()
	}
	t.Log(string(old.ToFileData()))
	new := &UniswapV2Pair{}
	err := new.FromFileData(old.ToFileData())
	if err != nil {
		t.Fatal(err)
	}
	if new.Address != old.Address ||
		new.Token0 != old.Token0 ||
		new.Token1 != old.Token1 ||
		new.Reserve0.Cmp(old.Reserve0) != 0 ||
		new.Reserve1.Cmp(old.Reserve1) != 0 ||
		new.Error != old.Error ||
		new.BlockNumber != old.BlockNumber ||
		new.TxIndex != old.TxIndex ||
		new.LogIndex != old.LogIndex {
		t.Fatalf("%+v %+v", new, old)
	}
}
