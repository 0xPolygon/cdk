//! Command line interface.
use std::path::PathBuf;

use clap::{Parser, Subcommand, ValueHint};

/// Command line interface.
#[derive(Parser)]
pub(crate) struct Cli {
    #[command(subcommand)]
    pub(crate) cmd: Commands,
}

#[derive(Subcommand)]
pub(crate) enum Commands {
    Run {
        /// The path to the configuration file.
        #[arg(long, short, value_hint = ValueHint::FilePath, default_value = "config/example-config.toml", env = "CDK_CONFIG_PATH")]
        cfg: PathBuf,
    },
}
