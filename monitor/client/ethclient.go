package client

import (
	"context"
	"fmt"

	"github.com/ethereum/go-ethereum/ethclient"
)

var ETHClient = map[string]*ethclient.Client{}

func GetETHClient(ctx context.Context, node string) (*ethclient.Client, error) {
	if ETHClient[node] != nil {
		return ETHClient[node], nil
	}
	cli, err := ethclient.DialContext(ctx, node)
	if err != nil {
		return nil, fmt.Errorf("eth client dial fail %s", err)
	}
	ETHClient[node] = cli
	return ETHClient[node], nil
}
