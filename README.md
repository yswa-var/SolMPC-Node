```
Branching:
- main: simulate mvp creation of transaction, distribution. handels tree structure of tilts. with multiple test cases (without solana/smart-contract integration)
- solana_integration: simulate mvp creation of transaction with solana integration for fewer test cases.
```

- **cmd/**: Initial setup and configuration.
  - **main.go**: Entry point for the application.
  - **config/config.go**: Application configuration.
  - **vrf_utils.go**: Logic for validator selection during final submission.

- **data/**: Validator database.
  - **validator.csv**: Registry of validators.

- **eddsa/**: Temporary testing files.

- **internals/**: Core logic and modules.
  - **distribution/**:
    - **distribution.go**: Main tilt distribution logic.
    - **distribution.md**: Explanation and plan for the distribution logic.
  - **exchange/**: Mimics a gossip protocol using CSVs.
    - **setup.go**: Cleans previous data.
    - **receiver.go**: Handles incoming communications between validators.
    - **sender.go**: Sends messages between validators.
  - **mpc/**:
    - **keygen.go**: Key generation for DKG (Distributed Key Generation).
    - **mpc_test.go**: Unit tests for the MPC package.
    - **party.go**: Logic for creating MPC parties.
    - **sign.go**: Generates the final signature.
  - **Transport/**: Temporary CSV files for inter-validator communication.
  - **Validators/**:
    - **validator.go**: Logic for creating validator parameters.
  - **solana_tx_test/**: Test module for submitting final signature to Solana.
  - **utils/**: Utility files for supporting the tilt distribution.
    - **create-tilt-flag.txt**: Flag for validators during tilt creation.
    - **current-tilt-data.txt**: Communicates transactions needing validation.
    - **distribution-dump.csv**: Facilitates manual validation of signed transactions.
    - **logger.go**: Logging utilities for type and format fixes.
    - **tilt-creator.go**: Helper functions for transaction creation.
    - **tiltdb.csv**: Database of created transactions for manual validation.

- **.env**: Paths and signatures configuration (requires refactoring).
- **go.mod, go.sum**: Dependencies and module management files.
- **lib.rs**: Solana smart contract code.
- **validator-keypair.json**: Devnet keypair for hosted smart contracts.

## Setup
1. Clone the repository.
2. Configure the `.env` file with appropriate paths and signatures.
3. install tmux "brew install tmux"
4. compile and run `cmd/run_validator.sh`

## Usage
1. Execute the main file: `go run cmd/main.go`.
2. Validators communicate via the exchange module using CSV files.
3. Final transactions are pushed to Solana devnet after signature generation.

## Testing
- Unit tests for the MPC module: `go test ./internals/mpc/...`
- Solana smart contract testing: Refer to `solana_tx_test`.
