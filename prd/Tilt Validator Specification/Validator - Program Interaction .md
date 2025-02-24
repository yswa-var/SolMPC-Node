# Validator - Program Interaction

---

## 1. Overview of the Validator’s Function

1. The **Validator** is responsible for:
    - Off-chain calculation of payout splits.
    - Submitting those splits to the on-chain program (the “Tilt” program) in a transaction.
2. The **Smart Contract** (on Solana) must:
    - Check that the caller is an authorized validator.
    - Verify the tilt is still active/not frozen.
    - Distribute lamports from the tilt’s PDA to recipients, or store the proposed distribution until enough validators approve it (in a multi-validator setup).

In other words, the validator acts as an “oracle” that finalizes or proposes final settlement. The rest of the system trusts that the validator has computed correct results.

---

## 2. Communication Flow: Off-Chain → On-Chain

### 2.1 Single Validator Model (Simplest)

1. **Validator Gathers Data**
    - Off-chain, the validator queries Solana RPC for each Tilt’s deposits, plus any user addresses that contributed.
    - The validator also factors in any external data needed (e.g., odds, off-chain calculations, or subscription receipts).
2. **Validator Computes Payout Splits**
    - For each tilt, the validator determines how to split the total lamports among the designated recipients.
3. **Validator Submits On-Chain Transaction**
    - The validator calls a program instruction, e.g., `execute_distribution(tiltPda, recipients[], amounts[])`.
    - This transaction must be signed by the validator’s keypair recognized by the contract.
4. **Contract Verifies & Executes**
    - The program checks:**(a)** `isValidator[caller] == true` (from global config).**(b)** `tilt.status == Active` (not closed/frozen).**(c)** `sum(amounts[]) <= tiltPda.lamports`.
    - If all checks pass, the program transfers lamports from the tilt’s PDA to each recipient and optionally marks the tilt `status = Closed`.

**Structure of Communication**:

- **Off-Chain**: The validator obtains data from standard Solana RPC calls, does any specialized calculation.
- **On-Chain**: A single instruction carrying the final distribution array (recipients + amounts).

**Pros**:

- Very straightforward.
- Minimal overhead.

**Cons**:

- Relies on a single validator’s correctness.
- If that validator is compromised, incorrect distributions could be executed (unless you add further safe checks).

---

### 2.2 Multiple Validators with Threshold (M-of-N)

If you want **multiple** off-chain validators to propose or confirm the distribution, you can incorporate threshold logic. The high-level flow might be:

1. **Off-chain** each validator calculates or verifies the distribution.
2. **On-chain** there is a “Distribution Proposal” mechanism in the tilt’s state.
    - **Step A**: A validator calls `propose_distribution(tiltPda, recipients[], amounts[])`.
        - This writes a “pending distribution” struct into the tilt account data (or into a sub-PDA) with a unique `distributionID`, the proposed amounts, and a list of validator approvals (starting with the proposer).
    - **Step B**: Other validators see that pending proposal (via RPC) and, if they agree, call `approve_distribution(tiltPda, distributionID)`.
        - Each call checks `isValidator[msg.sender]`, then marks that validator’s approval on-chain.
    - **Step C**: Once `>= requiredApprovals` (e.g., 2-of-3 or M-of-N) are collected, the on-chain program automatically or via another instruction finalizes the distribution, transferring funds to recipients.

**Structure of Communication**:

- **Off-Chain**:
    - Each validator still fetches deposit data and does the math. They can coordinate off-chain on the final amounts or propose separate distributions.
- **On-Chain**:
    - A two-step process:
        1. `propose_distribution(...)`: Store the proposed splits in the tilt’s account, marking it “pending.”
        2. `approve_distribution(...)`: Additional validators sign separate transactions that reference the stored proposal, incrementing an “approval count” in on-chain state.
    - Once the threshold is met, the program disburses the funds.

**Pros**:

- Safer in that no single validator can push an incorrect distribution if the design enforces threshold approvals.
- On-chain record of the distribution proposal (transparency).

**Cons**:

- More complex.
- Requires multiple transaction calls (proposal + approvals).

---

## 3. Validator Whitelisting & Authentication

Regardless of single or multiple validators, the contract must know who the **valid** validator(s) are. Typically:

1. **Global Config Account**
    - Stores a boolean or set/array like `isValidator[Pubkey] = true/false`.
    - Alternatively, store an array `validatorKeys[]`.
2. **Validator Check**
    - In the distribution instruction, the program reads the config account.
    - If the caller’s public key is not recognized, it rejects the transaction.

If you need to add or remove validators, your **Program Admin** can sign an `update_config` instruction to modify the list of valid keys in the global config.

---

## 4. Data Structures for Proposed Distributions

Below is an example of how you might store a distribution proposal on-chain (if using threshold logic).

```rust
pub struct DistributionProposal {
    pub distribution_id: u64,
    pub recipients: Vec<Pubkey>,     // or a fixed-size array if known
    pub amounts: Vec<u64>,           // match the length of recipients
    pub approvals: u8,               // how many validators approved
    pub max_approvals: u8,           // optional, total validators
    pub is_finalized: bool,
}

```

You could store one active proposal per tilt, or keep a vector of proposals in the tilt account (or in a dedicated proposal PDA) if you allow multiple distributions over time.

---

## 5. Putting It All Together: Example Workflows

### 5.1 Single Validator Workflow

1. **Program Admin**: sets `validatorPubkey = somePublicKey` in the global config.
2. **Users**: deposit lamports into the tilt.
3. **Validator**:
    1. Monitors the chain via RPC (e.g., `getAccountInfo(tiltPda)`), sees the total deposit.
    2. Calculates final distribution.
    3. Sends a transaction calling `execute_distribution(tiltPda, recipients[], amounts[])`, signing with `validatorPubkey`.
4. **Program**: checks `require(isValidator[caller] == true)`. If okay, it pays recipients and marks the tilt closed.

### 5.2 Multiple Validators with Threshold

1. **Program Admin**: sets `validatorKeys = [v1, v2, v3]` in global config, `requiredApprovals = 2`.
2. **Users**: deposit lamports into the tilt.
3. **Validator 1**:
    1. Off-chain collects deposit data.
    2. Calls `propose_distribution(tiltPda, recipients[], amounts[])`.
    3. The tilt’s data now has an entry: `DistributionProposal { distribution_id: 101, approvals: 1, recipients: [...], amounts: [...] }`.
4. **Validator 2**:
    1. Off-chain checks the proposed distribution.
    2. If correct, calls `approve_distribution(tiltPda, distribution_id=101)`.
    3. The program increments approvals to 2, sees that `2 >= requiredApprovals`, and finalizes the distribution.
    4. The program then transfers lamports to the recipients and sets `is_finalized = true`.
5. **Validator 3**:
    - Only needed if the first two disagree or if you want more than 2 approvals.

---

## 6. Implementation Details for Communication

### 6.1 How the Validator Knows What to Distribute

- **Source of Data**: The validator (or all validators) fetch tilt account data from a Solana node:
    - `getAccountInfo(tiltPda)` to see how many lamports are in the tilt’s PDA.
    - Possibly parse the tilt’s custom data to see how many deposits or references to who contributed.
- **Off-Chain Aggregation**: They might also fetch logs or event-like messages from transaction histories to see each deposit.
- **Any Off-Chain Inputs**: If your distribution depends on external events (e.g., sports results, random draws), the validator uses external data as well.

### 6.2 Submitting the Result

On Solana, *anyone* can pay the transaction fee (the fee payer can be different from the signer). The validator must still sign with the recognized “validatorPubkey,” but the actual lamports for the fee could come from a separate payer account if you want to sponsor transactions.

### 6.3 Security / Trust Considerations

- If using **Single Validator**:
    - Make sure that key is well-protected or multi-sig itself (e.g., the validator key might be a “smart wallet” that collects multiple signatures off-chain to confirm any distribution).
- If using **Multiple Validators**:
    - You distribute trust among the set of validator keys.
    - Attackers would need to compromise a threshold of them to push an invalid distribution.

### 6.4 Logging & Transparency

- Use `msg!()` or `log!()` in the Solana program to emit logs each time a distribution is executed (or proposed/approved).
- Off-chain indexers or explorers can parse these logs to show a human-readable event feed.

---

## 7. Summary: Recommended Validator Architecture

1. **Store the Validator Role** in your global config or in tilt-specific data, e.g., a `Vec<Pubkey>` or a boolean map.
2. **Implement an Instruction** `execute_distribution(...)` or `propose_distribution(...) + approve_distribution(...)` that:
    - **Checks** the caller is recognized as a validator (or among the validator set).
    - **Verifies** the tilt’s status (active, not frozen) and sufficient funds.
    - **Transfers** lamports from the tilt’s PDA to recipients or sets a final “distribution approved” state.
3. **Off-Chain**: Each validator calculates or verifies the distribution. They only pass the final distribution to the on-chain function.
4. **(Optional) Threshold**: If you need multiple approvals, store distribution proposals on-chain and track each validator’s signature or approval.
5. **Security**: If using a single validator key, consider making that key a multi-sig or using a “smart wallet” so that no single person can unilaterally sign malicious distributions.

This framework should give you a robust starting point for **how validators communicate with your Solana program**:

- all specialized logic happens off-chain,
- the final distribution is submitted in a Solana transaction,
- the contract checks the caller’s authority and any threshold requirements.

[Validator Questions](Validator - Program Interaction 191370a0a66d80a885caca683c0b98d3/Validator Questions 191370a0a66d80099a19c5876149ac41.md)