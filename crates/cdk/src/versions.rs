use colored::*;
use execute::Execute;
use std::io;
use std::process::{Command, Output};

fn version() -> Result<Output, io::Error> {
    let bin_path = crate::helpers::get_bin_path();

    // Run the node passing the config file path as argument
    let mut command = Command::new(bin_path.clone());
    command.args(&["version"]);

    command.execute_output()
}

pub(crate) fn versions() {
    // Get the version of the cdk-node binary.
    let output = version().unwrap();
    let version = String::from_utf8(output.stdout).unwrap();

    println!("{}", format!("{}", version.trim()).green());

    let versions = vec![
        (
            "zkEVM Contracts",
            "https://github.com/0xPolygonHermez/zkevm-contracts/releases/tag/v8.0.0-rc.4-fork.12",
        ),
        ("zkEVM Prover", "v8.0.0-RC12"),
        ("CDK Erigon", "hermeznetwork/cdk-erigon:0948e33"),
        (
            "zkEVM Pool Manager",
            "hermeznetwork/zkevm-pool-manager:v0.1.1",
        ),
        (
            "CDK Data Availability Node",
            "0xpolygon/cdk-data-availability:0.0.10",
        ),
    ];

    // Multi-line string to print the versions with colors.
    let formatted_versions: Vec<String> = versions
        .iter()
        .map(|(key, value)| format!("{}: {}", key.green(), value.blue()))
        .collect();

    println!("{}", formatted_versions.join("\n"));
}
