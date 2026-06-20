package config

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/feildrix/solami/internal/networks"
)

// Config stores non-secret Solami preferences.
type Config struct {
	ActiveNetworkID string                      `json:"active_network_id"`
	RPCURLs         map[string]string           `json:"rpc_urls"`
	AutoLockMinutes int                         `json:"auto_lock_minutes"`
	Theme           string                      `json:"theme"`
	Networks        map[string]networks.Network `json:"-"`
}

// Default returns the default non-secret configuration.
func Default() Config {
	rpcs := make(map[string]string)
	networkMap := make(map[string]networks.Network)
	for _, network := range networks.Defaults() {
		rpcs[network.ID] = network.RPCURL
		networkMap[network.ID] = network
	}
	return Config{
		ActiveNetworkID: networks.EthereumSepolia,
		RPCURLs:         rpcs,
		AutoLockMinutes: 5,
		Theme:           "default",
		Networks:        networkMap,
	}
}

// Load reads config from disk, creating parent directories when needed.
func Load(path string) (Config, error) {
	cfg := Default()
	data, err := os.ReadFile(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return cfg, nil
		}
		return Config{}, fmt.Errorf("read config: %w", err)
	}
	if err := json.Unmarshal(data, &cfg); err != nil {
		return Config{}, fmt.Errorf("decode config: %w", err)
	}
	networkMap := make(map[string]networks.Network)
	migrated := false
	for _, network := range networks.Defaults() {
		if rpcURL := cfg.RPCURLs[network.ID]; rpcURL != "" {
			if network.ID == networks.EthereumSepolia && rpcURL == networks.LegacyEthereumSepoliaRPCURL {
				rpcURL = network.RPCURL
				cfg.RPCURLs[network.ID] = rpcURL
				migrated = true
			}
			network.RPCURL = rpcURL
		}
		networkMap[network.ID] = network
	}
	cfg.Networks = networkMap
	if _, ok := cfg.Networks[cfg.ActiveNetworkID]; !ok {
		cfg.ActiveNetworkID = networks.EthereumSepolia
	}
	if cfg.AutoLockMinutes <= 0 {
		cfg.AutoLockMinutes = 5
		migrated = true
	}
	if migrated {
		if err := Save(path, cfg); err != nil {
			return Config{}, err
		}
	}
	return cfg, nil
}

// Save writes config to disk.
func Save(path string, cfg Config) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		return fmt.Errorf("create config dir: %w", err)
	}
	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return fmt.Errorf("encode config: %w", err)
	}
	if err := os.WriteFile(path, data, 0o600); err != nil {
		return fmt.Errorf("write config: %w", err)
	}
	return nil
}

// ActiveNetwork returns the currently selected network.
func (c Config) ActiveNetwork() networks.Network {
	network, ok := c.Networks[c.ActiveNetworkID]
	if !ok {
		network, _ = networks.ByID(networks.EthereumSepolia)
	}
	return network
}
