# System Architecture Overview

Tilt-Valid is a distributed validator system for Solana that implements secure, threshold-based transaction signing using Multi-Party Computation (MPC). This document provides a high-level overview of the system architecture.

## 🏗️ High-Level Architecture

```
┌─────────────────────────────────────────────────────────────────────────────┐
│                           Tilt-Valid System                                │
├─────────────────────────────────────────────────────────────────────────────┤
│                                                                             │
│  ┌─────────────┐    ┌─────────────┐    ┌─────────────┐                    │
│  │ Validator 1 │    │ Validator 2 │    │ Validator 3 │                    │
│  │             │    │             │    │             │                    │
│  │ • MPC Party │    │ • MPC Party │    │ • MPC Party │                    │
│  │ • Transport │    │ • Transport │    │ • Transport │                    │
│  │ • VRF       │    │ • VRF       │    │ • VRF       │                    │
│  │ • Logger    │    │ • Logger    │    │ • Logger    │                    │
│  └─────────────┘    └─────────────┘    └─────────────┘                    │
│         │                   │                   │                          │
│         └───────────────────┼───────────────────┘                          │
│                             │                                              │
│                    ┌─────────────────┐                                    │
│                    │  File-based     │                                    │
│                    │  Transport      │                                    │
│                    │  Layer          │                                    │
│                    └─────────────────┘                                    │
│                             │                                              │
│                    ┌─────────────────┐                                    │
│                    │  Solana         │                                    │
│                    │  Blockchain     │                                    │
│                    └─────────────────┘                                    │
│                                                                             │
└─────────────────────────────────────────────────────────────────────────────┘
```

## 🔧 Core Components

### 1. **Multi-Party Computation (MPC)**

- **Purpose**: Enables threshold-based signing without any single party having complete control
- **Implementation**: Uses TSS (Threshold Signature Scheme) library
- **Key Features**:
  - Distributed Key Generation (DKG)
  - Threshold EDDSA signing
  - No single point of failure
  - Cryptographic security guarantees

### 2. **Transport Layer**

- **Purpose**: Handles secure message exchange between validators
- **Implementation**: File-based communication using CSV files
- **Features**:
  - Asynchronous message passing
  - Broadcast and point-to-point communication
  - File watching for real-time updates
  - Thread-safe operations

### 3. **Distribution System**

- **Purpose**: Manages payment allocations and tilt structures
- **Implementation**: Hierarchical distribution algorithm
- **Features**:
  - Support for nested tilt structures
  - Business rules application
  - Recursive processing
  - Amount allocation optimization

### 4. **Verifiable Random Function (VRF)**

- **Purpose**: Provides fair, unpredictable validator selection
- **Implementation**: Cryptographic randomness generation
- **Features**:
  - Deterministic selection
  - Proof verification
  - Unpredictable but verifiable randomness

## 🔄 System Flow

### Phase 1: Initialization

1. **Validator Setup**: Each validator initializes with unique ID
2. **Configuration Loading**: Load environment and validator data
3. **Transport Initialization**: Set up file-based communication channels
4. **MPC Party Creation**: Initialize threshold signing participants

### Phase 2: Key Generation

1. **DKG Initiation**: Start distributed key generation process
2. **Message Exchange**: Validators exchange cryptographic messages
3. **Key Share Distribution**: Each validator receives a key share
4. **Public Key Derivation**: Compute shared public key

### Phase 3: Transaction Processing

1. **Tilt Data Loading**: Read distribution rules from CSV
2. **Amount Allocation**: Apply business rules to calculate payments
3. **Transaction Building**: Create Solana transaction with instructions
4. **Message Preparation**: Serialize transaction for signing

### Phase 4: Distributed Signing

1. **Threshold Signing**: Validators collaborate to sign transaction
2. **Message Coordination**: Exchange signing messages via transport
3. **Signature Generation**: Produce collective signature
4. **Verification**: Verify signature against public key

### Phase 5: Validator Selection

1. **VRF Generation**: Each validator generates random hash
2. **Hash Combination**: Combine all VRF hashes deterministically
3. **Validator Selection**: Select one validator for submission
4. **Role Assignment**: Assign verification and submission roles

### Phase 6: Transaction Submission

1. **Signature Verification**: Verify collective signature
2. **Transaction Submission**: Submit to Solana network
3. **Confirmation**: Wait for transaction confirmation
4. **Logging**: Record transaction details

## 🔒 Security Model

### Cryptographic Security

- **Threshold Cryptography**: No single validator can sign alone
- **Distributed Key Generation**: No party knows the complete private key
- **Verifiable Randomness**: Unpredictable but verifiable selection
- **Signature Verification**: Cryptographic proof of correctness

### Network Security

- **File-based Transport**: Secure message exchange
- **Message Validation**: Cryptographic message verification
- **Replay Protection**: Timestamp-based message validation
- **Error Handling**: Graceful failure recovery

### Operational Security

- **Validator Isolation**: Independent validator processes
- **Configuration Management**: Secure configuration loading
- **Logging**: Comprehensive audit trails
- **Error Recovery**: Automatic retry mechanisms

## 📊 Performance Characteristics

### Scalability

- **Horizontal Scaling**: Add more validators for increased security
- **Threshold Flexibility**: Configurable threshold requirements
- **Load Distribution**: VRF-based load balancing
- **Resource Efficiency**: Minimal computational overhead

### Reliability

- **Fault Tolerance**: System continues with threshold validators
- **Message Reliability**: File-based persistent messaging
- **Recovery Mechanisms**: Automatic retry and recovery
- **Monitoring**: Comprehensive logging and monitoring

### Latency

- **DKG Time**: ~30-60 seconds for key generation
- **Signing Time**: ~10-30 seconds for transaction signing
- **Selection Time**: ~1-5 seconds for validator selection
- **Submission Time**: ~2-10 seconds for blockchain submission

## 🎯 Design Principles

### 1. **Security First**

- No single point of failure
- Cryptographic security guarantees
- Verifiable randomness
- Threshold-based operations

### 2. **Simplicity**

- Clear component boundaries
- Minimal dependencies
- Straightforward communication patterns
- Easy to understand and debug

### 3. **Reliability**

- Fault-tolerant design
- Graceful error handling
- Comprehensive logging
- Automatic recovery mechanisms

### 4. **Extensibility**

- Modular architecture
- Plugin-friendly design
- Configurable components
- Easy to add new features

## 🔗 Integration Points

### Solana Integration

- **RPC Client**: Connection to Solana network
- **Transaction Building**: Solana instruction creation
- **Account Management**: Public key handling
- **Network Selection**: Devnet/Testnet/Mainnet support

### External Dependencies

- **TSS Library**: Threshold signature scheme implementation
- **Solana SDK**: Blockchain interaction
- **CSV Files**: Data storage and transport
- **Environment Variables**: Configuration management

## 📈 Future Enhancements

### Planned Features

- **Network Transport**: Replace file-based with network transport
- **Consensus Protocol**: Add Byzantine fault tolerance
- **Performance Optimization**: Reduce latency and improve throughput
- **Monitoring Dashboard**: Real-time system monitoring

### Scalability Improvements

- **Dynamic Validator Addition**: Add validators without restart
- **Load Balancing**: Intelligent workload distribution
- **Caching Layer**: Improve response times
- **Parallel Processing**: Concurrent transaction handling

---

**Next**: Read about [Component Details](./components.md) or [Data Flow](./data-flow.md) for more technical information.
