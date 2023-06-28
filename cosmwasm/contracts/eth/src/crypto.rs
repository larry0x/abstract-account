//! An Ethereum signature has a total length of 65 parts, consisting of three
//! parts:
//! - r: 32 bytes
//! - s: 32 bytes
//! - v: 1 byte
//!
//! r and s together are known as the recoverable signature. v is known as the
//! recovery id, which can take the value of one of 0, 1, 27, and 28.
//!
//! In order to verify a signature, we attempt to recover the signer's pubkey.
//! If the recovered key matches the signer's address, we consider the signature
//! valid.
//!
//! The address is the last 20 bytes of the hash keccak256(pubkey_bytes).
//!
//! Before a message is signed, it is prefixed with the bytes: b"\x19Ethereum Signed Message:\n".
//!
//! Adapted from
//! - sig verification:
//!   https://github.com/gakonst/ethers-rs/blob/master/ethers-core/src/types/signature.rs
//! - hash:
//!   https://github.com/gakonst/ethers-rs/blob/master/ethers-core/src/utils/hash.rs

use cosmwasm_std::Api;
use tiny_keccak::{Hasher, Keccak};

use crate::error::{ContractError, ContractResult};

pub fn verify(
    api:        &dyn Api,
    msg_bytes:  &[u8],
    sig_bytes:  &[u8],
    addr_bytes: &[u8],
) -> ContractResult<()> {
    let msg_hash_bytes = hash_message(msg_bytes);

    let recoverable_sig = &sig_bytes[..64];
    let recovery_id = normalize_recovery_id(sig_bytes[64])?;

    let pk_bytes = api.secp256k1_recover_pubkey(&msg_hash_bytes, recoverable_sig, recovery_id)?;

    let hash = keccak256(&pk_bytes[1..]);
    let recovered_addr = &hash[12..];

    if addr_bytes == recovered_addr {
        Ok(())
    } else {
        Err(ContractError::RecoveredPubkeyMismatch)
    }
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

fn normalize_recovery_id(id: u8) -> ContractResult<u8> {
    match id {
        0 | 1 => Ok(id),
        27 => Ok(0),
        28 => Ok(1),
        _ => Err(ContractError::InvalidRecoveryId),
    }
}

// ----------------------------------- Test ------------------------------------

#[cfg(test)]
mod tests {
    use std::str::FromStr;

    use cosmwasm_std::{testing::MockApi, Uint256};

    use super::*;

    #[test]
    fn verifying_ethereum_signature() {
        let api = MockApi::default();

        // example taken from ethers-rs:
        // https://github.com/gakonst/ethers-rs/tree/master/ethers-signers#examples
        let message = "hello world";
        let address = "0x63F9725f107358c9115BC9d86c72dD5823E9B1E6";

        let r = Uint256::from_str("49684349367057865656909429001867135922228948097036637749682965078859417767352").unwrap();
        let s = Uint256::from_str("26715700564957864553985478426289223220394026033170102795835907481710471636815").unwrap();
        let v = 28u8;

        let mut sig = vec![];
        sig.extend(r.to_be_bytes());
        sig.extend(s.to_be_bytes());
        sig.push(v);
        assert_eq!(sig.len(), 65);

        let address_bytes = hex::decode(&address[2..]).unwrap();
        let res = verify(&api, message.as_bytes(), &sig, &address_bytes);
        assert!(res.is_ok());

        // let's try an invalid case
        // we simply change the address to a different one
        let wrong_address = "0xd8dA6BF26964aF9D7eEd9e03E53415D37aA96045";

        let address_bytes = hex::decode(&wrong_address[2..]).unwrap();
        let res = verify(&api, message.as_bytes(), &sig, &address_bytes);
        assert!(res.is_err());
    }
}
