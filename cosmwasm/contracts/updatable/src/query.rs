use cosmwasm_std::{Binary, StdResult, Storage};

use crate::state::PUBKEY;

pub fn pubkey(store: &dyn Storage) -> StdResult<Binary> {
    PUBKEY.load(store)
}
