package main

import (
	"context"
	"monitor/action"
	"monitor/config"
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
		Node:             "wss://polygon.llamarpc.com",
		MulticallAddress: "0x275617327c958bD06b5D6b871E7f491D76113dd8",
		// Node:             "wss://dimensional-late-hill.discover.quiknode.pro/3a713b1cdb406ca2608e2a1b987eee47aedfcdaf/",
		// MulticallAddress: "0x9695fa23b27022c7dd752b7d64bb5900677ecc21",
		StoreFilePath:  "./data",
		MaxConcurrency: 5,
	}
	monitor := onchainmonitor.NewEVMMonitor(ctx, conf, []action.Action{
		action.NewProtocolData(ctx, conf),
		action.NewArbitrage(ctx, conf),
	})
	err := monitor.Init(ctx)
	if err != nil {
		panic(err)
	}

	for {
		select {
		case <-c:
			utils.Infof("program shutdown signal")
			close(c)
			monitor.ShutDown(ctx)
			return
		}
	}
}
