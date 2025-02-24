# Validator Sequence Diagram

[GO Lang & Libraries for Validator](Validator Sequence Diagram/GO Lang & Libraries for Validator )

[Tilt Calculation Example](Validator Sequence Diagram/Tilt Calculation Example )

sequenceDiagram
participant Curator (Off-Chain)
participant SubsetValidator1 (Off-Chain)
participant SubsetValidator2 (Off-Chain)
participant SubsetValidator3 (Off-Chain)
participant Aggregator (Off-Chain)
participant TiltProgram (On-Chain)
participant SolanaRPC (On-Chain)

```
note over Curator: (1) Possibly updates tilt's on-chain business rules or ephemeral weighting
Curator ->> TiltProgram: updateBusinessRules(...) (optional)
TiltProgram ->> TiltProgram: store updated rules in tilt's PDA

note over SubsetValidator1, SubsetValidator2, SubsetValidator3: (2) Subset of K=3 chosen via random or VRF approach

note over SubsetValidator1, SolanaRPC: (3) Data Gathering
SubsetValidator1 ->> SolanaRPC: getAccountInfo(tiltPDA)
SubsetValidator2 ->> SolanaRPC: getAccountInfo(tiltPDA)
SubsetValidator3 ->> SolanaRPC: getAccountInfo(tiltPDA)
SolanaRPC ->> SubsetValidator1: returns tilt's balance, rules, etc.
SolanaRPC ->> SubsetValidator2: returns tilt's balance, rules, etc.
SolanaRPC ->> SubsetValidator3: returns tilt's balance, rules, etc.

note over SubsetValidator1, SubsetValidator2, SubsetValidator3: (4) Compute final distribution using tilt's state

SubsetValidator1 ->> SubsetValidator1: compute (recipients[], amounts[])
SubsetValidator2 ->> SubsetValidator2: compute (recipients[], amounts[])
SubsetValidator3 ->> SubsetValidator3: compute (recipients[], amounts[])

note over SubsetValidator1, SubsetValidator2, SubsetValidator3: (5) compare results, ensure all match

SubsetValidator1 ->> Aggregator: distribution array match?
SubsetValidator2 ->> Aggregator: distribution array match?
SubsetValidator3 ->> Aggregator: distribution array match?

note over SubsetValidator1, SubsetValidator2, SubsetValidator3: (6) partial signatures on final message

SubsetValidator1 ->> Aggregator: partial_signature_1
SubsetValidator2 ->> Aggregator: partial_signature_2
SubsetValidator3 ->> Aggregator: partial_signature_3

note over Aggregator: (7) aggregator merges partial signatures into aggregatedSig

Aggregator ->> Aggregator: produce aggregatedSig from partial_signatures

note over Aggregator, TiltProgram: (8) aggregator broadcasts final distribution

Aggregator ->> TiltProgram: executeDistribution(tiltID, recipients[], amounts[], aggregatedSig)
TiltProgram ->> TiltProgram: verify threshold signature
TiltProgram ->> TiltProgram: check tiltBalance, sub-tilt logic if needed
alt signature invalid or mismatch
    TiltProgram ->> Aggregator: revert transaction
else signature valid
    TiltProgram ->> TiltProgram: transfer lamports to recipients
    TiltProgram ->> Aggregator: success, DistributionExecuted event
end

```

Below is a **step-by-step outline** of how a **group of validators** collaborates **off-chain** to **generate a payment distribution message** for a Solana tilt program. The focus is on the **computation** they perform to reach a final `(recipients, amounts)` array (plus any cryptographic signatures) that they will then submit on-chain.

---

## 1. Subset Formation & Data Gathering

1. **Subset Selection**
    - A random or protocol-defined mechanism (e.g., VRF-based) chooses a subset of K validators (out of N total) to finalize this tilt’s distribution.
    - If using ephemeral or threshold keys, the subset has or generates a group public key via a mini-DKG (Distributed Key Generation).
2. **Fetch On-Chain Tilt Data**
    - Each validator in the subset fetches **current tilt state** from Solana’s RPC, typically:
        - **Tilt’s deposit balance** (`tiltBalance`),
        - **Any sub-tilt references** or “business rules” in the tilt’s PDA,
        - **Time-based or status flags** (is the tilt active, partial, or closed?),
        - **KYC/Compliance flags** if relevant.
    - They confirm the version of the tilt’s “logic” (or ephemeral weighting) to ensure a consistent snapshot.
3. **Potential Off-Chain Inputs** (Optional)
    - If the tilt’s logic references external data (e.g., aggregator stats, curation scores, or external oracles), the subset queries those off-chain services as well.

---

## 2. Distribution Computation Logic

1. **Apply the Tilt’s Formula**
    - Using the tilt’s on-chain “business rules” or ephemeral weighting, each validator *independently* computes how to split `tiltBalance` among the set of potential recipients.
    - Example:
        1. 20% to protocol, 40% to curators, 40% to publishers.
        2. Within curators, sub-tilt references might get 75% of the curator share, local curators get 25%.
        3. Possibly break sub-tilts into multiple recipients, etc.
2. **Aggregate or Merge Sub-Tilts**
    - If sub-tilts also have distribution formulas, each validator must handle that (e.g., if sub-tilt #1 references other sub-tilts or has nested rules).
    - They end up with a single final list of `(recipient, amount)` pairs, ensuring no double-counting.
    - Sums are carefully checked to avoid exceeding `(tiltBalance - rentExemptReserve)`.
3. **Check Special Conditions**
    - If any recipients have `kycVerified = false`, that address might be skipped or flagged.
    - If there is a minimum deposit threshold or a time-based condition (like “only distribute after X slot”), each validator ensures it’s satisfied.
4. **Perform Rounding & Summation**
    - Convert percentages or basis points to integer lamports.
    - Handle leftover lamports due to rounding if necessary, often awarding them to the last recipient or the protocol, depending on design.

---

## 3. Reaching a Common Message

1. **Form the Distribution Message**
    - Once each validator completes local computation, they produce an identical final array, e.g.:
        
        ```json
        {
          "tiltID": 42,
          "recipients": [address1, address2, ...],
          "amounts": [100_000_000, 200_000_000, ...],
          "nonce": 123456,
          "version": tiltData.version
        }
        
        ```
        
    - The `nonce` or timestamp prevents replay. The `version` references the tilt’s current logic or ephemeral weighting snapshot.
2. **Consensus on the Final Array**
    - Subset validators compare results. If all match exactly, they proceed.
    - If any discrepancy arises, they identify which node’s computation is off or re-check data.
    - The aggregator node (or a separate “leader”) merges them if there are minor differences, or triggers a re-run if something is drastically inconsistent.

---

## 4. Partial Signatures & Aggregation

1. **Partial Signatures**
    - Each validator uses its threshold key share (BLS, threshold ECDSA, etc.) to **sign** the final `(tiltID, recipients[], amounts[], nonce, version, ...).`
    - They might produce a partial signature object: `{ validatorIndex, partialSigBytes }`.
2. **Validate Partial Sigs (Off-Chain)**
    - Each validator (or aggregator) checks that each partial signature is indeed correct for the distribution message.
    - If a partial signature fails, that node’s share is considered invalid or malicious.
3. **Aggregate**
    - Once at least **M** partial signatures are valid, the aggregator merges them into one short **aggregatedSig**.
    - They might store a list of signers or keep it for optional slash logic.

---

## 5. Broadcast to Solana

1. **Designated Broadcaster**
    - The subset chooses who broadcasts. That node includes in the Solana transaction:
        - `(tiltID, recipients[], amounts[], aggregatedSig, possibly version or tilt state hash)`.
    - This node pays the transaction fee.
2. **On-Chain Execution**
    - The tilt program verifies `aggregatedSig` with the ephemeral group public key or a known committee pubkey.
    - Checks basic constraints (balance, tilt status, etc.). If everything is valid, it executes lamport transfers from the tilt’s PDA to each recipient.
3. **Finalization**
    - The tilt program emits a `DistributionExecuted(tiltID, aggregatorAddress, totalPayout)` event.
    - Subset validators and watchers confirm the transaction was accepted without tampering.

---

## 6. Handling Mistakes & Mid-Process Updates

1. **If the Curator Updates Weights** (While partial sigs are collecting)
    - The aggregator or other validators detect a mismatch in `version` or tilt logic.
    - They discard old partial signatures, re-run the distribution logic, and produce a new final array & signature.
2. **If the Aggregator Fails to Broadcast**
    - Another node can broadcast the same aggregatedSig.
    - The tilt program does not care who the sender is, only that the signature is valid for that tilt distribution message.
3. **Edge Cases**
    - If sub-tilt references cause massive arrays, they might compress or chunk them.
    - If insufficient partial signatures appear (some validators offline), the system might define a re-randomization after a timeout.

---

### Summary

This **sequence** details how **a group of validators**:

1. **Collects tilt data** (the deposit balance, sub-tilt rules, etc.).
2. **Computes** the final distribution array.
3. **Ensures** everyone arrives at the same `(recipients, amounts)` message.
4. **Generates partial signatures**.
5. **Merges** them into a single aggregated signature.
6. **Broadcasts** that distribution to the Solana tilt program, which verifies the signature and executes the payout.

The crucial aspects for **computation** revolve around reading consistent state from the tilt’s account, applying the correct formula to handle sub-tilts or dynamic logic, and ensuring partial signatures reflect an **identical** final result.

Below is a **high-level specification** for implementing the validator logic in **Go**. It includes the major steps (subset selection, data gathering, computing distribution, partial signatures, and final aggregation) along with suggested **Go libraries** you might use for each component. This is just a **starting point**; you can adapt it to match your infrastructure, threshold-signature scheme, and naming conventions.

---

## 1. Overview

**Objective**: Implement a **validator node** that participates in the Tilt payment distribution flow. Each validator node:

1. Joins a selected **subset** of validators.
2. Gathers on-chain state from Solana.
3. Computes the final `(recipients[], amounts[])` distribution.
4. Generates a **partial signature** of the final message.
5. Sends partial signatures to an **aggregator** node to be merged into a single aggregated signature.

Finally, an **aggregator** broadcasts the signed transaction to the **Tilt program** on Solana.

---

## 2. Core Components and Recommended Libraries

1. **Solana RPC & On-Chain Data**
    - **Library**: [github.com/gagliardetto/solana-go](https://github.com/gagliardetto/solana-go) or [github.com/portto/solana-go-sdk](https://github.com/portto/solana-go-sdk)
    - **Usage**:
        - Establish an RPC client to connect to a Solana cluster (Mainnet, Testnet, Devnet, or localnet).
        - Fetch `AccountInfo` for the Tilt Program Derived Address (PDA), retrieving data such as:
            - Tilt’s balance
            - Business rules or ephemeral weighting
            - Sub-tilt references, etc.
2. **Threshold Signatures & Cryptography**
    - **Library (Possible Options)**:
        - [github.com/binance-chain/tss-lib](https://github.com/binance-chain/tss-lib) – for threshold ECDSA (commonly used for BTC/ETH, but can be adapted if you set up ECDSA-based solutions)
        - [crypto/ed25519](https://pkg.go.dev/crypto/ed25519) – built-in Go library for basic Ed25519 signing. *(Note: Go’s standard library does **not** provide a fully built-in threshold Ed25519; you may need third-party libs or multi-signature aggregator logic.)*
        - [golang.org/x/crypto](https://pkg.go.dev/golang.org/x/crypto) – extended cryptography packages, can be used for hashing, key derivation, etc.
    - **Usage**:
        - Each validator needs to hold a private key (or partial secret in a threshold scheme).
        - For threshold signatures (t-of-n), either use an **MPC-based** solution or a simpler approach like **multisig** on Solana (though that’s a different pattern than aggregated threshold sigs).
        - In an example aggregator flow, each validator signs the final distribution message with its private key, then sends the partial signature to the aggregator. The aggregator merges these signatures into a single valid signature if a threshold is met.
3. **Networking & Communication**
    - **Library**:
        - [github.com/libp2p/go-libp2p](https://github.com/libp2p/go-libp2p) – if you prefer a P2P network for validators to share data or partial signatures.
        - Or use standard **HTTP**/REST or **gRPC** with [google.golang.org/grpc](https://google.golang.org/grpc) to communicate partial signatures to the aggregator.
    - **Usage**:
        - Subset validators must exchange or send partial signatures to the aggregator.
        - The aggregator must broadcast the final transaction to the Solana cluster.
4. **Configuration & Logging**
    - **Library**:
        - [github.com/spf13/viper](https://github.com/spf13/viper) for managing configs.
        - [github.com/sirupsen/logrus](https://github.com/sirupsen/logrus) or [github.com/rs/zerolog](https://github.com/rs/zerolog) for structured logging.
    - **Usage**:
        - Store cluster endpoint, tilt program ID, aggregator addresses, threshold parameters, etc.
        - Uniform logging across all validator nodes and aggregator.

---

## 3. Data Structures

Below is a sample of the minimal data structures you might define in Go.

```go
package tilt

// Info about the tilt's state fetched from on-chain
type TiltState struct {
    Balance      uint64
    BusinessRules []byte
    // Additional fields as needed, e.g. ephemeral weighting, sub-tilt references
}

// The final distribution result that each validator computes
type Distribution struct {
    Recipients []string  // or use solana.PublicKey
    Amounts    []uint64
}

// Partial signature and metadata to correlate it with the final distribution
type PartialSignature struct {
    ValidatorID  string
    Signature    []byte
    Distribution Distribution
}

// Aggregated signature result
type AggregatedSignature struct {
    Signature    []byte  // single aggregated signature
    Distribution Distribution
}

```

---

## 4. Validator Node Flow

1. **Startup & Initialization**
    - Load config (cluster endpoint, tilt program ID, local private key).
    - Initialize Solana RPC client (e.g., `solana-go.NewClient(...)`).
    - Optionally join a P2P or register with an aggregator for partial signature submissions.
2. **(Optional) Subset Selection**
    - If the node is selected to be part of the k-of-n subset, it enters the distribution cycle.
    - Subset might be determined on-chain (e.g., VRF) or by an off-chain random selection process.
    - **Implementation detail**: You might store a local flag `IsSelected = true` if chosen.
3. **Data Gathering from Solana**
    
    ```go
    tiltAccountInfo, err := solanaClient.GetAccountInfo(context.Background(), tiltPDA)
    if err != nil {
        // handle error
    }
    tiltState := parseTiltState(tiltAccountInfo.Data)
    
    ```
    
    - The `parseTiltState(...)` function reads the program’s custom data format.
    - Retrieve `Balance`, `BusinessRules`, `DistributionHistory`, etc.
4. **Compute Distribution**
    - Using the tilt’s business rules, each validator arrives at the same `(recipients[], amounts[])`.
    - Example pseudocode:
        
        ```go
        func computeDistribution(state *TiltState) Distribution {
            // 1. Evaluate ephemeral weighting or rules
            // 2. Derive recipients (sub-tilts, creators, protocol fee, etc.)
            // 3. Return distribution
            return Distribution{ ... }
        }
        
        ```
        
    - Make sure all validators produce **identical** results.
5. **Partial Signature Generation**
    - Construct a **message** to sign that includes:
        - `tiltID`
        - `recipients[]`
        - `amounts[]`
        - Possibly a **nonce** or block hash to avoid replay
    - Sign with Ed25519 or ECDSA-based threshold scheme:
        
        ```go
        messageBytes := buildDistributionMessage(tiltID, distribution)
        sig := ed25519.Sign(privateKey, messageBytes)
        partialSig := PartialSignature{
            ValidatorID:  validatorID,
            Signature:    sig,
            Distribution: distribution,
        }
        
        ```
        
    - Transmit `partialSig` to the aggregator (via gRPC, P2P, or HTTP).
6. **Verification & Consensus**
    - Each validator or aggregator node can cross-check that all partial signatures are signing the **same** distribution array.
    - If any mismatch is detected, the aggregator might request a re-computation or raise an alert.
7. **Aggregator Receives Partial Signatures**
    - If using a **threshold signature library** (e.g., TSS, BLS, etc.), the aggregator will combine partial signatures into a single signature.
    - If using **multiple signatures** (like a Solana multisig approach), it might just collect them in a single transaction structure.
8. **Aggregator Broadcast**
    - The aggregator forms and broadcasts the final transaction:
        
        ```go
        tx := buildExecuteDistributionTransaction(
            tiltID,
            distribution.Recipients,
            distribution.Amounts,
            aggregatedSignature,
        )
        solanaClient.SendTransaction(context.Background(), tx)
        
        ```
        
    - The Tilt program verifies the aggregated signature on-chain and executes the distribution if valid.

---

## 5. Example Folder Structure

```
tilt/
 ┣ cmd/
 ┃ ┗ tilt-validator/
 ┃    ┣ main.go                 // CLI entry point for validator node
 ┣ pkg/
 ┃ ┣ aggregator/
 ┃ ┃ ┗ aggregator.go            // aggregator logic
 ┃ ┣ tss/
 ┃ ┃ ┗ partial_signature.go     // threshold signature logic
 ┃ ┣ solana/
 ┃ ┃ ┗ client.go                // wrapper around solana-go
 ┃ ┗ tiltlogic/
 ┃    ┣ distribution.go         // computeDistribution(...)
 ┃    ┗ state.go                // parseTiltState(...), data structs
 ┗ go.mod

```

---

## 6. Security & Best Practices

- **Key Management**: Protect validator private keys (potentially using HSM or a secure keystore).
- **Replay Protection**: Incorporate blockhash or a unique ID in the signed message.
- **Threshold**: Decide on `k-of-n` (e.g., 3-of-5) in your TSS config and ensure partial key shares are never exposed.
- **Data Consistency**: Double-check all validators have identical `TiltState` data before computing final distribution.

---

## 7. Extending the Spec

1. **Protocol Fee**: If 20% goes to the protocol, incorporate logic in `computeDistribution(...)` that calculates this portion for the foundation wallet or DAO.
2. **Re-tries & Timeouts**: If aggregator does not receive enough partial signatures within a certain window, it can alert or trigger a fallback aggregator.
3. **Governance Integration**: If on-chain governance modifies the rules, have your validator automatically fetch new logic from the tilt program’s account.
4. **Subset Selection**: If you want an on-chain VRF approach, integrate with a VRF oracle or a Dapp that commits the randomly selected validator set to Solana.

---

## Conclusion

This **Go-based validator specification** provides a blueprint for:

1. **Interacting with Solana** for on-chain data (using `solana-go`).
2. **Computing the distribution** deterministically.
3. **Generating partial signatures** (using threshold signature libraries or standard Ed25519 if you adopt a simpler aggregator approach).
4. **Aggregating** partial signatures and broadcasting the final transaction on-chain.

With the above structure and libraries, you can build out your own validator nodes, aggregator service, and robust distribution workflow for **Tilt**. Feel free to adapt naming conventions, data structures, and threshold signing logic to your specific requirements and security constraints.