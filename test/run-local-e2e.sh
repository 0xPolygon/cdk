#!/usr/bin/env bash

set -euo pipefail

GREEN='\033[0;32m'
RED='\033[0;31m'
NC='\033[0m'

log_info() {
    echo -e "${GREEN}[INFO]${NC} $*" | tee -a "$LOG_FILE"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $*" | tee -a "$LOG_FILE"
}

trap 'log_error "Script failed at line $LINENO"' ERR

if [ "$#" -lt 2 ]; then
    echo "Usage: $0 <test_type: fork9-cdk-validium | fork11-rollup | fork12-cdk-validium | fork12-rollup> <path/to/kurtosis-cdk/repo> [<path/to/e2e/repo>]"
    exit 1
fi

TEST_TYPE=$1
KURTOSIS_FOLDER=$2
E2E_FOLDER=${3:-""}

PROJECT_ROOT="$PWD"
ROOT_FOLDER="/tmp/cdk-e2e-run"
LOG_FOLDER="$ROOT_FOLDER/logs"
LOG_FILE="$LOG_FOLDER/run-local-e2e.log"

rm -rf "$ROOT_FOLDER"
mkdir -p "$LOG_FOLDER"

exec > >(tee -a "$LOG_FILE") 2>&1

log_info "Starting local E2E setup..."

# Build docker image if needed
if [ "$(docker images -q cdk:local | wc -l)" -eq 0 ]; then
    log_info "Building cdk:local docker image..."
    pushd "$PROJECT_ROOT" >/dev/null
    make build-docker
    popd >/dev/null
else
    log_info "Docker image cdk:local already exists."
fi

log_info "Using Kurtosis CDK repo: $KURTOSIS_FOLDER"
pushd "$KURTOSIS_FOLDER" >/dev/null

log_info "Cleaning any existing Kurtosis enclaves..."
kurtosis clean --all

ENCLAVE="cdk"
log_info "Starting Kurtosis enclave"

case "$TEST_TYPE" in
fork9-cdk-validium)
    kurtosis run --enclave "$ENCLAVE" --args-file "$PROJECT_ROOT/.github/test_fork9_cdk_validium_e2e_args.json" .
    ;;
fork11-rollup)
    kurtosis run --enclave "$ENCLAVE" --args-file "$PROJECT_ROOT/.github/test_fork11_rollup_e2e_args.json" .
    ;;
fork12-cdk-validium)
    kurtosis run --enclave "$ENCLAVE" --args-file "$PROJECT_ROOT/.github/test_fork12_cdk_validium_e2e_args.json" .
    ;;
fork12-rollup)
    kurtosis run --enclave "$ENCLAVE" --args-file "$PROJECT_ROOT/.github/test_fork12_rollup_e2e_args.json" .
    ;;
*)
    log_error "Unknown test type: $TEST_TYPE"
    exit 1
    ;;
esac

log_info "$ENCLAVE enclave started successfully."
popd >/dev/null

# Run tests if E2E_FOLDER is provided
if [ -n "$E2E_FOLDER" ]; then
    log_info "Running E2E tests using: $E2E_FOLDER"
    pushd "$E2E_FOLDER" >/dev/null

    log_info "Setting up E2E environment..."
    set -a
    source ./tests/.env
    set +a

    export BATS_LIB_PATH="$PWD/core/helpers/lib"
    export PROJECT_ROOT="$PWD"
    export ENCLAVE="$ENCLAVE"
    export DISABLE_L2_FUND="true"

    log_info "Running BATS E2E tests..."
    bats tests/cdk/access-list-e2e.bats tests/cdk/basic-e2e.bats tests/cdk/e2e.bats

    if [[ "$TEST_TYPE" == "fork9-cdk-validium" || "$TEST_TYPE" == "fork11-rollup" ]]; then
        bats tests/cdk/bridge-e2e.bats
    elif [[ "$TEST_TYPE" == "fork12-cdk-validium" || "$TEST_TYPE" == "fork12-rollup" ]]; then
        bats tests/aggkit/bridge-e2e.bats tests/aggkit/bridge-e2e-custom-gas.bats
    fi

    popd >/dev/null
    log_info "E2E tests completed. Logs saved to $LOG_FILE"
else
    log_info "No E2E repo provided. Skipping tests."
    exit 0
fi
