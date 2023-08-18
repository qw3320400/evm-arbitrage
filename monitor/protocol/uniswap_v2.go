package protocol

import (
	"context"
	"fmt"
	"math/big"
	"monitor/abi"
	"strings"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
)

/*
ai/ao = (ri+ai)/ro
ri/(ro-ao) = ai/ao
*/
type UniswapV2PairInfo struct {
	Address           common.Address
	Token0            common.Address
	Token1            common.Address
	Reserve0          *big.Int
	Reserve1          *big.Int
	Amount0Int        *big.Int
	Amount1Int        *big.Int
	Amount0Out        *big.Int
	Amount1Out        *big.Int
	Slash             float64 // TODO
	UpdateBlockNumber uint64
}

var (
	UniswapV2PairEventSwapSign = abi.UniswapV2PairABIInstance.Events["Swap"].ID.String()
)

func FilterUniswapV2PairFromReceipt(ctx context.Context, receipts []*types.Receipt) ([]*UniswapV2PairInfo, error) {
	datas := []*UniswapV2PairInfo{}
	for _, receipt := range receipts {
		for _, log := range receipt.Logs {
			if len(log.Topics) != 3 ||
				strings.EqualFold(log.Topics[0].String(), UniswapV2PairEventSwapSign) {
				continue
			}
			dataList, err := abi.UniswapV2PairABIInstance.Events["Swap"].Inputs.Unpack(log.Data)
			if err != nil {
				return nil, fmt.Errorf("unpack event data fail %s", err)
			}
			if len(dataList) != 4 {
				return nil, fmt.Errorf("unpack event data error %+v", dataList)
			}
			data := &UniswapV2PairInfo{
				Address:    log.Address,
				Amount0Int: dataList[0].(*big.Int),
				Amount1Int: dataList[1].(*big.Int),
				Amount0Out: dataList[2].(*big.Int),
				Amount1Out: dataList[3].(*big.Int),
			}
			datas = append(datas, data)
		}
	}
	return datas, nil
}
