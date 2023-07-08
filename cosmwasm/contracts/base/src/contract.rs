use cosmwasm_std::{
    entry_point, to_binary, Binary, Deps, DepsMut, Empty, Env, MessageInfo, Response, StdResult,
};

use absacc::AccountSudoMsg;

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
    _:    Env,
    _:    MessageInfo,
    msg:  InstantiateMsg,
) -> ContractResult<Response> {
    cw2::set_contract_version(deps.storage, CONTRACT_NAME, CONTRACT_VERSION)?;
    execute::init(deps.storage, &msg.pubkey)
}

#[entry_point]
pub fn sudo(deps: DepsMut, _env: Env, msg: AccountSudoMsg) -> ContractResult<Response> {
    match msg {
        AccountSudoMsg::BeforeTx {
            tx_bytes,
            cred_bytes,
            simulate,
            ..
        } => execute::before_tx(deps.as_ref(), &tx_bytes, cred_bytes.as_ref(), simulate),
        AccountSudoMsg::AfterTx {
            ..
        } => execute::after_tx(),
    }
}

#[entry_point]
pub fn execute(_: DepsMut, _: Env, _: MessageInfo, _: Empty) -> ContractResult<Response> {
    unreachable!("this contract does not have any execute method");
}

#[entry_point]
pub fn query(deps: Deps, _env: Env, msg: QueryMsg) -> StdResult<Binary> {
    match msg {
        QueryMsg::Pubkey {} => to_binary(&query::pubkey(deps.storage)?),
    }
}
