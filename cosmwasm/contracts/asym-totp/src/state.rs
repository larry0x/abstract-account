use cw_storage_plus::Item;

use crate::msg::Config;

pub const CONFIG: Item<Config> = Item::new("config");

/// The count value the last time an OTP was used. To prevent reuse of the OTP,
/// only one tx can be submitted per duration. To enforce this, we require the
/// count value of each tx must be strictly > prev_count.
pub const PREV_COUNT: Item<u64> = Item::new("prev_count");
