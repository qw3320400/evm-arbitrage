package protocol

import (
	"bytes"
	"context"
	"fmt"
	"math/big"
	"monitor/abi"
	"monitor/client"
	"monitor/storage"
	"strconv"
	"strings"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
)

var (
	UniswapV2PairEventSwapSign = abi.UniswapV2PairABIInstance.Events["Swap"].ID
	UniswapV2PairEventSyncSign = abi.UniswapV2PairABIInstance.Events["Sync"].ID

	_ storage.DataUpdate = &UniswapV2Pair{}
	_ DataConvert        = &UniswapV2Pair{}
)

/*
ai/ao = (ri+ai)/ro
ri/(ro-ao) = ai/ao
*/
type UniswapV2Pair struct {
	Address  common.Address
	Token0   common.Address
	Token1   common.Address
	Reserve0 *big.Int
	Reserve1 *big.Int
	Weight0  float64
	Weight1  float64
	Error    bool
	*StateFromLogUpdate
}

func (p *UniswapV2Pair) ToFileData() []byte {
	if p == nil {
		return []byte{}
	}
	return append([]byte(fmt.Sprintf("%s,%s,%s,%d,%d,%t@",
		p.Address,
		p.Token0,
		p.Token1,
		p.Reserve0,
		p.Reserve1,
		p.Error,
	)), p.StateFromLogUpdate.ToFileData()...)
}

func (p *UniswapV2Pair) FromFileData(body []byte) error {
	if p == nil {
		return fmt.Errorf("p is nil")
	}
	dataAndUpdate := bytes.Split(body, []byte("@"))
	if len(dataAndUpdate) != 2 {
		return fmt.Errorf("data format error %s", string(body))
	}
	dataBody := dataAndUpdate[0]
	words := bytes.Split(dataBody, []byte(","))
	if len(words) != 6 {
		return fmt.Errorf("data format error %s", string(body))
	}
	p.Address = common.HexToAddress(string(words[0]))
	if p.Address.Big().Cmp(big.NewInt(0)) == 0 {
		return fmt.Errorf("data format error %s", string(body))
	}
	p.Token0 = common.HexToAddress(string(words[1]))
	if p.Token0.Big().Cmp(big.NewInt(0)) == 0 {
		return fmt.Errorf("data format error %s", string(body))
	}
	p.Token1 = common.HexToAddress(string(words[2]))
	if p.Token1.Big().Cmp(big.NewInt(0)) == 0 {
		return fmt.Errorf("data format error %s", string(body))
	}
	p.Reserve0 = big.NewInt(0)
	_, ok := p.Reserve0.SetString(string(words[3]), 10)
	if !ok || p.Reserve0.Cmp(big.NewInt(0)) == 0 {
		return fmt.Errorf("data format error %s", string(body))
	}
	p.Reserve1 = big.NewInt(0)
	_, ok = p.Reserve1.SetString(string(words[4]), 10)
	if !ok || p.Reserve1.Cmp(big.NewInt(0)) == 0 {
		return fmt.Errorf("data format error %s", string(body))
	}
	var err error
	p.Error, err = strconv.ParseBool(string(words[5]))
	if err != nil {
		return fmt.Errorf("data format error %s", string(body))
	}
	p.StateFromLogUpdate = &StateFromLogUpdate{}
	return p.StateFromLogUpdate.FromFileData(dataAndUpdate[1])
}

func FilterUniswapV2PairFromLog(ctx context.Context, logs []*types.Log) (map[common.Address]*UniswapV2Pair, error) {
	datas := map[common.Address]*UniswapV2Pair{}
	for _, log := range logs {
		// utils.Infof("filter uniswapv2 log in blocknumber %d txindex %d logindex %d", log.BlockNumber, log.TxIndex, log.Index)
		if len(log.Topics) != 1 ||
			!strings.EqualFold(log.Topics[0].String(), UniswapV2PairEventSyncSign.String()) {
			continue
		}
		dataList, err := abi.UniswapV2PairABIInstance.Events["Sync"].Inputs.Unpack(log.Data)
		if err != nil {
			return nil, fmt.Errorf("unpack event data fail %s", err)
		}
		if len(dataList) != 2 {
			return nil, fmt.Errorf("unpack event data error %+v", dataList)
		}
		datas[log.Address] = &UniswapV2Pair{
			Address:  log.Address,
			Reserve0: dataList[0].(*big.Int),
			Reserve1: dataList[1].(*big.Int),
			StateFromLogUpdate: &StateFromLogUpdate{
				BlockNumber: log.BlockNumber,
				TxIndex:     log.TxIndex,
				LogIndex:    log.Index,
				Timestamp:   time.Now().Unix(),
			},
		}
	}
	return datas, nil
}

func NewUniswapV2PairInfoCalls(pair *UniswapV2Pair) []*client.ViewCall {
	return []*client.ViewCall{
		{
			ID:   "UniswapV2-" + pair.Address.String() + "-token0",
			To:   pair.Address,
			Data: abi.UniswapV2PairABIInstance.Methods["token0"].ID,
		},
		{
			ID:   "UniswapV2-" + pair.Address.String() + "-token1",
			To:   pair.Address,
			Data: abi.UniswapV2PairABIInstance.Methods["token1"].ID,
		},
	}
}

func NewUniswapV2PairStateCalls(pair *UniswapV2Pair) []*client.ViewCall {
	return []*client.ViewCall{
		{
			ID:   "UniswapV2-" + pair.Address.String() + "-reserve",
			To:   pair.Address,
			Data: abi.UniswapV2PairABIInstance.Methods["getReserves"].ID,
		},
	}
}

func UniswapV2PairCallResult(pairs map[common.Address]*UniswapV2Pair, results map[string]*abi.Multicall2Result) {
	for id, result := range results {
		keys := strings.Split(id, "-")
		if len(keys) != 3 || keys[0] != "UniswapV2" {
			continue
		}
		addr := common.HexToAddress(keys[1])
		if pairs[addr] == nil {
			pairs[addr] = &UniswapV2Pair{
				Address: addr,
			}
		}
		if !result.Success {
			pairs[addr].Error = true
			continue
		}
		switch keys[2] {
		case "token0":
			pairs[addr].Token0 = common.BytesToAddress(result.ReturnData)
		case "token1":
			pairs[addr].Token1 = common.BytesToAddress(result.ReturnData)
		case "reserve":
			if len(result.ReturnData) != 96 {
				pairs[addr].Error = true
				continue
			}
			pairs[addr].Reserve0 = &big.Int{}
			pairs[addr].Reserve0.SetBytes(result.ReturnData[:32])
			pairs[addr].Reserve1 = &big.Int{}
			pairs[addr].Reserve1.SetBytes(result.ReturnData[32:64])
		}
	}
}
