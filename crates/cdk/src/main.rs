//! Command line interface.
use alloy_rpc_client::ReqwestClient;
use alloy_rpc_client::{ClientBuilder, RpcCall};
use cdk_config::Config;
use clap::Parser;
use cli::Cli;
use execute::Execute;
use serde_json::Value;
use std::env;
use std::path::PathBuf;
use std::process::Command;
use url::Url;

pub mod allocs_render;
mod cli;
mod config_render;
mod logging;

const CDK_CLIENT_BIN: &str = "cdk-node";
const CDK_ERIGON_BIN: &str = "cdk-erigon";

#[tokio::main]
async fn main() -> anyhow::Result<()> {
    dotenvy::dotenv().ok();

    let cli = Cli::parse();

    // Read the config
    let config = read_config(cli.config.clone())?;

    // Initialize the logger
    logging::tracing(&config.log);

    println!(
        r#"🐼
  _____      _                            _____ _____  _  __
 |  __ \    | |                          / ____|  __ \| |/ /
 | |__) |__ | |_   _  __ _  ___  _ __   | |    | |  | | ' / 
 |  ___/ _ \| | | | |/ _` |/ _ \| '_ \  | |    | |  | |  <  
 | |  | (_) | | |_| | (_| | (_) | | | | | |____| |__| | . \ 
 |_|   \___/|_|\__, |\__, |\___/|_| |_|  \_____|_____/|_|\_\
                __/ | __/ |                                 
               |___/ |___/                                  
"#
    );

    match cli.cmd {
        cli::Commands::Node { components } => node(cli.config, components)?,
        cli::Commands::Erigon {} => erigon(config, cli.chain).await?,
        // _ => forward()?,
    }

    Ok(())
}

// read_config reads the configuration file and returns the configuration.
fn read_config(config_path: PathBuf) -> anyhow::Result<Config> {
    let config = std::fs::read_to_string(config_path)
        .map_err(|e| anyhow::anyhow!("Failed to read configuration file: {}", e))?;
    let config: Config = toml::from_str(&config)?;

    Ok(config)
}

/// This is the main node entrypoint.
///
/// This function starts everything needed to run an Agglayer node.
/// Starting by a Tokio runtime which can be used by the different components.
/// The configuration file is parsed and used to configure the node.
///
/// This function returns on fatal error or after graceful shutdown has
/// completed.
pub fn node(config_path: PathBuf, components: Option<String>) -> anyhow::Result<()> {
    // This is to find the erigon binary when running in development mode
    // otherwise it will use system path
    let mut bin_path = env::var("CARGO_MANIFEST_DIR").unwrap_or(CDK_CLIENT_BIN.into());
    if bin_path != CDK_CLIENT_BIN {
        bin_path = format!("{}/../../target/{}", bin_path, CDK_CLIENT_BIN);
    }

    let components_param = match components {
        Some(components) => format!("-components={}", components),
        None => "".to_string(),
    };

    // Run the node passing the config file path as argument
    let mut command = Command::new(bin_path.clone());
    command.args(&[
        "run",
        "-cfg",
        config_path.canonicalize()?.to_str().unwrap(),
        components_param.as_str(),
    ]);

    let output_result = command.execute_output();
    let output = match output_result {
        Ok(output) => output,
        Err(e) => {
            eprintln!(
                "Failed to execute command, trying to find executable in path: {}",
                bin_path
            );
            return Err(e.into());
        }
    };

    if let Some(exit_code) = output.status.code() {
        if exit_code == 0 {
            println!("Ok.");
        } else {
            eprintln!("Failed.");
        }
    } else {
        eprintln!("Interrupted!");
    }

    Ok(())
}

/// This is the main erigon entrypoint.
/// This function starts everything needed to run an Erigon node.
pub async fn erigon(config: Config, genesis_file: PathBuf) -> anyhow::Result<()> {
    // Render configuration files
    let chain_id = config.aggregator.chain_id.clone();
    let rpc_url = Url::parse(&config.sequence_sender.rpc_url).unwrap();
    let timestamp = get_timestamp(rpc_url).await.unwrap();
    let erigon_config_path = config_render::render(&config, genesis_file, timestamp)?;

    println!("Starting erigon with config: {:?}", erigon_config_path);

    // Run cdk-erigon in system path
    let output = Command::new(CDK_ERIGON_BIN)
        .args(&[
            "--config",
            erigon_config_path
                .path()
                .join(format!("dynamic-{}.yaml", chain_id))
                .to_str()
                .unwrap(),
        ])
        .execute_output()
        .unwrap();

    if let Some(exit_code) = output.status.code() {
        if exit_code != 0 {
            eprintln!(
                "Failed. Leaving configuration files in: {:?}",
                erigon_config_path
            );
            std::process::exit(1);
        }
    } else {
        eprintln!("Interrupted!");
    }

    Ok(())
}

/// Call the rpc server to retrieve the first batch timestamp
async fn get_timestamp(url: Url) -> Result<u64, anyhow::Error> {
    // Instantiate a new client over a transport.
    let client: ReqwestClient = ClientBuilder::default().http(url);

    // Prepare a request to the server.
    let request = client.request("zkevm_getBatchByNumber", vec!["0"]);

    // Poll the request to completion.
    let batch_json: serde_json::Map = request.await.unwrap();
    // Parse the string of data into serde_json::Value.
    let v: Value = serde_json::from_str(batch_json.as_str())?;

    Ok(v["timestamp"].as_u64().unwrap())
}
