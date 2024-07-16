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
    pub(crate) cfg: PathBuf,

    #[command(subcommand)]
    pub(crate) cmd: Commands,
}

#[derive(Subcommand)]
pub(crate) enum Commands {
    Rollup,
    Erigon,
}
