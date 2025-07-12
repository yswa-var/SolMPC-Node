# Quick Start Guide

Get Tilt-Valid up and running in under 10 minutes!

## 🚀 Prerequisites

- **Go 1.22+** installed on your system
- **Git** for cloning the repository
- **tmux** (optional, for running multiple validators)

## 📦 Installation

1. **Clone the repository**:

   ```bash
   git clone https://github.com/your-org/tilt-validator.git
   cd tilt-validator
   ```

2. **Install dependencies**:

   ```bash
   go mod download
   ```

3. **Set up environment** (optional):
   ```bash
   cp .env.example .env
   # Edit .env with your configuration
   ```

## 🏃‍♂️ Quick Run

### Option 1: Single Validator (for testing)

```bash
# Start validator 1 with simple tilt
go run cmd/main.go 1 --tilt-type=simple
```

### Option 2: Multiple Validators (recommended)

```bash
# Run the script to start 3 validators in tmux
./cmd/run_validators.sh
```

This will start 3 validators in separate tmux panes, allowing you to see all logs simultaneously.

## 🎯 What Happens Next

1. **Validator Initialization**: Each validator loads its configuration and connects to the network
2. **Distributed Key Generation (DKG)**: Validators collaborate to generate shared keys
3. **Transaction Creation**: System creates a Solana transaction based on tilt data
4. **Distributed Signing**: Validators sign the transaction using threshold cryptography
5. **Validator Selection**: VRF selects one validator to submit the transaction
6. **Transaction Submission**: Selected validator submits to Solana blockchain

## 📊 Expected Output

You should see logs like:

```
[INFO] Starting Validator ID: 1
[INFO] Initiating DKG process...
[SUCCESS] DKG completed. KeyShare length: 256
[INFO] Starting the signing process...
[SUCCESS] Signature generated: a1b2c3d4...
[SUCCESS] This validator (ID: 1) was selected for verification!
[SUCCESS] ✅ Signature verification successful!
Transaction sent! Signature: 5KJvsng9m...
```

## 🔧 Troubleshooting

### Common Issues

1. **"No such file or directory" errors**:

   - Ensure you're in the project root directory
   - Check that all required files exist in `data/` and `utils/`

2. **Transport errors**:

   - Verify that the transport path is writable
   - Check file permissions in the transport directory

3. **DKG timeout**:
   - Ensure all validators are running simultaneously
   - Check network connectivity between validators

### Getting Help

- Check the [Troubleshooting Guide](../user-guides/troubleshooting.md)
- Review the [Debugging Guide](../development/debugging.md)
- Open an issue on GitHub

## 🎓 Next Steps

1. **Learn the Architecture**: Read [System Architecture](../architecture/overview.md)
2. **Understand the Code**: Review [Code Structure](../development/code-structure.md)
3. **Explore Features**: Check out [Tilt Types](../user-guides/tilt-types.md)
4. **Contribute**: Read [Contributing Guidelines](../contributing/guidelines.md)

## 📚 Additional Resources

- [Installation Guide](./installation.md) - Detailed setup instructions
- [Configuration Guide](./configuration.md) - Environment and config setup
- [Running Validators](../user-guides/running-validators.md) - Advanced usage
- [API Reference](../api/) - Complete API documentation

---

**Need help?** Check the [troubleshooting guide](../user-guides/troubleshooting.md) or open an issue on GitHub.
