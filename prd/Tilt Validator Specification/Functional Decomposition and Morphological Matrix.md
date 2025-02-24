# Functional Decomposition and Morphological Matrix

Below is a **morphological matrix** that captures **alternative solution approaches** for each **major sub-function** in the off-chain validator workflow. Each row corresponds to a **functional area** (from the decomposition), and each column suggests **possible approaches** or solution variants.

> How to read this matrix: For each sub-function (row), select one approach (column) or combine approaches as needed. This helps you systematically explore different design possibilities.
> 

---

| **Sub-Function** | **Approach A** | **Approach B** | **Approach C** | **Approach D** |
| --- | --- | --- | --- | --- |
| **1. Subset Formation** | **On-Chain VRF Selection**- On-chain randomness (e.g., via an oracle) - Deterministic selection stored on-chain | **Off-Chain Random Draw**- Off-chain RNG/lottery - Verifiable cryptographic proof | **Weighted by Stake** - Higher-stake validators have greater selection probability | **Fixed Committee** - Predefined subset - Updated rarely via governance |
| **2. Data Gathering** | **Direct RPC Calls** - Each validator queries Solana directly - Ensures consistent tilt state | **Aggregator Node** - Dedicated aggregator fetches tilt state - Subset validators trust aggregator | **Shared Cache/Database** - A common data layer - Validators pull from a replicated store | **P2P Gossip** - Validators exchange Solana data via a peer-to-peer network before finalizing |
| **3. Distribution Computation** | **Local Formula Execution** - Each validator applies tilt logic independently | **Centralized Service** - Aggregator (or a single node) calculates distribution and shares result | **Hierarchical/Recursive** - Sub-tilts resolved in a tree structure, each node computing partial sums | **Pre-Computed Tables** - Common static distributions (e.g., known splits) - Minimal computation overhead |
| **4. Reaching a Common Message** | **Direct Comparison** - Subset validators share `(recipients, amounts)` arrays and check for byte-for-byte match | **Aggregator Reconciliation** - Aggregator collects partial results, identifies discrepancies | **Leader-Based Consensus** - One validator proposes the distribution, others confirm/deny | **Hash Comparison** - Validators share a hash of the final array for quick equality checks |
| **5. Partial Signatures & Aggregation** | **Threshold BLS** - BLS-based t-of-n partial sigs - Single aggregated signature after combination | **Threshold ECDSA (TSS)** - Using MPC libraries (e.g., `tss-lib`) - Final ECDSA signature | **Ed25519 Multisig** - Each signature is stored individually, requiring k-of-n verification on-chain | **Off-Chain Collation** - Simple approach: aggregator collects each signature, merges them in a specialized format |
| **6. Broadcasting the Final Distribution** | **Aggregator as Broadcaster** - The aggregator pays fees and submits the transaction | **Any Subset Member** - Whichever validator obtains aggregated sig first submits it | **External Relayer** - A dedicated service or “relayer” that handles all Solana transactions | **Multiple Parallel Broadcasts** - Several nodes simultaneously submit; first valid one finalizes |
| **7. Handling Mistakes & Mid-Process Updates** | **Version Check + Re-Run** - If tilt logic changes mid-process, discard partial sigs and recompute | **Fallback Aggregator** - A standby node can broadcast if the main aggregator fails | **Timeout & Re-Randomize** - If partial signatures not collected by deadline, pick new subset | **On-Chain Escalation** - If repeated failures occur, revert to an on-chain governance-based approach |

---

### How to Use This Matrix

1. **Identify Requirements**
    - For example, if your protocol requires maximum decentralization, you might lean toward a **P2P Gossip** approach for Data Gathering and a **Threshold BLS** signature scheme.
2. **Select One Approach per Row**
    - Each sub-function (row) needs at least one chosen approach (column). If you want a hybrid, combine or adapt them (e.g., direct RPC calls + shared cache).
3. **Implement & Integrate**
    - Once your design choices are made, map them back to the code modules or packages you’ll develop (e.g., “Threshold BLS” → integrate a BLS library in the partial-signature step).

Using a **morphological matrix** in this way ensures you systematically consider multiple ways of fulfilling each functional requirement in your off-chain validator flow.