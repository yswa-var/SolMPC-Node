# Tilt Calculation Example

Below is an **example** of how validators might calculate a **final distribution** for a tilt using **nested sub-tilt logic**. We’ll walk through the **numerical steps**—from the top-level splits, to a sub-tilt referencing further shares—**yielding a final flattened distribution** among actual wallet recipients.

---

## Example Tilt Scenario

- **Tilt Balance (`X`)**: 100 SOL
- **Top-Level Splits**:
    1. **Protocol**: 20%
    2. **Curators**: 40%
    3. **Publishers**: 40%
- **Curators’ Internal Rule**:
    - 25% of curators’ share goes to **local curator**.
    - 75% goes to **sub-tilts**.
    - There are **two sub-tilts** defined.
- **Sub-Tilt #1**:
    - Weighted at 66% of the “sub-tilts portion”.
    - Has its own distribution rules: split among **two local wallets**:
        - `subT1_creatorA`: 60%
        - `subT1_creatorB`: 40%
- **Sub-Tilt #2**:
    - Weighted at 34% of the “sub-tilts portion”.
    - Has a single local wallet recipient: `subT2_singleUser`.

**Note**: We assume no further recursion beyond sub-tilt #1 and #2, just a final wallet split.

---

## 1. Top-Level Split

1. **Protocol**: 20% of 100 SOL = `100 * 0.20 = 20 SOL`
2. **Curators**: 40% of 100 SOL = `100 * 0.40 = 40 SOL`
3. **Publishers**: 40% of 100 SOL = `100 * 0.40 = 40 SOL`

So after top-level splits:

- **Protocol** gets `20 SOL` (we’ll say this goes to a single address, `protocolWallet`).
- **Curators** pot = `40 SOL`.
- **Publishers** pot = `40 SOL` (perhaps going to `publisherWallet` if it’s a single entity, or multiple addresses if that’s your design).

---

## 2. Curators’ Internal Rule

From the **curators** pot (`40 SOL`):

- 25% to **local curator**
- 75% to **sub-tilts**

Therefore:

- **Local Curator** = `40 * 0.25 = 10 SOL`
- **Sub-Tilts** (combined) = `40 * 0.75 = 30 SOL`

Hence:

1. Local curator (`localCuratorWallet`) gets `10 SOL`.
2. We have `30 SOL` allocated to sub-tilts.

---

## 3. Allocating Among Sub-Tilts

We have **two sub-tilts** sharing that `30 SOL`, with ratio 66% and 34%:

- **Sub-Tilt #1**: 66%
- **Sub-Tilt #2**: 34%

### 3.1 Sub-Tilt #1 Amount

- **subT1_amount** = `30 SOL * 0.66 = 19.8 SOL`
    - Let’s keep that as `19.8 SOL` for now. We’ll see how we handle rounding.

### 3.2 Sub-Tilt #2 Amount

- **subT2_amount** = `30 SOL * 0.34 = 10.2 SOL`

Check total = `19.8 + 10.2 = 30.0` SOL. Good.

---

## 4. Resolving Sub-Tilt #1

Sub-Tilt #1’s own internal logic says:

- 60% to `subT1_creatorA`
- 40% to `subT1_creatorB`

Since subT1_amount = `19.8 SOL`:

1. **subT1_creatorA** = `19.8 * 0.60 = 11.88 SOL`
2. **subT1_creatorB** = `19.8 * 0.40 = 7.92 SOL`

---

## 5. Resolving Sub-Tilt #2

Sub-Tilt #2 is simpler: it references a single wallet `subT2_singleUser`, so:

- subT2_singleUser = `10.2 SOL`

No further splits needed here.

---

## 6. Final Flattened Distribution

We now combine all amounts:

1. **Protocol**: `20 SOL` → `protocolWallet`
2. **Curators**:
    - Local Curator: `10 SOL` → `localCuratorWallet`
    - Sub-Tilt #1:
        - `subT1_creatorA`: `11.88 SOL`
        - `subT1_creatorB`: `7.92 SOL`
    - Sub-Tilt #2:
        - `subT2_singleUser`: `10.2 SOL`
3. **Publishers**: `40 SOL` → `publisherWallet`

**Check Summation**:

- Protocol: 20
- Local Curator: 10
- subT1_creatorA: 11.88
- subT1_creatorB: 7.92
- subT2_singleUser: 10.2
- Publishers: 40

Sum = `20 + 10 + 11.88 + 7.92 + 10.2 + 40 = 100 (exact)`

*(We used decimals for clarity, but in real lamports math, you likely do integer basis points or rounding. Let’s see how that might look.)*

---

## 7. Rounding / Integer Considerations

### 7.1 In Lamports

1 SOL = 1,000,000,000 lamports. If dealing with integer math, you typically do:

- `protocolLamports = (100 * 1_000_000_000) * 2000 / 10000`
- etc.

But for an **illustrative** example, we used decimal SOL.

### 7.2 Handling Residual

If we do integer math, we might end up with a leftover lamport or two after rounding. One approach:

1. **Accumulate** leftover lamports in a “leftover pot” and assign them to either the protocol or the last recipient.
2. Or **truncate** each share (floor) and let 1–2 lamports remain undistributed until the next round.

---

## 8. KYC & Exclusions (Optional)

If we discover, for instance, `subT1_creatorB` fails KYC, the aggregator might:

- Set that line to `0 SOL`, distributing `7.92 SOL` to someone else or simply not distributing it.
- Or revert the entire distribution if policy says “any KYC failure blocks the distribution.”

---

## 9. Final Distribution Array

A final distribution might be a JSON:

```json
{
  "tiltID": 42,
  "version": 7,
  "nonce": 123456,
  "recipients": [
    "protocolWallet",
    "localCuratorWallet",
    "subT1_creatorA",
    "subT1_creatorB",
    "subT2_singleUser",
    "publisherWallet"
  ],
  "amounts": [
    20.0,
    10.0,
    11.88,
    7.92,
    10.2,
    40.0
  ]
}

```

*(In a real aggregator, you might convert these to integer lamports, e.g., `20000000000` for 20 SOL, etc.)*

---

### Conclusion

This **example calculation** shows how:

1. **Top-Level Splits** partition the tilt’s total among protocol, curators, publishers.
2. **Curators** further split into local vs. sub-tilts.
3. Each **sub-tilt** (like #1, #2) might have its own internal logic or multiple recipients, culminating in a flattened final distribution.
4. The aggregator or each validator carefully does the math, references the tilt’s stored ratio data, and ensures the sum is correct before signing off.

## Protocol Fee

In the **example distribution** scenario where the **protocol** receives a certain percentage (e.g., 20%), the **protocol** is typically the **underlying platform or network** that provides the infrastructure for the tilt system. You can think of it as a **“service fee”** or **“platform cut”** reserved for:

1. **Maintaining Core Infrastructure**
    - Covering operational costs, upgrades, or ongoing maintenance for the tilt program or the broader validator architecture.
2. **Protocol Treasury or Foundation**
    - Many blockchain-based platforms direct a portion of fees/revenue to a treasury or foundation that funds development, marketing, or grants.
3. **Platform Governance**
    - If the tilt system has a governance structure, the protocol’s allocation might sustain governance activities (audits, proposals, etc.) or be used as a community fund.

In a real implementation, you can **rename** that 20% portion and point it to:

- A **foundation wallet** or treasury address.
- A **DAO** that decides how funds are spent.
- A **maintenance/upgrade** account for covering future enhancements.

So, **“the protocol”** in this example is a **placeholder** for the **entity or treasury** that owns and maintains the tilt platform’s core functionality—ensuring it remains sustainable and can continue providing the service for creators, publishers, and sub-tilts.