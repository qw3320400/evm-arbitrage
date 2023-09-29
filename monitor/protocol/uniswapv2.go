package protocol

import (
	"bytes"
	"context"
	"fmt"
	"math"
	"math/big"
	"monitor/abi"
	"monitor/client"
	"monitor/storage"
	"monitor/utils"
	"strconv"
	"strings"
	"time"

	ethabi "github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
)

const (
	FeeBase = float64(10000)
)

var (
	UniswapV2PairEventSwapSign        = abi.UniswapV2PairABIInstance.Events["Swap"].ID
	UniswapV2PairEventSyncSign        = abi.UniswapV2PairABIInstance.Events["Sync"].ID
	UniswapV2PairEventSyncUint256Sign common.Hash

	_ storage.DataUpdate = &UniswapV2Pair{}
	_ DataConvert        = &UniswapV2Pair{}

	// swapEvent = cache.New(time.Minute, time.Hour)
	// syncEvent = cache.New(time.Minute, time.Hour)

	syncEventUint256 = `
[
    {
		"anonymous": false,
		"inputs": [{
			"indexed": false,
			"internalType": "uint256",
			"name": "reserve0",
			"type": "uint256"
		}, {
			"indexed": false,
			"internalType": "uint256",
			"name": "reserve1",
			"type": "uint256"
		}],
		"name": "Sync",
		"type": "event"
	}
]
	`
	syncEventUint256ABI *ethabi.ABI
)

func init() {
	md := &bind.MetaData{
		ABI: syncEventUint256,
	}
	var err error
	syncEventUint256ABI, err = md.GetAbi()
	if err != nil {
		panic(err)
	}
	UniswapV2PairEventSyncUint256Sign = syncEventUint256ABI.Events["Sync"].ID
}

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
	Fee      int64
	*StateFromLogUpdate
}

type UniswapV2SwapEvent struct {
	Address    common.Address
	Sender     common.Address
	Amount0In  *big.Int
	Amount1In  *big.Int
	Amount0Out *big.Int
	Amount1Out *big.Int
	To         common.Address
}

type UniswapV2SyncEvent struct {
	Address  common.Address
	Reserve0 *big.Int
	Reserve1 *big.Int
}

func (p *UniswapV2Pair) ToFileData() []byte {
	if p == nil {
		return []byte{}
	}
	if p.Token0.Big().Cmp(p.Token1.Big()) >= 0 {
		utils.Errorf("----- token not sort ?? %+v", p)
	}
	return append([]byte(fmt.Sprintf("%s,%s,%s,%d,%d,%t,%d@",
		p.Address,
		p.Token0,
		p.Token1,
		p.Reserve0,
		p.Reserve1,
		p.Error,
		p.Fee,
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
	if len(words) != 7 {
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
	p.Fee, err = strconv.ParseInt(string(words[6]), 10, 64)
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
		if len(log.Topics) != 1 || !isSyncTopic(log.Topics[0]) {
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
			Fee:      30,
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

func FilterUniswapV2FeeFromLog(ctx context.Context, logs []*types.Log) (map[common.Address]int64, error) {
	var (
		fees         = map[common.Address]int64{}
		swapEventMap = map[string]*UniswapV2SwapEvent{}
		syncEventMap = map[string]*UniswapV2SyncEvent{}
	)
	for _, log := range logs {
		if len(log.Topics) != 3 ||
			!strings.EqualFold(log.Topics[0].String(), UniswapV2PairEventSwapSign.String()) {
			continue
		}
		dataList, err := abi.UniswapV2PairABIInstance.Events["Swap"].Inputs.Unpack(log.Data)
		if err != nil {
			return nil, fmt.Errorf("unpack event data fail %s", err)
		}
		key := fmt.Sprintf("%d:%d:%d", log.BlockNumber, log.TxIndex, log.Index)
		event := &UniswapV2SwapEvent{
			Address:    log.Address,
			Sender:     common.BytesToAddress(log.Topics[1].Bytes()),
			To:         common.BytesToAddress(log.Topics[2].Bytes()),
			Amount0In:  dataList[0].(*big.Int),
			Amount1In:  dataList[1].(*big.Int),
			Amount0Out: dataList[2].(*big.Int),
			Amount1Out: dataList[3].(*big.Int),
		}
		swapEventMap[key] = event

		if _, ok := fees[event.Address]; ok {
			continue
		}
		syncEvent, ok := syncEventMap[key]
		if !ok {
			continue
		}
		if syncEvent.Address != event.Address {
			continue
		}
		fees[event.Address] = CalculatePairFee(event.Amount0In, event.Amount1In, event.Amount0Out, event.Amount1Out, syncEvent.Reserve0, syncEvent.Reserve1)
	}
	for _, log := range logs {
		if len(log.Topics) != 1 || !isSyncTopic(log.Topics[0]) {
			continue
		}
		dataList, err := abi.UniswapV2PairABIInstance.Events["Sync"].Inputs.Unpack(log.Data)
		if err != nil {
			return nil, fmt.Errorf("unpack event data fail %s", err)
		}
		if len(dataList) != 2 {
			return nil, fmt.Errorf("unpack event data error %+v", dataList)
		}
		key := fmt.Sprintf("%d:%d:%d", log.BlockNumber, log.TxIndex, log.Index+1)
		event := &UniswapV2SyncEvent{
			Address:  log.Address,
			Reserve0: dataList[0].(*big.Int),
			Reserve1: dataList[1].(*big.Int),
		}
		syncEventMap[key] = event

		if _, ok := fees[event.Address]; ok {
			continue
		}
		swapEvent, ok := swapEventMap[key]
		if !ok {
			continue
		}
		if swapEvent.Address != event.Address {
			continue
		}
		fees[event.Address] = CalculatePairFee(swapEvent.Amount0In, swapEvent.Amount1In, swapEvent.Amount0Out, swapEvent.Amount1Out, event.Reserve0, event.Reserve1)
	}
	return fees, nil
}

func CalculatePairFee(amount0In, amount1In, amount0Out, amount1Out, reserve0, reserve1 *big.Int) int64 {
	a0i, _ := amount0In.Float64()
	a1i, _ := amount1In.Float64()
	a0o, _ := amount0Out.Float64()
	a1o, _ := amount1Out.Float64()
	r0, _ := reserve0.Float64()
	r1, _ := reserve1.Float64()

	pr0 := r0 + a0o - a0i
	pr1 := r1 + a1o - a1i

	ret := int64(0)
	if pr0 < pr1 {
		ret = int64(math.Ceil(math.Abs((r0*r1/pr1-pr0)/(r0-pr0)) * FeeBase))
	} else {
		ret = int64(math.Ceil(math.Abs((r0*r1/pr0-pr1)/(r1-pr1)) * FeeBase))
	}
	// stable ? 1.000010154287709
	if ret > 200 {
		dk := (math.Pow(r0, 3)*r1 + math.Pow(r1, 3)*r0) / (math.Pow(pr0, 3)*pr1 + math.Pow(pr1, 3)*pr0)
		if dk > 1 && dk < 1.0001 {
			ret = 5
		}
	}
	if ret > 0 && ret < int64(FeeBase) {
		return fixFee(ret)
	}
	return 30
}

func fixFee(fee int64) int64 {
	switch fee {
	case 14:
		return 15
	case 19:
		return 20
	case 24:
		return 25
	case 29:
		return 30
	default:
		return fee
	}
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

func GetAmountsOut(tokenIn common.Address, amountIn float64, pairPath []*UniswapV2Pair) float64 {
	var (
		pAmtOut  float64 = amountIn
		tokenOut         = tokenIn
	)
	for _, pair := range pairPath {
		r0, _ := pair.Reserve0.Float64()
		r1, _ := pair.Reserve1.Float64()
		if pair.Token0 == tokenOut {
			pAmtOut = GetAmountOut(pAmtOut, r0, r1, float64(pair.Fee))
			tokenOut = pair.Token1
		} else if pair.Token1 == tokenOut {
			pAmtOut = GetAmountOut(pAmtOut, r1, r0, float64(pair.Fee))
			tokenOut = pair.Token0
		} else {
			return 0
		}
	}
	if tokenOut != tokenIn {
		return 0
	}
	return pAmtOut
}

func GetAmountOut(amountIn, reserveIn, reserveOut, fee float64) float64 {
	if amountIn <= 0 || reserveIn <= 0 || reserveOut <= 0 {
		return 0
	}
	amountInWithFee := amountIn * (FeeBase - fee)
	numerator := amountInWithFee * reserveOut
	denominator := reserveIn*FeeBase + amountInWithFee
	return numerator / denominator
}

func isSyncTopic(topic common.Hash) bool {
	if strings.EqualFold(topic.String(), UniswapV2PairEventSyncSign.String()) ||
		strings.EqualFold(topic.String(), UniswapV2PairEventSyncUint256Sign.String()) {
		return true
	}
	return false
}
