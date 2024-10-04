use anyhow::{Context, Result};
use serde::{Deserialize, Serialize};
use serde_json::{self, Value};
use std::collections::HashMap;
use std::fs::File;
use std::io::Read;
use std::path::Path;

#[derive(Serialize, Deserialize, Debug, Clone)]
struct Input {
    #[serde(rename = "contractName", skip_serializing_if = "Option::is_none")]
    contract_name: Option<String>,
    #[serde(rename = "accountName", skip_serializing_if = "Option::is_none")]
    account_name: Option<String>,
    balance: String,
    nonce: String,
    address: String,
    #[serde(skip_serializing_if = "Option::is_none")]
    bytecode: Option<String>,
    #[serde(skip_serializing_if = "Option::is_none")]
    storage: Option<HashMap<String, String>>,
}

#[derive(Serialize, Deserialize, Debug)]
pub struct Wrapper {
    pub root: String,
    #[serde(rename = "L1Config")]
    pub l1_config: L1Config,
    genesis: Vec<Input>,
    #[serde(rename = "rollupCreationBlockNumber")]
    pub rollup_creation_block_number: u64,
}

#[derive(Serialize, Deserialize, Debug)]
pub struct L1Config {
    #[serde(rename = "chainId")]
    pub chain_id: u64,
    #[serde(rename = "polygonZkEVMGlobalExitRootAddress")]
    pub zkevm_global_exit_root_address: String,
    #[serde(rename = "polygonRollupManagerAddress")]
    pub rollup_manager_address: String,
    #[serde(rename = "polTokenAddress")]
    pub pol_token_address: String,
    #[serde(rename = "polygonZkEVMAddress")]
    pub zkevm_address: String,
}

#[derive(Serialize, Deserialize, Debug)]
struct Output {
    #[serde(rename = "contractName", skip_serializing_if = "Option::is_none")]
    contract_name: Option<String>,
    #[serde(rename = "accountName", skip_serializing_if = "Option::is_none")]
    account_name: Option<String>,
    balance: Option<String>,
    nonce: Option<String>,
    code: Option<String>,
    storage: Option<Value>,
}

pub struct Rendered {
    pub output: String,
    pub wrapper: Wrapper,
}

pub fn render_allocs(genesis_file_path: &str) -> Result<Rendered> {
    let path = Path::new(genesis_file_path);
    let display = path.display();

    let mut file = File::open(&path).with_context(|| format!("couldn't open {}", display))?;

    let mut data = String::new();
    file.read_to_string(&mut data)
        .with_context(|| format!("couldn't read {}", display))?;

    let wrapper: Wrapper = serde_json::from_str(&data)
        .with_context(|| format!("couldn't parse JSON from {}", display))?;

    let mut outputs: HashMap<String, Output> = HashMap::new();

    for input in wrapper.genesis.clone() {
        let output = Output {
            contract_name: input.contract_name,
            account_name: input.account_name,
            balance: Some(input.balance),
            nonce: Some(input.nonce),
            code: input.bytecode,
            storage: input.storage.map(|s| serde_json::to_value(s).unwrap()),
        };
        outputs.insert(input.address, output);
    }

    // outputs.sort_by(|a, b| a.contract_name.cmp(&b.contract_name));

    Ok(Rendered {
        output: serde_json::to_string_pretty(&outputs)
            .with_context(|| "couldn't serialize outputs to JSON")?,
        wrapper,
    })
}
