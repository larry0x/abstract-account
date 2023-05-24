use cosmwasm_std::{
    entry_point, to_binary, Binary, Deps, DepsMut, Env, MessageInfo, Response, StdResult,
};

use abstract_account::AccountSudoMsg;

use crate::{
    error::ContractResult,
    execute,
    msg::{ExecuteMsg, InstantiateMsg, QueryMsg},
    query,
    CONTRACT_NAME, CONTRACT_VERSION,
};

#[entry_point]
pub fn instantiate(
    deps: DepsMut,
    _env: Env,
    _info: MessageInfo,
    msg: InstantiateMsg,
) -> ContractResult<Response> {
    cw2::set_contract_version(deps.storage, CONTRACT_NAME, CONTRACT_VERSION)?;
    execute::init(deps.storage, &msg.pubkey)
}

#[entry_point]
pub fn sudo(deps: DepsMut, _env: Env, msg: AccountSudoMsg) -> ContractResult<Response> {
    match msg {
        AccountSudoMsg::BeforeTx {
            sign_bytes,
            credential,
            ..
        } => execute::before_tx(deps.as_ref(), &sign_bytes, &credential),
        AccountSudoMsg::AfterTx {
            ..
        } => execute::after_tx(),
    }
}

#[entry_point]
pub fn execute(
    deps: DepsMut,
    env: Env,
    info: MessageInfo,
    msg: ExecuteMsg,
) -> ContractResult<Response> {
    match msg {
        ExecuteMsg::UpdatePubkey {
            new_pubkey,
        } => execute::update_pubkey(deps.storage, &info.sender, &env.contract.address, &new_pubkey),
    }
}

#[entry_point]
pub fn query(deps: Deps, _env: Env, msg: QueryMsg) -> StdResult<Binary> {
    match msg {
        QueryMsg::Pubkey {} => to_binary(&query::pubkey(deps.storage)?),
    }
}
