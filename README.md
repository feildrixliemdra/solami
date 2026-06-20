# Solami 

Solami is a TUI-first, multi-chain crypto wallet designed entirely for the terminal. Built with Go, it provides a secure, interactive, and responsive interface to manage your assets without leaving your command line.

```text
   _____       __                 _
  / ___/____  / /___ _____ ___  _(_)
  \__ \/ __ \/ / __ `/ __ `__ \/ / /
 ___/ / /_/ / / /_/ / / / / / / / /
/____/\____/_/\__,_/_/ /_/ /_/_/_/
```

## Features

- **TUI-First Experience**: Fully interactive full-screen terminal interface powered by Charm's `bubbletea`, `bubbles`, `lipgloss`, and `huh`.
- **Single Mnemonic**: Uses one 12-word recovery phrase to derive accounts for multiple chains (Ethereum and Solana).
- **Secure Local Storage**: Keystore file encrypted locally using AES-256-GCM. Derives key from password using scrypt KDF.
- **Multi-Chain Balance**: Real-time balance checks for Ethereum (Sepolia testnet) and Solana (Devnet).
- **Automatic & Manual Refresh**: Balances auto-refresh every 15 seconds in the background or instantly by pressing `r`.
- **In-TUI Network Switcher**: Seamlessly switch active networks from the dashboard.

## Derivation Paths

Solami derives separate key pairs using different cryptographic curves from the same seed phrase:
- **Ethereum** (secp256k1): `m/44'/60'/0'/0/0`
- **Solana** (ed25519, SLIP-0010): `m/44'/501'/0'/0'`

## Supported Networks

Currently, Solami is in its v1 stage and only supports the following networks:
- **Ethereum**: Mainnet & Sepolia Testnet
- **Solana**: Mainnet Beta & Devnet

*(Note: Sepolia and Devnet are prioritized during active development and staging).*

## Future Roadmap

Planned features and improvements for upcoming releases:
- **Additional Blockchains**: Support for Bitcoin (BTC) and other Layer 1/2 networks.
- **Transaction History**: View recent transactions and logs directly in the TUI dashboard.
- **Hardware Wallet Integration**: Connect and sign transactions via Ledger/Trezor devices.
- **DEX Swaps**: Direct token swapping inside the TUI using DEX aggregators.
- **Terminal QR Codes**: Render ASCII QR codes of wallet addresses for easier receiving.
- **Custom RPC Endpoints**: Add custom RPC nodes dynamically via settings screen.

## Project Structure

```text
solami/
├── cmd/               # CLI commands (cobra)
├── internal/
│   ├── app/           # App engine setup
│   ├── chains/        # Chain client implementations (Ethereum / Solana)
│   ├── config/        # User configuration (viper)
│   ├── networks/      # Network configurations
│   ├── storage/       # File path resolver
│   ├── tui/           # Bubble Tea model, view, update loop
│   └── wallet/        # BIP-39 mnemonic, keystore encryption, HD derivation
├── main.go            # Entry point
└── Makefile           # Build and run commands
```

## Getting Started

### Prerequisites

- Go 1.24.0 or later
- Make utility

### Build & Install

Clone the repository and build the binary:

```bash
git clone https://github.com/feildrix/solami.git
cd solami
make build
```

The compiled binary will be placed inside `bin/solami`.

### Running the App

You can start the terminal wallet with:

```bash
./bin/solami start
```

Or run it directly using the Makefile:

```bash
make start
```

## Keyboard Shortcuts

- **`Up / Down` (or `k / j`)**: Navigate through menu items.
- **`Enter`**: Select/confirm action.
- **`Esc`**: Go back to the dashboard or previous screen.
- **`r` (Dashboard)**: Manually refresh account balance.
- **`Ctrl + C`**: Quit the application.

## Security

- **Encryption**: Key material is encrypted locally on disk at `~/.solami/wallets/default.wallet`. Plaintext secrets are never stored in memory or logged.
- **Confirmation Gating**: Password confirmation is requested before critical actions.

## License

This project is licensed under the [MIT License](LICENSE).
