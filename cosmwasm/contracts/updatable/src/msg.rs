use cosmwasm_schema::cw_serde;
use cosmwasm_std::Binary;

#[cw_serde]
pub enum ExecuteMsg {
    /// Change the pubkey associated with this account.
    ///
    /// Only callable by the account itself.
    UpdatePubkey {
        new_pubkey: Binary,
    },
}
