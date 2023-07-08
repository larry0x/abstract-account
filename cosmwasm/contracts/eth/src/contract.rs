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
    _env: Env,
    _info: MessageInfo,
    msg: InstantiateMsg,
) -> ContractResult<Response> {
    cw2::set_contract_version(deps.storage, CONTRACT_NAME, CONTRACT_VERSION)?;
    execute::init(deps.storage, &msg.ethereum_address)
}

#[entry_point]
pub fn sudo(deps: DepsMut, _env: Env, msg: AccountSudoMsg) -> ContractResult<Response> {
    match msg {
        AccountSudoMsg::BeforeTx {
            tx_bytes,
            credential,
            simulate,
            ..
        } => execute::before_tx(deps.as_ref(), &tx_bytes, credential.as_ref(), simulate),
        AccountSudoMsg::AfterTx {
            ..
        } => base::execute::after_tx().map_err(Into::into),
    }
}

#[entry_point]
pub fn execute(
    _deps: DepsMut,
    _env: Env,
    _info: MessageInfo,
    _msg: Empty,
) -> ContractResult<Response> {
    unreachable!("this contract does not have any execute method");
}

#[entry_point]
pub fn query(deps: Deps, _env: Env, msg: QueryMsg) -> StdResult<Binary> {
    match msg {
        QueryMsg::EthereumAddress {} => to_binary(&query::ethereum_address(deps.storage)?),
    }
}
