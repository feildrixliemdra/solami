package wallet

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

const testMnemonic = "test test test test test test test test test test test junk"

func TestMnemonicGenerationAndValidation(t *testing.T) {
	mnemonic, err := GenerateMnemonic()
	if err != nil {
		t.Fatalf("GenerateMnemonic failed: %v", err)
	}
	if len(strings.Fields(mnemonic)) != 12 {
		t.Fatalf("expected 12 words, got %d", len(strings.Fields(mnemonic)))
	}
	if err := ValidateMnemonic(mnemonic); err != nil {
		t.Fatalf("generated mnemonic should validate: %v", err)
	}
	if err := ValidateMnemonic("not a valid mnemonic"); !errors.Is(err, ErrInvalidMnemonic) {
		t.Fatalf("expected ErrInvalidMnemonic, got %v", err)
	}
}

func TestDeriveAccountExpectedAddresses(t *testing.T) {
	account, err := DeriveAccount(testMnemonic)
	if err != nil {
		t.Fatalf("DeriveAccount failed: %v", err)
	}
	if account.EthereumAddress != "0xf39Fd6e51aad88F6F4ce6aB8827279cffFb92266" {
		t.Fatalf("unexpected ethereum address: %s", account.EthereumAddress)
	}
	if account.SolanaAddress != "oeYf6KAJkLYhBuR8CiGc6L4D4Xtfepr85fuDgA9kq96" {
		t.Fatalf("unexpected solana address: %s", account.SolanaAddress)
	}
}

func TestEncryptionRoundTripAndWrongPassword(t *testing.T) {
	plaintext := []byte(`{"mnemonic":"` + testMnemonic + `"}`)
	payload, err := Encrypt("correct horse", plaintext)
	if err != nil {
		t.Fatalf("Encrypt failed: %v", err)
	}
	decrypted, err := Decrypt("correct horse", payload)
	if err != nil {
		t.Fatalf("Decrypt failed: %v", err)
	}
	if !constantTimeEqual(plaintext, decrypted) {
		t.Fatalf("decrypted payload mismatch")
	}
	if _, err := Decrypt("wrong", payload); !errors.Is(err, ErrInvalidPassword) {
		t.Fatalf("expected ErrInvalidPassword, got %v", err)
	}
}

func TestStoreEncryptsWalletFile(t *testing.T) {
	dir := t.TempDir()
	store := NewStore(filepath.Join(dir, "default.wallet"))
	plain, err := store.Import(testMnemonic, "password", false)
	if err != nil {
		t.Fatalf("Import failed: %v", err)
	}
	data, err := os.ReadFile(store.Path())
	if err != nil {
		t.Fatalf("ReadFile failed: %v", err)
	}
	if bytesContainSecret(data, testMnemonic) {
		t.Fatalf("wallet file contains plaintext mnemonic")
	}
	unlocked, err := store.Unlock("password")
	if err != nil {
		t.Fatalf("Unlock failed: %v", err)
	}
	if unlocked.Account != plain.Account {
		t.Fatalf("unlocked account mismatch")
	}
	if _, err := store.Unlock("wrong"); !errors.Is(err, ErrInvalidPassword) {
		t.Fatalf("expected ErrInvalidPassword, got %v", err)
	}
}
