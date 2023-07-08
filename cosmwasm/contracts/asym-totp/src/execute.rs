use account_base::{error::ContractError as BaseError, execute::sha256};
use cosmwasm_std::{from_binary, to_binary, Binary, DepsMut, Env, Response};

use crate::{
    error::{ContractError, ContractResult},
    msg::{Config, Credential, SignDoc},
    state::{CONFIG, PREV_COUNT},
};

pub fn init(deps: DepsMut, env: Env, cfg: &Config) -> ContractResult<Response> {
    let prev_count = env.block.time.seconds() / cfg.duration_secs - 1;

    CONFIG.save(deps.storage, cfg)?;
    PREV_COUNT.save(deps.storage, &prev_count)?;

    Ok(Response::new()
        .add_attribute("method", "init")
        .add_attribute("pubkey", cfg.pubkey.to_base64())
        .add_attribute("otp_pubkey", cfg.otp_pubkey.to_base64())
        .add_attribute("duration_secs", cfg.duration_secs.to_string())
        .add_attribute("previous_count", prev_count.to_string()))
}

pub fn before_tx(
    deps: DepsMut,
    env: Env,
    tx_bytes: &Binary,
    cred_bytes: Option<&Binary>,
    simulate: bool,
) -> ContractResult<Response> {
    let cfg = CONFIG.load(deps.storage)?;

    // bytes to be signed by the account privkey
    let tx_bytes_hash = sha256(tx_bytes);

    // bytes to be signed by the OTP privkey
    let count = env.block.time.seconds() / cfg.duration_secs;
    let sign_doc = SignDoc {
        chain_id: env.block.chain_id,
        account: env.contract.address.into(),
        count,
    };
    let sign_bytes = to_binary(&sign_doc)?;
    let sign_bytes_hash = sha256(&sign_bytes);

    // verify signautre and OTP
    // skip if in simulation mode
    if !simulate {
        let cred_bytes = cred_bytes.ok_or(BaseError::SignatureNotFound)?;
        let cred: Credential = from_binary(cred_bytes)?;

        if !deps.api.secp256k1_verify(&tx_bytes_hash, &cred.signature, &cfg.pubkey)? {
            return Err(BaseError::InvalidSignature.into());
        }

        if !deps.api.secp256k1_verify(&sign_bytes_hash, &cred.otp, &cfg.otp_pubkey)? {
            return Err(ContractError::InvalidOTP);
        }
    }

    // make sure the OTP hasn't been reused
    PREV_COUNT.update(deps.storage, |prev_count| {
        if prev_count >= count {
            return Err(ContractError::ReusedOTP);
        }

        Ok(count)
    })?;

    Ok(Response::new()
        .add_attribute("method", "before_tx")
        .add_attribute("count", count.to_string()))
}
