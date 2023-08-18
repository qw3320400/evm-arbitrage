package abi

import "github.com/ethereum/go-ethereum/accounts/abi"

var (
	Multicall2ABIInstance    *abi.ABI
	UniswapV2PairABIInstance *abi.ABI
)

func init() {
	var err error
	Multicall2ABIInstance, err = Multicall2MetaData.GetAbi()
	if err != nil {
		panic(err)
	}
	UniswapV2PairABIInstance, err = UniswapV2PairMetaData.GetAbi()
	if err != nil {
		panic(err)
	}
}
