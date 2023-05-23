use cosmwasm_schema::{cw_serde, QueryResponses};
use cosmwasm_std::Binary;

#[cw_serde]
pub struct InstantiateMsg {
    pub pubkey: Binary,
}

#[cw_serde]
pub enum ExecuteMsg {
    /// Change the pubkey associated with this account.
    ///
    /// Only callable by the account itself.
    UpdatePubkey {
        new_pubkey: Binary,
    },
}

#[cw_serde]
#[derive(QueryResponses)]
pub enum QueryMsg {
    /// Query the pubkey associated with this account.
    #[returns(Binary)]
    Pubkey {},
}
