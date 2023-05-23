#[derive(Debug, thiserror::Error)]
pub enum ContractError {
    #[error(transparent)]
    Std(#[from] cosmwasm_std::StdError),

    #[error(transparent)]
    Verification(#[from] cosmwasm_std::VerificationError),

    #[error("grant found for grantee `{grantee}` and type_url `{type_url}`")]
    GrantExpired {
        type_url: String,
        grantee: String,
    },

    #[error("no grant found for grantee `{grantee}` and type_url `{type_url}`")]
    GrantNotFound {
        type_url: String,
        grantee: String,
    },

    #[error("signature is invalid")]
    InvalidSignature,

    #[error("cannot create a new grant that has already expired")]
    NewGrantExpired,

    #[error("only the contract itself can call this method")]
    Unauthorized,
}

pub type ContractResult<T> = Result<T, ContractError>;
