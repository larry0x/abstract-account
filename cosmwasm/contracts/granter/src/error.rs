#[derive(Debug, thiserror::Error)]
pub enum ContractError {
    #[error(transparent)]
    Base(#[from] account_base::error::ContractError),

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

    #[error("cannot create a new grant that has already expired")]
    NewGrantExpired,
}

pub type ContractResult<T> = Result<T, ContractError>;
