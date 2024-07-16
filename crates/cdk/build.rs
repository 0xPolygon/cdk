use std::env;
use std::path::PathBuf;
use std::process::Command;

fn main() {
    // Determine the directory where the build script is located
    let dir = env::var("CARGO_MANIFEST_DIR").unwrap();
    let build_path = PathBuf::from(dir + "/../..");
    println!("cargo:rerun-if-changed=build.rs");

    // Optionally, specify the directory where your Makefile is located
    // For this example, it's assumed to be the same as the build script's directory
    // If your Makefile is in a different directory, adjust `build_path` accordingly

    // Call the make command
    let output = Command::new("/bin/make")
        .arg("build") // Create a new make command
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
