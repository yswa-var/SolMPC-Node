We aim to build **voting as a service**, a fully decentralized, threshold-secure transaction validation and signing system for the Solana blockchain, powered by Multi-Party Computation (MPC) and Verifiable Random Functions (VRF). Our goal is to eliminate single points of failure and trust, ensuring that no individual node ever controls the signing key. Instead, a configurable threshold of validators—chosen fairly at random via VRF—collaborates to generate each on-chain signature. Alongside core signing, to offer this infrastructure as a turnkey Voting-as-a-Service

***we might need to remove the sub tilt system as we are intending to use this only for Voting-as-a-Service
infra ***
**What We Intend to Make**

1. **Threshold MPC Signing Network**

   * A network of `n` validator nodes, of which any `t` can jointly produce a valid Ed25519 signature without exposing their individual key shares.
   * A VRF-based selection mechanism that randomly and verifiably chooses which subset of validators will handle each transaction or vote tally.

3. **User-Facing Service Layer**

   * REST and WebSocket APIs for client apps or a browser-based front end, enabling wallet-based or OAuth user authentication, ballot submission, and real-time status updates.
4. **Election/Voting-as-a-Service**

   * A turnkey offering where any organization (DAO, corporation, public body) can spin up a secure, anonymous ballot with provable tallying, audit logs, and result anchoring on Solana.

**What We Have Ready**

* **Core Cryptographic Primitives**

  * A custom MPC implementation supporting t-of-n threshold signing with working DKG.
  * A Distributed Key Generation (DKG) protocol that produces shared public keys and individual secret shares without any single party knowing the full key.
  * ⚠️ **Gap**: MPC signatures are generated but not used for actual Solana transactions (still uses regular wallet signing).

* **Pseudo-VRF Selection** 

  * A basic validator selection mechanism using timestamp-based hashing.
  * ⚠️ **Gap**: Current implementation is NOT true VRF (just SHA256 of timestamp), though proper VRF implementations exist in codebase but are unused.

* **Solana Integration**

  * Transaction building and submission pipelines using Solana Go SDK.
  * ⚠️ **Gap**: Transactions are signed with regular Ed25519 keys, NOT MPC-generated signatures, defeating the purpose of distributed signing.

* **Legacy Tilt System**

  * Full implementation of payment distribution patterns (simple, one\_subtilt, two\_subtilts, nested).
  * ⚠️ **Gap**: This contradicts the voting-as-a-service pivot and should be removed/replaced with voting logic.

* **Prototype Transport Layer**

  * A basic file-based message exchange using CSV polling for validator coordination.
  * ⚠️ **Gap**: Extremely primitive for production use (1ms file polling), needs replacement with proper networking.

---

## 🔧 **Priority Fixes Needed**

**Critical (Breaks Core Functionality):**
1. **Fix MPC-Solana Integration**: Make Solana transactions use MPC signatures instead of regular wallet signatures
2. **Implement Real VRF**: Replace timestamp-based pseudo-randomness with proper VRF (code exists but unused)
3. **Remove Tilt System**: Replace payment distribution logic with voting/ballot logic

**High Priority (Production Readiness):**
4. **Upgrade Transport Layer**: Replace CSV file polling with proper P2P networking (libp2p/gRPC)
5. **Add Vote Tallying Logic**: Core voting functionality missing
6. **Implement Ballot Management**: Create, manage, and close voting sessions

**Medium Priority (Service Layer):**
7. **Add REST/WebSocket APIs**: User-facing service endpoints
8. **Implement Authentication**: Wallet-based or OAuth user auth
9. **Add Audit Logging**: Track all votes and validator actions
