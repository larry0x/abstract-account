use cosmwasm_std::{Addr, Binary, Deps, Response, Storage};
use sha2::{Digest, Sha256};

use crate::{
    error::{ContractError, ContractResult},
    state::PUBKEY,
};

pub fn init(store: &mut dyn Storage, pubkey: &Binary) -> ContractResult<Response> {
    PUBKEY.save(store, pubkey)?;

    Ok(Response::new()
        .add_attribute("method", "init")
        .add_attribute("pubkey", pubkey.to_base64()))
}

pub fn before_tx(
    deps: Deps,
    pubkey: Option<&Binary>,
    sign_bytes: &Binary,
    signature: &Binary,
) -> ContractResult<Response> {
    let sign_bytes_hash = sha256(sign_bytes);
    let self_pubkey = PUBKEY.load(deps.storage)?;
    let pubkey = pubkey.unwrap_or(&self_pubkey);

    if *pubkey != self_pubkey {
        return Err(ContractError::PubKeyMismatch);
    }

    if !deps.api.secp256k1_verify(&sign_bytes_hash, signature, &self_pubkey)? {
        return Err(ContractError::InvalidSignature);
    }

    Ok(Response::new()
        .add_attribute("method", "before_tx"))
}

pub fn after_tx() -> ContractResult<Response> {
    Ok(Response::new()
        .add_attribute("method", "after_tx"))
}

pub fn update_pubkey(
    store: &mut dyn Storage,
    sender: &Addr,
    contract: &Addr,
    new_pubkey: &Binary,
) -> ContractResult<Response> {
    // only the account itself can update its pubkey
    assert_self(sender, contract)?;

    PUBKEY.save(store, new_pubkey)?;

    Ok(Response::new()
        .add_attribute("method", "update_pubkey")
        .add_attribute("new_pubkey", new_pubkey.to_base64()))
}

fn assert_self(sender: &Addr, contract: &Addr) -> ContractResult<()> {
    if sender != contract {
        return Err(ContractError::Unauthorized);
    }

    Ok(())
}

fn sha256(msg: &[u8]) -> Vec<u8> {
    let mut hasher = Sha256::new();
    hasher.update(msg);
    hasher.finalize().to_vec()
}
