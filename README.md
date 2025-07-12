# Tilt-Valid: MPC-Based Distributed Validator System for Solana

## 🔐 Overview

**Tilt-Valid** is a modular, threshold-based validator framework built for the Solana blockchain. It enables secure, distributed transaction signing via **Multi-Party Computation (MPC)**, ensuring no single validator has access to the full private key. This design increases robustness and decentralization in validator operations.

🚀 **Live Demo Transaction**:  
[Solana Explorer Devnet Transaction →](https://explorer.solana.com/tx/3gqBAfy8JSNrLVFENK6d5HgrC1He2KXSHTjKtH9VqqjLhDsvDjeUZneGn2jfWTcu6csdefixH7111rvEVtjMkKaL?cluster=devnet)

<img width="1440" alt="Tilt-Valid Architecture" src="https://github.com/user-attachments/assets/63616258-4de4-43c8-b873-f2f32276041a" />

---

## ✨ Key Features

- **Threshold MPC**: Uses a `t-of-n` scheme to collaboratively sign transactions.
- **Secure DKG (Distributed Key Generation)**: Keys are generated without revealing any secret to a single validator.
- **VRF-Based Validator Selection**: Ensures fair and verifiable leader election.
- **Solana Native Integration**: Builds and submits real Solana transactions.
- **Pluggable Tilt Distributions**: Modular support for `simple`, `subtilt`, and `nested` payout structures.
- **CLI Operable**: Quickly simulate validator clusters from the command line.
- **MVP with Working Demo**: Fully functional prototype running on Solana Devnet.

---

## ⚙️ How It Works

1. **Validator Initialization**
   - Validators join the network with their own ID and config.
2. **Distributed Key Generation**
   - Validators jointly compute a shared public key and retain private key shares.
3. **Transaction Construction**
   - A tilt (payment structure) is parsed and turned into Solana instructions.
4. **MPC Signing**
   - Validators jointly compute a signature without revealing secrets.
5. **VRF Selection**
   - Verifiable randomness selects a leader to broadcast the transaction.
6. **Submission**
   - The leader sends the signed transaction to the Solana network.

---

## 🚀 Running a Validator Node

```bash
go run cmd/main.go <validator_id> --tilt-type=<tilt_type>
````

**Arguments:**

* `<validator_id>`: Unique ID for each validator (e.g., 1, 2, 3)
* `--tilt-type=`: Choose one of:

  * `simple`: Single recipient tilt
  * `one_subtilt`: One-level nested
  * `two_subtilts`: Two sub-tilts (original behavior)
  * `nested`: Fully nested recursive structure

---

## 🛡 Security Design

* ✅ **No single-point signing**: Private keys are never reconstructed.
* ✅ **Threshold fault tolerance**: System is functional even if `n - t` nodes are offline.
* ✅ **Replay protection**: Nonce and blockhash management.
* ✅ **VRF-based validator election**: Prevents manipulation in leader selection.

---

## 📦 Project Structure

```bash
tilt-valid/
├── cmd/                    # Main entrypoint for running validators
├── internal/
│   ├── mpc/                # Threshold signing and DKG logic
│   ├── exchange/           # Transport layer (file-based, to be upgraded)
│   ├── distribution/       # Tilt parsing and instruction generation
│   └── vrf/                # Verifiable Random Function logic
├── data/validators.csv     # Cluster configuration
├── utils/                  # Helper functions and constants
├── docs/                   # Diagrams, specs, and explainer docs
└── scripts/                # Scripts to simulate full cluster
```

---

## 🔧 Configuration

Via config files or environment variables:

* Validator paths and identities
* Threshold value `t`
* Tilt type
* Logging level
* Future: Replace file-exchange with libp2p or gRPC transport

---

## 🌐 Live Demo Preview

Run a 3-node validator cluster on Solana Devnet:

```bash
./scripts/run_cluster.sh
```

Then view the example transaction here:
🔗 [https://explorer.solana.com/tx/3gqBAfy8JSNrLVFENK6d5HgrC1He2KXSHTjKtH9VqqjLhDsvDjeUZneGn2jfWTcu6csdefixH7111rvEVtjMkKaL?cluster=devnet](https://explorer.solana.com/tx/3gqBAfy8JSNrLVFENK6d5HgrC1He2KXSHTjKtH9VqqjLhDsvDjeUZneGn2jfWTcu6csdefixH7111rvEVtjMkKaL?cluster=devnet)

---

## 🛠 Future Enhancements

| Feature                        | Status | Description                          |
| ------------------------------ | ------ | ------------------------------------ |
| P2P Transport Layer            | 🧪 MVP | Replace file I/O with libp2p/gRPC    |
| Signature Audit Logging        | 🔜     | Track validator participation logs   |
| ZK or VSS MPC Integration      | 🔜     | Explore zk-MPC and verifiable shares |
| WASM Client SDK                | 🔜     | For browser or JS-based usage        |
| Replay Attack Protection       | ✅      | Handles recent blockhash and nonce   |
| VRF Upgrade with Commit-Reveal | 🔜     | Fair validator selection with proof  |

---

## 🤝 Contributing

We welcome contributions!

### Setup:

```bash
git clone https://github.com/YOUR_USERNAME/tilt-valid
cd tilt-valid
go mod tidy
```

### Key Areas to Explore

* `internal/mpc/`: Enhance signing protocol or add ZK
* `exchange/`: Swap to libp2p/pubsub transport
* `distribution/`: Add new tilt types or structure
* `cmd/main.go`: CLI or orchestration logic

---

## 📫 Contact & Credits

Maintainer: [Yashaswa Varshney](https://github.com/yswa-var)
Email: [yswa.var@gmail.com](mailto:yswa.var@gmail.com)
Built as part of exploring secure Solana infrastructure with MPC and distributed coordination.

---

## 📄 License

MIT — feel free to fork, improve, and contribute.

