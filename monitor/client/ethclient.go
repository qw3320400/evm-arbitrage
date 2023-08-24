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

type ViewCall struct {
	ID   string
	To   common.Address
	Data []byte
}

func (e *ETHClient) MultiViewCall(ctx context.Context, opt *bind.TransactOpts, calls []*ViewCall) (map[string]*abi.Multicall2Result, error) {
	if len(calls) == 0 {
		return map[string]*abi.Multicall2Result{}, nil
	}
	to := common.HexToAddress(e.multicallAddress)
	input, err := abi.Multicall2ABIInstance.Methods["tryAggregate"].Inputs.Pack(false, NewViewMulticall2Calls(calls))
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
	if len(resStruct) != len(calls) {
		return nil, fmt.Errorf("return results less than calls %d %d", len(resStruct), len(calls))
	}
	ret := make(map[string]*abi.Multicall2Result, len(resStruct))
	for i, one := range resStruct {
		ret[calls[i].ID] = &abi.Multicall2Result{
			Success:    one.Success,
			ReturnData: one.ReturnData,
		}
	}
	return ret, nil
}

func NewViewMulticall2Calls(calls []*ViewCall) []abi.Multicall2Call {
	viewcalls := []abi.Multicall2Call{}
	for _, call := range calls {
		viewcall := abi.Multicall2Call{
			Target:   call.To,
			CallData: call.Data,
		}
		viewcalls = append(viewcalls, viewcall)
	}
	return viewcalls
}
