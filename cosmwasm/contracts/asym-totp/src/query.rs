use cosmwasm_std::{StdResult, Storage};

use crate::{
    msg::Config,
    state::{CONFIG, PREV_COUNT},
};

pub fn config(store: &dyn Storage) -> StdResult<Config> {
    CONFIG.load(store)
}

pub fn previous_count(store: &dyn Storage) -> StdResult<u64> {
    PREV_COUNT.load(store)
}
