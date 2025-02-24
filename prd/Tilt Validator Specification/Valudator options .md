# Valudator options

Below are two contrasting ways of ensuring that the entity submitting a distribution is â€œauthorizedâ€ (i.e., not malicious) in your Solana Tilt program. Which method you choose depends on whether you truly want a â€œpermissionlessâ€ approach or a role-based (whitelisted) approach.

1. Permissionless Approach: On-Chain Correctness Checks

In a fully permissionless design, any address can submit the distribution transaction. The program does not rely on the callerâ€™s â€œauthorityâ€ but rather on whether the proposed distribution itself is valid. In other words:

1. No Whitelist of Pubkeys

â€¢ There is no â€œauthorized validatorPubkeyâ€ field in the tilt account.

â€¢ Anyone can sign and send the execute_distribution(...) instruction.

2. On-Chain Checks Ensure Validity

â€¢ Sum of Amounts: Must not exceed (tiltPDA.lamports - rent_exempt_reserve).

â€¢ Tilt Status: Must be Active (not Frozen/Closed).

â€¢ Arithmetic Safety: The program uses checked_add/checked_sub. If an invalid array leads to overflow, the transaction fails.

â€¢ Rent-Exemption: If the tilt is not being fully closed, distribution must keep enough lamports in the PDA to remain rent-exempt.

3. If Proposal Is Invalid, Instruction Fails

â€¢ If the distribution is out of bounds (e.g., trying to distribute 1 lamport more than allowed), or the tilt is frozen, the on-chain logic rejects the transaction as unauthorized.

Why This Works Without a Whitelist

â€¢ Attackers can try to submit a malicious distribution, but it will fail the programâ€™s checks (sum of amounts, tilt status, etc.).

â€¢ A valid distribution, by definition, cannot harm the system, because it respects all constraints.

â€¢ Thus, â€œauthorizationâ€ is effectively replaced by â€œcorrectness verificationâ€. The â€œvalidatorâ€ does not need a recognized keyâ€”only a correct distribution payload that passes the tiltâ€™s checks.

2. Role-Based (Whitelisted) Approach: Checking Caller Pubkey

If you do want only a specific set of addresses (or a single â€œvalidatorâ€) to finalize distributionsâ€”i.e., you do not want it permissionlessâ€”store an authorized validator list in your global config or tilt account:

1. Store isValidator[Pubkey] = true/false

â€¢ Either in a global config or in each tiltâ€™s data (depending on your design).

2. Check Caller

â€¢ In your execute_distribution(...) instruction, do:

let caller_pubkey = ctx.accounts.caller.key();

require!(

tilt_config.is_validator[caller_pubkey] == true,

TiltError::Unauthorized

);

3. Then Perform the Usual On-Chain Checks

â€¢ Even if the caller is on the whitelist, you still confirm the sum of amounts <= tiltâ€™s lamports, tilt is Active, etc.

4. Outcome

â€¢ If the caller is not in the validator list, the instruction fails with an â€œUnauthorizedâ€ error.

Why Use a Whitelist?

â€¢ Sometimes you want to ensure only a known â€œtrusted aggregatorâ€ can finalize distributions.

â€¢ This might be necessary if your distribution logic depends on off-chain computations that canâ€™t be easily validated by the on-chain code.

Trade-Off

â€¢ Whitelists limit decentralization. If that single â€œauthorized validatorâ€ key is lost or compromised, the system is stuck or vulnerable.

â€¢ A permissionless design eliminates that single point of failure but relies on robust on-chain checks (or cryptographic proofs) to guarantee no malicious distribution can succeed.

3. Which Model to Choose?

1. Truly Permissionless

â€¢ Pro: No reliance on a single key or set of keys. Anyone can finalize distribution.

â€¢ Pro: More decentralized, no gatekeeping.

â€¢ Con: Must code the on-chain logic so that distributions canâ€™t be â€œcheatedâ€ (the program itself must calculate or verify correctness).

2. Whitelisted / Role-Based

â€¢ Pro: Tighter control; only known parties can push final distributions.

â€¢ Con: Key management risk (if the validatorâ€™s key is lost, no one can distribute). Less decentralized.

Most decentralized Solana programs aiming for trust-minimized logic choose the permissionless approach, where the program itself enforces correctness and does not rely on any single â€œauthorizedâ€ key. However, if your use case truly requires a specific aggregator or small set of known validators (e.g., specialized off-chain data feeds), you can store a validator list and perform a signature check on every distribution instruction.

4. Summary

â€¢ Permissionless: No single â€œauthorized validator.â€ The tilt contract verifies the distribution is valid (sum of amounts <= tiltâ€™s balance, correct tilt status, etc.), and any user can submit it.

â€¢ Role-Based: The program checks if the callerâ€™s public key is in a whitelist. If not, the transaction fails with Unauthorized.

Hence, to ensure a validator is â€œauthorizedâ€ under a role-based model, you store that validatorâ€™s pubkey on-chain and require a signature check. Alternatively, in a fully permissionless model, thereâ€™s no concept of â€œonly this key can finalize.â€ Instead, the program uses strict validity checks so that only correct distributions are accepted, no matter who submits them.