use super::DEFAULT_IP;
use serde::Deserialize;
use std::net::SocketAddr;

#[derive(Deserialize, Debug, Clone, Copy)]
#[serde(rename_all = "PascalCase")]
#[allow(dead_code)]
pub struct TelemetryConfig {
    #[serde(rename = "PrometheusAddr", default = "default_metrics_api_addr")]
    pub addr: SocketAddr,
}

impl Default for TelemetryConfig {
    fn default() -> Self {
        Self {
            addr: default_metrics_api_addr(),
        }
    }
}

const fn default_metrics_api_addr() -> SocketAddr {
    SocketAddr::V4(std::net::SocketAddrV4::new(DEFAULT_IP, 3000))
}
