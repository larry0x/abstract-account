#[derive(Debug, thiserror::Error)]
pub enum ContractError {
    #[error(transparent)]
    Base(#[from] account_base::error::ContractError),

    #[error(transparent)]
    Std(#[from] cosmwasm_std::StdError),

    #[error(transparent)]
    Verification(#[from] cosmwasm_std::VerificationError),

    #[error("invalid one-time password")]
    InvalidOTP,

    #[error("reused one-time password")]
    ReusedOTP,
}

pub type ContractResult<T> = Result<T, ContractError>;
