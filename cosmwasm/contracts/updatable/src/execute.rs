use cosmwasm_std::{Addr, Binary, Response, Storage};

use account_base::{error::ContractResult, execute::assert_self, state::PUBKEY};

pub fn update_pubkey(
    store:      &mut dyn Storage,
    sender:     &Addr,
    contract:   &Addr,
    new_pubkey: &Binary,
) -> ContractResult<Response> {
    // only the account itself can update its pubkey
    assert_self(sender, contract)?;

    PUBKEY.save(store, new_pubkey)?;

    Ok(Response::new()
        .add_attribute("method", "update_pubkey")
        .add_attribute("new_pubkey", new_pubkey.to_base64()))
}
