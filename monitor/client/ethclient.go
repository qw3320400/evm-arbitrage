package client

import (
	"context"
	"fmt"
	"monitor/abi"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
)

type ETHClient struct {
	*ethclient.Client

	multicallAddress string
}

var ETHClientMap = map[string]*ETHClient{}

func GetETHClient(ctx context.Context, node, multicallAddress string) (*ETHClient, error) {
	if ETHClientMap[node] != nil {
		return ETHClientMap[node], nil
	}
	cli, err := ethclient.DialContext(ctx, node)
	if err != nil {
		return nil, fmt.Errorf("eth client dial fail %s", err)
	}
	ETHClientMap[node] = &ETHClient{cli, multicallAddress}
	return ETHClientMap[node], nil
}

func (e *ETHClient) Multicall2(ctx context.Context, opt *bind.TransactOpts, msgs ...*ethereum.CallMsg) ([]*abi.Multicall2Result, error) {
	to := common.HexToAddress(e.multicallAddress)
	input, err := abi.Multicall2ABIInstance.Methods["tryAggregate"].Inputs.Pack(false, NewMulticall2Calls(msgs...))
	if err != nil {
		return nil, fmt.Errorf("pack input fail %s", err)
	}
	resBody, err := e.Client.CallContract(ctx, ethereum.CallMsg{
		To:   &to,
		Data: append(abi.Multicall2ABIInstance.Methods["tryAggregate"].ID, input...),
	}, nil)
	if err != nil {
		return nil, fmt.Errorf("call contract fail %s", err)
	}
	res, err := abi.Multicall2ABIInstance.Methods["tryAggregate"].Outputs.Unpack(resBody)
	if err != nil {
		return nil, fmt.Errorf("unpack result fail %s", err)
	}
	if len(res) <= 0 {
		return nil, fmt.Errorf("res length is 0")
	}
	resStruct, ok := res[0].([]struct {
		Success    bool    "json:\"success\""
		ReturnData []uint8 "json:\"returnData\""
	})
	if !ok {
		return nil, fmt.Errorf("res type error %+v", res)
	}
	ret := make([]*abi.Multicall2Result, len(resStruct))
	for i, one := range resStruct {
		ret[i] = &abi.Multicall2Result{
			Success:    one.Success,
			ReturnData: one.ReturnData,
		}
	}
	return ret, nil
}

func NewMulticall2Calls(msgs ...*ethereum.CallMsg) []abi.Multicall2Call {
	calls := []abi.Multicall2Call{}
	for _, msg := range msgs {
		call := abi.Multicall2Call{
			Target:   *msg.To,
			CallData: msg.Data,
		}
		calls = append(calls, call)
	}
	return calls
}
