//! CDK configuration.
//!
//! The CDK is configured via its TOML configuration file, `cdk.toml`
//! by default, which is deserialized into the [`Config`] struct.
use serde::Deserialize;

pub(crate) const DEFAULT_IP: std::net::Ipv4Addr = std::net::Ipv4Addr::new(0, 0, 0, 0);

pub(crate) mod aggregator;
pub(crate) mod l1;
pub mod log;
pub(crate) mod network_config;
pub(crate) mod sequence_sender;
pub(crate) mod telemetry;

pub use log::Log;
use sequence_sender::SequenceSender;

/// The Agglayer configuration.
#[derive(Deserialize, Debug)]
#[cfg_attr(any(test, feature = "testutils"), derive(Default))]
pub struct Config {
    /// The log configuration.
    #[serde(rename = "Log", default)]
    pub log: Log,

    #[serde(rename = "ForkUpgradeBatchNumber")]
    pub fork_upgrade_batch_number: Option<u64>,

    #[serde(rename = "NetworkConfig", default)]
    pub network_config: network_config::NetworkConfig,

    #[serde(rename = "Aggregator", default)]
    pub aggregator: aggregator::Aggregator,

    #[serde(rename = "SequenceSender", default)]
    pub sequence_sender: SequenceSender,
}
