# Validator questions Feb 10

---

## 1. Membership & Subset Selection

1. **How do we handle new validators joining or old validators leaving?**
    - Must the entire pool perform a new DKG each time membership changes?
    - Or do we rely on ephemeral keys per subset so global changes don’t force re-derivation?
2. **Is the random subset truly unpredictable and tamper-resistant?**
    - Do we depend on an on-chain VRF oracle, or do off-chain processes produce seeds that could be manipulated?
    - If a subset fails to finalize, can an attacker spam re-randomization to eventually get a malicious majority?

---

## 2. Threshold Signing & Verification

1. **Which threshold signature algorithm (BLS, ECDSA, FROST, etc.) are we using, and how is it verified on-chain?**
    - Is there a suitable library for BPF on Solana that can handle pairing-based checks (for BLS) or ECDSA?
    - Is the compute cost feasible under Solana’s constraints?
2. **How does the tilt program reference or store ephemeral subset keys?**
    - Do we store a single ephemeral public key in the tilt PDA?
    - If so, who sets it? Is there a “committee manager” instruction that updates the tilt state each time?

---

## 3. Off-Chain Collaboration & Broadcast

1. **How do subset validators exchange partial signatures and confirm each other’s correctness?**
    - Are they using a peer-to-peer network with gossip? A central aggregator node?
    - How do we prevent a malicious aggregator from withholding partial signatures or forging them?
2. **What if the designated broadcaster fails to submit** (due to inaction or malice)?**
    - Is there a fallback aggregator or an option for any other subset node to broadcast the aggregated signature?
    - Does the subset have a time-based re-try or re-randomization if no broadcast occurs?

---

## 4. Slashing & Accountability

1. **Do we have a slashing mechanism for validators who sign fraudulent distributions?**
    - Must partial signatures be published so watchers can detect double-signing or contradictory states?
    - If M-of-N colludes to produce a malicious final signature, can the system automatically slash them, or is that a catastrophic event requiring governance intervention?
2. **How do we handle disputes about the final distribution**?**
    - If an honest minority claims the aggregator is ignoring sub-tilt weights or forging amounts, can they prove it on-chain?
    - Is there a “challenge” window or a mechanism to revert if an obviously incorrect signature is posted?

---

## 5. Scalability & Maintenance

1. **What happens when we have very large distribution arrays** (hundreds of recipients)?**
    - Do we break them into multiple instructions or use a specialized approach to avoid transaction size limits?
    - Does partial signature collection become unwieldy if the aggregator must unify large data sets?
2. **How do we manage long-term key rotations or changes to the threshold** (e.g., from M-of-N to M’-of-N’)?**
- Do we have a planned “upgrade path” for the validator pool?
- If the tilt program references a group pubkey, how do we securely migrate to a new pubkey without invalidating ongoing distributions?

---

###