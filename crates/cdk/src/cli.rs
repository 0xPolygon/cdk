//! Command line interface.
use std::path::PathBuf;

use clap::{Parser, Subcommand, ValueHint};

/// Command line interface.
#[derive(Parser)]
#[command(author, version, about, long_about = None)]
pub(crate) struct Cli {
    #[command(subcommand)]
    pub(crate) cmd: Commands,
}

#[derive(Subcommand)]
pub(crate) enum Commands {
    Node {
        /// The path to the configuration file.
        #[arg(
            long,
            short,
            value_hint = ValueHint::FilePath,
            env = "CDK_CONFIG_PATH"
        )]
        config: PathBuf,

        /// Components to run.
        #[arg(
            long,
            short,
            value_hint = ValueHint::CommandString,
            env = "CDK_COMPONENTS",
        )]
        components: Option<String>,
    },
    Erigon {
        /// The path to the configuration file.
        #[arg(
            long,
            short,
            value_hint = ValueHint::FilePath,
            env = "CDK_CONFIG_PATH"
        )]
        config: PathBuf,

        /// The path to a chain specification file.
        #[arg(
            long,
            short = 'g',
            value_hint = ValueHint::FilePath,
            env = "CDK_GENESIS_PATH"
        )]
        chain: PathBuf,
    },
    Versions,
}
