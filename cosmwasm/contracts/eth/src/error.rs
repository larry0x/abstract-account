
#[derive(Debug, thiserror::Error)]
pub enum ContractError {
    #[error(transparent)]
    Base(#[from] account_base::error::ContractError),

    #[error(transparent)]
    FromHex(#[from] hex::FromHexError),

    #[error(transparent)]
    Std(#[from] cosmwasm_std::StdError),

    #[error(transparent)]
    RecoverPubkey(#[from] cosmwasm_std::RecoverPubkeyError),

    #[error("recovery id can only be one of 0, 1, 27, 28")]
    InvalidRecoveryId,

    #[error("recovered pubkey doesn't match with with signer address")]
    RecoveredPubkeyMismatch,
}

pub type ContractResult<T> = Result<T, ContractError>;
