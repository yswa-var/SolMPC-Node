# Validator Distribution Logic

Below is a **more detailed look** at how validators can compute the **final distribution** off-chain, especially if your tilt supports **nested sub-tilts**, dynamic weighting, and multiple “business rules.” This approach can accommodate the scenario where a tilt references sub-tilts that themselves might have local or further sub-tilts, ensuring we arrive at a **flattened** array of `(recipient, amount)` for the final distribution.

---

## 1. High-Level Computation Goals

1. **Apply Global or Top-Level Splits**
    - Typically, each tilt might define top-level shares: e.g., **protocol** gets 20%, **curators** get 40%, **publishers** get 40%.
    - We then refine the “curators’ share” based on sub-tilt references (e.g., “75% to sub-tilts, 25% to local curators”).
2. **Resolve Sub-Tilts Recursively**
    - A sub-tilt can either be treated as a single “line item” or it may also have further sub-rules.
    - We want a final, *flattened* result that enumerates every unique recipient (like a user, curator, or actual wallet) and how many lamports they get.
3. **Respect Dynamic or Versioned Logic**
    - If the tilt’s “business rules” (ratios, conditions) have changed, validators must incorporate the latest version from the tilt’s PDA.
    - Each sub-tilt might similarly store its own local splits.
4. **Apply Additional Constraints**
    - Possibly skip or block recipients who fail KYC.
    - Round amounts properly so that the total sum never exceeds `(tiltBalance - rentExemptReserve)`.

---

## 2. Parsing the Tilt’s On-Chain “Business Rules”

Below is an example data layout you might find in the tilt’s PDA:

```rust
pub struct TiltBusinessRules {
    pub version: u64,                 // For consistency checks
    pub protocol_bps: u16,            // e.g. 2000 = 20%
    pub curator_bps: u16,            // e.g. 4000 = 40%
    pub publisher_bps: u16,          // e.g. 4000 = 40%
    pub sub_tilt_rules: Vec<SubTilt> // each entry references another tilt
}

pub struct SubTilt {
    pub sub_tilt_pubkey: Pubkey,      // Points to another tilt
    pub ratio_bps: u16,              // e.g. 7500 = 75% of the curator portion
}

```

The aggregator or each validator in the subset would:

1. **Load** `TiltBusinessRules` for the main tilt.
2. **Confirm** it sums to 100% (like 2000 + 4000 + 4000 = 10000 bps).
3. **Check** the sub-tilts’ ratio as well. For instance, `ratio_bps = 7500` implies 75% of the curator portion is allocated to that sub-tilt.

---

## 3. Step-by-Step Distribution Algorithm

Let’s assume a scenario:

- The tilt’s balance is `X` lamports.
- `X` has not been distributed yet.
- The top-level splits:
    1. Protocol: 20%
    2. Curators: 40%
    3. Publishers: 40%
- The “curators” share includes local curators plus sub-tilt references. In the example: 75% of the curators’ share is sub-tilts, 25% is local curators.

### 3.1 Outline

1. **Compute Top-Level Splits**
    - `protocol_amount = X * (protocol_bps / 10_000)`
    - `curator_amount = X * (curator_bps / 10_000)`
    - `publisher_amount = X * (publisher_bps / 10_000)`
    - Make sure `protocol_amount + curator_amount + publisher_amount = X` (minus rounding differences).
2. **Compute Local Curator vs. Sub-Tilts**
    - Within `curator_amount`, 25% might go to local curator(s), 75% to sub-tilts.
    - Let’s say `local_curator_amt = curator_amount * 0.25`, `sub_tilts_amt = curator_amount * 0.75`.
3. **For Each Sub-Tilt**
    - We have an array `sub_tilt_rules = Vec<SubTilt>`, each with `ratio_bps`. Suppose we only have 1 sub-tilt with ratio = 100% of the sub-tilts portion. Then `sub_tilt1_amt = sub_tilts_amt`.
    - If there were multiple sub-tilts, sum their ratio_bps and allocate sub_tilts_amt among them proportionally.
4. **Resolve Each Sub-Tilt**
    - Now, each sub-tilt is itself a tilt. We must do the same logic:
        1. Grab that sub-tilt’s `TiltBusinessRules` from on-chain.
        2. Let’s call the sub-tilt’s balance for distribution “the sub-tilt’s portion.” In this case, `sub_tilt1_amt`.
        3. Apply that sub-tilt’s own top-level splits or further recursion until we eventually reach only “real recipients” or deeper sub-tilts.
5. **Flatten Results**
    - Each sub-tilt might produce multiple `(recipient, amount)` pairs.
    - We merge them with our top-level recipients array. If a recipient key is repeated, sum amounts.
    - Similarly, local curators might produce 1 or many recipients if there’s a local multi-sig or distribution among multiple local curators.
6. **Sum & Round**
    - We ensure the sum of all final amounts does not exceed `(tiltBalance - rentExemptReserve)`.
    - Apply rounding rules if needed. Possibly accumulate “leftover lamports” to protocol or the last recipient.
7. **KYC Filtering** (Optional)
    - For each `(recipient, amount)`, if `kycVerified[recipient] != true`, set that amount to 0 or skip it.
    - Or if partial skipping is disallowed, the aggregator might fail the distribution entirely.
8. **Output**
    - The final distribution is a **flat list** of `(recipient, final_amount)`.
    - Possibly remove zero-amount lines for cleanliness.

---

### 3.2 Example Pseudocode

Below is a simplified approach to handle **nested** sub-tilts:

```python
def compute_distribution(tilt_pubkey, total_amt):
    tilt = fetchTiltDataFromRPC(tilt_pubkey)
    # tilt has: protocol_bps, curator_bps, publisher_bps, sub_tilts, etc.

    # 1. Top-level splits
    protocol_amt = (total_amt * tilt.protocol_bps) // 10000
    curator_amt  = (total_amt * tilt.curator_bps)  // 10000
    publisher_amt= (total_amt * tilt.publisher_bps)// 10000

    distribution_map = {}

    # 2. Add protocol recipient(s)
    if tilt.protocol_recipient:
        distribution_map[tilt.protocol_recipient] = protocol_amt

    # 3. Deeper logic for curators
    #    Possibly (local_curator_share, sub_tilts_share)...

    local_curator_amt, sub_tilts_amt = split_curator_amt(curator_amt, tilt)

    # if local_curator_amt > 0, add local curators...
    # distribution_map[ tilt.local_curator_pubkey ] += local_curator_amt

    # 4. For sub-tilts:
    # E.g., each sub_tilt has ratio_bps summing to 10000 across them
    sum_ratio = sum(st.ratio_bps for st in tilt.sub_tilts)
    results = {}
    for st in tilt.sub_tilts:
        st_part = (sub_tilts_amt * st.ratio_bps) // sum_ratio
        sub_tilt_results = compute_distribution(st.sub_tilt_pubkey, st_part)
        # merge sub_tilt_results into distribution_map

    # 5. Publishers
    # distribution_map[ tilt.publisher_pubkey ] += publisher_amt
    # or if multiple publishers, break down similarly

    # 6. Return the final flattened map
    return distribution_map

```

- **Note**: This function is **recursive**. If a sub-tilt references further sub-tilts, we keep going down. Eventually, we reach the “leaves” where real recipients are assigned amounts, or no further sub-tilts exist.

---

## 4. Edge Cases & Complexities

1. **Infinite or Cyclical References**
    - The tilt or sub-tilts might incorrectly reference each other in a cycle. The aggregator must detect and fail.
    - Typically, your on-chain code ensures no cycles are allowed, or the aggregator times out.
2. **Exceeding Transaction Size**
    - If a single tilt’s final distribution has hundreds of recipients, we might need chunking or multiple instructions.
    - The aggregator or partial signers might produce partial results, or do a multi-step distribution if necessary.
3. **Partial or Additional Layers**
    - Some designs might only do a single level of sub-tilt references (like “sub-tilts portion = 75%”), no deeper recursion. This is simpler.
    - Others might do indefinite nesting, which the aggregator must handle systematically.
4. **Time or Slot Conditions**
    - The tilt might specify “only distribute after slot X.” Each validator checks if the current slot > X.
    - Or if deposit < Y lamports, skip certain recipients. The aggregator can embed these conditions in the logic.

---

## 5. Final Checks & Signature

After computing the final array of `(recipient, amount)`:

1. **Sum** the amounts to ensure it doesn’t exceed `(tiltBalance - rentExemptReserve)`.
2. **Eliminate** zero-amount lines.
3. **Check** `kycVerified[recipient]` if required.
4. **Form** the final message structure, e.g.:
    
    ```json
    {
      "tiltID": 42,
      "version": tilt.version,
      "recipients": [ A, B, C, ... ],
      "amounts": [ 1000, 2000, 3000, ... ],
      "nonce": 123456
    }
    
    ```
    
5. **Partial Sign** this exact message. An aggregator merges partial sigs into one aggregated signature.

---

## 6. Summary

Validators, or the aggregator node, must:

1. **Fetch** the tilt’s data (and sub-tilt data if nested).
2. **Recursively compute** each portion: top-level splits, sub-tilts within the curator share, etc.
3. **Flatten** or merge results so the final distribution is a single list of `(recipient, lamports)`.
4. **Check** advanced conditions (KYC, min deposit, time constraints).
5. **Sign** and produce an aggregated signature. The tilt program then **verifies** it on-chain and executes the distribution.

This deeper dive shows the **algorithmic** flow: from reading multiple PDAs for sub-tilts, to applying ratio logic, to consolidating the final list, ensuring no step can be manipulated without validators noticing.