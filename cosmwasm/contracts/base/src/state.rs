use cosmwasm_std::Binary;
use cw_storage_plus::Item;

pub const PUBKEY: Item<Binary> = Item::new("pk");
