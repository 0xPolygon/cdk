use regex::Regex;
use reqwest::blocking::get;
use std::env;
use std::fs::File;
use std::io::{self, Write};
use std::path::Path;
use std::path::PathBuf;
use std::process::Command;
use serde_json::Value;

fn main() {
    let _ = build_versions();

    let build_script_disabled = env::var("BUILD_SCRIPT_DISABLED")
        .map(|v| v == "1")
        .unwrap_or(false); // run by default

    if build_script_disabled {
        println!("cargo:warning=Build script is disabled. Skipping build.");
        return;
    }

    // Determine the directory where the build script is located
    let dir = env::var("CARGO_MANIFEST_DIR").unwrap();
    let build_path = PathBuf::from(dir + "/../..");
    println!("cargo:rerun-if-changed=*.go");

    // Optionally, specify the directory where your Makefile is located
    // For this example, it's assumed to be the same as the build script's directory
    // If your Makefile is in a different directory, adjust `build_path` accordingly

    // Call the make command
    let output = Command::new("make")
        .arg("build-go") // Create a new make command
        .current_dir(build_path) // Set the current directory for the command
        .output() // Execute the command and capture the output
        .expect("Failed to execute make command");

    // Check the output and react accordingly
    if !output.status.success() {
        // If the make command failed, print the error and exit
        let error_message = String::from_utf8_lossy(&output.stderr);
        panic!("Make command failed with error: {}", error_message);
    }

    // Optionally, print the output of the make command
    println!(
        "Make command output: {}",
        String::from_utf8_lossy(&output.stdout)
    );

    // Here you can also add additional commands to inform Cargo about
    // how to rerun the build script. For example, to rerun this script
    // only when a specific file changes:
    // println!("cargo:rerun-if-changed=path/to/file");
}

// build_versions retrieves the versions from the Starlark file and embeds them in the binary.
fn build_versions() -> io::Result<()> {
    // URL of the Starlark file
    let url = "https://raw.githubusercontent.com/0xPolygon/kurtosis-cdk/refs/heads/main/input_parser.star";

    // Download the file content
    let response = get(url).expect("Failed to send request");
    let content = response.text().expect("Failed to read response text");

    // Extract the relevant lines (skip the first 30 lines, take the next 15)
    let raw_versions = content
        .lines()
        .skip(30)
        .take(15)
        .collect::<Vec<&str>>()
        .join("\n");

    // Remove the declaration `DEFAULT_IMAGES = `
    let raw_versions = raw_versions.replace("DEFAULT_IMAGES = ", "");

    // Clean up the content by removing comments and unnecessary spaces
    let re_comments = Regex::new(r"#.*$").unwrap(); // Regex to remove comments
    let re_trailing_commas = Regex::new(r",(\s*})").unwrap(); // Regex to fix trailing commas

    let cleaned_versions = raw_versions
        .lines()
        .map(|line| re_comments.replace_all(line, "").trim().to_string()) // Remove comments and trim spaces
        .filter(|line| !line.is_empty()) // Filter out empty lines
        .collect::<Vec<_>>()
        .join("\n");

    // Fix improperly placed trailing commas
    let cleaned_versions = re_trailing_commas.replace_all(&cleaned_versions, "$1");

    // Attempt to parse the cleaned content as JSON
    let versions_json: Value = match serde_json::from_str(&cleaned_versions) {
        Ok(json) => json,
        Err(e) => {
            eprintln!("Failed to parse JSON: {}", e); // Print the error
            eprintln!("Input string was: {}", cleaned_versions); // Print the input causing the error
            return Err(io::Error::new(io::ErrorKind::InvalidData, "JSON parsing failed"));
        }
    };

    // Define the output file path for the JSON
    let dest_path = Path::new(".").join("versions.json");
    let mut file = File::create(&dest_path)?;
    file.write_all(
        format!(
            "{}\n",
            serde_json::to_string_pretty(&versions_json).unwrap() // Pretty-print JSON to the file
        )
        .as_bytes(),
    )?;

    Ok(())
}
