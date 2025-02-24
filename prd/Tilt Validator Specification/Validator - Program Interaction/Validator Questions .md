# Validator Questions

---

## 1. Roles & Governance

1. **How are validators added or removed?**
    - Is there a dedicated instruction (`addValidator`, `removeValidator`)?
    - Which role(s) can perform this update, and how is it authenticated?
2. **What is the process for rotating the `Program Admin` key if it’s compromised?**
    - Do you allow the admin key to be changed (or is it immutable once set)?
    - What additional controls (e.g., multi-signature) protect this update?
3. **Is there a scenario where a tilt can be frozen or halted by someone other than the Admin or Compliance role?**
    - Under what exact conditions is a tilt frozen?
    - How do you manage the unfreeze process?
4. **How do we handle *emergency upgrades* or *emergency stops*?**
    - Is there a global “pause” or “circuit breaker” that halts critical instructions like distribution?

---

## 2. Validator Logic & Security

1. **Single vs. Multiple Validators**
    - **Is a single validator enough** for your use case? Or do you anticipate multiple validators providing separate calculations and needing a threshold?
    - How is the trust model communicated to end-users?
2. **Validator Key Management**
    - If using a **single validator** approach, **how is that key secured** (e.g., hardware wallet, multi-sig wallet)?
    - If using **multiple validators** with threshold approvals, **how do you handle** inconsistent proposals (two validators propose different distributions)?
3. **Validator Data Sources**
    - What off-chain data does each validator rely on to compute a final distribution (e.g., deposit logs, user contributions, external events)?
    - Do multiple validators share the same data aggregator or do they fetch from different sources?
4. **Validator Incentives**
    - Are validators incentivized or paid for their service?
    - What happens if a validator goes offline or fails to submit a distribution?
5. **Conflict Resolution**
    - In a multi-validator setup, if proposals differ, is there an on-chain process to reconcile them, or is that entirely off-chain?
    - Are partial approvals or multiple proposals at the same time allowed?

---

## 3. On-Chain Storage & Data Structures

1. **Distribution Proposals**
    - If using a multi-step approach, **where and how** are proposals stored?
    - Is there a maximum number of pending proposals or distributions per tilt?
2. **Handling Large Recipient Lists**
    - If a tilt has hundreds (or thousands) of recipients, does storing them in a single instruction or single on-chain structure cause **transaction size or account size** issues?
3. **Upgradability vs. Data Compatibility**
    - If the program is upgradeable, how do you handle changes to the tilt account data structure without breaking existing accounts?
    - Is there a migration path?
4. **Off-Chain Indexing**
    - Do you rely on off-chain indexers (e.g., Graph-like solutions or custom scripts) to reconstruct deposit histories?
    - Is there a fallback if those indexers go down or become inconsistent?

---

## 4. Instruction Execution & Edge Cases

1. **Partially Filled Distribution**
    - If the sum of payouts exceeds the tilt’s current balance (maybe due to a deposit failing?), **does the program revert** or can it distribute partially?
    - How do you handle scenario where recipients have changed or someone is no longer eligible at the time of distribution?
2. **Race Conditions**
    - Could two valid transactions attempt to distribute the same tilt funds simultaneously?
    - Do you need concurrency control (like setting the tilt to a “pending distribution” state until the transaction finishes)?
3. **Compliance & KYC**
    - If the distribution requires KYC checks, **is it done for every recipient** or just the major beneficiary?
    - How do you store/verify compliance flags on-chain (versus an off-chain list)?
4. **Transaction Fee Payment**
    - Who pays the transaction fees for distribution calls?
    - Do you allow “meta-transactions” or a sponsor to cover fees for others?

---

## 5. Testing & Audit Considerations

1. **Unit Tests**
    - Have you covered all instructions (createTilt, deposit, executeDistribution, freezeTilt, etc.) with realistic tests?
    - Did you test abnormal cases like zero deposit, repeated calls, or huge distribution arrays?
2. **Security Audits**
    - Will you engage a third-party audit for the program code, specifically around the distribution logic and role-based checks?
    - How do you plan to handle any critical vulnerabilities found post-deployment?
3. **Simulation**
    - Can you do local or testnet simulations of various complex scenarios (multiple depositors, partial deposits, multiple distribution attempts)?
    - Do you test concurrency (two distributions hitting the same tilt simultaneously)?

---

## 6. User Experience & Operational Flow

1. **User Onboarding**
    - How do users discover which tilt they’re depositing into?
    - Do you provide a front-end that reads the tilt data from on-chain and shows them the deposit state?
2. **Distribution Transparency**
    - How does a user see the final distribution results?
    - Are logs and final amounts easily accessible or do they need to parse on-chain logs manually?
3. **Recovery / Dispute Mechanism**
    - If an incorrect distribution is posted, is there a mechanism to revert or dispute it before finalization?
    - How do you handle user disputes (e.g., “I was supposed to get 10 SOL but only got 1 SOL”)?
4. **Notifications**
    - Do you intend to notify depositors that a distribution is about to occur or has occurred?
    - Is there an off-chain push or event system for that?

---

## 7. Performance & Scalability

1. **Throughput**
    - How many tilts might be created or distributed at once?
    - Can your program handle a high volume of deposit transactions without running into Solana’s account lock or transaction size limits?
2. **Data Footprint**
    - Each tilt and distribution requires storing some state on-chain. **Do you have a pruning or archival strategy** for stale or closed tilts?
    - Could large amounts of historical tilt data lead to excessive storage costs?
3. **RPC Bottlenecks**
    - Does the validator process frequently query the chain for updates?
    - Might that cause performance issues or the need for a specialized indexing service?

---

## 8. Future Extensions

1. **Multiple Escrow Assets**
    - If you eventually want to support SPL tokens, not just SOL, how does that affect deposit and distribution logic?
    - Does your current approach for lamports easily extend to tokens?
2. **Integrating Other Oracles**
    - If you plan to use external data sources (e.g., sports results, price feeds), how do you foresee integrating them on-chain?
    - Are you using a specialized oracle program or your own?
3. **Upgrading to Full “DAO” Governance**
    - Do you want to eventually transition the “Program Admin” role to a DAO governance program (like Realms on Solana)?
    - If so, how do you plan for that handover of upgrade authority or config account authority?

---

###