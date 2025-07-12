# Tilt-Valid: Secure Distributed Validator System for Solana

## Overview

Tilt-Valid is a distributed validator system for Solana blockchain that enables secure, threshold-based transaction signing and validation using multi-party computation (MPC). The system allows a set of validators to collectively authorize transactions without any single validator having complete control over the signing keys.

<img width="1440" alt="Screenshot 2025-03-27 at 3 45 35 PM" src="https://github.com/user-attachments/assets/63616258-4de4-43c8-b873-f2f32276041a" />
example transaction https://explorer.solana.com/tx/3gqBAfy8JSNrLVFENK6d5HgrC1He2KXSHTjKtH9VqqjLhDsvDjeUZneGn2jfWTcu6csdefixH7111rvEVtjMkKaL?cluster=devnet

## Key Features

- **Threshold Multi-Party Computation (MPC)**: Uses a t-of-n threshold scheme where at least t validators must participate to generate signatures
- **Distributed Key Generation (DKG)**: Secure generation of shared keys without any single party knowing the complete key
- **VRF-based Validator Selection**: Fair, verifiable random selection of validators for transaction verification
- **Solana Transaction Integration**: Creates, signs, and submits transactions to the Solana blockchain
- **Configurable Tilt Types**: Supports various distribution patterns for payments (simple, one_subtilt, two_subtilts, nested)

## Architecture

The system consists of several key components:

- **Validators**: Independent nodes that participate in the consensus and signing process
- **Transport Layer**: Handles secure message exchange between validators
- **MPC Protocol**: Implements threshold signing using distributed key shares
- **Distribution System**: Manages payment allocations to recipients
- **VRF Selection**: Uses Verifiable Random Functions to select validators for verification

## How It Works

1. **Validator Setup**: Each validator initializes with a unique ID and connects to the validator network

2. **Distributed Key Generation**:

   - Validators collectively generate a shared public key
   - Each validator receives a key share without revealing it to others
   - Requires threshold number of validators to participate

3. **Transaction Creation**:

   - System reads tilt data and allocates amounts to recipients
   - Creates Solana instructions with appropriate recipient accounts and amounts
   - Builds a complete transaction with recent blockhash

4. **Distributed Signing**:

   - Transaction data is hashed and distributed to validators
   - Validators collaborate to sign the transaction without revealing their key shares
   - Produces a valid signature only when threshold validators participate

5. **Validator Selection & Verification**:

   - Each validator generates a VRF hash
   - Based on these hashes, one validator is randomly selected
   - Selected validator verifies the collective signature and submits the transaction

6. **Transaction Submission**:
   - The selected validator sends the signed transaction to the Solana network
   - Transaction is processed on the blockchain

## Usage

Run a validator node with:

```bash
go run cmd/main.go <validator_id> [--tilt-type=<tilt_type>]
```

Where:

- `<validator_id>` is the unique identifier for the validator
- `<tilt_type>` is one of: simple, one_subtilt, two_subtilts, nested
  - simple: A simple tilt with one recipient
  - one_subtilt: Tilt with one sub-tilt
  - two_subtilts: Tilt with two sub-tilts (matches original behavior)
  - nested: A nested tilt structure with multiple levels

## Security Features

- No single validator has complete control over signing keys
- Threshold scheme ensures security even if some validators are compromised
- VRF selection provides unpredictable, fair validator selection
- Distributed architecture eliminates single points of failure

## Dependencies

- Solana Go SDK
- Ed25519 cryptography
- Custom MPC implementation for secure multi-party computation

## Configuration

Configure the system through the config file to set:

- Validator paths
- Threshold requirements
- Network parameters
- Database locations for tilt data

This distributed architecture ensures high security and availability for Solana transaction signing and validation.

## Contributing

We welcome contributions! To get started:

- **Setup:**

  1. Clone the repository and run `go mod tidy` to install dependencies.
  2. Review the project structure and key files (see below).
  3. Use the provided scripts (e.g., `cmd/run_validators.sh`) to run the system locally.

- **Key Files & Directories:**

  - `cmd/main.go`: Main entry point and flow logic
  - `internal/mpc/`: Multi-party computation (MPC) logic
  - `internal/exchange/`: File-based transport layer
  - `internal/distribution/`: Payment distribution logic
  - `utils/`: Utility functions and tilt data helpers
  - `data/validators.csv`: Validator configuration

- **How to Contribute:**
  - Fork the repo and create a feature branch.
  - Make your changes with clear, concise commits.
  - Add or update tests if needed.
  - Open a pull request with a description of your changes.
  - Follow Go best practices and keep code/documentation clear.

For more details, see [CONTRIBUTING.md](CONTRIBUTING.md).
