#[derive(Debug, thiserror::Error)]
pub enum ContractError {
    // #[error(transparent)]
    // FromDecStr(#[from] ethers::abi::ethereum_types::FromDecStrErr),

    #[error(transparent)]
    FromHex(#[from] rustc_hex::FromHexError),

    #[error(transparent)]
    Signature(#[from] ethers_core::types::SignatureError),

    #[error(transparent)]
    Std(#[from] cosmwasm_std::StdError),
}

pub type ContractResult<T> = Result<T, ContractError>;
