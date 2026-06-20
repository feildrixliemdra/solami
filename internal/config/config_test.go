package config

import (
	"path/filepath"
	"testing"

	"github.com/feildrix/solami/internal/networks"
)

func TestLoadDefaultConfigWhenMissing(t *testing.T) {
	cfg, err := Load(filepath.Join(t.TempDir(), "config.json"))
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}
	if cfg.ActiveNetworkID != networks.EthereumSepolia {
		t.Fatalf("unexpected active network: %s", cfg.ActiveNetworkID)
	}
	if cfg.AutoLockMinutes != 5 {
		t.Fatalf("unexpected auto lock: %d", cfg.AutoLockMinutes)
	}
	if len(cfg.Networks) != len(networks.Defaults()) {
		t.Fatalf("expected default networks")
	}
}

func TestSaveAndLoadConfig(t *testing.T) {
	path := filepath.Join(t.TempDir(), "config.json")
	cfg := Default()
	cfg.ActiveNetworkID = networks.SolanaDevnet
	cfg.RPCURLs[networks.SolanaDevnet] = "http://127.0.0.1:8899"
	if err := Save(path, cfg); err != nil {
		t.Fatalf("Save failed: %v", err)
	}
	loaded, err := Load(path)
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}
	if loaded.ActiveNetworkID != networks.SolanaDevnet {
		t.Fatalf("unexpected active network: %s", loaded.ActiveNetworkID)
	}
	if loaded.ActiveNetwork().RPCURL != "http://127.0.0.1:8899" {
		t.Fatalf("custom rpc not applied: %s", loaded.ActiveNetwork().RPCURL)
	}
}

func TestLoadMigratesLegacySepoliaRPC(t *testing.T) {
	path := filepath.Join(t.TempDir(), "config.json")
	cfg := Default()
	cfg.RPCURLs[networks.EthereumSepolia] = networks.LegacyEthereumSepoliaRPCURL
	if err := Save(path, cfg); err != nil {
		t.Fatalf("Save failed: %v", err)
	}
	loaded, err := Load(path)
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}
	if got := loaded.RPCURLs[networks.EthereumSepolia]; got != networks.EthereumSepoliaRPCURL {
		t.Fatalf("expected migrated rpc %q, got %q", networks.EthereumSepoliaRPCURL, got)
	}
}
