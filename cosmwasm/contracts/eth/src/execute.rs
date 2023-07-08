use cosmwasm_std::{Binary, Response, Storage, Deps};

use account_base::error::ContractError as BaseError;

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
    sig_bytes: Option<&Binary>,
    simulate:  bool,
) -> ContractResult<Response> {
    // load the ethereum address
    let addr_str = ETHEREUM_ADDRESS.load(deps.storage)?;
    let addr_bytes = hex::decode(&addr_str[2..])?;

    // validate the ethereum signature
    // skip if in simulation mode
    if !simulate {
        let sig_bytes = sig_bytes.ok_or(BaseError::SignatureNotFound)?;
        crypto::verify(deps.api, tx_bytes, sig_bytes, &addr_bytes)?;
    }

    Ok(Response::new()
        .add_attribute("method", "before_tx"))
}
