use cosmwasm_schema::{cw_serde, QueryResponses};
use cosmwasm_std::Binary;
use cw_utils::Expiration;

#[cw_serde]
pub struct Grant {
    pub expiry: Option<Expiration>,
}

#[cw_serde]
pub struct Credential {
    pub pubkey:    Binary,
    pub signature: Binary,
}

#[cw_serde]
pub enum ExecuteMsg {
    /// Grant another pubkey the ability to sign messages on this account's
    /// behalf.
    ///
    /// Only callable by the account itself.
    Grant {
        type_url: String,
        grantee:  Binary,
        expiry:   Option<Expiration>,
    },

    /// Revoke a grant that has been given to a grantee.
    ///
    /// Only callable by the account itself.
    Revoke {
        type_url: String,
        grantee:  Binary,
    },
}

#[cw_serde]
#[derive(QueryResponses)]
pub enum QueryMsg {
    /// Query the pubkey associated with this account.
    #[returns(Binary)]
    Pubkey {},

    /// Query a single grant
    #[returns(GrantResponse)]
    Grant {
        type_url: String,
        grantee:  Binary,
    },

    /// List all grants of all message types and all grantees
    #[returns(Vec<GrantResponse>)]
    Grants {
        start_after: Option<GrantPaginationParam>,
        limit:       Option<u32>,
    },
}

#[cw_serde]
pub struct GrantPaginationParam {
    pub type_url: String,
    pub grantee:  Binary,
}

#[cw_serde]
pub struct GrantResponse {
    pub type_url: String,
    pub grantee:  Binary,
    pub expiry:   Option<Expiration>,
}
