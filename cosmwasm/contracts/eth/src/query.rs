use cosmwasm_std::{Storage, StdResult};

use crate::state::ETHEREUM_ADDRSS;

pub fn ethereum_address(store: &dyn Storage) -> StdResult<String> {
    ETHEREUM_ADDRSS.load(store)
}
