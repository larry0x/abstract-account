#[derive(Debug, thiserror::Error)]
pub enum ContractError {
    #[error(transparent)]
    Ecdsa(#[from] k256::ecdsa::Error),

    #[error(transparent)]
    FromHex(#[from] hex::FromHexError),

    #[error(transparent)]
    Std(#[from] cosmwasm_std::StdError),

    #[error("invalid ethereum signature")]
    InvalidSignature,
}

pub type ContractResult<T> = Result<T, ContractError>;
