package trader

import (
	"context"
	"crypto/ecdsa"
	"fmt"
	"math/big"
	"monitor/client"
	"monitor/config"
	"monitor/protocol"
	"monitor/utils"
	"strings"
	"time"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
)

var (
	_ utils.Keeper = &Trader{}

	swapMetaData = `[{"type":"function","name":"swap2","stateMutability":"nonpayable","inputs":[{"internalType":"uint256","type":"uint256","name":"amountIn"},{"internalType":"struct Swaper.Route[]","type":"tuple[]","name":"routes","components":[{"internalType":"address","type":"address","name":"pair"},{"internalType":"bool","type":"bool","name":"direction"},{"internalType":"uint256","type":"uint256","name":"fee"}]}],"outputs":[]}]`
	swapABI      *abi.ABI
)

func init() {
	md := &bind.MetaData{
		ABI: swapMetaData,
	}
	var err error
	swapABI, err = md.GetAbi()
	if err != nil {
		panic(err)
	}
}

type Trader struct {
	config *config.Config

	gasPrice    *big.Int
	ethGasPrice *big.Int
	signer      types.Signer
	privateKey  *ecdsa.PrivateKey
}

func NewTrader(ctx context.Context, conf *config.Config) *Trader {
	return &Trader{
		config:      conf,
		gasPrice:    big.NewInt(120000000),
		ethGasPrice: big.NewInt(20000000000),
	}
}

func (t *Trader) Init(ctx context.Context) error {
	var err error
	cli, err := client.GetETHClient(ctx, t.config.Node, t.config.MulticallAddress)
	if err != nil {
		return fmt.Errorf("get eth client fail %s", err)
	}
	chainID, err := cli.ChainID(ctx)
	if err != nil {
		return fmt.Errorf("get chain id fail %s", err)
	}
	t.signer = types.LatestSignerForChainID(chainID)
	t.privateKey, err = crypto.HexToECDSA(strings.TrimPrefix(t.config.PrivateKey, "0x"))
	if err != nil {
		return fmt.Errorf("hex to ecdsa fail %s", err)
	}

	go t.loopWatcher(ctx)
	return nil
}

func (*Trader) ShutDown(context.Context) {

}

func (t *Trader) loopWatcher(ctx context.Context) {
	for {
		err := t.fetchGasPrice(ctx)
		if err != nil {
			utils.Warnf("fetch gas price fail %s", err)
		}

		err = t.fetchETHGasPrice(ctx)
		if err != nil {
			utils.Warnf("fetch eth gas price fail %s", err)
		}

		<-time.After(time.Second * 5)
	}
}

func (t *Trader) fetchGasPrice(ctx context.Context) error {
	cli, err := client.GetETHClient(ctx, t.config.Node, t.config.MulticallAddress)
	if err != nil {
		return fmt.Errorf("get eth client fail %s", err)
	}
	gasPrice, err := cli.SuggestGasPrice(ctx)
	if err != nil {
		return fmt.Errorf("get suggest gas price fail %s", err)
	}
	if gasPrice.Cmp(big.NewInt(0)) > 0 {
		t.gasPrice = gasPrice
	}
	utils.Infof("current suggest gas price is %f gwei", t.GasPrice()/1000000000)
	return nil
}

func (t *Trader) fetchETHGasPrice(ctx context.Context) error {
	ecli, err := client.GetETHClient(ctx, t.config.ETHNode, common.Address{})
	if err != nil {
		return fmt.Errorf("get eth client fail %s", err)
	}
	eGasPrice, err := ecli.SuggestGasPrice(ctx)
	if err != nil {
		return fmt.Errorf("get eth suggest gas price fail %s", err)
	}
	if eGasPrice.Cmp(big.NewInt(0)) > 0 {
		t.ethGasPrice = eGasPrice
	}
	utils.Infof("current eth suggest gas price is %f gwei", t.ETHGasPrice()/1000000000)
	return nil
}

func (t *Trader) GasPrice() float64 {
	gp, _ := t.gasPrice.Float64()
	return gp
}

func (t *Trader) ETHGasPrice() float64 {
	gp, _ := t.ethGasPrice.Float64()
	return gp
}

type Route struct {
	Pair      common.Address
	Direction bool
	Fee       *big.Int
}

func (t *Trader) SwapV2(ctx context.Context, inputAmount, outputAmount float64, pairPath []*protocol.UniswapV2Pair) error {
	gasPrice := int64(t.GasPrice())
	if gasPrice <= 0 {
		return fmt.Errorf("gas price error %d", gasPrice)
	}
	call := ethereum.CallMsg{
		From:     t.config.FromAddress,
		To:       &t.config.SwapAddress,
		Gas:      uint64(70000 + len(pairPath)*100000),
		GasPrice: big.NewInt(gasPrice),
	}
	var (
		routes = make([]Route, 0, len(pairPath))
		inAddr = t.config.WETHAddress
	)
	for _, pair := range pairPath {
		route := Route{
			Pair: pair.Address,
			Fee:  big.NewInt(pair.Fee),
		}
		if pair.Token0 == inAddr {
			route.Direction = true
			inAddr = pair.Token1
		} else {
			route.Direction = false
			inAddr = pair.Token0
		}
		routes = append(routes, route)
	}
	param, err := swapABI.Methods["swap2"].Inputs.Pack(big.NewInt(int64(inputAmount)), routes)
	if err != nil {
		return fmt.Errorf("pack input param fail %s", err)
	}
	call.Data = append(swapABI.Methods["swap2"].ID, param...)

	cli, err := client.GetETHClient(ctx, t.config.Node, t.config.MulticallAddress)
	if err != nil {
		return fmt.Errorf("get eth client fail %s %s", err, common.Bytes2Hex(call.Data))
	}
	gasUsed, err := cli.EstimateGas(ctx, call)
	if err != nil {
		return fmt.Errorf("estimate gas fail %s %s", err, common.Bytes2Hex(call.Data))
	}
	utils.Warnf("---- estimate gas result %d %s", gasUsed, common.Bytes2Hex(call.Data))

	nonce, err := cli.NonceAt(ctx, call.From, nil)
	if err != nil {
		return fmt.Errorf("get nonce fail %s", err)
	}
	tx := types.NewTransaction(nonce, *call.To, big.NewInt(0), uint64(float64(gasUsed)*1.1), call.GasPrice, call.Data)
	tx, err = types.SignTx(tx, t.signer, t.privateKey)
	if err != nil {
		return fmt.Errorf("sign tx fail %s", err)
	}
	return cli.SendTransaction(ctx, tx)
}
