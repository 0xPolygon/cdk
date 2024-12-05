use ethers::types::Address;
use serde::Deserialize;
use url::Url;

#[derive(Deserialize, Debug, Clone)]
pub struct EthTxManager {
    #[serde(rename = "Etherman")]
    pub etherman: Etherman,
}

impl Default for EthTxManager {
    fn default() -> Self {
        Self {
            etherman: Etherman::default(),
        }
    }
}

#[derive(Deserialize, Debug, Clone)]
pub struct Etherman {
    #[serde(rename = "URL", default)]
    pub url: String,
}

impl Default for Etherman {
    fn default() -> Self {
        Self {
            url: "http://localhost:8545".to_string(),
        }
    }
}

/// The Aggregator configuration.
#[derive(Deserialize, Debug, Clone)]
pub struct Aggregator {
    #[serde(rename = "ChainID", default)]
    pub chain_id: String,
    #[serde(rename = "Host", default)]
    pub host: String,
    #[serde(rename = "Port", default)]
    pub port: String,
    #[serde(rename = "RetryTime", default)]
    pub retry_time: String,
    #[serde(rename = "VerifyProofInterval", default)]
    pub verify_proof_interval: String,
    #[serde(rename = "ProofStatePollingInterval", default)]
    pub proof_state_polling_interval: String,
    #[serde(rename = "TxProfitabilityCheckerType", default)]
    pub tx_profitability_checker_type: String,
    #[serde(rename = "TxProfitabilityMinReward", default)]
    pub tx_profitability_min_reward: String,
    #[serde(rename = "IntervalAfterWhichBatchConsolidateAnyway", default)]
    pub interval_after_which_batch_consolidate_anyway: String,
    #[serde(rename = "ForkId", default)]
    pub fork_id: u64,
    #[serde(rename = "CleanupLockedProofsInterval", default)]
    pub cleanup_locked_proofs_interval: String,
    #[serde(rename = "GeneratingProofCleanupThreshold", default)]
    pub generating_proof_cleanup_threshold: String,
    #[serde(rename = "GasOffset", default)]
    pub gas_offset: u64,
    #[serde(rename = "RPCURL", default = "default_url")]
    pub rpc_url: Url,
    #[serde(rename = "WitnessURL", default = "default_url")]
    pub witness_url: Url,
    #[serde(rename = "SenderAddress", default = "default_address")]
    pub sender_address: Address,
    #[serde(rename = "SettlementBackend", default)]
    pub settlement_backend: String,
    #[serde(rename = "AggLayerTxTimeout", default)]
    pub agg_layer_tx_timeout: String,
    #[serde(rename = "AggLayerURL", default = "default_url")]
    pub agg_layer_url: Url,
    #[serde(rename = "UseFullWitness", default)]
    pub use_full_witness: bool,
    #[serde(rename = "SyncModeOnlyEnabled", default)]
    pub sync_mode_only_enabled: bool,

    #[serde(rename = "EthTxManager", default)]
    pub eth_tx_manager: EthTxManager,
}

fn default_url() -> Url {
    Url::parse("http://localhost:8546").unwrap()
}

fn default_address() -> Address {
    "0x0000000000000000000000000000000000000000"
        .parse()
        .unwrap()
}

impl Default for Aggregator {
    fn default() -> Self {
        // Values are coming from https://github.com/0xPolygon/agglayer/blob/main/config/default.go#L11
        Self {
            chain_id: "1".to_string(),
            host: "localhost".to_string(),
            port: "8545".to_string(),
            retry_time: "10s".to_string(),
            verify_proof_interval: "1m".to_string(),
            proof_state_polling_interval: "10s".to_string(),
            tx_profitability_checker_type: "default".to_string(),
            tx_profitability_min_reward: "0.1".to_string(),
            interval_after_which_batch_consolidate_anyway: "5m".to_string(),
            fork_id: 0,
            cleanup_locked_proofs_interval: "1h".to_string(),
            generating_proof_cleanup_threshold: "10m".to_string(),
            gas_offset: 0,
            rpc_url: default_url(),
            witness_url: default_url(),
            sender_address: default_address(),
            settlement_backend: "default".to_string(),
            agg_layer_tx_timeout: "30s".to_string(),
            agg_layer_url: Url::parse("http://localhost:8547").unwrap(),
            use_full_witness: false,
            sync_mode_only_enabled: false,
            eth_tx_manager: EthTxManager {
                etherman: Etherman {
                    url: "http://localhost:9093".to_string(),
                },
            },
        }
    }
}
