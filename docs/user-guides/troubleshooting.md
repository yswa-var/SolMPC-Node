# Troubleshooting Guide

This guide helps you resolve common issues when running Tilt-Valid.

## 🚨 Common Issues

### 1. File Not Found Errors

#### Problem

```
[ERROR] Failed to load validators: open data/validators.csv: no such file or directory
[ERROR] Failed to open file: permission denied
```

#### Solution

```bash
# Create required directories and files
mkdir -p data internal utils

# Create validator data
cat > data/validators.csv << EOF
ID,Name,stake,active,VRFHash
1,bcvs,100.5,true,0
2,bbdj,50.2,true,0
3,sujskd,20.0,true,0
EOF

# Create transport files
touch internal/Transport1.csv internal/Transport2.csv internal/Transport3.csv

# Fix permissions
chmod 644 data/validators.csv internal/Transport*.csv
chmod 755 data internal utils
```

### 2. DKG Timeout

#### Problem

```
[ERROR] Failed to perform DKG: timeout waiting for messages
[ERROR] DKG timed out: context deadline exceeded
```

#### Solution

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

### 3. Configuration Errors

#### Problem

```
[ERROR] Error in loading config
[ERROR] Invalid configuration: missing required fields
```

#### Solution

```bash
# Check .env file exists
ls -la .env

# Create .env if missing
cp .env.example .env

# Verify required variables
grep -E "VALIDATOR_PATH|TRANSPORT_PATH|TILT_DB" .env

# Check file paths exist
ls -la $(grep VALIDATOR_PATH .env | cut -d'=' -f2)
ls -la $(grep TRANSPORT_PATH .env | cut -d'=' -f2)
```

### 4. Transport Errors

#### Problem

```
[ERROR] Error opening file: permission denied
[ERROR] Failed to write message: no space left on device
```

#### Solution

```bash
# Fix file permissions
chmod 644 internal/Transport*.csv
chmod 755 internal/

# Check disk space
df -h

# Clear old transport files
rm -f internal/Transport*.csv
touch internal/Transport1.csv internal/Transport2.csv internal/Transport3.csv
```

### 5. Go Module Issues

#### Problem

```
[ERROR] go: module not found
[ERROR] go: cannot find module providing package
```

#### Solution

```bash
# Clean module cache
go clean -modcache

# Re-download dependencies
go mod download

# Update go.mod
go mod tidy

# Verify dependencies
go mod verify
```

## 🔍 Debugging Techniques

### 1. Enable Debug Logging

```bash
# Set debug environment variable
export DEBUG=true

# Run with debug output
go run cmd/main.go 1 --tilt-type=simple 2>&1 | tee debug.log
```

### 2. Verbose Output

```bash
# Run with verbose output
go run cmd/main.go 1 --tilt-type=simple -v

# Monitor file changes
watch -n 1 "ls -la internal/Transport*.csv"
```

### 3. Process Monitoring

```bash
# Check running processes
ps aux | grep main.go

# Monitor CPU and memory
top -p $(pgrep -f "go run cmd/main.go")

# Check file descriptors
lsof -p $(pgrep -f "go run cmd/main.go")
```

### 4. Network Diagnostics

```bash
# Check file system
df -h
ls -la internal/

# Monitor file I/O
iotop -p $(pgrep -f "go run cmd/main.go")

# Check network connections
netstat -tulpn | grep go
```

## 🛠️ Advanced Troubleshooting

### 1. Memory Issues

#### Problem

```
[ERROR] Out of memory
[ERROR] Cannot allocate memory
```

#### Solution

```bash
# Check memory usage
free -h

# Increase swap space (Linux)
sudo fallocate -l 2G /swapfile
sudo chmod 600 /swapfile
sudo mkswap /swapfile
sudo swapon /swapfile

# Set Go memory limits
export GOMEMLIMIT=1GiB
export GOGC=100
```

### 2. Performance Issues

#### Problem

```
[WARNING] Slow performance detected
[ERROR] Operation timed out
```

#### Solution

```bash
# Optimize system settings
sudo sysctl -w net.core.rmem_max=16777216
sudo sysctl -w net.core.wmem_max=16777216

# Set process priority
nice -n -10 go run cmd/main.go 1 --tilt-type=simple

# Use tmpfs for transport files (Linux)
sudo mount -t tmpfs -o size=100M tmpfs /path/to/internal/
```

### 3. Security Issues

#### Problem

```
[ERROR] Permission denied
[ERROR] Access denied
```

#### Solution

```bash
# Fix file permissions
chmod 600 internal/Transport*.csv
chmod 600 data/validators.csv
chmod 600 utils/tiltdb.csv

# Create dedicated user
sudo useradd -r -s /bin/false tilt-validator
sudo chown -R tilt-validator:tilt-validator /path/to/tilt-validator

# Run as dedicated user
sudo -u tilt-validator go run cmd/main.go 1 --tilt-type=simple
```

## 📊 Diagnostic Commands

### 1. System Information

```bash
# Check system resources
top
htop
iostat

# Check disk usage
df -h
du -sh *

# Check memory
free -h
cat /proc/meminfo
```

### 2. Process Information

```bash
# Check running processes
ps aux | grep main.go

# Check process tree
pstree -p $(pgrep -f "go run cmd/main.go")

# Check open files
lsof -p $(pgrep -f "go run cmd/main.go")
```

### 3. File System

```bash
# Check file permissions
ls -la data/ internal/ utils/

# Check file contents
head -10 data/validators.csv
head -10 internal/Transport1.csv

# Monitor file changes
inotifywait -m internal/ -e modify,create,delete
```

### 4. Network

```bash
# Check network connections
netstat -tulpn | grep go

# Check DNS resolution
nslookup api.devnet.solana.com

# Test connectivity
curl -I https://api.devnet.solana.com
```

## 🔧 Recovery Procedures

### 1. Complete Reset

```bash
# Stop all validators
pkill -f "go run cmd/main.go"

# Clear all data
rm -rf data/* internal/Transport*.csv utils/tiltdb.csv

# Recreate files
mkdir -p data internal utils
cat > data/validators.csv << EOF
ID,Name,stake,active,VRFHash
1,bcvs,100.5,true,0
2,bbdj,50.2,true,0
3,sujskd,20.0,true,0
EOF
touch internal/Transport1.csv internal/Transport2.csv internal/Transport3.csv

# Restart validators
./cmd/run_validators.sh
```

### 2. Partial Reset

```bash
# Clear only transport files
rm -f internal/Transport*.csv
touch internal/Transport1.csv internal/Transport2.csv internal/Transport3.csv

# Restart validators
./cmd/run_validators.sh
```

### 3. Configuration Reset

```bash
# Backup current config
cp .env .env.backup

# Create new config
cp .env.example .env

# Edit new config
nano .env

# Restart validators
./cmd/run_validators.sh
```

## 📞 Getting Help

### 1. Collect Information

Before asking for help, collect this information:

```bash
# System information
uname -a
go version
df -h
free -h

# Process information
ps aux | grep main.go

# Log files
tail -50 validator*.log

# Configuration
cat .env
ls -la data/ internal/ utils/
```

### 2. Search Existing Issues

- Check GitHub Issues for similar problems
- Search the documentation for solutions
- Look for known issues in the release notes

### 3. Create Bug Report

When creating a bug report, include:

- **Error Message**: Exact error text
- **Steps to Reproduce**: Detailed steps
- **Environment**: OS, Go version, system specs
- **Logs**: Relevant log output
- **Configuration**: Relevant config files

### 4. Contact Support

- **GitHub Issues**: For bugs and feature requests
- **GitHub Discussions**: For questions and help
- **Documentation**: Check the [docs](../) directory first

## 🎯 Prevention Tips

### 1. Regular Maintenance

```bash
# Clean old files regularly
find . -name "*.log" -mtime +7 -delete
find . -name "Transport*.csv" -size +10M -delete

# Monitor disk space
df -h | grep -E "Use%|/dev"

# Check process health
ps aux | grep main.go | grep -v grep
```

### 2. Monitoring

```bash
# Set up monitoring script
cat > monitor.sh << 'EOF'
#!/bin/bash
while true; do
    echo "$(date): Checking validators..."
    ps aux | grep main.go | grep -v grep || echo "No validators running"
    ls -la internal/Transport*.csv
    sleep 30
done
EOF
chmod +x monitor.sh
./monitor.sh &
```

### 3. Backup

```bash
# Backup important files
cp data/validators.csv data/validators.csv.backup
cp .env .env.backup

# Create backup script
cat > backup.sh << 'EOF'
#!/bin/bash
DATE=$(date +%Y%m%d_%H%M%S)
tar -czf backup_$DATE.tar.gz data/ .env utils/
echo "Backup created: backup_$DATE.tar.gz"
EOF
chmod +x backup.sh
```

## 📚 Additional Resources

- [Quick Start Guide](../getting-started/quick-start.md) - Basic setup
- [Installation Guide](../getting-started/installation.md) - Detailed installation
- [Running Validators](./running-validators.md) - How to run the system
- [Development Setup](../development/setup.md) - Development environment
- [Contributing Guidelines](../contributing/guidelines.md) - How to contribute

---

**Still having issues?** Open an issue on GitHub with the information collected above.
