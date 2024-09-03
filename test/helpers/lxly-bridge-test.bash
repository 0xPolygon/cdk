#!/usr/bin/env bash
# Error code reference https://hackmd.io/WwahVBZERJKdfK3BbKxzQQ
function deposit () {
    2>&1 echo "Attempting to deposit $amount wei to net $destination_net for token $token_addr"

    if [[ $dry_run == "true" ]]; then
        cast calldata $bridge_sig $destination_net $destination_addr $amount $token_addr $is_forced $meta_bytes
    else
        if [[ $token_addr == "0x0000000000000000000000000000000000000000" ]]; then
            set -x
            cast send --legacy --private-key $skey --value $amount --rpc-url $rpc_url $bridge_addr $bridge_sig $destination_net $destination_addr $amount $token_addr $is_forced $meta_bytes
            set +x
        else
            set -x
            cast send --legacy --private-key $skey --rpc-url $rpc_url $bridge_addr $bridge_sig $destination_net $destination_addr $amount $token_addr $is_forced $meta_bytes
            set +x
        fi
    fi
}

function claim() {
    readonly bridge_deposit_file=$(mktemp)
    readonly claimable_deposit_file=$(mktemp)
    2>&1 echo "Getting full list of deposits"
    set -x
    2>&1 curl -s "$bridge_api_url/bridges/$destination_addr?limit=100&offset=0" | jq '.' | tee $bridge_deposit_file
    set +x
    2>&1 echo "Looking for claimable deposits"
    2>&1 jq '[.deposits[] | select(.ready_for_claim == true and .claim_tx_hash == "" and .dest_net == '$destination_net')]' $bridge_deposit_file | tee $claimable_deposit_file
    readonly claimable_count=$(jq '. | length' $claimable_deposit_file)
    if [[ $claimable_count == 0 ]]; then
        2>&1 echo "We have no claimable deposits at this time"
        exit
    fi
    if [[ $rpc_network_id != $destination_net ]]; then
        2>&1 echo "The bridge on the current rpc has network id $rpc_network_id but you are claming a transaction on network $destination_net - are you sure you're using the right RPC??"
        exit 1
    fi
    2>&1 echo "We have $claimable_count claimable deposits on network $destination_net. Let's get this party started."
    readonly current_deposit=$(mktemp)
    readonly current_proof=$(mktemp)
    while read deposit_idx; do
        2>&1 echo "Starting claim for tx index: "$deposit_idx
        2>&1 echo "Deposit info:"
        2>&1 jq --arg idx $deposit_idx '.[($idx | tonumber)]' $claimable_deposit_file | tee $current_deposit

        curr_deposit_cnt=$(jq -r '.deposit_cnt' $current_deposit)
        curr_network_id=$(jq -r '.network_id' $current_deposit)
        2>&1 echo "Proof:"
        set -x
        2>&1 curl -s "$bridge_api_url/merkle-proof?deposit_cnt=$curr_deposit_cnt&net_id=$curr_network_id" | jq '.' | tee $current_proof
        set  +x

        in_merkle_proof="$(jq -r -c '.proof.merkle_proof' $current_proof | tr -d '"')"
        in_rollup_merkle_proof="$(jq -r -c '.proof.rollup_merkle_proof' $current_proof | tr -d '"')"
        in_global_index=$(jq -r '.global_index' $current_deposit)
        in_main_exit_root=$(jq -r '.proof.main_exit_root' $current_proof)
        in_rollup_exit_root=$(jq -r '.proof.rollup_exit_root' $current_proof)
        in_orig_net=$(jq -r '.orig_net' $current_deposit)
        in_orig_addr=$(jq -r '.orig_addr' $current_deposit)
        in_dest_net=$(jq -r '.dest_net' $current_deposit)
        in_dest_addr=$(jq -r '.dest_addr' $current_deposit)
        in_amount=$(jq -r '.amount' $current_deposit)
        in_metadata=$(jq -r '.metadata' $current_deposit)

        if [[ $dry_run == "true" ]]; then
            cast calldata $claim_sig "$in_merkle_proof" "$in_rollup_merkle_proof" $in_global_index $in_main_exit_root $in_rollup_exit_root $in_orig_net $in_orig_addr $in_dest_net $in_dest_addr $in_amount $in_metadata
            set -x
            cast call --rpc-url $rpc_url $bridge_addr $claim_sig "$in_merkle_proof" "$in_rollup_merkle_proof" $in_global_index $in_main_exit_root $in_rollup_exit_root $in_orig_net $in_orig_addr $in_dest_net $in_dest_addr $in_amount $in_metadata
            set +x
        else
            set -x
            cast send --legacy --rpc-url $rpc_url --private-key $skey $bridge_addr $claim_sig "$in_merkle_proof" "$in_rollup_merkle_proof" $in_global_index $in_main_exit_root $in_rollup_exit_root $in_orig_net $in_orig_addr $in_dest_net $in_dest_addr $in_amount $in_metadata
            set +x
        fi


    done < <(seq 0 $((claimable_count - 1)) )
}
