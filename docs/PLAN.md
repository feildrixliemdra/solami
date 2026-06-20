# Solami — TUI-First Multi-Chain CLI Wallet

## Product Plan

### Overview

Solami is a TUI-first crypto wallet built in Go. It should feel similar to MetaMask in functionality, but designed entirely for the terminal.

The app should run from one command:

```bash
solami start
```

After running the command, the user enters a full-screen interactive terminal wallet interface.

Initial supported networks:

- Ethereum
- Solana

The wallet should use one mnemonic/seed phrase to derive both Ethereum and Solana accounts, so users only need to back up one recovery phrase.

---

# Product Goals

## Primary Goals

1. Create a new wallet from a single mnemonic
2. Import an existing mnemonic
3. Derive Ethereum and Solana accounts from the same mnemonic
4. Store wallet data securely using local encrypted storage
5. View ETH and SOL balances
6. Send ETH and SOL from the TUI
7. Switch networks inside the TUI
8. Provide a responsive, easy-to-use terminal experience

## Non-Goals for v1

- Browser extension support
- Hardware wallet support
- Token swaps
- NFT support
- Staking
- Smart contract deployment
- Multi-signature wallets
- Full dApp connection support

---

# Core UX Direction

Solami should not feel like a traditional command-heavy CLI.

It should feel like a terminal app.

The only required command should be:

```bash
solami start
```

Optional future commands may exist for debugging or automation, but the main user experience should happen inside the TUI.

---

# Recommended Go Libraries

## TUI Framework

Use:

```text
github.com/charmbracelet/bubbletea
github.com/charmbracelet/bubbles
github.com/charmbracelet/lipgloss
github.com/charmbracelet/huh
```

Purpose:

- Full-screen terminal UI
- Keyboard navigation
- Responsive layouts
- Forms
- Inputs
- Confirmation screens
- Styling

## CLI Entrypoint

Use:

```text
github.com/spf13/cobra
```

Purpose:

- Provide the `solami start` command
- Future extensibility if needed

The CLI layer should stay thin. Most functionality should live inside the TUI application.

## Configuration

Use:

```text
github.com/spf13/viper
```

Purpose:

- Store selected network
- Store RPC URLs
- Store active account
- Read config file

## Ethereum

Use:

```text
github.com/ethereum/go-ethereum
```

Purpose:

- Ethereum RPC client
- Address handling
- Balance fetching
- Gas estimation
- Transaction signing
- Transaction broadcasting

## Solana

Use:

```text
github.com/gagliardetto/solana-go
```

Purpose:

- Solana RPC client
- Public/private key handling
- Balance fetching
- Transaction signing
- Transaction broadcasting

## Mnemonic and HD Wallet

Use:

```text
github.com/tyler-smith/go-bip39
github.com/tyler-smith/go-bip32
github.com/miguelmota/go-ethereum-hdwallet
```

The implementation agent should verify the best Solana derivation library/approach because Ethereum and Solana use different key types.

---

# Single Mnemonic Strategy

Solami should use one mnemonic to derive both Ethereum and Solana wallets.

This is important for UX because the user only needs one recovery phrase.

## Ethereum Derivation Path

```text
m/44'/60'/0'/0/0
```

## Solana Derivation Path

Recommended default:

```text
m/44'/501'/0'/0'
```

The agent should verify compatibility with common Solana wallets such as Phantom and Solflare before final implementation.

## Important Technical Note

Ethereum uses secp256k1 keys.

Solana uses ed25519 keys.

Because of this, the implementation should not assume the same private key format works for both chains.

The mnemonic can be shared, but the derived key type and derivation process are different per chain.

Recommended approach:

- Generate one BIP39 mnemonic
- Derive Ethereum account using Ethereum-compatible HD derivation
- Derive Solana account using Solana-compatible ed25519 derivation
- Store only encrypted wallet material locally
- Never store plaintext mnemonic

---

# App Flow

## First Launch

When the user runs:

```bash
solami start
```

If no wallet exists, show onboarding:

```text
Welcome to Solami

[ Create New Wallet ]
[ Import Existing Wallet ]
[ Exit ]
```

## Create Wallet Flow

Steps:

1. Generate mnemonic
2. Show recovery phrase
3. Ask user to confirm they saved it
4. Ask user to re-enter selected words from the phrase
5. Ask user to create password
6. Encrypt wallet data
7. Derive Ethereum and Solana addresses
8. Enter dashboard

## Import Wallet Flow

Steps:

1. User enters mnemonic
2. Validate mnemonic
3. Ask user to create password
4. Derive Ethereum and Solana addresses
5. Encrypt wallet data
6. Enter dashboard

---

# Main TUI Screens

## Dashboard

```text
┌ Solami ───────────────────────────────────┐
│ Active Network: Ethereum Sepolia          │
│ Address: 0x1234...abcd                    │
│ Balance: 0.42 ETH                         │
├───────────────────────────────────────────┤
│ > Send                                    │
│   Receive                                 │
│   Switch Network                          │
│   Accounts                                │
│   Settings                                │
│   Lock Wallet                             │
└───────────────────────────────────────────┘
```

## Network Switcher

Supported v1 networks:

```text
Ethereum
- Mainnet
- Sepolia

Solana
- Mainnet Beta
- Devnet
```

The user should be able to switch networks without leaving the TUI.

## Send Flow

Steps:

1. Choose recipient address
2. Enter amount
3. Fetch fee/gas estimate
4. Show confirmation screen
5. Ask password to unlock signing
6. Sign transaction locally
7. Broadcast transaction
8. Show success/failure result

Confirmation screen:

```text
Confirm Transaction

Network: Ethereum Sepolia
From: 0x1234...abcd
To: 0xabcd...7890
Amount: 0.01 ETH
Estimated Gas: 0.0003 ETH
Total: 0.0103 ETH

[ Confirm ] [ Cancel ]
```

## Receive Screen

Show:

- Full address
- Copy-friendly address text
- Network name
- Warning to only send assets from the selected network

Optional future feature:

- Terminal QR code

## Settings Screen

Settings:

- Default network
- Ethereum RPC URL
- Solana RPC URL
- Auto-lock timeout
- Theme
- Export recovery phrase
- Export private key

Sensitive actions require password confirmation.

---

# Security Requirements

## Must-Have Security Features

1. Never store plaintext mnemonic
2. Encrypt all wallet secrets locally
3. Require password before signing transactions
4. Require password before exporting mnemonic/private key
5. Show transaction confirmation before signing
6. Mask sensitive input
7. Lock wallet after inactivity
8. Avoid logging secrets
9. Avoid printing private keys unless explicitly requested

## Encryption

Recommended:

- AES-256-GCM for encrypted wallet data
- Argon2id or scrypt for deriving encryption key from password

The implementation agent should choose a secure, maintained Go package for password-based key derivation.

## Local Storage

Default directory:

```text
~/.solami/
```

Suggested structure:

```text
~/.solami/
  config.json
  wallets/
    default.wallet
```

Wallet file should contain encrypted data only.

---

# Suggested Architecture

```text
solami/

├── cmd/
│   ├── root.go
│   └── start.go
│
├── internal/
│   ├── app/
│   │   └── app.go
│   │
│   ├── tui/
│   │   ├── model.go
│   │   ├── update.go
│   │   ├── view.go
│   │   ├── screens/
│   │   │   ├── onboarding.go
│   │   │   ├── dashboard.go
│   │   │   ├── send.go
│   │   │   ├── receive.go
│   │   │   ├── network.go
│   │   │   └── settings.go
│   │   └── components/
│   │       ├── menu.go
│   │       ├── form.go
│   │       ├── spinner.go
│   │       └── modal.go
│   │
│   ├── wallet/
│   │   ├── mnemonic.go
│   │   ├── derivation.go
│   │   ├── keystore.go
│   │   ├── encryption.go
│   │   └── account.go
│   │
│   ├── chains/
│   │   ├── ethereum/
│   │   │   ├── client.go
│   │   │   ├── balance.go
│   │   │   ├── send.go
│   │   │   ├── signer.go
│   │   │   └── fees.go
│   │   │
│   │   └── solana/
│   │       ├── client.go
│   │       ├── balance.go
│   │       ├── send.go
│   │       └── signer.go
│   │
│   ├── config/
│   │   └── config.go
│   │
│   └── storage/
│       ├── wallet_store.go
│       └── config_store.go
│
├── docs/
├── tests/
├── go.mod
└── main.go
```

---

# Development Roadmap

## Phase 1 — Project Foundation

Deliverables:

- Go module setup
- Cobra root command
- `solami start` command
- Bubble Tea app shell
- Basic navigation
- Config directory setup

## Phase 2 — Wallet Creation and Import

Deliverables:

- Generate mnemonic
- Import mnemonic
- Validate mnemonic
- Password creation
- Encrypted wallet storage
- Unlock wallet flow

## Phase 3 — Account Derivation

Deliverables:

- Derive Ethereum address from mnemonic
- Derive Solana address from mnemonic
- Display both addresses in TUI
- Confirm derivation compatibility with common wallets

## Phase 4 — Network and Balance

Deliverables:

- Ethereum RPC integration
- Solana RPC integration
- Network switcher
- ETH balance
- SOL balance
- Loading/error states

## Phase 5 — Send Transactions

Deliverables:

- Send ETH on Sepolia first
- Send SOL on Devnet first
- Transaction confirmation screen
- Password-gated signing
- Broadcast transaction
- Show explorer link or transaction hash

## Phase 6 — TUI Polish

Deliverables:

- Responsive layout
- Keyboard shortcuts
- Better errors
- Empty states
- Loading states
- Theme polish
- README
- Demo GIF/video

---

# MVP Scope

The first working version should support:

1. `solami start`
2. Create wallet
3. Import wallet from mnemonic
4. One mnemonic for ETH and SOL
5. Encrypted local wallet storage
6. Show ETH Sepolia balance
7. Show SOL Devnet balance
8. Send ETH on Sepolia
9. Send SOL on Devnet
10. Switch networks in TUI

Mainnet support can exist in config, but testnet/devnet should be prioritized during development.

---

# UX Principles

1. TUI-first, not command-first
2. Clear confirmation before dangerous actions
3. Friendly errors
4. Minimal typing
5. Keyboard-friendly navigation
6. No scary raw blockchain errors unless in debug mode
7. Never surprise the user when money is involved

---

# Success Criteria

Solami v1 is successful if a user can:

1. Install the binary
2. Run `solami start`
3. Create/import one wallet
4. See both Ethereum and Solana addresses
5. View testnet balances
6. Send testnet ETH and SOL
7. Use the app comfortably without needing command documentation

---

# Important Implementation Notes for Agent

The most important technical research item is shared mnemonic derivation.

The agent should deeply verify:

1. Best Go library for Solana ed25519 derivation from BIP39 mnemonic
2. Compatibility with Phantom/Solflare derivation paths
3. Compatibility with MetaMask Ethereum derivation path
4. Safe password-based encryption method
5. Secure local keystore format

Do not proceed with transaction signing until wallet derivation and encryption are verified.
