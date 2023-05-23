use cosmwasm_schema::cw_serde;
use cosmwasm_std::Binary;

#[cw_serde]
pub struct StargateMsg {
    pub type_url: String,
    pub value:    Binary,
}

/// Any contract must implement this sudo message (both variants) in order to
/// qualify as an abstract account.
#[cw_serde]
pub enum AccountSudoMsg {
    /// Called by the AnteHandler's BeforeTxDecorator before a tx is executed.
    ///
    /// The account is provided with the messages (may need to be unmarshaled)
    /// and the credentials (signer's pubkey, sign bytes, and signature) which
    /// the account may use to determine whether the tx is authenticated.
    BeforeTx {
        msgs:       Vec<StargateMsg>,
        pubkey:     Option<Binary>,
        sign_bytes: Binary,
        signature:  Binary,
    },

    /// Called by the PostHandler's AfterTxDecorator after the tx is executed.
    ///
    /// The account is informed whether the tx had been successful.
    AfterTx {
        success: bool,
    },
}
