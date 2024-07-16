//! CDK configuration.
//!
//! The CDK is configured via its TOML configuration file, `cdk.toml`
//! by default, which is deserialized into the [`Config`] struct.
use serde::Deserialize;

pub(crate) const DEFAULT_IP: std::net::Ipv4Addr = std::net::Ipv4Addr::new(0, 0, 0, 0);

pub(crate) mod auth;
pub(crate) mod layer1;
pub mod log;
pub(crate) mod rpc;
pub(crate) mod telemetry;

pub use auth::{AuthConfig, GcpKmsConfig, LocalConfig, PrivateKey};
pub use layer1::Layer1;
pub use log::Log;
pub use rpc::RpcConfig;

/// The Agglayer configuration.
#[derive(Deserialize, Debug)]
#[cfg_attr(any(test, feature = "testutils"), derive(Default))]
pub struct Config {
    /// A map of Zkevm node RPC endpoints for each rollup.
    ///
    /// The log configuration.
    #[serde(rename = "Log")]
    pub log: Log,

    /// The local RPC server configuration.
    #[serde(rename = "RPC")]
    pub rpc: RpcConfig,
}

impl Config {
    /// Get the target RPC socket address from the configuration.
    pub fn rpc_addr(&self) -> std::net::SocketAddr {
        std::net::SocketAddr::from((self.rpc.host, self.rpc.port))
    }
}
