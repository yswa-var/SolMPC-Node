# Dynamic Distributions - Ephemeral values vs business rules

Below is a **side-by-side comparison** of the two approaches—**(A) ephemeral weighting** vs. **(B) storing explicit “business rules” in the PDA**—and a **recommendation** on which method might serve you best. Each approach allows the tilt’s **curator** to manage sub-tilt dependencies, but they differ in how the information is stored, updated, and validated.

---

## 1. Ephemeral Weights vs. Business Rules: Core Concepts

### (A) Ephemeral Weights

- **Concept**: At any given moment, the curator assigns percentages (like “75% sub-tilt, 25% local”) stored in the tilt’s account or quickly updated off-chain. Validators read these ephemeral ratios at distribution time.
- **Flexibility**: The curator can tweak them frequently, at any moment.
- **Distribution**: The validator subset references these ephemeral ratios, calculates final amounts, and signs off.
- **Risks**: Potential confusion if partial signatures are collected for one ratio, then the curator changes it mid-process. The aggregator or validators must re-verify or discard old partial signatures.

### (B) Business Rules in PDA

- **Concept**: Instead of ephemeral numeric percentages, you store structured “logic” or “formulas” in the tilt’s PDA. For example, “40% to curators, but sub-tilt X gets 75% of curator portion.”
- **Versioning**: The curator updates the rules by calling `updateBusinessRules(...)`. Each update increments a version number or timestamp.
- **Distribution**: Validators read the current rules from the PDA (which includes sub-tilt references, ratio_of_curators_share, etc.) and compute final amounts accordingly.
- **Predictability**: Everyone references the same structured logic. If the curator updates it, that triggers a new version, so partial signers know exactly which version is in effect.

---

## 2. Key Differences

| **Aspect** | **(A) Ephemeral Weights** | **(B) Business Rules in PDA** |
| --- | --- | --- |
| **Data Representation** | Often raw percentages or immediate weighting fields (“75% to sub-tilt, 25% local”). | A structured set of rules or formulas (like `protocol_cut=20%, curator_cut=40%, sub-tilt portion=75%`). |
| **Updates** | Curator can change ephemeral weights at any moment (possibly multiple times a day). | Curator changes are explicit on-chain updates via an `updateBusinessRules()` function, versioned. |
| **Validation Over Time** | Potential for partial signatures to become obsolete if weights change mid-process. | Each distribution references a definitive version of the rules, so partial signatures remain consistent. |
| **On-Chain Complexity** | Possibly simpler data but can cause confusion if ephemeral updates are frequent. | More structured data (logic, arrays of sub-tilt references), but results in clearer, auditable changes. |
| **Event/Audit Trail** | Must rely on ephemeral or a single numeric field that changes. Harder to maintain version logs. | Each rule change triggers an on-chain event (`RulesUpdated`), making it easier to see the tilt’s history. |
| **Security / Collisions** | If a curator changes weights just before the aggregator broadcasts, partial signatures may break. | The tilt program & validators see a stable set of rules each time. No partial sig confusion if version # is consistent. |
| **Implementation** | Potentially simpler to implement at first, but demands clear “snapshot” handling for partial sigs. | Slightly more up-front coding to store formulas or logic, but leads to a more robust and stable approach. |

---

## 3. Pros & Cons

### (A) Ephemeral Weights

**Pros**

1. **Immediate Flexibility**: The curator can instantly shift sub-tilt weighting at any time (like “75%” → “80%”) without formal version updates.
2. **Minimal Overhead**: Might only store a few ephemeral fields or a numeric ratio in the tilt’s account.

**Cons**

1. **Partial Signer Confusion**: If ephemeral weights are changed mid-collection of signatures, old partial signatures become invalid or stale.
2. **Less Transparency**: Harder to keep a clear record of changes if the ephemeral field is overwritten repeatedly.
3. **Fewer On-Chain Guarantees**: It’s easier for the aggregator to claim they used a “previous ephemeral ratio” while the curator changed it, leading to disputes.

### (B) Business Rules in PDA

**Pros**

1. **Stable Versioning**: A tilt’s “rules” are updated in discrete on-chain transactions, each with a clear event and version number.
2. **Predictable Finalization**: Subset validators always reference the current `tiltBusinessRules.version`. No ambiguous mid-process changes.
3. **Better Auditability**: Each update is logged. Observers see exactly how sub-tilt references changed over time.
4. **Easier Partial Sig Flow**: If the rules update, the aggregator must wait for the new version. Otherwise, they use the old version. Fewer collisions.

**Cons**

1. **Slightly More Setup**: A “business rules” struct is more formal, requiring consistent logic in the tilt account.
2. **Less Immediate**: The curator must do an on-chain transaction to update rules, which might cost fees or require a standard process.

---

## 4. Best Practice Recommendation

In most scenarios, **(B) Storing Business Rules in the PDA** is the **preferred** approach because:

1. **Clarity & Consistency**: Instead of ephemeral changes that can disrupt partial signatures, you get a stable snapshot each time you update.
2. **Reduced Risk of Partial Sig Race**: No fiasco of ephemeral values mid-collection. The aggregator references version X, partial signers confirm the distribution is correct for version X, done.
3. **Better Audit Trail**: Observers, compliance, or even your own team can see exactly how tilt-level logic has evolved.

**(A) Ephemeral Weights** can still work if:

- You have a small team or a simpler system where ephemeral changes are not frequent or partial signers can easily re-collect signatures.
- You are comfortable with the concept of “live” data that can invalidate partial signatures at any moment.

However, ephemeral weighting is more prone to confusion, especially as you scale up the number of validators or want robust accountability.

**Conclusion**:
**Storing business rules**—the structured ratio (protocol vs. curators vs. publishers) and sub-tilt references—in the **PDA** offers **a more secure, stable, and auditable** design. It ensures that at distribution time, the tilt’s logic is explicitly versioned, letting the off-chain validator subset produce partial signatures in a single, consistent pass. This approach **minimizes** the risk of disputes or partial signer collisions and **maximizes** clarity for all participants.