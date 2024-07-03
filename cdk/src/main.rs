use execute::Execute;
use std::process::Command;

const ERIGON_PATH: &str = "erigon";

fn main() {
    let mut command = Command::new(ERIGON_PATH);

    let output = command.execute_output().unwrap();

    if let Some(exit_code) = output.status.code() {
        if exit_code == 0 {
            println!("Ok.");
        } else {
            eprintln!("Failed.");
        }
    } else {
        eprintln!("Interrupted!");
    }
}
