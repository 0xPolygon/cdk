use crate::l1::L1;
use serde::Deserialize;

/// The L1 configuration.
#[derive(Deserialize, Debug, Clone)]
pub struct NetworkConfig {
    #[serde(rename = "L1", default)]
    pub l1: L1,
}

impl Default for NetworkConfig {
    fn default() -> Self {
        Self { l1: L1::default() }
    }
}
