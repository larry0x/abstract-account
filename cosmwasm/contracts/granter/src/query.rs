use cosmwasm_std::{Binary, Order, StdResult, Storage};
use cw_storage_plus::Bound;

use crate::{
    msg::{GrantPaginationParam, GrantResponse},
    state::GRANTS,
};

pub const DEFAULT_LIMIT: u32 = 10;
pub const MAX_LIMIT: u32 = 30;

pub fn grant(store: &dyn Storage, type_url: String, grantee: Binary) -> StdResult<GrantResponse> {
    let grant = GRANTS.load(store, (&type_url, &grantee))?;
    Ok(GrantResponse {
        type_url,
        grantee,
        expiry: grant.expiry,
    })
}

pub fn grants(
    store: &dyn Storage,
    start_after: Option<GrantPaginationParam>,
    limit: Option<u32>,
) -> StdResult<Vec<GrantResponse>> {
    let start = start_after
        .as_ref()
        .map(|p| Bound::exclusive((p.type_url.as_str(), p.grantee.as_slice())));
    let limit = limit.unwrap_or(DEFAULT_LIMIT).min(MAX_LIMIT) as usize;

    GRANTS
        .range(store, start, None, Order::Ascending)
        .take(limit)
        .map(|item| {
            let ((type_url, grantee), grant) = item?;
            Ok(GrantResponse {
                type_url,
                grantee: grantee.into(),
                expiry: grant.expiry,
            })
        })
        .collect()
}
