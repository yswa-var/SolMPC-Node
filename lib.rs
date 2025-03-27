use anchor_lang::prelude::*;

declare_id!("EM7AAngMgQPXizeuwAKaBvci79DhRxJMBYjRVoJWYEH3");

#[program]
pub mod payment_validator {
    use super::*;

    pub fn validate_payment_distribution(
        ctx: Context<ValidatePayment>, 
        total_amount: u64, 
        receivers: Vec<Pubkey>, 
        amounts: Vec<u64>
    ) -> Result<()> {
        // Validate that number of receivers matches number of amounts
        require!(
            receivers.len() == amounts.len(), 
            PaymentError::MismatchedReceiversAndAmounts
        );

        // Calculate sum of amounts
        let sum: u64 = amounts.iter().sum();

        // Validate that sum matches total amount
        require!(
            sum == total_amount, 
            PaymentError::TotalAmountMismatch
        );

        Ok(())
    }
}

#[derive(Accounts)]
pub struct ValidatePayment<'info> {
    #[account(mut)]
    pub sender: Signer<'info>,
}

#[error_code]
pub enum PaymentError {
    #[msg("Number of receivers does not match number of amounts")]
    MismatchedReceiversAndAmounts,
    #[msg("Total amount does not match sum of individual amounts")]
    TotalAmountMismatch,
}