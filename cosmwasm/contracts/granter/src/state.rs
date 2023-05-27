use cw_storage_plus::Map;

use crate::msg::Grant;

// (type_url, grantee_pubkey) => grant
pub const GRANTS: Map<(&str, &[u8]), Grant> = Map::new("g");
