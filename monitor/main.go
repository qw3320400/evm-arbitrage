package main

import (
	"context"
	"monitor/action"
	"monitor/arbitrage"
	"monitor/config"
	"monitor/datakeeper"
	"monitor/onchainmonitor"
	"monitor/utils"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	c := make(chan os.Signal, 1)
	signal.Notify(c, syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)

	ctx := context.Background()
	conf := &config.Config{
		Node:             "wss://distinguished-long-frog.base-mainnet.discover.quiknode.pro/9733b4ce6e9bbd6556771ea11f7a910d7ba0c50a/",
		MulticallAddress: "0xcA11bde05977b3631167028862bE2a173976CA11",
		WETHAddress:      "0x4200000000000000000000000000000000000006",
		// Node:             "wss://dimensional-late-hill.discover.quiknode.pro/3a713b1cdb406ca2608e2a1b987eee47aedfcdaf/",
		// MulticallAddress: "0x9695fa23b27022c7dd752b7d64bb5900677ecc21",
		// WETHAddress:      "0xc02aaa39b223fe8d0a0e5c4f27ead9083c756cc2",
		StoreFilePath:  "./data",
		MaxConcurrency: 5,
	}
	keepers := []utils.Keeper{
		datakeeper.NewFileDataKeeper(ctx, conf.StoreFilePath),
		onchainmonitor.NewEVMMonitor(ctx, conf, []action.Action{
			action.NewProtocolData(ctx, conf),
		}),
		arbitrage.NewArbitrage(ctx, conf),
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
