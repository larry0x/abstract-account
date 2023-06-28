use cosmwasm_schema::{cw_serde, QueryResponses};

/// The credential is an Ethereum signature, which consists of three parts: r, s,
/// and v.
///
/// r and s are 256-bit unsigned integers, which are represented as Strings here.
#[cw_serde]
pub struct Credential {
    pub r: String,
    pub s: String,
    pub v: u64,
}

#[cw_serde]
pub struct InstantiateMsg {
    /// Address of the Ethereum wallet that controls this account, in hex encoding
    pub ethereum_address: String,
}

#[cw_serde]
#[derive(QueryResponses)]
pub enum QueryMsg {
    #[returns(String)]
    EthereumAddress {},
}
