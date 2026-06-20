package wallet

import "errors"

var (
	// ErrInvalidMnemonic indicates malformed BIP39 input.
	ErrInvalidMnemonic = errors.New("invalid mnemonic")
	// ErrInvalidPassword indicates that wallet decryption failed.
	ErrInvalidPassword = errors.New("invalid password")
	// ErrWalletExists indicates a wallet file already exists.
	ErrWalletExists = errors.New("wallet already exists")
	// ErrWalletNotFound indicates no wallet file exists.
	ErrWalletNotFound = errors.New("wallet not found")
)
