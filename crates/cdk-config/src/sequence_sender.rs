use serde::Deserialize;

/// The SequenceSender configuration.
#[derive(Deserialize, Debug, Clone)]
pub struct SequenceSender {
    #[serde(rename = "WaitPeriodSendSequence")]
    pub wait_period_send_sequence: String,
    #[serde(rename = "LastBatchVirtualizationTimeMaxWaitPeriod")]
    pub last_batch_virtualization_time_max_wait_period: String,
    #[serde(rename = "MaxTxSizeForL1")]
    pub max_tx_size_for_l1: u32,
    #[serde(rename = "L2Coinbase")]
    pub l2_coinbase: String,
    #[serde(rename = "SequencesTxFileName")]
    pub sequences_tx_file_name: String,
    #[serde(rename = "GasOffset")]
    pub gas_offset: u64,
    #[serde(rename = "WaitPeriodPurgeTxFile")]
    pub wait_period_purge_tx_file: String,
    #[serde(rename = "MaxPendingTx")]
    pub max_pending_tx: u32,
    #[serde(rename = "MaxBatchesForL1")]
    pub max_batches_for_l1: u32,
    #[serde(rename = "BlockFinality")]
    pub block_finality: String,
    #[serde(rename = "RPCURL")]
    pub rpc_url: String,
    #[serde(rename = "GetBatchWaitInterval")]
    pub get_batch_wait_interval: String,
}

// Default trait implementation
impl Default for SequenceSender {
    fn default() -> Self {
        Self {
            wait_period_send_sequence: "1s".to_string(),
            last_batch_virtualization_time_max_wait_period: "1s".to_string(),
            max_tx_size_for_l1: 1000,
            l2_coinbase: "0x".to_string(),
            sequences_tx_file_name: "sequences_tx.json".to_string(),
            gas_offset: 0,
            wait_period_purge_tx_file: "1s".to_string(),
            max_pending_tx: 1000,
            max_batches_for_l1: 100,
            block_finality: "1s".to_string(),
            rpc_url: "http://localhost:8545".to_string(),
            get_batch_wait_interval: "1s".to_string(),
        }
    }
}
