use anchor_lang::prelude::*;

declare_id!("3WwkA4gXYDR1fQ1BHfe5J4XoFYZ5kdSPe2kGkDE3uC22");

#[program]
pub mod distribution {
    use super::*;

    /// Initializes the distribution system with business rules, receivers, and sub-tilts as strings.
    pub fn initialize(
        ctx: Context<Initialize>,
        business_rules: [u8; 10],
        receivers: [Receiver; 10],
        sub_tilts: Vec<String>,
    ) -> Result<()> {
        // Validate that business rules sum to 100
        let total: u32 = business_rules.iter().map(|&x| x as u32).sum();
        require!(total == 100, ErrorCode::InvalidBusinessRules);

        // Set up DistributionAccount
        let distribution = &mut ctx.accounts.distribution;
        distribution.sender = ctx.accounts.sender.key();
        distribution.business_rules = business_rules;
        distribution.is_initialized = true;

        // Set up ReceiverList
        let receiver_list = &mut ctx.accounts.receiver_list;
        receiver_list.receivers = receivers;

        // Set up SubTiltList
        let sub_tilt_list = &mut ctx.accounts.sub_tilt_list;
        sub_tilt_list.sub_tilts = sub_tilts;

        Ok(())
    }

    /// Distributes funds based on the stored receivers and logs sub-tilts.
    pub fn distribute(ctx: Context<Distribute>) -> Result<()> {
        let distribution = &ctx.accounts.distribution;

        // Validate initialization and sender
        require!(
            distribution.is_initialized,
            ErrorCode::AccountNotInitialized
        );
        require!(
            distribution.sender == ctx.accounts.sender.key(),
            ErrorCode::InvalidSender
        );

        // Log transfers for receivers
        for receiver in &ctx.accounts.receiver_list.receivers {
            msg!("Transferring {} to {}", receiver.amount, receiver.pubkey);
        }

        // Log sub-tilts
        for sub_tilt in &ctx.accounts.sub_tilt_list.sub_tilts {
            msg!("SubTilt Entry: {}", sub_tilt);
        }

        Ok(())
    }
}

// --- Account Structures ---

#[derive(Accounts)]
pub struct Initialize<'info> {
    #[account(
    init_if_needed, // This will only create the account if it doesn't exist
    payer = sender,
    space = 8 + 32 + 10 + 1,
    seeds = [b"distribution_1", sender.key().as_ref()],
    bump
)]
    pub distribution: Account<'info, DistributionAccount>,

    #[account(
    init_if_needed,
    payer = sender,
    space = 8 + 10 * (32 + 8),
    seeds = [b"receiver_list", sender.key().as_ref()],
    bump
)]
    pub receiver_list: Account<'info, ReceiverList>,

    #[account(
    init_if_needed,
    payer = sender,
    space = 8 + 1024,
    seeds = [b"sub_tilt_list", sender.key().as_ref()],
    bump
)]
    pub sub_tilt_list: Account<'info, SubTiltList>,

    #[account(mut)]
    pub sender: Signer<'info>,
    pub system_program: Program<'info, System>,
}

#[derive(Accounts)]
pub struct Distribute<'info> {
    #[account(seeds = [b"distribution", sender.key().as_ref()], bump)]
    pub distribution: Account<'info, DistributionAccount>,

    #[account(seeds = [b"receiver_list", sender.key().as_ref()], bump)]
    pub receiver_list: Account<'info, ReceiverList>,

    #[account(seeds = [b"sub_tilt_list", sender.key().as_ref()], bump)]
    pub sub_tilt_list: Account<'info, SubTiltList>,

    pub sender: Signer<'info>,
    pub system_program: Program<'info, System>,
}

#[account]
pub struct DistributionAccount {
    pub sender: Pubkey,
    pub business_rules: [u8; 10],
    pub is_initialized: bool,
}

#[account]
pub struct ReceiverList {
    pub receivers: [Receiver; 10],
}

#[account]
pub struct SubTiltList {
    pub sub_tilts: Vec<String>,
}

// --- Data Structures ---

#[derive(AnchorSerialize, AnchorDeserialize, Clone)]
pub struct Receiver {
    pub pubkey: Pubkey,
    pub amount: u64,
}

// --- Error Codes ---

#[error_code]
pub enum ErrorCode {
    #[msg("Business rules must sum to 100")]
    InvalidBusinessRules,
    #[msg("Account not initialized")]
    AccountNotInitialized,
    #[msg("Invalid sender")]
    InvalidSender,
}
