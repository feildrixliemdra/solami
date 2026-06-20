package wallet

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

// PlainWallet is the decrypted wallet payload.
type PlainWallet struct {
	Mnemonic string    `json:"mnemonic"`
	Account  Account   `json:"account"`
	Created  time.Time `json:"created"`
}

// Store persists and unlocks the encrypted default wallet.
type Store struct {
	path string
}

// NewStore creates a wallet store.
func NewStore(path string) Store {
	return Store{path: path}
}

// Exists reports whether the encrypted wallet exists.
func (s Store) Exists() (bool, error) {
	_, err := os.Stat(s.path)
	if err == nil {
		return true, nil
	}
	if errors.Is(err, os.ErrNotExist) {
		return false, nil
	}
	return false, fmt.Errorf("stat wallet: %w", err)
}

// CreateNew generates and persists a new encrypted wallet.
func (s Store) CreateNew(password string) (PlainWallet, error) {
	mnemonic, err := GenerateMnemonic()
	if err != nil {
		return PlainWallet{}, err
	}
	return s.Import(mnemonic, password, false)
}

// Import validates, derives, encrypts, and persists an imported mnemonic.
func (s Store) Import(mnemonic string, password string, overwrite bool) (PlainWallet, error) {
	if err := ValidateMnemonic(mnemonic); err != nil {
		return PlainWallet{}, err
	}
	exists, err := s.Exists()
	if err != nil {
		return PlainWallet{}, err
	}
	if exists && !overwrite {
		return PlainWallet{}, ErrWalletExists
	}
	account, err := DeriveAccount(mnemonic)
	if err != nil {
		return PlainWallet{}, err
	}
	wallet := PlainWallet{
		Mnemonic: mnemonic,
		Account:  account,
		Created:  time.Now().UTC(),
	}
	if err := s.Save(wallet, password); err != nil {
		return PlainWallet{}, err
	}
	return wallet, nil
}

// Save encrypts and writes a wallet.
func (s Store) Save(wallet PlainWallet, password string) error {
	data, err := json.Marshal(wallet)
	if err != nil {
		return fmt.Errorf("encode wallet: %w", err)
	}
	payload, err := Encrypt(password, data)
	zero(data)
	if err != nil {
		return err
	}
	encoded, err := json.MarshalIndent(payload, "", "  ")
	if err != nil {
		return fmt.Errorf("encode encrypted wallet: %w", err)
	}
	if err := os.MkdirAll(filepath.Dir(s.path), 0o700); err != nil {
		return fmt.Errorf("create wallet dir: %w", err)
	}
	if err := os.WriteFile(s.path, encoded, 0o600); err != nil {
		return fmt.Errorf("write wallet: %w", err)
	}
	return nil
}

// Unlock decrypts the wallet with a password.
func (s Store) Unlock(password string) (PlainWallet, error) {
	data, err := os.ReadFile(s.path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return PlainWallet{}, ErrWalletNotFound
		}
		return PlainWallet{}, fmt.Errorf("read wallet: %w", err)
	}
	var payload EncryptedPayload
	if err := json.Unmarshal(data, &payload); err != nil {
		return PlainWallet{}, fmt.Errorf("decode encrypted wallet: %w", err)
	}
	plaintext, err := Decrypt(password, payload)
	if err != nil {
		return PlainWallet{}, err
	}
	var wallet PlainWallet
	if err := json.Unmarshal(plaintext, &wallet); err != nil {
		zero(plaintext)
		return PlainWallet{}, fmt.Errorf("decode wallet: %w", err)
	}
	zero(plaintext)
	return wallet, nil
}

// Path returns the encrypted wallet file path.
func (s Store) Path() string {
	return s.path
}
