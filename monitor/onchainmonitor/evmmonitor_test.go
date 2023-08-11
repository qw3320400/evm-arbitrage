package onchainmonitor

import (
	"context"
	"fmt"
	"math/big"
	"monitor/action"
	"monitor/client"
	"monitor/config"
	"monitor/utils"
	"testing"
	"time"
)

func TestEVMMonitorLoop(t *testing.T) {
	ctx := context.Background()
	conf := &config.Config{
		Node:     "https://polygon.llamarpc.com",
		Loop:     true,
		LoopTime: time.Millisecond * 100,
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
	acts := []action.Action{
		NewTestAction(ctx, conf),
	}
	e := NewEVMMonitor(ctx, conf, acts)
	err := e.Init(ctx)
	if err != nil {
		t.Fatal(err)
	}
	time.Sleep(time.Second * 10)
}

var _ action.Action = &TestAction{}

type TestAction struct {
	config *config.Config
}

func NewTestAction(ctx context.Context, conf *config.Config) *TestAction {
	return &TestAction{
		config: conf,
	}
}

func (*TestAction) Init(context.Context) error {
	return nil
}

func (a *TestAction) OnNewBlockHandler(ctx context.Context, blockNumber uint64) error {
	cli, err := client.GetETHClient(ctx, a.config.Node)
	if err != nil {
		return fmt.Errorf("get eth client fail %s", err)
	}
	block, err := cli.BlockByNumber(ctx, big.NewInt(int64(blockNumber)))
	if err != nil {
		return fmt.Errorf("get block fail %s", err)
	}
	utils.Infof("block number %d, tx number %d", blockNumber, len(block.Transactions()))
	return nil
}
