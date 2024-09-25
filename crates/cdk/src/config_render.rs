use anyhow::Error;
use std::fs;
use std::path::PathBuf;
use tempfile::{tempdir, TempDir};

pub fn render(
    chain_id: String,
    l2_sequencer_rpc_url: String,
    datastreamer_host: String,
    l1_rpc_url: String,
    sequencer_address: String,
    genesis_file: PathBuf,
) -> Result<TempDir, Error> {
    // Create a temporary directory
    let tmp_dir = tempdir()?;

    let res = crate::allocs_render::render_allocs(genesis_file.to_str().unwrap())?;
    // Write the three files to disk
    fs::write(
        tmp_dir
            .path()
            .join(format!("dynamic-{}-allocs.json", chain_id.clone())),
        res.output,
    )?;
    fs::write(
        tmp_dir
            .path()
            .join(format!("dynamic-{}-chainspec.json", chain_id.clone())),
        render_chainspec(chain_id.clone()),
    )?;
    fs::write(
        tmp_dir
            .path()
            .join(format!("dynamic-{}-conf.json", chain_id.clone())),
        render_conf(res.wrapper.root, 1000000000000000000),
    )?;

    let zkevm_address = res.wrapper.l1_config.zkevm_address;
    let rollup_address = res.wrapper.l1_config.rollup_manager_address;
    let ger_manager_address = res.wrapper.l1_config.zkevm_global_exit_root_address;
    let matic_contract_address = res.wrapper.l1_config.pol_token_address;

    fs::write(
        tmp_dir
            .path()
            .join(format!("dynamic-{}.yaml", chain_id.clone())),
        render_yaml(
            chain_id,
            l2_sequencer_rpc_url,
            datastreamer_host,
            l1_rpc_url,
            sequencer_address,
            zkevm_address,
            rollup_address,
            ger_manager_address,
            matic_contract_address,
        ),
    )?;

    Ok(tmp_dir)
}

fn render_chainspec(chain_id: String) -> String {
    format!(
        r#"
{{
    "ChainName": "dynamic-{chain_id}",
    "chainId": {chain_id},
    "consensus": "ethash",
    "homesteadBlock": 0,
    "daoForkBlock": 0,
    "eip150Block": 0,
    "eip155Block": 0,
    "byzantiumBlock": 0,
    "constantinopleBlock": 0,
    "petersburgBlock": 0,
    "istanbulBlock": 0,
    "muirGlacierBlock": 0,
    "berlinBlock": 0,
    "londonBlock": 9999999999999999999999999999999999999999999999999,
    "arrowGlacierBlock": 9999999999999999999999999999999999999999999999999,
    "grayGlacierBlock": 9999999999999999999999999999999999999999999999999,
    "terminalTotalDifficulty": 58750000000000000000000,
    "terminalTotalDifficultyPassed": false,
    "shanghaiTime": 9999999999999999999999999999999999999999999999999,
    "cancunTime": 9999999999999999999999999999999999999999999999999,
    "pragueTime": 9999999999999999999999999999999999999999999999999,
    "ethash": {{}}
}}
    "#
    )
}

fn render_conf(root: String, gas_limit: u64) -> String {
    format!(
        r#"
{{
  "root": {:?},
  "timestamp": 0,
  "gasLimit": {:?},
  "difficulty": 0
}}
    "#,
        root, gas_limit
    )
}

// render_config renders the configuration file for the Erigon node.
fn render_yaml(
    chain_id: String,
    l2_sequencer_rpc_url: String,
    datastreamer_host: String,
    l1_rpc_url: String,
    sequencer_address: String,
    zkevm_address: String,
    rollup_address: String,
    ger_manager_address: String,
    matic_contract_address: String,
) -> String {
    format!(
        r#"
datadir: ./data/dynamic-{chain_id}
chain: dynamic-{chain_id}
http: true
private.api.addr: localhost:9092
zkevm.l2-chain-id: {chain_id}
zkevm.l2-sequencer-rpc-url: {l2_sequencer_rpc_url}
zkevm.l2-datastreamer-url: {datastreamer_host}
zkevm.l1-chain-id: 271828
zkevm.l1-rpc-url: {l1_rpc_url}

zkevm.address-sequencer: {sequencer_address}
zkevm.address-zkevm: {zkevm_address}
zkevm.address-rollup: {rollup_address}
zkevm.address-ger-manager: {ger_manager_address}

zkevm.l1-rollup-id: 1
zkevm.l1-matic-contract-address: {matic_contract_address}
zkevm.l1-first-block: 23
zkevm.rpc-ratelimit: 250
txpool.disable: true
torrent.port: 42070
zkevm.datastream-version: 2

externalcl: true
http.api: [eth, debug, net, trace, web3, erigon, zkevm]
http.addr: 0.0.0.0
http.vhosts: any
http.corsdomain: any
ws: true
"#
    )
}
