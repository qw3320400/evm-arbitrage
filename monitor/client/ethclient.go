package client

import (
	"context"
	"fmt"
	"monitor/abi"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/ethclient"
)

type ETHClient struct {
	*ethclient.Client

	multicallAddress common.Address
}

var ETHClientMap = map[string]*ETHClient{}

func GetETHClient(ctx context.Context, node string, multicallAddress common.Address) (*ETHClient, error) {
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
	input, err := abi.Multicall2ABIInstance.Methods["tryAggregate"].Inputs.Pack(false, NewViewMulticall2Calls(calls))
	if err != nil {
		return nil, fmt.Errorf("pack input fail %s", err)
	}
	resBody, err := e.Client.CallContract(ctx, ethereum.CallMsg{
		To:   &e.multicallAddress,
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

func (e *ETHClient) EstimateGasLast(ctx context.Context, msg ethereum.CallMsg) (uint64, error) {
	var hex hexutil.Uint64
	err := e.Client.Client().CallContext(ctx, &hex, "eth_estimateGas", toCallArg(msg), "latest")
	if err != nil {
		return 0, err
	}
	return uint64(hex), nil
}

func toCallArg(msg ethereum.CallMsg) interface{} {
	arg := map[string]interface{}{
		"from": msg.From,
		"to":   msg.To,
	}
	if len(msg.Data) > 0 {
		arg["data"] = hexutil.Bytes(msg.Data)
	}
	if msg.Value != nil {
		arg["value"] = (*hexutil.Big)(msg.Value)
	}
	if msg.Gas != 0 {
		arg["gas"] = hexutil.Uint64(msg.Gas)
	}
	if msg.GasPrice != nil {
		arg["gasPrice"] = (*hexutil.Big)(msg.GasPrice)
	}
	return arg
}
