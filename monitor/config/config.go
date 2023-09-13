package config

import "github.com/ethereum/go-ethereum/common"

type Config struct {
	Node             string
	MulticallAddress string
	WETHAddress      common.Address
	StoreFilePath    string
	MaxConcurrency   uint64
}
