#!/usr/bin/env bash

set -euo pipefail

GREEN='\033[0;32m'
RED='\033[0;31m'
NC='\033[0m' # No Color

log_info() {
    echo -e "${GREEN}[INFO]${NC} $*" | tee -a "$LOG_FILE"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $*" | tee -a "$LOG_FILE"
}

trap 'log_error "Script failed at line $LINENO"' ERR

if [ "$#" -ne 3 ]; then
    echo "Usage: $0 <test_type: fork9-cdk-validium | fork11-rollup | fork12-cdk-validium | fork12-rollup> <path/to/kurtosis/cdk/repo> <path/to/e2e/repo>"
    exit 1
fi

TEST_TYPE=$1
KURTOSIS_FOLDER=$2
E2E_FOLDER=$3

PROJECT_ROOT="$PWD"
ROOT_FOLDER="/tmp/cdk-e2e-run"
LOG_FOLDER="$ROOT_FOLDER/logs"
LOG_FILE="$LOG_FOLDER/run-local-e2e.log"

rm -rf "$ROOT_FOLDER"
mkdir -p "$LOG_FOLDER"

exec > >(tee -a "$LOG_FILE") 2>&1

log_info "Starting local E2E setup..."

# Build cdk Docker Image if it doesn't exist
if [ "$(docker images -q cdk:local | wc -l)" -eq 0 ]; then
    log_info "Building cdk:local docker image..."
    pushd "$PROJECT_ROOT" > /dev/null
    make build-docker
    popd > /dev/null
else
    log_info "Docker image cdk:local already exists."
fi

log_info "Using provided Kurtosis CDK repo at: $KURTOSIS_FOLDER"

pushd "$KURTOSIS_FOLDER" > /dev/null
log_info "Cleaning any existing Kurtosis enclaves..."
kurtosis clean --all

ENCLAVE="cdk"

# Start Kurtosis Enclave 
log_info "Starting Kurtosis enclave"

if [ "$TEST_TYPE" == "fork9-cdk-validium" ]; then
    kurtosis run --enclave "$ENCLAVE" --args-file "$PROJECT_ROOT/.github/test_fork9_cdk_validium_e2e_args.json" .
elif [ "$TEST_TYPE" == "fork11-rollup" ]; then
    kurtosis run --enclave "$ENCLAVE" --args-file "$PROJECT_ROOT/.github/test_fork11_rollup_e2e_args.json" .
elif [ "$TEST_TYPE" == "fork12-cdk-validium" ]; then
    kurtosis run --enclave "$ENCLAVE" --args-file "$PROJECT_ROOT/.github/test_fork12_cdk_validium_e2e_args.json" .
elif [ "$TEST_TYPE" == "fork12-rollup" ]; then
    kurtosis run --enclave "$ENCLAVE" --args-file "$PROJECT_ROOT/.github/test_fork12_rollup_e2e_args.json" .
else
    log_error "Unknown test type: $TEST_TYPE"
    exit 1
fi

log_info "$ENCLAVE enclave started successfully."
popd > /dev/null

log_info "Using provided Agglayer E2E repo at: $E2E_FOLDER"

pushd "$E2E_FOLDER" > /dev/null

# Setup environment
log_info "Setting up e2e environment..."
set -a
source ./tests/.env
set +a

export BATS_LIB_PATH="$PWD/core/helpers/lib"
export PROJECT_ROOT="$PWD"
export ENCLAVE="$ENCLAVE"
export DISABLE_L2_FUND="true"

log_info "Running BATS E2E tests..."
bats ./tests/cdk

popd > /dev/null
log_info "E2E tests executed. Logs saved to $LOG_FILE"
