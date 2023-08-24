package onchainmonitor

import (
	"context"
	"monitor/config"
	"testing"
	"time"
)

func TestEVMMonitorLoop(t *testing.T) {
	ctx := context.Background()
	conf := &config.Config{
		Node: "https://polygon.llamarpc.com",
	}
	e := NewEVMMonitor(ctx, conf, nil)
	err := e.Init(ctx)
	if err != nil {
		t.Fatal(err)
	}
	time.Sleep(time.Second * 10)
}

func TestEVMMonitorSubscribe(t *testing.T) {
	ctx := context.Background()
	conf := &config.Config{
		Node: "wss://polygon.llamarpc.com",
	}
	e := NewEVMMonitor(ctx, conf, nil)
	err := e.Init(ctx)
	if err != nil {
		t.Fatal(err)
	}
	time.Sleep(time.Second * 10)
}
