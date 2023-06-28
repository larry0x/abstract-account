use std::str::FromStr;

use cosmwasm_std::{from_binary, Binary, Response, Storage};
use ethers::types::{Address, Signature, U256};

use crate::{error::ContractResult, msg::Credential, state::ETHEREUM_ADDRSS};

pub fn init(store: &mut dyn Storage, address_str: &String) -> ContractResult<Response> {
    // validate ethereum address
    let _ = Address::from_str(address_str)?;

    ETHEREUM_ADDRSS.save(store, address_str)?;

    Ok(Response::new()
        .add_attribute("method", "init")
        .add_attribute("ethereum_address", address_str))
}

pub fn before_tx(
    store: &dyn Storage,
    tx_bytes: &Binary,
    cred_bytes: &Binary,
) -> ContractResult<Response> {
    // load the ethereum address
    let address_str = ETHEREUM_ADDRSS.load(store)?;
    let address = Address::from_str(&address_str)?;

    // parse the ethereum signature
    let cred: Credential = from_binary(cred_bytes)?;
    let sig = Signature {
        r: U256::from_dec_str(&cred.r)?,
        s: U256::from_dec_str(&cred.s)?,
        v: cred.v,
    };

    // validate the ethereum signature
    sig.verify(tx_bytes.as_slice(), address)?;

    Ok(Response::new()
        .add_attribute("method", "before_tx"))
}

pub fn after_tx() -> ContractResult<Response> {
    Ok(Response::new()
        .add_attribute("method", "after_tx"))
}
