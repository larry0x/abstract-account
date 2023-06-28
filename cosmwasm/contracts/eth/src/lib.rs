#[cfg(not(feature = "library"))]
pub mod contract;
pub mod crypto;
pub mod error;
pub mod execute;
pub mod msg;
pub mod query;
pub mod state;

pub const CONTRACT_NAME: &str = "account-eth";
pub const CONTRACT_VERSION: &str = env!("CARGO_PKG_VERSION");
