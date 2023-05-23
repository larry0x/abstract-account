use cosmwasm_std::Binary;
use cw_storage_plus::{Item, Map};

use crate::msg::Grant;

pub const PUBKEY: Item<Binary> = Item::new("pk");

// (type_url, grantee_pubkey) => grant
pub const GRANTS: Map<(&str, &[u8]), Grant> = Map::new("g");
