use cosmwasm_std::{
    entry_point, to_binary, Binary, Deps, DepsMut, Empty, Env, MessageInfo, Response, StdResult,
};

use absacc::AccountSudoMsg;
use account_base as base;

use crate::{
    error::ContractResult,
    execute,
    msg::{InstantiateMsg, QueryMsg},
    query,
    CONTRACT_NAME, CONTRACT_VERSION,
};

#[entry_point]
pub fn instantiate(
    deps: DepsMut,
    env:  Env,
    _:    MessageInfo,
    msg:  InstantiateMsg,
) -> ContractResult<Response> {
    cw2::set_contract_version(deps.storage, CONTRACT_NAME, CONTRACT_VERSION)?;
    execute::init(deps, env, &msg)
}

#[entry_point]
pub fn sudo(deps: DepsMut, env: Env, msg: AccountSudoMsg) -> ContractResult<Response> {
    match msg {
        AccountSudoMsg::BeforeTx {
            tx_bytes,
            cred_bytes,
            simulate,
            ..
        } => execute::before_tx(deps, env, &tx_bytes, cred_bytes.as_ref(), simulate),
        AccountSudoMsg::AfterTx {
            ..
        } => base::execute::after_tx().map_err(Into::into),
    }
}

#[entry_point]
pub fn execute(_: DepsMut, _: Env, _: MessageInfo, _: Empty) -> ContractResult<Response> {
    unreachable!("this contract does not have any execute method");
}

#[entry_point]
pub fn query(deps: Deps, _env: Env, msg: QueryMsg) -> StdResult<Binary> {
    match msg {
        QueryMsg::Config {} => to_binary(&query::config(deps.storage)?),
        QueryMsg::PreviousCount {} => to_binary(&query::previous_count(deps.storage)?),
    }
}
