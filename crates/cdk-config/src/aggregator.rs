use ethers::types::Address;
use serde::Deserialize;
use url::Url;

/// The StreamClient configuration.
#[derive(Deserialize, Debug, Clone)]
pub struct StreamClient {
    #[serde(rename = "Server")]
    pub server: String,
}

#[derive(Deserialize, Debug, Clone)]
pub struct EthTxManager {
    #[serde(rename = "Etherman")]
    pub etherman: Etherman,
}

#[derive(Deserialize, Debug, Clone)]
pub struct Etherman {
    #[serde(rename = "URL")]
    pub url: String,
}

/// The Aggregator configuration.
#[derive(Deserialize, Debug, Clone)]
pub struct Aggregator {
    #[serde(rename = "ChainID")]
    pub chain_id: String,
    #[serde(rename = "Host")]
    pub host: String,
    #[serde(rename = "Port")]
    pub port: String,
    #[serde(rename = "RetryTime")]
    pub retry_time: String,
    #[serde(rename = "VerifyProofInterval")]
    pub verify_proof_interval: String,
    #[serde(rename = "ProofStatePollingInterval")]
    pub proof_state_polling_interval: String,
    #[serde(rename = "TxProfitabilityCheckerType")]
    pub tx_profitability_checker_type: String,
    #[serde(rename = "TxProfitabilityMinReward")]
    pub tx_profitability_min_reward: String,
    #[serde(rename = "IntervalAfterWhichBatchConsolidateAnyway")]
    pub interval_after_which_batch_consolidate_anyway: String,
    #[serde(rename = "ForkId")]
    pub fork_id: u64,
    #[serde(rename = "CleanupLockedProofsInterval")]
    pub cleanup_locked_proofs_interval: String,
    #[serde(rename = "GeneratingProofCleanupThreshold")]
    pub generating_proof_cleanup_threshold: String,
    #[serde(rename = "GasOffset")]
    pub gas_offset: u64,
    #[serde(rename = "WitnessURL")]
    pub witness_url: Url,
    #[serde(rename = "SenderAddress")]
    pub sender_address: Address,
    // #[serde(rename = "SequencerPrivateKey")]
    // pub sequencer_private_key: SequencerPrivateKey,
    #[serde(rename = "SettlementBackend")]
    pub settlement_backend: String,
    #[serde(rename = "AggLayerTxTimeout")]
    pub agg_layer_tx_timeout: String,
    #[serde(rename = "AggLayerURL")]
    pub agg_layer_url: Url,
    #[serde(rename = "UseL1BatchData")]
    pub use_l1_batch_data: bool,
    #[serde(rename = "UseFullWitness")]
    pub use_full_witness: bool,
    #[serde(rename = "MaxWitnessRetrievalWorkers")]
    pub max_witness_retrieval_workers: u32,
    #[serde(rename = "SyncModeOnlyEnabled")]
    pub sync_mode_only_enabled: bool,

    #[serde(rename = "StreamClient")]
    pub stream_client: StreamClient,

    #[serde(rename = "EthTxManager")]
    pub eth_tx_manager: EthTxManager,
}

#[cfg(any(test, feature = "testutils"))]
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
            witness_url: Url::parse("http://localhost:8546").unwrap(),
            sender_address: "0x0000000000000000000000000000000000000000"
                .parse()
                .unwrap(),
            settlement_backend: "default".to_string(),
            agg_layer_tx_timeout: "30s".to_string(),
            agg_layer_url: Url::parse("http://localhost:8547").unwrap(),
            use_l1_batch_data: true,
            use_full_witness: false,
            max_witness_retrieval_workers: 4,
            sync_mode_only_enabled: false,
            stream_client: StreamClient {
                server: "localhost:9092".to_string(),
            },
            eth_tx_manager: EthTxManager {
                etherman: Etherman {
                    url: "http://localhost:9093".to_string(),
                },
            },
        }
    }
}
