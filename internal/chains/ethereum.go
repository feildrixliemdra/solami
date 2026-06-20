package chains

import (
	"context"
	"fmt"
	"math/big"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/feildrix/solami/internal/networks"
	"github.com/feildrix/solami/internal/wallet"
)

const ethDecimals = 18

// EthereumClient implements Ethereum native asset operations.
type EthereumClient struct {
	network networks.Network
}

// NewEthereumClient creates an Ethereum client.
func NewEthereumClient(network networks.Network) EthereumClient {
	return EthereumClient{network: network}
}

// Balance fetches the native ETH balance.
func (c EthereumClient) Balance(ctx context.Context, account wallet.Account) (Balance, error) {
	client, err := ethclient.DialContext(ctx, c.network.RPCURL)
	if err != nil {
		return Balance{}, fmt.Errorf("connect ethereum rpc: %w", err)
	}
	defer client.Close()
	address := common.HexToAddress(account.EthereumAddress)
	balance, err := client.BalanceAt(ctx, address, nil)
	if err != nil {
		return Balance{}, fmt.Errorf("fetch ethereum balance: %w", err)
	}
	return Balance{Atomic: balance, Formatted: formatDecimalAmount(balance, ethDecimals, 6), Symbol: c.network.Symbol}, nil
}

// EstimateFee estimates a native ETH transfer fee.
func (c EthereumClient) EstimateFee(ctx context.Context, account wallet.Account, to string, amount string) (FeeEstimate, error) {
	client, err := ethclient.DialContext(ctx, c.network.RPCURL)
	if err != nil {
		return FeeEstimate{}, fmt.Errorf("connect ethereum rpc: %w", err)
	}
	defer client.Close()
	from := common.HexToAddress(account.EthereumAddress)
	toAddress := common.HexToAddress(to)
	value, err := parseDecimalAmount(amount, ethDecimals)
	if err != nil {
		return FeeEstimate{}, err
	}
	gasPrice, err := client.SuggestGasPrice(ctx)
	if err != nil {
		return FeeEstimate{}, fmt.Errorf("suggest ethereum gas price: %w", err)
	}
	gasLimit, err := client.EstimateGas(ctx, ethereum.CallMsg{From: from, To: &toAddress, Value: value})
	if err != nil {
		gasLimit = 21000
	}
	fee := new(big.Int).Mul(gasPrice, new(big.Int).SetUint64(gasLimit))
	return FeeEstimate{Atomic: fee, Formatted: formatDecimalAmount(fee, ethDecimals, 8), Symbol: c.network.Symbol}, nil
}

// Send signs and broadcasts a native ETH transfer.
func (c EthereumClient) Send(ctx context.Context, plain wallet.PlainWallet, to string, amount string) (SendResult, error) {
	if !c.network.SendEnabled {
		return SendResult{}, ErrSendDisabled
	}

	if !common.IsHexAddress(to) {
		return SendResult{}, fmt.Errorf("invalid ethereum address")
	}

	client, err := ethclient.DialContext(ctx, c.network.RPCURL)
	if err != nil {
		return SendResult{}, fmt.Errorf("connect ethereum rpc: %w", err)
	}

	defer client.Close()
	privateKey, err := wallet.DeriveEthereumPrivateKey(plain.Mnemonic)
	if err != nil {
		return SendResult{}, err
	}

	from := common.HexToAddress(plain.Account.EthereumAddress)
	toAddress := common.HexToAddress(to)
	value, err := parseDecimalAmount(amount, ethDecimals)
	if err != nil {
		return SendResult{}, err
	}

	nonce, err := client.PendingNonceAt(ctx, from)
	if err != nil {
		return SendResult{}, fmt.Errorf("fetch ethereum nonce: %w", err)
	}

	gasPrice, err := client.SuggestGasPrice(ctx)
	if err != nil {
		return SendResult{}, fmt.Errorf("suggest ethereum gas price: %w", err)
	}

	gasLimit, err := client.EstimateGas(ctx, ethereum.CallMsg{From: from, To: &toAddress, Value: value})
	if err != nil {
		gasLimit = 21000
	}
	chainID, err := client.ChainID(ctx)
	if err != nil {
		return SendResult{}, fmt.Errorf("fetch ethereum chain id: %w", err)
	}
	tx := types.NewTransaction(nonce, toAddress, value, gasLimit, gasPrice, nil)
	signed, err := types.SignTx(tx, types.LatestSignerForChainID(chainID), privateKey)
	if err != nil {
		return SendResult{}, fmt.Errorf("sign ethereum transaction: %w", err)
	}
	if err := client.SendTransaction(ctx, signed); err != nil {
		return SendResult{}, fmt.Errorf("broadcast ethereum transaction: %w", err)
	}
	hash := signed.Hash().Hex()
	return SendResult{Hash: hash, ExplorerURL: c.network.ExplorerURL + hash}, nil
}
