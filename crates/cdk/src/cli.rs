//! Command line interface.
use std::path::PathBuf;

use clap::{Parser, Subcommand, ValueHint};

/// Command line interface.
#[derive(Parser)]
pub(crate) struct Cli {
    /// The path to the configuration file.
    #[arg(
        long,
        short,
        value_hint = ValueHint::FilePath,
        global = true,
        default_value = "config/example-config.toml",
        env = "CDK_CONFIG_PATH"
    )]
    pub(crate) config: PathBuf,

    /// The path to a chain specification file.
    #[arg(
        long,
        short = 'g',
        value_hint = ValueHint::FilePath,
        global = true,
        default_value = "config/genesis.json",
        env = "CDK_GENESIS_PATH"
    )]
    pub(crate) chain: PathBuf,

    #[command(subcommand)]
    pub(crate) cmd: Commands,
}

#[derive(Subcommand)]
pub(crate) enum Commands {
    Node,
    Erigon,
}
