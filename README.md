# SolMPC-Node: Multi-Party Computation Validator for Solana

## Overview

**SolMPC-Node** is a distributed validator system that uses **Multi-Party Computation (MPC)** for threshold signing on Solana. Validators jointly generate keys and sign transactions without any single party holding the complete private key.

## Current Implementation

- **MPC Threshold Signing**: 2-of-3 EdDSA key generation and transaction signing
- **Ballot Processing**: Demo voting system with vote tallying and result submission
- **VRF Validator Selection**: Verifiable random function for leader election
- **Solana Integration**: Creates and submits real transactions to Solana devnet

## Quick Start

```bash
# Run 3 validators in tmux
./cmd/run_validators.sh

# Or run single validator
cd cmd && go run *.go 1
```

## Architecture

```
SolMPC-Node/
├── cmd/                    # Validator entrypoint and CLI
├── internal/
│   ├── mpc/                # MPC threshold signing (EdDSA)
│   ├── exchange/           # File-based message transport
│   └── vrf/                # VRF leader selection
└── data/validators.csv     # Validator configuration
```

## Issues to Fix

### 🔧 **Transport Layer**
- **Current**: File-based message exchange between validators
- **Issue**: Not suitable for production, introduces delays and race conditions
- **Fix Needed**: Implement proper P2P networking (libp2p/gRPC)

### 🏗️ **Solana Program Integration** 
- **Current**: Using system program for demo transactions
- **Issue**: No actual voting program deployed on-chain
- **Fix Needed**: Deploy custom Solana program for vote storage and verification

### ⚡ **Performance & Scalability**
- **Current**: 2-of-3 threshold with file I/O bottlenecks
- **Issue**: Doesn't scale beyond demo, slow consensus
- **Fix Needed**: Optimize MPC rounds, async message handling, configurable thresholds

### 🔐 **Production Security**
- **Current**: Fixed test recipients, basic key management
- **Issue**: Not production-ready for real assets
- **Fix Needed**: Proper key rotation, hardware security modules, audit trails

---

## License

MIT
