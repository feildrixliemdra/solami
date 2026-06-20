package chains

import (
	"context"
	"fmt"
	"math/big"

	"github.com/feildrix/solami/internal/networks"
	"github.com/feildrix/solami/internal/wallet"
	"github.com/gagliardetto/solana-go"
	"github.com/gagliardetto/solana-go/programs/system"
	"github.com/gagliardetto/solana-go/rpc"
)

const solDecimals = 9

// SolanaClient implements Solana native asset operations.
type SolanaClient struct {
	network networks.Network
}

// NewSolanaClient creates a Solana client.
func NewSolanaClient(network networks.Network) SolanaClient {
	return SolanaClient{network: network}
}

// Balance fetches the native SOL balance.
func (c SolanaClient) Balance(ctx context.Context, account wallet.Account) (Balance, error) {
	client := rpc.New(c.network.RPCURL)
	pubkey, err := solana.PublicKeyFromBase58(account.SolanaAddress)
	if err != nil {
		return Balance{}, fmt.Errorf("invalid solana address: %w", err)
	}
	result, err := client.GetBalance(ctx, pubkey, rpc.CommitmentFinalized)
	if err != nil {
		return Balance{}, fmt.Errorf("fetch solana balance: %w", err)
	}
	lamports := new(big.Int).SetUint64(result.Value)
	return Balance{Atomic: lamports, Formatted: formatDecimalAmount(lamports, solDecimals, 6), Symbol: c.network.Symbol}, nil
}

// EstimateFee estimates a native SOL transfer fee.
func (c SolanaClient) EstimateFee(ctx context.Context, account wallet.Account, to string, amount string) (FeeEstimate, error) {
	if _, err := parseDecimalAmount(amount, solDecimals); err != nil {
		return FeeEstimate{}, err
	}
	if _, err := solana.PublicKeyFromBase58(to); err != nil {
		return FeeEstimate{}, fmt.Errorf("invalid solana address: %w", err)
	}
	fee := big.NewInt(5000)
	return FeeEstimate{Atomic: fee, Formatted: formatDecimalAmount(fee, solDecimals, 9), Symbol: c.network.Symbol}, nil
}

// Send signs and broadcasts a native SOL transfer.
func (c SolanaClient) Send(ctx context.Context, plain wallet.PlainWallet, to string, amount string) (SendResult, error) {
	if !c.network.SendEnabled {
		return SendResult{}, ErrSendDisabled
	}
	privateKey, err := wallet.DeriveSolanaPrivateKey(plain.Mnemonic)
	if err != nil {
		return SendResult{}, err
	}
	from := privateKey.PublicKey()
	toAddress, err := solana.PublicKeyFromBase58(to)
	if err != nil {
		return SendResult{}, fmt.Errorf("invalid solana address: %w", err)
	}
	lamports, err := parseDecimalAmount(amount, solDecimals)
	if err != nil {
		return SendResult{}, err
	}
	if !lamports.IsUint64() {
		return SendResult{}, fmt.Errorf("%w: amount too large", ErrInvalidAmount)
	}
	client := rpc.New(c.network.RPCURL)
	blockhash, err := client.GetLatestBlockhash(ctx, rpc.CommitmentFinalized)
	if err != nil {
		return SendResult{}, fmt.Errorf("fetch solana blockhash: %w", err)
	}
	tx, err := solana.NewTransaction(
		[]solana.Instruction{
			system.NewTransferInstruction(lamports.Uint64(), from, toAddress).Build(),
		},
		blockhash.Value.Blockhash,
		solana.TransactionPayer(from),
	)
	if err != nil {
		return SendResult{}, fmt.Errorf("build solana transaction: %w", err)
	}
	_, err = tx.Sign(func(key solana.PublicKey) *solana.PrivateKey {
		if key.Equals(from) {
			return &privateKey
		}
		return nil
	})
	if err != nil {
		return SendResult{}, fmt.Errorf("sign solana transaction: %w", err)
	}
	signature, err := client.SendTransaction(ctx, tx)
	if err != nil {
		return SendResult{}, fmt.Errorf("broadcast solana transaction: %w", err)
	}
	return SendResult{Hash: signature.String(), ExplorerURL: c.network.ExplorerURL + signature.String()}, nil
}
