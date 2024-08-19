//! CDK configuration.
//!
//! The CDK is configured via its TOML configuration file, `cdk.toml`
//! by default, which is deserialized into the [`Config`] struct.
use serde::Deserialize;

pub(crate) const DEFAULT_IP: std::net::Ipv4Addr = std::net::Ipv4Addr::new(0, 0, 0, 0);

pub(crate) mod layer1;
pub mod log;
pub(crate) mod telemetry;

pub use layer1::Layer1;
pub use log::Log;

/// The Agglayer configuration.
#[derive(Deserialize, Debug)]
#[cfg_attr(any(test, feature = "testutils"), derive(Default))]
pub struct Config {
    /// A map of Zkevm node RPC endpoints for each rollup.
    ///
    /// The log configuration.
    #[serde(rename = "Log")]
    pub log: Log,

    #[serde(rename = "ForkUpgradeBatchNumber")]
    pub fork_upgrade_batch_number: Option<u64>,
}
