A validator node for the Tilt Validator Specification on Solana, using Go for off-chain computation and distributed signing!!

## Task Checklist

### 1. **Setup**
- Install the latest Go version.
- Set up Git repository and workspace.
- Run `go mod init <project-name>`.

### 2. **Project Structure & Configuration**
- Create directory structure (`cmd/`, `pkg/`, `internal/`).
- Create `config.yaml`.
- Implement config loader using Viper.

### 3. **Logging & Utility Functions**
- Develop logger using Logrus or Zerolog.
- Add utility functions.

### 4. **Solana Network Integration**
- Create Solana client (`client.go`, `rpc_client.go`, `transaction.go`).
- Test Solana RPC calls.

### 5. **Tilt Business Logic**
- Implement state management (`state.go`).
- Compute distribution (`distribution.go`).
- Enforce rules (`business_rules.go`).

### 6. **Threshold Signature Scheme (TSS)**
- Develop TSS functionalities (`partial_signature.go`, `key_share.go`).
- Implement signature aggregation (`aggregator.go`, `threshold_signature.go`).

### 7. **Internal Models & Handlers**
- Define models (`tilt.go`, `validator.go`, `distribution.go`).
- Create handlers (`distribution_handler.go`, `validator_handler.go`).

### 8. **Main Application CLI**
- Implement main command (`main.go`).
- Initialize and run services.

### 9. **Testing and CI**
- Write unit tests (`distribution_test.go`, `business_rules_test.go`).
- Write integration tests (`solana_test.go`, `validator_test.go`).
- Set up CI/CD pipeline with GitHub Actions.

### 10. **Deployment & Security**
- Deploy to Solana testnet.
- Monitor logs and performance.
- Conduct security audit.


## incode todo's
- TODO: save these keys to HSM
- 
