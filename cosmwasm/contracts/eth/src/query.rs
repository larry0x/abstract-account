use cosmwasm_std::{Storage, StdResult};

use crate::state::ETHEREUM_ADDRESS;

pub fn ethereum_address(store: &dyn Storage) -> StdResult<String> {
    ETHEREUM_ADDRESS.load(store)
}
