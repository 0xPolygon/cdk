use serde::Deserialize;
use std::path::PathBuf;

// ExecutionEngine configuration
#[derive(Deserialize, Debug)]
pub struct ExecutionEngine {
    #[serde(rename = "Type")]
    pub type_: String,
    #[serde(rename = "DataDir")]
    pub data_dir: PathBuf,
    #[serde(rename = "HTTP")]
    pub http: bool,
    #[serde(rename = "HTTPAddr")]
    pub http_addr: String,
    #[serde(rename = "HTTPAPI")]
    pub http_api: Vec<String>,
    #[serde(rename = "HTTPVHosts")]
    pub http_vhosts: String,
    #[serde(rename = "HTTPCorsDomain")]
    pub http_cors_domain: String,
    #[serde(rename = "WS")]
    pub ws: bool,
    #[serde(rename = "PrivateAPIAddr")]
    pub private_api_addr: String,
    #[serde(rename = "ZkevmRPCRateLimit")]
    pub zkevm_rpc_rate_limit: u32,
    #[serde(rename = "ZkevmDatastreamVersion")]
    pub zkevm_datastream_version: u32,
    #[serde(rename = "L1FirstBlock")]
    pub l1_first_block: u32,
}

impl Default for ExecutionEngine {
    fn default() -> Self {
        Self {
            type_: "ExecutionEngine".to_string(),
            data_dir: PathBuf::from("./data/dynamic-1337"),
            http: true,
            http_addr: "0.0.0.0".to_string(),
            http_api: vec![
                "eth".to_string(),
                "debug".to_string(),
                "net".to_string(),
                "trace".to_string(),
                "web3".to_string(),
                "erigon".to_string(),
                "zkevm".to_string(),
            ],
            http_vhosts: "any".to_string(),
            http_cors_domain: "any".to_string(),
            ws: true,
            private_api_addr: "localhost:9092".to_string(),
            zkevm_rpc_rate_limit: 250,
            zkevm_datastream_version: 2,
            l1_first_block: 23,
        }
    }
}
