package chains

import (
	"context"
	"errors"
	"fmt"
	"math/big"
	"strings"

	"github.com/feildrix/solami/internal/networks"
	"github.com/feildrix/solami/internal/wallet"
)

var (
	// ErrSendDisabled indicates that v1 does not permit sending on this network.
	ErrSendDisabled = errors.New("sending is disabled for this network")
	// ErrInvalidAmount indicates a malformed or non-positive amount.
	ErrInvalidAmount = errors.New("invalid amount")
)

// Balance contains a formatted native-asset balance.
type Balance struct {
	Atomic    *big.Int
	Formatted string
	Symbol    string
}

// FeeEstimate contains a formatted native-asset fee estimate.
type FeeEstimate struct {
	Atomic    *big.Int
	Formatted string
	Symbol    string
}

// SendResult contains the broadcast transaction identifier.
type SendResult struct {
	Hash        string
	ExplorerURL string
}

// Client provides chain operations for one network.
type Client interface {
	Balance(ctx context.Context, account wallet.Account) (Balance, error)
	EstimateFee(ctx context.Context, account wallet.Account, to string, amount string) (FeeEstimate, error)
	Send(ctx context.Context, plain wallet.PlainWallet, to string, amount string) (SendResult, error)
}

// NewClient creates a chain client for a network.
func NewClient(network networks.Network) (Client, error) {
	if strings.TrimSpace(network.RPCURL) == "" {
		return nil, fmt.Errorf("network %s has empty RPC URL", network.Name)
	}
	switch network.Chain {
	case networks.ChainEthereum:
		return NewEthereumClient(network), nil
	case networks.ChainSolana:
		return NewSolanaClient(network), nil
	default:
		return nil, fmt.Errorf("unsupported chain %q", network.Chain)
	}
}
