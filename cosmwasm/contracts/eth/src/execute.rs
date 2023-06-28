use std::str::FromStr;

use cosmwasm_std::{from_binary, Binary, Response, Storage};
use ethers_core::types::{Address, Signature, U256};

use crate::{error::ContractResult, msg::Credential, state::ETHEREUM_ADDRESS};

pub fn init(store: &mut dyn Storage, address_str: &String) -> ContractResult<Response> {
    // validate ethereum address
    let _ = Address::from_str(address_str)?;

    ETHEREUM_ADDRESS.save(store, address_str)?;

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
    let address_str = ETHEREUM_ADDRESS.load(store)?;
    let address = Address::from_str(&address_str)?;

    // parse the ethereum signature
    let cred: Credential = from_binary(cred_bytes)?;
    let sig = Signature {
        r: U256::from_dec_str(&cred.r).unwrap(),
        s: U256::from_dec_str(&cred.s).unwrap(),
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

// ----------------------------------- Test ------------------------------------

#[cfg(test)]
mod tests {
    use cosmwasm_std::{testing::mock_dependencies, to_binary};

    use super::*;

    #[test]
    fn verifying_ethereum_signature() {
        let mut deps = mock_dependencies();

        // examples taken from ethers-rs example:
        // https://github.com/gakonst/ethers-rs/tree/master/ethers-signers#examples
        let message = "hello world";
        let address = "0x63F9725f107358c9115BC9d86c72dD5823E9B1E6";
        let cred = Credential {
            r: "49684349367057865656909429001867135922228948097036637749682965078859417767352".into(),
            s: "26715700564957864553985478426289223220394026033170102795835907481710471636815".into(),
            v: 28,
        };

        ETHEREUM_ADDRESS.save(deps.as_mut().storage, &address.into()).unwrap();

        let res = before_tx(
            deps.as_ref().storage,
            &message.as_bytes().into(),
            &to_binary(&cred).unwrap(),
        );
        assert!(res.is_ok());

        // let's try an invalid case
        // we simply change the address to a different one
        let wrong_addrss = "0xd8dA6BF26964aF9D7eEd9e03E53415D37aA96045";

        ETHEREUM_ADDRESS.save(deps.as_mut().storage, &wrong_addrss.into()).unwrap();

        let res = before_tx(
            deps.as_ref().storage,
            &message.as_bytes().into(),
            &to_binary(&cred).unwrap(),
        );
        assert!(res.is_err());
    }
}
