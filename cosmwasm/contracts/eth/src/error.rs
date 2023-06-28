
#[derive(Debug, thiserror::Error)]
pub enum ContractError {
    #[error(transparent)]
    FromHex(#[from] hex::FromHexError),

    #[error(transparent)]
    Std(#[from] cosmwasm_std::StdError),

    #[error(transparent)]
    RecoverPubkey(#[from] cosmwasm_std::RecoverPubkeyError),

    #[error("recovered pubkey doesn't match with with signer address")]
    RecoveredPubkeyMismatch,
}

pub type ContractResult<T> = Result<T, ContractError>;
