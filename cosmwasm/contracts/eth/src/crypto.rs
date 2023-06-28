//! Adapted from
//! - sig verification:
//!   https://github.com/gakonst/ethers-rs/blob/master/ethers-core/src/types/signature.rs
//! - hash:
//!   https://github.com/gakonst/ethers-rs/blob/master/ethers-core/src/utils/hash.rs

use cosmwasm_schema::cw_serde;
use cosmwasm_std::{Api, Uint256};
use tiny_keccak::{Hasher, Keccak};

use crate::error::{ContractError, ContractResult};

/// The credential is an Ethereum signature, which consists of three parts: r, s,
/// and v.
///
/// r and s are 256-bit unsigned integers, which are represented as Strings here.
#[cw_serde]
pub struct Credential {
    pub r: Uint256,
    pub s: Uint256,
    pub v: u64,
}

pub fn verify(api: &dyn Api, msg: &[u8], addr: &str, cred: &Credential) -> ContractResult<()> {
    let msg_hash_bytes = hash_message(msg);
    let addr_bytes = hex::decode(&addr[2..])?;

    // prepare recovery id
    let recovery_id = normalize_recovery_id(cred.v);

    // prepare recoverable signature
    let mut sig_bytes = vec![];
    sig_bytes.extend(cred.r.to_be_bytes());
    sig_bytes.extend(cred.s.to_be_bytes());

    // recover public key
    let pk_bytes = api.secp256k1_recover_pubkey(&msg_hash_bytes, &sig_bytes, recovery_id)?;

    // derive address from pubkey
    let hash = keccak256(&pk_bytes[1..]);
    let recovered_addr = &hash[12..];

    if addr_bytes != recovered_addr {
        return Err(ContractError::RecoveredPubkeyMismatch);
    }

    Ok(())
}

fn hash_message(msg: &[u8]) -> [u8; 32] {
    const PREFIX: &str = "\x19Ethereum Signed Message:\n";

    let mut bytes = vec![];
    bytes.extend_from_slice(PREFIX.as_bytes());
    bytes.extend_from_slice(msg.len().to_string().as_bytes());
    bytes.extend_from_slice(msg);

    keccak256(&bytes)
}

fn keccak256(bytes: &[u8]) -> [u8; 32] {
    let mut output = [0u8; 32];

    let mut hasher = Keccak::v256();
    hasher.update(bytes);
    hasher.finalize(&mut output);

    output
}

fn normalize_recovery_id(v: u64) -> u8 {
    match v {
        0 => 0,
        1 => 1,
        27 => 0,
        28 => 1,
        v if v >= 35 => ((v - 1) % 2) as _,
        _ => 4,
    }
}

// ----------------------------------- Test ------------------------------------

#[cfg(test)]
mod tests {
    use std::str::FromStr;

    use cosmwasm_std::testing::MockApi;

    use super::*;

    #[test]
    fn verifying_ethereum_signature() {
        let api = MockApi::default();

        // example taken from ethers-rs:
        // https://github.com/gakonst/ethers-rs/tree/master/ethers-signers#examples
        let message = "hello world";
        let address = "0x63F9725f107358c9115BC9d86c72dD5823E9B1E6";
        let cred = Credential {
            r: Uint256::from_str("49684349367057865656909429001867135922228948097036637749682965078859417767352").unwrap(),
            s: Uint256::from_str("26715700564957864553985478426289223220394026033170102795835907481710471636815").unwrap(),
            v: 28,
        };

        let res = verify(&api, message.as_bytes(), address, &cred);
        assert!(res.is_ok());

        // let's try an invalid case
        // we simply change the address to a different one
        let wrong_address = "0xd8dA6BF26964aF9D7eEd9e03E53415D37aA96045";

        let res = verify(&api, message.as_bytes(), wrong_address, &cred);
        assert!(res.is_err());
    }
}

