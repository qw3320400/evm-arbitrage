package config

import "github.com/ethereum/go-ethereum/common"

type Config struct {
	Node             string
	MulticallAddress common.Address
	WETHAddress      common.Address
	StoreFilePath    string
	FromAddress      common.Address
	PrivateKey       string
	SwapAddress      common.Address
	MinRecieve       float64
	ETHNode          string
}
