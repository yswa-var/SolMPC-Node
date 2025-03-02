# Tilt Validator Node

This project implements a **validator node** that participates in the Tilt payment distribution flow on the Solana blockchain. The system ensures **secure and tamper-resistant payouts** through a decentralised, threshold-based approach. Validators work collaboratively off-chain to compute distributions, with the final result secured by an on-chain verified aggregated signature. This approach minimises trust and maximises transparency within the Tilt ecosystem.

## Key Features

* **Decentralised Security**: Implements an **M-of-N threshold signature scheme**, preventing any single validator from unilaterally pushing malicious distributions.
* **Off-Chain Computation, On-Chain Enforcement**: Complex distribution logic is computed off-chain, while final settlement is enforced by a single on-chain transaction with a verified aggregated signature.
* **Random Subset Selection**: A smaller group (K-of-N) is randomly chosen from the entire pool of validators for each distribution. This prevents collusion and ensures a dynamic validation process.
* **Flexible Distribution Logic**: Accommodates various payout formulas, curation algorithms and sub-tilt dependencies.
* **Go Implementation**: The validator node is implemented in Go, leveraging various libraries for Solana interaction, cryptography, networking, and configuration.

## Core Components

1.  **Solana RPC & On-Chain Data**:
    * Libraries: `github.com/gagliardetto/solana-go` or `github.com/portto/solana-go-sdk`.
    * Usage: Connects to a Solana cluster and fetches `AccountInfo` for the Tilt Program Derived Address (PDA), retrieving data such as the tilt's balance, business rules, and sub-tilt references.

2.  **Threshold Signatures & Cryptography**:
    * Libraries: `github.com/binance-chain/tss-lib` (for threshold ECDSA), `crypto/ed25519` (built-in Go library for basic Ed25519 signing).
    * Usage: Holds a private key (or partial secret) and generates partial signatures of the final distribution message.

3.  **Networking & Communication**:
    * Libraries: `github.com/libp2p/go-libp2p` (for P2P networks), standard HTTP/REST or gRPC with `google.golang.org/grpc`.
    * Usage: Exchanges partial signatures with other validators and sends them to the aggregator. The aggregator broadcasts the final transaction to the Solana cluster.

4.  **Configuration & Logging**:
    * Libraries: `github.com/spf13/viper` (for managing configs), `github.com/sirupsen/logrus` or `github.com/rs/zerolog` (for structured logging).
    * Usage: Stores cluster endpoint, tilt program ID, aggregator addresses, threshold parameters, and ensures uniform logging across all nodes.
