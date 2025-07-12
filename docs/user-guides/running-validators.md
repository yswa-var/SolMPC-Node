# Running Validators Guide

This guide explains how to run validators in the Tilt-Valid system, from basic single-validator operation to complex multi-validator setups.

## 🚀 Quick Start

### Single Validator (Testing)

```bash
# Start validator 1 with simple tilt
go run cmd/main.go 1 --tilt-type=simple
```

### Multiple Validators (Production)

```bash
# Run 3 validators using the provided script
./cmd/run_validators.sh
```

## 📋 Prerequisites

Before running validators, ensure you have:

1. **Go 1.22+** installed
2. **Required files** in place:

   - `data/validators.csv` - Validator configuration
   - `utils/tiltdb.csv` - Tilt data (will be created automatically)
   - Transport files in `internal/` directory

3. **Environment setup**:
   ```bash
   # Check if .env exists, create if needed
   if [ ! -f .env ]; then
     cp .env.example .env
   fi
   ```

## 🎯 Validator Modes

### 1. Simple Mode

- **Purpose**: Testing and development
- **Validators**: 1 validator
- **Tilt Type**: Simple distribution
- **Use Case**: Quick testing, debugging

```bash
go run cmd/main.go 1 --tilt-type=simple
```

### 2. Production Mode

- **Purpose**: Full distributed operation
- **Validators**: 3 validators (minimum)
- **Tilt Type**: Complex distributions
- **Use Case**: Real-world deployment

```bash
./cmd/run_validators.sh
```

### 3. Custom Mode

- **Purpose**: Custom configurations
- **Validators**: Any number
- **Tilt Type**: Any supported type
- **Use Case**: Specific requirements

```bash
# Custom validator setup
go run cmd/main.go 1 --tilt-type=two_subtilts
go run cmd/main.go 2 --tilt-type=two_subtilts
go run cmd/main.go 3 --tilt-type=two_subtilts
```

## 🔧 Command Line Options

### Basic Syntax

```bash
go run cmd/main.go <validator_id> [--tilt-type=<tilt_type>]
```

### Parameters

| Parameter      | Type    | Required | Description                                       |
| -------------- | ------- | -------- | ------------------------------------------------- |
| `validator_id` | integer | Yes      | Unique validator identifier (1, 2, 3, etc.)       |
| `--tilt-type`  | string  | No       | Type of tilt distribution (default: two_subtilts) |

### Tilt Types

| Type           | Description            | Complexity | Use Case                         |
| -------------- | ---------------------- | ---------- | -------------------------------- |
| `simple`       | Single recipient       | Low        | Testing, simple distributions    |
| `one_subtilt`  | One sub-distribution   | Medium     | Basic hierarchical distributions |
| `two_subtilts` | Two sub-distributions  | High       | Complex distributions (default)  |
| `nested`       | Multiple nested levels | Very High  | Advanced hierarchical structures |

## 🏃‍♂️ Running Methods

### Method 1: Direct Execution

```bash
# Terminal 1 - Validator 1
go run cmd/main.go 1 --tilt-type=simple

# Terminal 2 - Validator 2
go run cmd/main.go 2 --tilt-type=simple

# Terminal 3 - Validator 3
go run cmd/main.go 3 --tilt-type=simple
```

### Method 2: Using tmux Script

```bash
# Run all validators in tmux panes
./cmd/run_validators.sh
```

### Method 3: Background Processes

```bash
# Start validators in background
nohup go run cmd/main.go 1 --tilt-type=simple > validator1.log 2>&1 &
nohup go run cmd/main.go 2 --tilt-type=simple > validator2.log 2>&1 &
nohup go run cmd/main.go 3 --tilt-type=simple > validator3.log 2>&1 &

# Check if they're running
ps aux | grep main.go

# View logs
tail -f validator*.log
```

## 📊 Expected Output

### Successful Run

```
[INFO] Starting Validator ID: 1
===== Starting Validator ID: 1 =====

[INFO] Initiating DKG process...
===== Distributed Key Generation (DKG) =====
[SUCCESS] DKG completed. KeyShare length: 256
[INFO] DKG completed in 45.23 seconds

[INFO] Starting the signing process...
===== Signing Process =====
[SUCCESS] Signature generated: a1b2c3d4e5f6...
[INFO] Signing process completed.

===== VRF-based Validator Selection =====
[INFO] Generating VRF hash...
[INFO] Generated VRF hash: 1234567890abcdef...
[SUCCESS] This validator (ID: 1) was selected for verification!
[SUCCESS] ✅ Signature verification successful!
Transaction sent! Signature: 5KJvsng9m...
```

### Error Output

```
[ERROR] Failed to load validators: open data/validators.csv: no such file or directory
[ERROR] Error in loading config
[ERROR] Failed to perform DKG: timeout waiting for messages
```

## 🔍 Monitoring Validators

### 1. Log Monitoring

```bash
# Monitor all validator logs
tail -f validator*.log

# Filter by validator
grep "Validator 1" validator*.log

# Search for errors
grep "ERROR" validator*.log

# Search for success messages
grep "SUCCESS" validator*.log
```

### 2. Process Monitoring

```bash
# Check if validators are running
ps aux | grep main.go

# Check CPU and memory usage
top -p $(pgrep -f "go run cmd/main.go")

# Monitor file changes
watch -n 1 "ls -la internal/Transport*.csv"
```

### 3. Transport File Monitoring

```bash
# Monitor transport files for messages
tail -f internal/Transport*.csv

# Check file sizes
ls -lh internal/Transport*.csv

# Monitor file changes
fswatch -o internal/ | xargs -n1 -I{} echo "Transport file changed"
```

## 🛠️ Troubleshooting

### Common Issues

#### 1. "No such file or directory" Errors

**Problem**: Missing required files

```bash
[ERROR] Failed to load validators: open data/validators.csv: no such file or directory
```

**Solution**:

```bash
# Create required directories and files
mkdir -p data internal utils
echo "ID,Name,stake,active,VRFHash" > data/validators.csv
echo "1,bcvs,100.5,true,0" >> data/validators.csv
echo "2,bbdj,50.2,true,0" >> data/validators.csv
echo "3,sujskd,20.0,true,0" >> data/validators.csv
touch internal/Transport1.csv internal/Transport2.csv internal/Transport3.csv
```

#### 2. DKG Timeout

**Problem**: Validators not communicating

```bash
[ERROR] Failed to perform DKG: timeout waiting for messages
```

**Solution**:

```bash
# Check if all validators are running
ps aux | grep main.go

# Verify transport files are being written
ls -la internal/Transport*.csv

# Check file permissions
chmod 644 internal/Transport*.csv

# Restart all validators simultaneously
pkill -f "go run cmd/main.go"
./cmd/run_validators.sh
```

#### 3. Transport Errors

**Problem**: File permission or path issues

```bash
[ERROR] Error opening file: permission denied
```

**Solution**:

```bash
# Fix file permissions
chmod 644 internal/Transport*.csv
chmod 755 internal/

# Check file paths in .env
cat .env | grep TRANSPORT_PATH

# Verify paths exist
ls -la $(grep TRANSPORT_PATH .env | cut -d'=' -f2)
```

#### 4. Configuration Errors

**Problem**: Environment or config issues

```bash
[ERROR] Error in loading config
```

**Solution**:

```bash
# Check .env file
cat .env

# Verify all required variables are set
grep -E "VALIDATOR_PATH|TRANSPORT_PATH|TILT_DB" .env

# Recreate .env if needed
cp .env.example .env
```

### Advanced Troubleshooting

#### 1. Debug Mode

```bash
# Enable debug logging
export DEBUG=true
go run cmd/main.go 1 --tilt-type=simple
```

#### 2. Verbose Output

```bash
# Run with verbose output
go run cmd/main.go 1 --tilt-type=simple 2>&1 | tee validator1_debug.log
```

#### 3. Network Diagnostics

```bash
# Check file system
df -h
ls -la internal/

# Check process resources
top -p $(pgrep -f "go run cmd/main.go")

# Monitor file I/O
iotop -p $(pgrep -f "go run cmd/main.go")
```

## 🔧 Configuration

### Environment Variables

Create or edit `.env` file:

```env
# Solana Configuration
SOLANA_PRODUCT_ID=EM7AAngMgQPXizeuwAKaBvci79DhRxJMBYjRVoJWYEH3

# File Paths
VALIDATOR_PATH=/path/to/your/tilt-validator/data/
TRANSPORT_PATH=/path/to/your/tilt-validator/internal/
TILT_DB=/path/to/your/tilt-validator/utils/tiltdb.csv
DISTRIBUTION_DUMP=/path/to/your/tilt-validator/utils/distribution-dump.csv
```

### Validator Configuration

Edit `data/validators.csv`:

```csv
ID,Name,stake,active,VRFHash
1,bcvs,100.5,true,0
2,bbdj,50.2,true,0
3,sujskd,20.0,true,0
```

## 📈 Performance Optimization

### 1. Resource Allocation

```bash
# Set process priority
nice -n -10 go run cmd/main.go 1 --tilt-type=simple

# Limit CPU usage
cpulimit -l 50 -p $(pgrep -f "go run cmd/main.go")
```

### 2. File System Optimization

```bash
# Use tmpfs for transport files (Linux)
sudo mount -t tmpfs -o size=100M tmpfs /path/to/internal/

# Optimize file system
sudo tune2fs -O has_journal /dev/sda1
```

### 3. Network Optimization

```bash
# Optimize network settings
sudo sysctl -w net.core.rmem_max=16777216
sudo sysctl -w net.core.wmem_max=16777216
```

## 🔒 Security Considerations

### 1. File Permissions

```bash
# Secure file permissions
chmod 600 internal/Transport*.csv
chmod 600 data/validators.csv
chmod 600 utils/tiltdb.csv
```

### 2. Process Isolation

```bash
# Run validators in separate users
sudo useradd validator1
sudo useradd validator2
sudo useradd validator3

# Run as different users
sudo -u validator1 go run cmd/main.go 1 --tilt-type=simple
```

### 3. Network Security

```bash
# Firewall rules
sudo ufw allow from 127.0.0.1 to any port 8080
sudo ufw deny from any to any port 8080
```

## 📚 Additional Resources

- [Quick Start Guide](../getting-started/quick-start.md) - Get up and running quickly
- [Troubleshooting Guide](./troubleshooting.md) - Common issues and solutions
- [Configuration Guide](../getting-started/configuration.md) - Detailed configuration options
- [Monitoring Guide](./monitoring.md) - System monitoring and alerting

---

**Need help?** Check the [troubleshooting guide](./troubleshooting.md) or open an issue on GitHub.
