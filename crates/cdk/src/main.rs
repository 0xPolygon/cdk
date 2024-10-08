//! Command line interface.
use cdk_config::Config;
use clap::Parser;
use cli::Cli;
use execute::Execute;
use std::env;
use std::path::PathBuf;
use std::process::Command;
use std::sync::Arc;

mod cli;
mod logging;

const CDK_CLIENT_PATH: &str = "cdk-node";
const CDK_ERIGON_PATH: &str = "cdk-erigon";

fn main() -> anyhow::Result<()> {
    dotenvy::dotenv().ok();

    let cli = Cli::parse();

    match cli.cmd {
        cli::Commands::Node {} => node(cli.config)?,
        cli::Commands::Erigon {} => erigon(cli.config)?,
        // _ => forward()?,
    }

    Ok(())
}

/// This is the main node entrypoint.
///
/// This function starts everything needed to run an Agglayer node.
/// Starting by a Tokio runtime which can be used by the different components.
/// The configuration file is parsed and used to configure the node.
///
/// This function returns on fatal error or after graceful shutdown has
/// completed.
pub fn node(config_path: PathBuf) -> anyhow::Result<()> {
    // Load the configuration file
    let config_read = std::fs::read_to_string(config_path.clone());
    let toml_str = match config_read {
        Ok(toml) => toml,
        Err(e) => {
            eprintln!(
                "Failed to read configuration file, from path: {}",
                config_path.to_str().unwrap()
            );
            return Err(e.into());
        }
    };
    let config: Arc<Config> = Arc::new(toml::from_str(&toml_str)?);

    let mut bin_path = env::var("CARGO_MANIFEST_DIR").unwrap_or(CDK_CLIENT_PATH.into());
    if bin_path != CDK_CLIENT_PATH {
        bin_path = format!("{}/../../{}", bin_path, CDK_CLIENT_PATH);
    }

    // Run the node passing the config file path as argument
    let mut command = Command::new(bin_path.clone());
    command.args(&["run", "-cfg", config_path.canonicalize()?.to_str().unwrap()]);

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

    // Initialize the logger
    logging::tracing(&config.log);

    Ok(())
}

/// This is the main erigon entrypoint.
/// This function starts everything needed to run an Erigon node.
pub fn erigon(config_path: PathBuf) -> anyhow::Result<()> {
    // Load the configuration file
    let _config: Arc<Config> = Arc::new(toml::from_str(&std::fs::read_to_string(
        config_path.clone(),
    )?)?);

    let mut bin_path = env::var("CARGO_MANIFEST_DIR").unwrap_or(CDK_ERIGON_PATH.into());
    if bin_path != CDK_ERIGON_PATH {
        bin_path = format!("{}/../../{}", bin_path, CDK_ERIGON_PATH);
    }

    let mut command = Command::new(bin_path);

    // TODO: 1. Prepare erigon config files or flags

    command.args(&["--config", config_path.to_str().unwrap()]);

    let output = command.execute_output().unwrap();

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
