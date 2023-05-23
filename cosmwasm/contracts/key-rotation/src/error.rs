#[derive(Debug, thiserror::Error)]
pub enum ContractError {
    #[error(transparent)]
    Std(#[from] cosmwasm_std::StdError),

    #[error(transparent)]
    Verification(#[from] cosmwasm_std::VerificationError),

    #[error("signature is invalid")]
    InvalidSignature,

    #[error("pubkey included in the contract does not match the account")]
    PubKeyMismatch,

    #[error("only the contract itself can call this method")]
    Unauthorized,
}

pub type ContractResult<T> = Result<T, ContractError>;
