//! Command line interface.
use cdk_config::Config;
use clap::Parser;
use cli::Cli;
use std::path::PathBuf;
use std::sync::Arc;

mod cli;
mod logging;

fn main() -> anyhow::Result<()> {
    dotenvy::dotenv().ok();

    let cli = Cli::parse();

    match cli.cmd {
        cli::Commands::Run { cfg } => run(cfg)?,
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
pub fn run(cfg: PathBuf) -> anyhow::Result<()> {
    // Load the configuration file
    let config: Arc<Config> = Arc::new(toml::from_str(&std::fs::read_to_string(cfg)?)?);

    // Initialize the logger
    logging::tracing(&config.log);

    Ok(())
}
