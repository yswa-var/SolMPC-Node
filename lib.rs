use anchor_lang::prelude::*;
use anchor_lang::solana_program::program::invoke;
use anchor_lang::solana_program::system_instruction;

declare_id!("3WwkA4gXYDR1fQ1BHfe5J4XoFYZ5kdSPe2kGkDE3uC22");

#[program]
pub mod sub_tilts_hybrid {
    use super::*;

    /// Stores the rules URI on-chain for reference
    pub fn set_rules_uri(
        ctx: Context<SetRulesUri>,
        rules_uri: String,
    ) -> Result<()> {
        let routing_rules = &mut ctx.accounts.routing_rules;
        routing_rules.owner = ctx.accounts.owner.key();
        routing_rules.rules_uri = rules_uri;
        
        Ok(())
    }

    /// Executes payments based on pre-calculated amounts from the client
    /// The client performs the recursive calculations off-chain using rules fetched from URI
    pub fn execute_payments<'a, 'b, 'c, 'info>(
        ctx: Context<'a, 'b, 'c, 'info, ExecutePayments<'info>>, 
        payment_amounts: Vec<u64>,
        total_amount: u64,
    ) -> Result<()> {
        let payer = &ctx.accounts.payer;
        let system_program = &ctx.accounts.system_program;
        
        // Verify we have the right number of recipients
        require!(
            ctx.remaining_accounts.len() == payment_amounts.len(),
            CustomError::RecipientCountMismatch
        );
        
        // Verify total amounts add up to expected total (basic validation)
        let sum: u64 = payment_amounts.iter().sum();
        require!(
            sum == total_amount,
            CustomError::InvalidTotalAmount
        );
        
        // Execute each payment
        for (i, recipient) in ctx.remaining_accounts.iter().enumerate() {
            let amount = payment_amounts[i];
            
            let transfer_ix = system_instruction::transfer(
                &payer.key(), 
                &recipient.key(), 
                amount
            );
            invoke(
                &transfer_ix,
                &[payer.to_account_info(), recipient.clone(), system_program.to_account_info()],
            )?;
        }
        
        Ok(())
    }

    /// Verify a specific payment distribution against rules hash for integrity verification
    pub fn verify_distribution<'a, 'b, 'c, 'info>(
        ctx: Context<'a, 'b, 'c, 'info, VerifyDistribution<'info>>, 
        payment_amounts: Vec<u64>,
        rules_hash: [u8; 32],  // Hash of the rules used for this distribution
        total_amount: u64
    ) -> Result<()> {
        let routing_rules = &ctx.accounts.routing_rules;
        
        // Verify the specified rules hash matches the stored rules hash (if provided)
        if routing_rules.rules_hash != [0u8; 32] {
            require!(
                routing_rules.rules_hash == rules_hash,
                CustomError::RulesHashMismatch
            );
        }
        
        // Verify total amount
        let sum: u64 = payment_amounts.iter().sum();
        require!(sum == total_amount, CustomError::InvalidTotalAmount);
        
        // Note: Detailed rule verification is done client-side with the rules fetched from URI
        
        Ok(())
    }

    /// Update the rules hash for integrity verification
    pub fn update_rules_hash(
        ctx: Context<UpdateRulesHash>,
        rules_hash: [u8; 32]
    ) -> Result<()> {
        let routing_rules = &mut ctx.accounts.routing_rules;
        routing_rules.rules_hash = rules_hash;
        
        Ok(())
    }
}

#[derive(Accounts)]
pub struct SetRulesUri<'info> {
    #[account(
        init,
        payer = owner,
        space = RoutingRules::SIZE
    )]
    pub routing_rules: Account<'info, RoutingRules>,
    #[account(mut)]
    pub owner: Signer<'info>,
    pub system_program: Program<'info, System>,
}

#[derive(Accounts)]
pub struct UpdateRulesHash<'info> {
    #[account(
        mut,
        has_one = owner @ CustomError::Unauthorized
    )]
    pub routing_rules: Account<'info, RoutingRules>,
    #[account(mut)]
    pub owner: Signer<'info>,
}

#[derive(Accounts)]
pub struct ExecutePayments<'info> {
    #[account(
        has_one = owner @ CustomError::Unauthorized
    )]
    pub routing_rules: Account<'info, RoutingRules>,
    #[account(mut)]
    pub payer: Signer<'info>,

    ///CHECK: Used to transfer funds to recipients
    pub owner: UncheckedAccount<'info>,
    pub system_program: Program<'info, System>,
    // Recipient accounts are provided as remaining_accounts
}

#[derive(Accounts)]
pub struct VerifyDistribution<'info> {
    pub routing_rules: Account<'info, RoutingRules>,
    // Recipient accounts are provided as remaining_accounts
}

#[account]
pub struct RoutingRules {
    pub owner: Pubkey,
    pub rules_uri: String,         // URI pointing to the off-chain rules (IPFS, Arweave, etc.)
    pub rules_hash: [u8; 32],      // Hash of the rules for integrity verification
}

impl RoutingRules {
    // 8 bytes for discriminator + 32 bytes for owner + 
    // space for URI (estimated max 200 chars) + 32 bytes for hash
    pub const SIZE: usize = 8 + 32 + 4 + 200 + 32;
}

#[error_code]
pub enum CustomError {
    #[msg("The total percentage of rules at this level must equal 100.")]
    InvalidTotalPercentage,
    #[msg("The number of provided recipient accounts does not match the expected count.")]
    RecipientCountMismatch,
    #[msg("A math error occurred during calculation.")]
    MathError,
    #[msg("Unauthorized: Only the owner can perform this action.")]
    Unauthorized,
    #[msg("The provided recipient account does not match the expected address.")]
    RecipientMismatch, 
    #[msg("Invalid path indices provided.")]
    InvalidIndices,
    #[msg("The sum of payment amounts does not match the expected total.")]
    InvalidTotalAmount,
    #[msg("The provided rules hash does not match the stored hash.")]
    RulesHashMismatch,
}