use ethers::types::Address;
use serde::Deserialize;

/// The L1 configuration.
#[derive(Deserialize, Debug, Clone)]
pub struct L1 {
    #[serde(rename = "L1ChainID", alias = "ChainID", default)]
    pub l1_chain_id: String,
    #[serde(rename = "PolAddr", default)]
    pub pol_addr: Address,
    #[serde(rename = "ZkEVMAddr", default)]
    pub zk_evm_addr: Address,
    #[serde(rename = "RollupManagerAddr", default)]
    pub rollup_manager_addr: Address,
    #[serde(rename = "GlobalExitRootManagerAddr", default)]
    pub global_exit_root_manager_addr: Address,
}

impl Default for L1 {
    fn default() -> Self {
        // Values are coming from https://github.com/0xPolygon/agglayer/blob/main/config/default.go#L11
        Self {
            l1_chain_id: "1337".to_string(),
            pol_addr: "0x5b06837A43bdC3dD9F114558DAf4B26ed49842Ed"
                .parse()
                .unwrap(),
            zk_evm_addr: "0x2F50ef6b8e8Ee4E579B17619A92dE3E2ffbD8AD2"
                .parse()
                .unwrap(),
            rollup_manager_addr: "0x1Fe038B54aeBf558638CA51C91bC8cCa06609e91"
                .parse()
                .unwrap(),
            global_exit_root_manager_addr: "0x1f7ad7caA53e35b4f0D138dC5CBF91aC108a2674"
                .parse()
                .unwrap(),
        }
    }
}
