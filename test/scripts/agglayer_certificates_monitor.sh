#!/usr/bin/env bash
# This script monitors the agglayer certificates progress of pessimistic proof.

function parse_params(){
    # Check if the required arguments are provided.
    if [ "$#" -lt 3 ]; then
    echo "Usage: $0 <settle_certificates_target> <timeout> <l2_network_id>"
    exit 1
    fi

    # The number of batches to be verified.
    settle_certificates_target="$1"

    # The script timeout (in seconds).
    timeout="$2"

    # The network id of the L2 network.
    l2_rpc_network_id="$3"
}

function check_timeout(){
    local _end_time=$1
    current_time=$(date +%s)
    if ((current_time > _end_time)); then
        echo "[$(date '+%Y-%m-%d %H:%M:%S')] ❌ Exiting... Timeout reached not found the expected numbers of settled certs!"
        exit 1
    fi
}

function check_num_certificates(){
    readonly agglayer_rpc_url="$(kurtosis port print cdk agglayer agglayer)"

    cast_output=$(cast rpc --rpc-url "$agglayer_rpc_url" "interop_getLatestKnownCertificateHeader" "$l2_rpc_network_id" 2>&1)

    if [ $? -ne 0 ]; then
        echo "[$(date '+%Y-%m-%d %H:%M:%S')] Error executing command cast rpc: $cast_output"
        return
    fi

    height=$(extract_certificate_height "$cast_output")
    [[ -z "$height" ]] && {
        echo "Error: Failed to extract certificate height: $height." >&3
        return
    }

    status=$(extract_certificate_status "$cast_output")
    [[ -z "$status" ]] && {
        echo "Error: Failed to extract certificate status." >&3
        return
    }

    echo "[$(date '+%Y-%m-%d %H:%M:%S')] Last known agglayer certificate height: $height, status: $status" >&3

    if (( height > settle_certificates_target - 1 )); then
        echo "[$(date '+%Y-%m-%d %H:%M:%S')] ✅ Success! The number of settled certificates has reached the target." >&3
        exit 0
    fi

    if (( height == settle_certificates_target - 1 )); then
        if [ "$status" == "Settled" ]; then
            echo "[$(date '+%Y-%m-%d %H:%M:%S')] ✅ Success! The number of settled certificates has reached the target." >&3
            exit 0
        fi
        
        echo "[$(date '+%Y-%m-%d %H:%M:%S')] ⚠️ Warning! The number of settled certificates is one less than the target." >&3
    fi
}

function extract_certificate_height() {
    local cast_output="$1"
    echo "$cast_output" | jq -r '.height'
}

function extract_certificate_status() {
    local cast_output="$1"
     echo "$cast_output" | jq -r '.status'
}

# MAIN

parse_params $*
start_time=$(date +%s)
end_time=$((start_time + timeout))
echo "[$(date '+%Y-%m-%d %H:%M:%S')] Start monitoring agglayer certificates progress..."
while true; do
    check_num_certificates
    check_timeout $end_time
    sleep 10
done
