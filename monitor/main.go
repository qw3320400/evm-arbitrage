package main

import (
	"context"
	"monitor/action"
	"monitor/arbitrage"
	"monitor/config"
	"monitor/datakeeper"
	"monitor/onchainmonitor"
	"monitor/trader"
	"monitor/utils"
	"net/http"
	_ "net/http/pprof"
	"os"
	"os/signal"
	"runtime/debug"
	"syscall"

	"github.com/ethereum/go-ethereum/common"
)

func main() {
	defer func() {
		if err := recover(); err != nil {
			utils.Errorf("%s\n%s", err, debug.Stack())
		}
	}()

	c := make(chan os.Signal, 1)
	signal.Notify(c, syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)

	go http.ListenAndServe(":8080", nil)

	ctx := context.Background()
	conf := &config.Config{
		Node:             "wss://linea-mainnet.infura.io/ws/v3/b13093bf13104631811fcef50795a4a9",
		MulticallAddress: common.HexToAddress("0xcA11bde05977b3631167028862bE2a173976CA11"),
		WETHAddress:      common.HexToAddress("0xe5D7C2a44FfDDf6b295A15c148167daaAf5Cf34f"),
		// Node:             "wss://dimensional-late-hill.discover.quiknode.pro/3a713b1cdb406ca2608e2a1b987eee47aedfcdaf/",
		// MulticallAddress: "0x9695fa23b27022c7dd752b7d64bb5900677ecc21",
		// WETHAddress:      common.HexToAddress("0xc02aaa39b223fe8d0a0e5c4f27ead9083c756cc2"),
		StoreFilePath: "./data_linea",
		FromAddress:   common.HexToAddress(os.Getenv("ADDRESS")),
		PrivateKey:    os.Getenv("PRIVATEKEY"),
		SwapAddress:   common.HexToAddress("0x1bB1BaCEA254Be45aA38446A03292337E82B8F8f"),
		MinRecieve:    0.0001,
		ETHNode:       "https://eth.llamarpc.com",
	}
	if len(conf.FromAddress) == 0 || len(conf.PrivateKey) == 0 {
		panic("missing env config ADDRESS or PRIVATEKEY")
	}
	traderKeeper := trader.NewTrader(ctx, conf)
	keepers := []utils.Keeper{
		traderKeeper,
		datakeeper.NewFileDataKeeper(ctx, conf.StoreFilePath),
		onchainmonitor.NewEVMMonitor(ctx, conf, []action.Action{
			action.NewProtocolData(ctx, conf),
		}),
		arbitrage.NewArbitrage(ctx, conf, traderKeeper),
	}
	for _, keeper := range keepers {
		err := keeper.Init(ctx)
		if err != nil {
			panic(err)
		}
	}

	for {
		select {
		case <-c:
			utils.Infof("program shutdown signal")
			close(c)
			for _, keeper := range keepers {
				keeper.ShutDown(ctx)
			}
			return
		}
	}
}
