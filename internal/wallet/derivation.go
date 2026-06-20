package wallet

import (
	"crypto/ecdsa"
	"crypto/ed25519"
	"crypto/hmac"
	"crypto/sha512"
	"encoding/binary"
	"fmt"
	"strings"

	"github.com/ethereum/go-ethereum/accounts"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/gagliardetto/solana-go"
	bip32 "github.com/tyler-smith/go-bip32"
	bip39 "github.com/tyler-smith/go-bip39"
)

const (
	// EthereumDerivationPath is the v1 Ethereum account path.
	EthereumDerivationPath = "m/44'/60'/0'/0/0"
	// SolanaDerivationPath is the v1 Solana account path used by common wallets.
	SolanaDerivationPath = "m/44'/501'/0'/0'"
)

// Account contains the derived public addresses for the default account.
type Account struct {
	EthereumAddress string `json:"ethereum_address"`
	SolanaAddress   string `json:"solana_address"`
}

// DeriveAccount derives the default Ethereum and Solana addresses.
func DeriveAccount(mnemonic string) (Account, error) {
	eth, err := DeriveEthereumAddress(mnemonic)
	if err != nil {
		return Account{}, err
	}
	sol, err := DeriveSolanaAddress(mnemonic)
	if err != nil {
		return Account{}, err
	}
	return Account{EthereumAddress: eth, SolanaAddress: sol}, nil
}

// DeriveEthereumAddress derives the v1 Ethereum address.
func DeriveEthereumAddress(mnemonic string) (string, error) {
	if err := ValidateMnemonic(mnemonic); err != nil {
		return "", err
	}
	privateKey, err := DeriveEthereumPrivateKey(mnemonic)
	if err != nil {
		return "", err
	}
	return crypto.PubkeyToAddress(privateKey.PublicKey).Hex(), nil
}

// DeriveEthereumPrivateKey derives the v1 Ethereum signing key.
func DeriveEthereumPrivateKey(mnemonic string) (*ecdsa.PrivateKey, error) {
	if err := ValidateMnemonic(mnemonic); err != nil {
		return nil, err
	}
	seed := bip39.NewSeed(mnemonic, "")
	key, err := bip32.NewMasterKey(seed)
	if err != nil {
		return nil, fmt.Errorf("create ethereum master key: %w", err)
	}
	path, err := accounts.ParseDerivationPath(EthereumDerivationPath)
	if err != nil {
		return nil, fmt.Errorf("parse ethereum derivation path: %w", err)
	}
	for _, index := range path {
		key, err = key.NewChildKey(index)
		if err != nil {
			return nil, fmt.Errorf("derive ethereum child key: %w", err)
		}
	}
	return crypto.ToECDSA(key.Key)
}

// DeriveSolanaAddress derives the v1 Solana address using SLIP-0010 ed25519 hardened derivation.
func DeriveSolanaAddress(mnemonic string) (string, error) {
	privateKey, err := DeriveSolanaPrivateKey(mnemonic)
	if err != nil {
		return "", err
	}
	return privateKey.PublicKey().String(), nil
}

// DeriveSolanaPrivateKey derives the v1 Solana signing key.
func DeriveSolanaPrivateKey(mnemonic string) (solana.PrivateKey, error) {
	if err := ValidateMnemonic(mnemonic); err != nil {
		return nil, err
	}
	seed := bip39.NewSeed(mnemonic, "")
	key, err := deriveEd25519Seed(seed, SolanaDerivationPath)
	if err != nil {
		return nil, err
	}
	edKey := ed25519.NewKeyFromSeed(key)
	return solana.PrivateKey(edKey), nil
}

func deriveEd25519Seed(seed []byte, path string) ([]byte, error) {
	key, chainCode := slip10Master(seed)
	segments, err := parseHardenedPath(path)
	if err != nil {
		return nil, err
	}
	for _, segment := range segments {
		key, chainCode = slip10Child(key, chainCode, segment)
	}
	return key, nil
}

func slip10Master(seed []byte) ([]byte, []byte) {
	mac := hmac.New(sha512.New, []byte("ed25519 seed"))
	mac.Write(seed)
	sum := mac.Sum(nil)
	return sum[:32], sum[32:]
}

func slip10Child(key []byte, chainCode []byte, index uint32) ([]byte, []byte) {
	data := make([]byte, 1+32+4)
	copy(data[1:], key)
	binary.BigEndian.PutUint32(data[33:], index+0x80000000)
	mac := hmac.New(sha512.New, chainCode)
	mac.Write(data)
	sum := mac.Sum(nil)
	return sum[:32], sum[32:]
}

func parseHardenedPath(path string) ([]uint32, error) {
	if path == "m" {
		return nil, nil
	}
	parts := strings.Split(path, "/")
	if len(parts) == 0 || parts[0] != "m" {
		return nil, fmt.Errorf("path must start with m")
	}
	out := make([]uint32, 0, len(parts)-1)
	for _, part := range parts[1:] {
		if !strings.HasSuffix(part, "'") {
			return nil, fmt.Errorf("ed25519 path segment %q must be hardened", part)
		}
		var value uint32
		if _, err := fmt.Sscanf(strings.TrimSuffix(part, "'"), "%d", &value); err != nil {
			return nil, fmt.Errorf("parse path segment %q: %w", part, err)
		}
		out = append(out, value)
	}
	return out, nil
}
