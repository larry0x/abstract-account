use cosmwasm_std::{Binary, Deps, Response, Storage, DepsMut, Env, MessageInfo, BlockInfo, from_binary};
use cw_utils::Expiration;

use abstract_account::Any;
use account_base::{
    error::ContractError as BaseError,
    execute::{assert_self, sha256},
    state::PUBKEY,
};

use crate::{
    error::{ContractError, ContractResult},
    msg::{Credential, Grant},
    state::GRANTS,
};

pub fn before_tx(
    deps: Deps,
    block: &BlockInfo,
    msgs: &[Any],
    sign_bytes: &Binary,
    credential_bytes: &Binary,
) -> ContractResult<Response> {
    let sign_bytes_hash = sha256(sign_bytes);
    let pubkey = PUBKEY.load(deps.storage)?;

    let credential: Credential = from_binary(credential_bytes)?;

    // if the signature is signed by the account's own pubkey, the simply move
    // on to verify the signature
    // if it's signed by another pubkey, we need to make sure that this pubkey
    // has, for each message involved, a non-expired grant to send it
    let signer_is_self = credential.pubkey == pubkey;

    if !signer_is_self {
        assert_has_grant(deps.storage, block, msgs, &credential.pubkey)?;
    }

    if !deps.api.secp256k1_verify(&sign_bytes_hash, &credential.signature, &credential.pubkey)? {
        return Err(BaseError::InvalidSignature.into());
    }

    Ok(Response::new()
        .add_attribute("method", "before_tx")
        .add_attribute("signer_is_self", signer_is_self.to_string())
        .add_attribute("signer", credential.pubkey.to_base64()))
}

pub fn grant(
    deps: DepsMut,
    env: Env,
    info: MessageInfo,
    type_url: String,
    grantee: Binary,
    expiry: Option<Expiration>,
) -> ContractResult<Response> {
    // only the account itself can make grants
    assert_self(&info.sender, &env.contract.address)?;

    // the grant can't be already expired
    if let Some(expiry) = expiry.as_ref() {
        if expiry.is_expired(&env.block) {
            return Err(ContractError::NewGrantExpired);
        }
    }

    GRANTS.save(deps.storage, (&type_url, &grantee), &Grant { expiry })?;

    Ok(Response::new()
        .add_attribute("method", "grant")
        .add_attribute("granter", env.contract.address)
        .add_attribute("grantee", grantee.to_base64())
        .add_attribute("type_url", type_url))
}

pub fn revoke(
    deps: DepsMut,
    env: Env,
    info: MessageInfo,
    type_url: String,
    grantee: Binary,
) -> ContractResult<Response> {
    // only the account itself can revoke grants
    assert_self(&info.sender, &env.contract.address)?;

    GRANTS.remove(deps.storage, (&type_url, &grantee));

    Ok(Response::new()
        .add_attribute("method", "revoke")
        .add_attribute("granter", env.contract.address)
        .add_attribute("grantee", grantee.to_base64())
        .add_attribute("type_url", type_url))
}

fn assert_has_grant(
    store: &dyn Storage,
    block: &BlockInfo,
    msgs: &[Any],
    grantee: &Binary,
) -> ContractResult<()> {
    for msg in msgs {
        let Some(grant) = GRANTS.may_load(store, (&msg.type_url, grantee))? else {
            return Err(ContractError::GrantNotFound {
                type_url: msg.type_url.clone(),
                grantee: grantee.to_base64(),
            });
        };

        if let Some(expiry) = grant.expiry {
            if expiry.is_expired(block) {
                return Err(ContractError::GrantExpired {
                    type_url: msg.type_url.clone(),
                    grantee: grantee.to_base64(),
                });
            }
        }
    }

    Ok(())
}
