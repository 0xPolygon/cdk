#!/usr/bin/env bash
# This script monitors the agglayer certificates progress of pessimistic proof.

function parse_params(){
    # Check if the required arguments are provided.
    if [ "$#" -lt 2 ]; then
    echo "Usage: $0 <settle_certificates_target> <timeout>"
    exit 1
    fi

    # The number of batches to be verified.
    settle_certificates_target="$1"

    # The script timeout (in seconds).
    timeout="$2"
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
    local _cmd="echo 'select status, count(*) from certificate_info group by status;' | sqlite3 /tmp/aggsender.sqlite"
    local _outcmd=$(mktemp)
    kurtosis service exec cdk cdk-node-001 "$_cmd" > $_outcmd
    if [ $? -ne 0 ]; then
        echo "[$(date '+%Y-%m-%d %H:%M:%S')] Error executing command kurtosis service: $_cmd"
        # clean temp file
        rm $_outcmd
        return 
    fi
    local num_certs=$(cat $_outcmd |  tail -n +2 | cut -f 2 -d '|' | xargs |tr ' ' '+' | bc)
    # Get the number of settled certificates "4|0"
    local _num_settle_certs=$(cat $_outcmd |  tail -n +2 | grep ^${aggsender_status_settled} | cut -d'|' -f2)
    [ -z "$_num_settle_certs" ] && _num_settle_certs=0
    # clean temp file
    rm $_outcmd 
    echo "[$(date '+%Y-%m-%d %H:%M:%S')] Num certificates on aggsender: $num_certs.  Settled certificates : $_num_settle_certs"
    if [ $num_certs -ge $settle_certificates_target ]; then
        echo "[$(date '+%Y-%m-%d %H:%M:%S')] ✅ Exiting... $num_certs certificates were settled! (total certs $num_certs)"
        exit 0
    fi
}

# MAIN
declare -A aggsender_status_map=(
    [0]="Pending"
    [1]="Proven"
    [2]="Candidate"
    [3]="InError"
    [4]="Settled"
)

readonly aggsender_status_settled=4


parse_params $*
start_time=$(date +%s)
end_time=$((start_time + timeout))
echo "[$(date '+%Y-%m-%d %H:%M:%S')] Start monitoring agglayer certificates progress..."
while true; do
    check_num_certificates
    check_timeout $end_time
    sleep 10
done
