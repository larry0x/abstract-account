use cosmwasm_std::{Binary, Response, Storage, Deps};

use crate::{crypto, error::ContractResult, state::ETHEREUM_ADDRESS};

pub fn init(store: &mut dyn Storage, addr_str: &String) -> ContractResult<Response> {
    ETHEREUM_ADDRESS.save(store, addr_str)?;

    Ok(Response::new()
        .add_attribute("method", "init")
        .add_attribute("ethereum_address", addr_str))
}

pub fn before_tx(
    deps:      Deps,
    tx_bytes:  &Binary,
    sig_bytes: &Binary,
) -> ContractResult<Response> {
    // load the ethereum address
    let addr_str = ETHEREUM_ADDRESS.load(deps.storage)?;
    let addr_bytes = hex::decode(&addr_str[2..])?;

    // validate the ethereum signature
    crypto::verify(deps.api, &tx_bytes, &sig_bytes, &addr_bytes)?;

    Ok(Response::new()
        .add_attribute("method", "before_tx"))
}
