use crate::allocs_render::Rendered;
use anyhow::Error;
use cdk_config::Config;
use std::fs;
use std::path::PathBuf;
use tempfile::{tempdir, TempDir};

pub fn render(config: &Config, genesis_file: PathBuf, timestamp: u64) -> Result<TempDir, Error> {
    // Create a temporary directory
    let tmp_dir = tempdir()?;
    let chain_id = config.aggregator.chain_id.clone();
    let res = crate::allocs_render::render_allocs(genesis_file.to_str().unwrap())?;
    // Write the three files to disk
    fs::write(
        tmp_dir
            .path()
            .join(format!("dynamic-{}-allocs.json", chain_id.clone())),
        res.output.clone(),
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
        render_conf(res.wrapper.root.clone(), timestamp),
    )?;

    let contents = render_yaml(config, res);
    fs::write(
        tmp_dir
            .path()
            .join(format!("dynamic-{}.yaml", chain_id.clone())),
        contents,
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

fn render_conf(root: String, timestamp: u64) -> String {
    format!(
        r#"
{{
  "root": {:?},
  "timestamp": {:?},
  "gasLimit": 0,
  "difficulty": 0
}}
    "#,
        root, timestamp
    )
}

// render_config renders the configuration file for the Erigon node.
fn render_yaml(config: &Config, res: Rendered) -> String {
    format!(
        r#"
chain: dynamic-{chain_id}
zkevm.l2-chain-id: {chain_id}
zkevm.l2-sequencer-rpc-url: {l2_sequencer_rpc_url}
zkevm.l2-datastreamer-url: {datastreamer_host}
zkevm.l1-chain-id: {l1_chain_id}
zkevm.l1-rpc-url: {l1_rpc_url}

zkevm.address-sequencer: {sequencer_address}
zkevm.address-zkevm: {zkevm_address}
zkevm.address-rollup: {rollup_address}
zkevm.address-ger-manager: {ger_manager_address}

zkevm.l1-matic-contract-address: {pol_token_address}
externalcl: true

zkevm.l1-first-block: {l1_first_block}
datadir: ./data/dynamic-{chain_id}
http: {http}
private.api.addr: {private_api_addr}
zkevm.rpc-ratelimit: {zkevm_rate_limit}
zkevm.datastream-version: {zkevm_datastream_version}
http.api: [{http_api}]
http.addr: {http_addr}
http.vhosts: {http_vhosts}
http.corsdomain: {http_cors_domain}
ws: {ws}
"#,
        chain_id = config.aggregator.chain_id.clone(),
        l2_sequencer_rpc_url = config.aggregator.witness_url.to_string(),
        datastreamer_host = config.aggregator.stream_client.server,
        l1_rpc_url = config.aggregator.eth_tx_manager.etherman.url,
        l1_chain_id = config.network_config.l1.l1_chain_id,
        sequencer_address = config.sequence_sender.l2_coinbase,
        zkevm_address = res.wrapper.l1_config.zkevm_address,
        rollup_address = res.wrapper.l1_config.rollup_manager_address,
        ger_manager_address = res.wrapper.l1_config.zkevm_global_exit_root_address,
        pol_token_address = res.wrapper.l1_config.pol_token_address,
        l1_first_block = config.execution_engine.l1_first_block,
        http = config.execution_engine.http,
        private_api_addr = config.execution_engine.private_api_addr,
        zkevm_rate_limit = config.execution_engine.zkevm_rpc_rate_limit,
        zkevm_datastream_version = config.execution_engine.zkevm_datastream_version,
        http_api = config.execution_engine.http_api.join(", "),
        http_addr = config.execution_engine.http_addr,
        http_vhosts = config.execution_engine.http_vhosts,
        http_cors_domain = config.execution_engine.http_cors_domain,
        ws = config.execution_engine.ws
    )
}
