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
    // Load the versions from the versions.json file in the crate directory
    // and parse it using serde_json.
    let versions = include_str!("../versions.json");
    let versions_json: serde_json::Value = serde_json::from_str(versions).unwrap();

    // Convert the JSON object to a HashMap.
    let versions_map = versions_json.as_object().unwrap();

    // Get the version of the cdk-node binary.
    let output = version().unwrap();
    let version = String::from_utf8(output.stdout).unwrap();

    println!("{}", format!("{}", version.trim()).green());

    // Multi-line string to print the versions with colors.
    let formatted_versions: Vec<String> = versions_map
        .iter()
        .map(|(key, value)| format!("{}: {}", key.green(), value.to_string().blue()))
        .collect();

    println!("{}", "Supported up to fork12".yellow());
    println!("{}", formatted_versions.join("\n"));
}
