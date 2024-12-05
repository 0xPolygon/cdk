use regex::Regex;
use reqwest::blocking::get;
use std::env;
use std::fs::File;
use std::io::Write;
use std::path::Path;
use std::path::PathBuf;
use std::process::Command;

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
fn build_versions() -> std::io::Result<()> {
    // Retrieve the contents of the file from the URL
    let url = "https://raw.githubusercontent.com/0xPolygon/kurtosis-cdk/refs/heads/main/input_parser.star";
    let response = get(url).expect("Failed to send request");
    let content = response.text().expect("Failed to read response text");

    // Write the contents to a file
    let out_dir = std::env::var("OUT_DIR").unwrap();
    let dest_path = Path::new(&out_dir).join("input_parser.star");
    let mut file = File::create(&dest_path)?;
    file.write_all(content.as_bytes())?;

    // Get the corresponding lines from the contents of the starlark file
    let versions = content
        .lines()
        .skip(30)
        .take(15)
        .collect::<Vec<&str>>()
        .join("\n");

    // Replace the string DEFAULT_IMAGES = from the versions string
    let versions = versions.replace("DEFAULT_IMAGES = ", "");

    // Remove all comments to the end of the line using a regexp
    let re = Regex::new(r"\s#\s.*\n").unwrap();
    let versions = re.replace_all(&versions, "");
    // Replace the trailing comma on the last line
    let versions = versions.replace(", }", " }");

    // The versions string is a JSON object we can parse
    let versions_json: serde_json::Value = serde_json::from_str(&versions).unwrap();

    // Write the versions to a file
    let dest_path = Path::new(".").join("versions.json");
    let mut file = File::create(&dest_path)?;
    file.write_all(
        format!(
            "{}\n",
            serde_json::to_string_pretty(&versions_json).unwrap()
        )
        .as_bytes(),
    )?;

    Ok(())
}
