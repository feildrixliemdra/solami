package networks

// Chain identifies a supported blockchain family.
type Chain string

const (
	// ChainEthereum identifies Ethereum-compatible networks.
	ChainEthereum Chain = "ethereum"
	// ChainSolana identifies Solana-compatible networks.
	ChainSolana Chain = "solana"
)

// Network describes a supported Solami network.
type Network struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Chain       Chain  `json:"chain"`
	Symbol      string `json:"symbol"`
	RPCURL      string `json:"rpc_url"`
	ExplorerURL string `json:"explorer_url"`
	SendEnabled bool   `json:"send_enabled"`
}

const (
	// LegacyEthereumSepoliaRPCURL is kept so config loading can migrate older installs.
	LegacyEthereumSepoliaRPCURL = "https://rpc.sepolia.org"
	// EthereumSepoliaRPCURL is the default v1 Sepolia JSON-RPC endpoint.
	EthereumSepoliaRPCURL = "https://ethereum-sepolia-rpc.publicnode.com"
	// EthereumMainnetRPCURL is the default v1 Ethereum mainnet JSON-RPC endpoint.
	EthereumMainnetRPCURL = "https://ethereum-rpc.publicnode.com"
)

const (
	// EthereumSepolia is the default v1 network.
	EthereumSepolia = "ethereum-sepolia"
	// EthereumMainnet is view-only in v1.
	EthereumMainnet = "ethereum-mainnet"
	// SolanaDevnet supports v1 sending.
	SolanaDevnet = "solana-devnet"
	// SolanaMainnet is view-only in v1.
	SolanaMainnet = "solana-mainnet"
)

// Defaults returns the v1 supported networks.
func Defaults() []Network {
	return []Network{
		{
			ID:          EthereumSepolia,
			Name:        "Ethereum Sepolia",
			Chain:       ChainEthereum,
			Symbol:      "ETH",
			RPCURL:      EthereumSepoliaRPCURL,
			ExplorerURL: "https://sepolia.etherscan.io/tx/",
			SendEnabled: true,
		},
		{
			ID:          SolanaDevnet,
			Name:        "Solana Devnet",
			Chain:       ChainSolana,
			Symbol:      "SOL",
			RPCURL:      "https://api.devnet.solana.com",
			ExplorerURL: "https://explorer.solana.com/tx/?cluster=devnet",
			SendEnabled: true,
		},
		{
			ID:          EthereumMainnet,
			Name:        "Ethereum Mainnet",
			Chain:       ChainEthereum,
			Symbol:      "ETH",
			RPCURL:      EthereumMainnetRPCURL,
			ExplorerURL: "https://etherscan.io/tx/",
			SendEnabled: false,
		},
		{
			ID:          SolanaMainnet,
			Name:        "Solana Mainnet Beta",
			Chain:       ChainSolana,
			Symbol:      "SOL",
			RPCURL:      "https://api.mainnet-beta.solana.com",
			ExplorerURL: "https://explorer.solana.com/tx/",
			SendEnabled: false,
		},
	}
}

// ByID returns a network by ID.
func ByID(id string) (Network, bool) {
	for _, network := range Defaults() {
		if network.ID == id {
			return network, true
		}
	}
	return Network{}, false
}
