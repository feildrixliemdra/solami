package storage

import (
	"fmt"
	"os"
	"path/filepath"
)

// Paths contains Solami's local data file paths.
type Paths struct {
	BaseDir   string
	Config    string
	WalletDir string
	Wallet    string
}

// DefaultPaths returns the default ~/.solami data layout.
func DefaultPaths() (Paths, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return Paths{}, fmt.Errorf("resolve home directory: %w", err)
	}
	return NewPaths(filepath.Join(home, ".solami")), nil
}

// NewPaths builds a Solami data layout rooted at baseDir.
func NewPaths(baseDir string) Paths {
	return Paths{
		BaseDir:   baseDir,
		Config:    filepath.Join(baseDir, "config.json"),
		WalletDir: filepath.Join(baseDir, "wallets"),
		Wallet:    filepath.Join(baseDir, "wallets", "default.wallet"),
	}
}

// Ensure creates the Solami data directories.
func (p Paths) Ensure() error {
	if err := os.MkdirAll(p.WalletDir, 0o700); err != nil {
		return fmt.Errorf("create wallet directory: %w", err)
	}
	return nil
}
