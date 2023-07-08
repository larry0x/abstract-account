use cosmwasm_schema::{cw_serde, QueryResponses};
use cosmwasm_std::Binary;

#[cw_serde]
pub struct Config {
    /// Public key that will sign the tx
    pub pubkey: Binary,
    /// Public key that will sign the OTP SignDoc (defined later in this file)
    pub otp_pubkey: Binary,
    /// The OTP time duration, in seconds
    pub duration_secs: u64,
}

/// The tx sender must sign the JSON-encoded string of this data using the OTP
/// private key.
#[cw_serde]
pub struct SignDoc {
    /// Identifies which chain this OTP is intended for
    pub chain_id: String,
    /// Address of the AbstractAccount that this OTP is used to authenticate
    pub account: String,
    /// How many OTP durations has passed since the UNIX epoch
    pub count: u64,
}

/// The tx sender must provide the JSON-encoded string of this data in the tx's
/// signature field.
#[cw_serde]
pub struct Credential {
    pub signature: Binary,
    pub otp:       Binary,
}

pub type InstantiateMsg = Config;

#[cw_serde]
#[derive(QueryResponses)]
pub enum QueryMsg {
    #[returns(Config)]
    Config {},

    #[returns(u64)]
    PreviousCount {},
}
