package trader

import (
	"context"
	"crypto/ecdsa"
	"fmt"
	"math"
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

	swapMetaData = `
[
    {
        "type":"function",
        "name":"swap",
        "stateMutability":"nonpayable",
        "inputs":[
            {
                "internalType":"bytes",
                "type":"bytes",
                "name":"params"
            }
        ],
        "outputs":[

        ]
    }
]
	`
	swapABI *abi.ABI
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
	var logFeeTime = time.Now()
	for {
		err := t.fetchGasPrice(ctx)
		if err != nil {
			utils.Warnf("fetch gas price fail %s", err)
		}

		// err = t.fetchETHGasPrice(ctx)
		// if err != nil {
		// 	utils.Warnf("fetch eth gas price fail %s", err)
		// }

		if now := time.Now(); now.Sub(logFeeTime) > time.Second*5 {
			utils.Infof("current suggest gas price is %f gwei, eth gas price is %f gwei", t.GasPrice()/math.Pow10(9), t.ETHGasPrice()/math.Pow10(9))
			logFeeTime = now
		}
		<-time.After(time.Second * 2)
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

func (t *Trader) SwapV2(ctx context.Context, inputAmount float64, pairPath []*protocol.UniswapV2Pair) error {
	gasPrice := int64(t.GetPreferGasPrice())
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
		inAddr          = t.config.WETHAddress
		paramStr string = fmt.Sprintf("%020x", big.NewInt(int64(inputAmount)))
	)
	for _, pair := range pairPath {
		var boolToInt int
		if pair.Token0 == inAddr {
			boolToInt = 1
			inAddr = pair.Token1
		} else {
			inAddr = pair.Token0
		}
		paramStr += fmt.Sprintf("%040x%02x%04x", pair.Address, boolToInt, big.NewInt(pair.Fee))
	}
	param, err := swapABI.Methods["swap"].Inputs.Pack(common.FromHex(paramStr))
	if err != nil {
		return fmt.Errorf("abi pack fail %s", err)
	}
	call.Data = append(swapABI.Methods["swap"].ID, param...)

	cli, err := client.GetETHClient(ctx, t.config.Node, t.config.MulticallAddress)
	if err != nil {
		return fmt.Errorf("get eth client fail %s %s", err, common.Bytes2Hex(call.Data))
	}
	gasUsed, err := cli.EstimateGas(ctx, call)
	if err != nil {
		return fmt.Errorf("estimate gas fail %s %s", err, common.Bytes2Hex(call.Data))
	}

	nonce, err := cli.NonceAt(ctx, call.From, nil)
	if err != nil {
		return fmt.Errorf("get nonce fail %s", err)
	}
	tx := types.NewTransaction(nonce, *call.To, big.NewInt(0), uint64(float64(gasUsed)*1.1), call.GasPrice, call.Data)
	tx, err = types.SignTx(tx, t.signer, t.privateKey)
	if err != nil {
		return fmt.Errorf("sign tx fail %s", err)
	}
	// final check
	amountOut := protocol.GetAmountsOut(t.config.WETHAddress, inputAmount, pairPath)
	estimateFee := t.EstimateFeeByGas(len(pairPath), gasUsed)
	// if amountOut-inputAmount < estimateFee*2 {
	// 	return fmt.Errorf("final check fail amountIn %f amountOut %f estimateFee %f gasUsed %d gasPrice %f gwei eth gasPrice %f gwei", inputAmount/math.Pow10(18), amountOut/math.Pow10(18), estimateFee/math.Pow10(18), gasUsed, float64(gasPrice)/math.Pow10(9), t.ETHGasPrice()/math.Pow10(9))
	// }
	if amountOut-inputAmount < estimateFee*1.2 {
		return fmt.Errorf("final check fail amountIn %f amountOut %f estimateFee %f gasUsed %d gasPrice %f gwei eth gasPrice %f gwei", inputAmount/math.Pow10(18), amountOut/math.Pow10(18), estimateFee/math.Pow10(18), gasUsed, float64(gasPrice)/math.Pow10(9), t.ETHGasPrice()/math.Pow10(9))
	}
	utils.Warnf("final check pass amountIn %f amountOut %f estimateFee %f gasUsed %d gasPrice %f gwei eth gasPrice %f gwei", inputAmount/math.Pow10(18), amountOut/math.Pow10(18), estimateFee/math.Pow10(18), gasUsed, float64(gasPrice)/math.Pow10(9), t.ETHGasPrice()/math.Pow10(9))
	// return nil
	return cli.SendTransaction(ctx, tx)
}

func (t *Trader) EstimateFee(length int) float64 {
	// linea chain
	gasPrice := t.GetPreferGasPrice()
	gas := swapGas(length)
	fee := gas * gasPrice
	return fee
}

func (t *Trader) EstimateFeeByGas(length int, gas uint64) float64 {
	// linea chain
	gasPrice := t.GetPreferGasPrice()
	fee := float64(gas) * gasPrice
	return fee
}

func (t *Trader) GetPreferGasPrice() float64 {
	return t.GasPrice() / 10
}

func swapGas(length int) float64 {
	switch length {
	case 2:
		return float64(210000)
	case 3:
		return float64(250000)
	case 4:
		return float64(320000)
	case 5:
		return float64(360000)
	default:
		return float64(80000 + length*60000)
	}
}

func swapBaseEthGas(length int) float64 {
	switch length {
	case 2:
		return float64(2100)
	case 3:
		return float64(2400)
	case 4:
		return float64(2600)
	case 5:
		return float64(2800)
	default:
		return float64(1800 + length*200)
	}
}
