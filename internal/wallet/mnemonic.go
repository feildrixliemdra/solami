package wallet

import (
	"fmt"

	bip39 "github.com/tyler-smith/go-bip39"
)

// GenerateMnemonic creates a new 12-word BIP39 mnemonic.
func GenerateMnemonic() (string, error) {
	entropy, err := bip39.NewEntropy(128)
	if err != nil {
		return "", fmt.Errorf("generate entropy: %w", err)
	}
	mnemonic, err := bip39.NewMnemonic(entropy)
	if err != nil {
		return "", fmt.Errorf("generate mnemonic: %w", err)
	}
	return mnemonic, nil
}

// ValidateMnemonic checks whether a mnemonic is a valid BIP39 phrase.
func ValidateMnemonic(mnemonic string) error {
	if !bip39.IsMnemonicValid(mnemonic) {
		return ErrInvalidMnemonic
	}
	return nil
}
