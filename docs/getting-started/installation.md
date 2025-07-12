# Installation Guide

This guide provides detailed instructions for installing and setting up Tilt-Valid on your system.

## 🛠️ System Requirements

### Minimum Requirements

- **Operating System**: Linux, macOS, or Windows (with WSL)
- **Go Version**: 1.22 or higher
- **Memory**: 4GB RAM
- **Storage**: 2GB free space
- **Network**: Internet connection for dependencies

### Recommended Requirements

- **Operating System**: Linux (Ubuntu 20.04+ or CentOS 8+)
- **Go Version**: 1.22.2 or higher
- **Memory**: 8GB RAM
- **Storage**: 5GB free space
- **CPU**: 4+ cores
- **Network**: Stable internet connection

## 📦 Installation Methods

### Method 1: From Source (Recommended)

#### Step 1: Install Go

**Linux (Ubuntu/Debian)**:

```bash
# Download Go
wget https://go.dev/dl/go1.22.2.linux-amd64.tar.gz

# Extract to /usr/local
sudo tar -C /usr/local -xzf go1.22.2.linux-amd64.tar.gz

# Add to PATH
echo 'export PATH=$PATH:/usr/local/go/bin' >> ~/.bashrc
source ~/.bashrc

# Verify installation
go version
```

**macOS**:

```bash
# Using Homebrew
brew install go

# Or download from golang.org
# Visit https://golang.org/dl/ and download the macOS installer
```

**Windows**:

```bash
# Download from golang.org
# Visit https://golang.org/dl/ and download the Windows installer
# Follow the installation wizard
```

#### Step 2: Clone Repository

```bash
# Clone the repository
git clone https://github.com/your-org/tilt-validator.git
cd tilt-validator

# Verify the clone
ls -la
```

#### Step 3: Install Dependencies

```bash
# Download Go modules
go mod download

# Verify dependencies
go mod verify

# Check for any missing dependencies
go mod tidy
```

#### Step 4: Build the Project

```bash
# Build the main application
go build -o tilt-validator cmd/main.go

# Verify the build
./tilt-validator --help
```

### Method 2: Using Docker

#### Step 1: Install Docker

**Linux**:

```bash
# Install Docker
curl -fsSL https://get.docker.com -o get-docker.sh
sudo sh get-docker.sh

# Add user to docker group
sudo usermod -aG docker $USER
```

**macOS**:

```bash
# Install Docker Desktop
brew install --cask docker
```

**Windows**:

```bash
# Download Docker Desktop from https://www.docker.com/products/docker-desktop
# Follow the installation wizard
```

#### Step 2: Create Dockerfile

Create a `Dockerfile` in the project root:

```dockerfile
FROM golang:1.22-alpine

WORKDIR /app

# Copy go mod files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Build the application
RUN go build -o tilt-validator cmd/main.go

# Create necessary directories
RUN mkdir -p data internal utils

# Expose port if needed
EXPOSE 8080

# Run the application
CMD ["./tilt-validator"]
```

#### Step 3: Build and Run

```bash
# Build the Docker image
docker build -t tilt-validator .

# Run the container
docker run -it --rm tilt-validator 1 --tilt-type=simple
```

### Method 3: Using Package Managers

#### Using Go Install

```bash
# Install directly using go install
go install github.com/your-org/tilt-validator/cmd/main.go@latest

# Run the installed binary
tilt-validator 1 --tilt-type=simple
```

## 🔧 Configuration Setup

### Step 1: Create Environment File

```bash
# Copy example environment file
cp .env.example .env

# Edit the environment file
nano .env
```

### Step 2: Configure Environment Variables

Edit `.env` with your configuration:

```env
# Solana Configuration
SOLANA_PRODUCT_ID=EM7AAngMgQPXizeuwAKaBvci79DhRxJMBYjRVoJWYEH3

# File Paths (adjust these to your installation path)
VALIDATOR_PATH=/path/to/your/tilt-validator/data/
TRANSPORT_PATH=/path/to/your/tilt-validator/internal/
TILT_DB=/path/to/your/tilt-validator/utils/tiltdb.csv
DISTRIBUTION_DUMP=/path/to/your/tilt-validator/utils/distribution-dump.csv

# Optional: Debug mode
DEBUG=false
```

### Step 3: Create Required Directories

```bash
# Create data directories
mkdir -p data
mkdir -p internal
mkdir -p utils

# Set proper permissions
chmod 755 data internal utils
```

### Step 4: Initialize Test Data

```bash
# Create validator data
cat > data/validators.csv << EOF
ID,Name,stake,active,VRFHash
1,bcvs,100.5,true,0
2,bbdj,50.2,true,0
3,sujskd,20.0,true,0
EOF

# Create transport files
touch internal/Transport1.csv internal/Transport2.csv internal/Transport3.csv

# Set file permissions
chmod 644 internal/Transport*.csv
chmod 644 data/validators.csv
```

## 🧪 Verification

### Step 1: Test Single Validator

```bash
# Test with single validator
go run cmd/main.go 1 --tilt-type=simple
```

Expected output:

```
[INFO] Starting Validator ID: 1
===== Starting Validator ID: 1 =====
[INFO] Initiating DKG process...
[SUCCESS] DKG completed. KeyShare length: 256
...
```

### Step 2: Test Multiple Validators

```bash
# Test with multiple validators
./cmd/run_validators.sh
```

Expected output:

```
# Three tmux panes with validators running
# Each showing similar output to single validator
```

### Step 3: Verify File Creation

```bash
# Check that files were created
ls -la utils/tiltdb.csv
ls -la internal/Transport*.csv

# Check file contents
head -5 utils/tiltdb.csv
head -5 internal/Transport1.csv
```

## 🔧 Post-Installation Setup

### Step 1: Install Development Tools (Optional)

```bash
# Install Go tools for development
go install golang.org/x/tools/gopls@latest
go install github.com/go-delve/delve/cmd/dlv@latest
go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
go install golang.org/x/tools/cmd/goimports@latest
```

### Step 2: Set Up IDE (Optional)

**VS Code**:

```bash
# Install Go extension
code --install-extension golang.go
```

**GoLand**:

- Download from https://www.jetbrains.com/go/
- Import the project directory

### Step 3: Configure Git Hooks (Optional)

```bash
# Install pre-commit hooks
cp .git/hooks/pre-commit.sample .git/hooks/pre-commit

# Make executable
chmod +x .git/hooks/pre-commit
```

## 🐛 Troubleshooting

### Common Installation Issues

#### 1. Go Not Found

**Problem**: `go: command not found`

**Solution**:

```bash
# Check if Go is installed
which go

# If not found, install Go
# Follow the installation instructions above

# Verify PATH
echo $PATH | grep go
```

#### 2. Permission Denied

**Problem**: `permission denied` errors

**Solution**:

```bash
# Fix file permissions
chmod 755 data internal utils
chmod 644 internal/Transport*.csv
chmod 644 data/validators.csv

# Check ownership
ls -la data/ internal/ utils/
```

#### 3. Module Download Errors

**Problem**: `go mod download` fails

**Solution**:

```bash
# Clear module cache
go clean -modcache

# Set GOPROXY if needed
export GOPROXY=https://proxy.golang.org,direct

# Try again
go mod download
```

#### 4. Build Errors

**Problem**: `go build` fails

**Solution**:

```bash
# Check Go version
go version

# Update Go if needed
# Download latest version from golang.org

# Clean and rebuild
go clean
go mod tidy
go build cmd/main.go
```

### Platform-Specific Issues

#### Linux Issues

**Problem**: Missing dependencies

**Solution**:

```bash
# Install build essentials
sudo apt-get update
sudo apt-get install build-essential

# Install additional dependencies
sudo apt-get install git curl wget
```

#### macOS Issues

**Problem**: Xcode command line tools missing

**Solution**:

```bash
# Install Xcode command line tools
xcode-select --install

# Verify installation
xcode-select -p
```

#### Windows Issues

**Problem**: Path issues

**Solution**:

```bash
# Add Go to PATH
# Edit System Environment Variables
# Add C:\Go\bin to PATH

# Verify in PowerShell
$env:PATH -split ';' | Where-Object { $_ -like '*go*' }
```

## 📊 Performance Tuning

### System Optimization

```bash
# Increase file descriptor limits (Linux)
echo "* soft nofile 65536" | sudo tee -a /etc/security/limits.conf
echo "* hard nofile 65536" | sudo tee -a /etc/security/limits.conf

# Optimize network settings
sudo sysctl -w net.core.rmem_max=16777216
sudo sysctl -w net.core.wmem_max=16777216
```

### Go Runtime Optimization

```bash
# Set Go runtime variables
export GOMAXPROCS=4
export GOGC=100
export GOMEMLIMIT=1GiB
```

## 🔒 Security Considerations

### File Permissions

```bash
# Secure file permissions
chmod 600 internal/Transport*.csv
chmod 600 data/validators.csv
chmod 600 utils/tiltdb.csv
chmod 700 data internal utils
```

### User Isolation

```bash
# Create dedicated user
sudo useradd -r -s /bin/false tilt-validator

# Change ownership
sudo chown -R tilt-validator:tilt-validator /path/to/tilt-validator
```

## 📚 Next Steps

After successful installation:

1. **Read the Documentation**: Start with [Quick Start Guide](quick-start.md)
2. **Run Tests**: Execute `go test ./...`
3. **Explore Features**: Try different tilt types
4. **Join the Community**: Check our [Contributing Guidelines](../contributing/guidelines.md)

## 📞 Getting Help

If you encounter issues during installation:

- **Check Logs**: Look for error messages in console output
- **Review Documentation**: Check relevant docs in `docs/`
- **Search Issues**: Look for similar problems in GitHub issues
- **Ask Questions**: Use GitHub Discussions for help

---

**Need help?** Check the [troubleshooting guide](../user-guides/troubleshooting.md) or open an issue on GitHub.
