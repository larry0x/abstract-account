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
    deps:      Deps,
    tx_bytes:  &Binary,
    signature: Option<&Binary>,
    simulate:  bool,
) -> ContractResult<Response> {
    let tx_bytes_hash = sha256(tx_bytes);
    let pubkey = PUBKEY.load(deps.storage)?;

    // skip the signature validation in simulation mode
    if !simulate {
        let Some(sig_bytes) = signature else {
            return Err(ContractError::SignatureNotFound);
        };

        if !deps.api.secp256k1_verify(&tx_bytes_hash, sig_bytes, &pubkey)? {
            return Err(ContractError::InvalidSignature);
        }
    }

    Ok(Response::new()
        .add_attribute("method", "before_tx"))
}

pub fn after_tx() -> ContractResult<Response> {
    Ok(Response::new()
        .add_attribute("method", "after_tx"))
}

// this function is not used in this base contract directly, but is used by
// several other account contracts that extend base, so we put it here
pub fn assert_self(sender: &Addr, contract: &Addr) -> ContractResult<()> {
    if sender != contract {
        return Err(ContractError::Unauthorized);
    }

    Ok(())
}

pub fn sha256(msg: &[u8]) -> Vec<u8> {
    let mut hasher = Sha256::new();
    hasher.update(msg);
    hasher.finalize().to_vec()
}
