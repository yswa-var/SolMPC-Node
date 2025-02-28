# Solana and Anchor Setup Commands

Based on the terminal output, here's a documentation of the commands used and their results:

## Environment Check and Initial Setup
```bash
ls                          # List directory contents
source "$HOME/.cargo/env"   # Load Cargo environment variables
cargo --version             # Check Cargo version (1.85.0)
```

## Installing Anchor CLI
```bash
cargo install --git https://github.com/coral-xyz/anchor anchor-cli --locked
anchor clean                # Clean project directories
```

## Build Attempts and Troubleshooting
```bash
anchor build                # Failed due to missing build-sbf command
cargo install cargo-build-sbf            # Failed - package not found
cargo install cargo-build-sbf --git https://github.com/solana-labs/cargo-build-sbf  # Failed - authentication issues
```

## Solana Setup
```bash
solana --version            # Check Solana version (1.18.20)
sh -c "$(curl -sSfL https://release.anza.xyz/v1.18.10/install)"  # Install Solana v1.18.10
```

## Rust Configuration
```bash
rustup install stable       # Ensure stable Rust is installed
rustup component add rustfmt clippy  # Add Rust components
rustup target add bpfel-unknown-unknown  # Failed - target not supported
rustup install stable-x86_64-apple-darwin  # Install x86_64 toolchain
rustup override set stable-x86_64-apple-darwin  # Set x86_64 toolchain for project
```

## Environment Variables Setup
```bash
export SOLANA_VERSION=v1.18.10
export ANCHOR_CLI_VERSION=0.30.1
export BPF_TOOLCHAIN=$HOME/.local/share/solana/install/active_release/bin/sdk/bpf
export PATH="$BPF_TOOLCHAIN/bin:$PATH"
```

## Persisting Environment Variables
```bash
echo 'export PATH="$HOME/.local/share/solana/install/active_release/bin:$PATH"' >> ~/.zshrc
echo 'export SOLANA_VERSION=v1.18.10' >> ~/.zshrc
echo 'export ANCHOR_CLI_VERSION=0.30.1' >> ~/.zshrc
echo 'export BPF_TOOLCHAIN=$HOME/.local/share/solana/install/active_release/bin/sdk/bpf' >> ~/.zshrc
echo 'export PATH="$BPF_TOOLCHAIN/bin:$PATH"' >> ~/.zshrc
source ~/.zshrc
```

## Successful Build and Test
```bash
cargo --list | grep build-sbf  # Confirm build-sbf is available
anchor build                   # Successfully built the project
solana-test-validator         # Start local Solana test validator
```

The developer successfully set up a Solana development environment on macOS by:
1. Installing the Anchor CLI
2. Installing Solana v1.18.10
3. Configuring the x86_64 Rust toolchain
4. Setting up environment variables
5. Building the project
6. Running a local Solana test validator