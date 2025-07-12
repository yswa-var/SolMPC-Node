# Development Setup Guide

This guide will help you set up a development environment for contributing to Tilt-Valid.

## 🛠️ Prerequisites

### Required Software

- **Go 1.22+**: [Download from golang.org](https://golang.org/dl/)
- **Git**: [Download from git-scm.com](https://git-scm.com/)
- **tmux** (optional): For running multiple validators
- **VS Code** (recommended): With Go extension

### System Requirements

- **Operating System**: Linux, macOS, or Windows (with WSL)
- **Memory**: 4GB RAM minimum, 8GB recommended
- **Storage**: 2GB free space
- **Network**: Internet connection for dependencies

## 🚀 Development Environment Setup

### 1. Clone the Repository

```bash
# Clone the repository
git clone https://github.com/your-org/tilt-validator.git
cd tilt-validator

# Verify the clone
ls -la
```

### 2. Install Dependencies

```bash
# Download Go modules
go mod download

# Verify dependencies
go mod verify
```

### 3. Set Up Environment Variables

```bash
# Copy example environment file
cp .env.example .env

# Edit the environment file
nano .env
```

**Example `.env` file**:

```env
# Solana Configuration
SOLANA_PRODUCT_ID=EM7AAngMgQPXizeuwAKaBvci79DhRxJMBYjRVoJWYEH3

# File Paths
VALIDATOR_PATH=/path/to/your/tilt-validator/data/
TRANSPORT_PATH=/path/to/your/tilt-validator/internal/
TILT_DB=/path/to/your/tilt-validator/utils/tiltdb.csv
DISTRIBUTION_DUMP=/path/to/your/tilt-validator/utils/distribution-dump.csv
```

### 4. Create Required Directories

```bash
# Create data directories
mkdir -p data
mkdir -p internal/Transport1.csv internal/Transport2.csv internal/Transport3.csv
mkdir -p utils

# Set proper permissions
chmod 755 data internal utils
```

### 5. Initialize Test Data

```bash
# Create validator data
echo "ID,Name,stake,active,VRFHash" > data/validators.csv
echo "1,bcvs,100.5,true,0" >> data/validators.csv
echo "2,bbdj,50.2,true,0" >> data/validators.csv
echo "3,sujskd,20.0,true,0" >> data/validators.csv

# Create transport files
touch internal/Transport1.csv internal/Transport2.csv internal/Transport3.csv
```

## 🔧 IDE Setup

### VS Code Configuration

1. **Install Go Extension**:

   - Open VS Code
   - Go to Extensions (Ctrl+Shift+X)
   - Search for "Go" by Google
   - Install the extension

2. **Configure Go Tools**:

   ```bash
   # Install Go tools
   go install golang.org/x/tools/gopls@latest
   go install github.com/go-delve/delve/cmd/dlv@latest
   go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
   ```

3. **VS Code Settings** (`.vscode/settings.json`):
   ```json
   {
     "go.useLanguageServer": true,
     "go.lintTool": "golangci-lint",
     "go.lintFlags": ["--fast"],
     "go.formatTool": "goimports",
     "go.testFlags": ["-v"],
     "files.associations": {
       "*.go": "go"
     }
   }
   ```

### GoLand Configuration

1. **Import Project**:

   - Open GoLand
   - Import project from existing sources
   - Select the `tilt-validator` directory

2. **Configure Run Configurations**:
   - Go to Run → Edit Configurations
   - Add new Go Build configuration
   - Set working directory to project root
   - Set program arguments: `1 --tilt-type=simple`

## 🧪 Testing Setup

### 1. Run Unit Tests

```bash
# Run all tests
go test ./...

# Run tests with coverage
go test -cover ./...

# Run tests with verbose output
go test -v ./...
```

### 2. Run Integration Tests

```bash
# Test the complete flow
./cmd/run_validators.sh

# Or test individual components
go run cmd/main.go 1 --tilt-type=simple
```

### 3. Benchmark Tests

```bash
# Run benchmarks
go test -bench=. ./...

# Run benchmarks with memory profiling
go test -bench=. -benchmem ./...
```

## 🔍 Debugging Setup

### 1. Delve Debugger

```bash
# Install delve
go install github.com/go-delve/delve/cmd/dlv@latest

# Debug the main application
dlv debug cmd/main.go -- 1 --tilt-type=simple
```

### 2. VS Code Debugging

Create `.vscode/launch.json`:

```json
{
  "version": "0.2.0",
  "configurations": [
    {
      "name": "Debug Validator 1",
      "type": "go",
      "request": "launch",
      "mode": "debug",
      "program": "${workspaceFolder}/cmd/main.go",
      "args": ["1", "--tilt-type=simple"],
      "cwd": "${workspaceFolder}"
    },
    {
      "name": "Debug Validator 2",
      "type": "go",
      "request": "launch",
      "mode": "debug",
      "program": "${workspaceFolder}/cmd/main.go",
      "args": ["2", "--tilt-type=simple"],
      "cwd": "${workspaceFolder}"
    }
  ]
}
```

## 📊 Monitoring and Logging

### 1. Enable Debug Logging

```bash
# Set debug environment variable
export DEBUG=true

# Run with debug logging
go run cmd/main.go 1 --tilt-type=simple
```

### 2. Monitor File Changes

```bash
# Watch for file changes during development
fswatch -o . | xargs -n1 -I{} go run cmd/main.go 1 --tilt-type=simple
```

### 3. Log Analysis

```bash
# Filter logs by validator
grep "Validator 1" logs.txt

# Search for errors
grep "ERROR" logs.txt

# Monitor transport files
tail -f internal/Transport*.csv
```

## 🚀 Development Workflow

### 1. Making Changes

```bash
# Create a new branch
git checkout -b feature/your-feature-name

# Make your changes
# ... edit files ...

# Test your changes
go test ./...
go run cmd/main.go 1 --tilt-type=simple

# Commit your changes
git add .
git commit -m "Add your feature description"
```

### 2. Code Quality Checks

```bash
# Run linter
golangci-lint run

# Format code
go fmt ./...

# Run go vet
go vet ./...

# Check for security issues
gosec ./...
```

### 3. Performance Testing

```bash
# Profile CPU usage
go test -cpuprofile=cpu.prof -bench=. ./...

# Profile memory usage
go test -memprofile=mem.prof -bench=. ./...

# Analyze profiles
go tool pprof cpu.prof
go tool pprof mem.prof
```

## 🔧 Common Development Tasks

### 1. Adding New Tilt Types

1. **Update tilt creator** (`utils/tilt-creator.go`)
2. **Add test cases** (`internal/distribution/distribution_test.go`)
3. **Update documentation** (`docs/user-guides/tilt-types.md`)

### 2. Modifying MPC Protocol

1. **Update party implementation** (`internal/mpc/party.go`)
2. **Modify key generation** (`internal/mpc/keygen.go`)
3. **Update signing logic** (`internal/mpc/sign.go`)

### 3. Enhancing Transport Layer

1. **Modify sender** (`internal/exchange/sender.go`)
2. **Update receiver** (`internal/exchange/reciver.go`)
3. **Add new transport types** (`internal/exchange/`)

### 4. Adding New Validators

1. **Update validator data** (`data/validators.csv`)
2. **Modify run script** (`cmd/run_validators.sh`)
3. **Update configuration** (`cmd/config/config.go`)

## 🐛 Troubleshooting

### Common Issues

1. **"No such file or directory"**:

   ```bash
   # Check file paths in .env
   cat .env

   # Verify directories exist
   ls -la data/ internal/ utils/
   ```

2. **Transport errors**:

   ```bash
   # Check file permissions
   ls -la internal/Transport*.csv

   # Fix permissions if needed
   chmod 644 internal/Transport*.csv
   ```

3. **DKG timeout**:

   ```bash
   # Check if all validators are running
   ps aux | grep main.go

   # Verify transport files are being written
   tail -f internal/Transport*.csv
   ```

4. **Go module issues**:

   ```bash
   # Clean module cache
   go clean -modcache

   # Re-download dependencies
   go mod download
   ```

### Getting Help

- **Check logs**: Look for error messages in console output
- **Review documentation**: Check relevant docs in `docs/`
- **Search issues**: Look for similar problems in GitHub issues
- **Ask questions**: Use GitHub Discussions for help

## 📚 Additional Resources

- [Code Structure](./code-structure.md) - Understanding the codebase
- [Adding Features](./adding-features.md) - How to add new functionality
- [Testing Guide](./testing.md) - Writing and running tests
- [Debugging](./debugging.md) - Troubleshooting common issues
- [Contributing Guidelines](../contributing/guidelines.md) - How to contribute

---

**Need help?** Check the [troubleshooting guide](../user-guides/troubleshooting.md) or open an issue on GitHub.
