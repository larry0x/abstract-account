use cosmwasm_std::{from_binary, Binary, Response, Storage, Deps};

use crate::{crypto, error::ContractResult, state::ETHEREUM_ADDRESS};

pub fn init(store: &mut dyn Storage, address_str: &String) -> ContractResult<Response> {
    ETHEREUM_ADDRESS.save(store, address_str)?;

    Ok(Response::new()
        .add_attribute("method", "init")
        .add_attribute("ethereum_address", address_str))
}

pub fn before_tx(
    deps: Deps,
    tx_bytes: &Binary,
    cred_bytes: &Binary,
) -> ContractResult<Response> {
    // load the ethereum address
    let address_str = ETHEREUM_ADDRESS.load(deps.storage)?;

    // parse the ethereum signature
    let cred: crypto::Credential = from_binary(cred_bytes)?;

    // validate the ethereum signature
    crypto::verify(deps.api, tx_bytes.as_slice(), &address_str, &cred)?;

    Ok(Response::new()
        .add_attribute("method", "before_tx"))
}

pub fn after_tx() -> ContractResult<Response> {
    Ok(Response::new()
        .add_attribute("method", "after_tx"))
}
