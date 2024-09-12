#!/usr/bin/env bash
# Error code reference https://hackmd.io/WwahVBZERJKdfK3BbKxzQQ
function deposit () {
    if [[ $token_addr == "0x0000000000000000000000000000000000000000" ]]; then
        echo "Checking the current ETH balance: " >&3
        cast balance -e --rpc-url $l1_rpc_url $current_addr >&3
    else
        echo "Checking the current token balance for token at $token_addr: " >&3
        cast call --rpc-url $l1_rpc_url $token_addr 'balanceOf(address)(uint256)' $current_addr >&3
    fi

    echo "Attempting to deposit $amount wei to net $destination_net for token $token_addr" >&3

    if [[ $dry_run == "true" ]]; then
        cast calldata $bridge_sig $destination_net $destination_addr $amount $token_addr $is_forced $meta_bytes
    else
        if [[ $token_addr == "0x0000000000000000000000000000000000000000" ]]; then
            cast send --legacy --private-key $skey --value $amount --rpc-url $l1_rpc_url $bridge_addr $bridge_sig $destination_net $destination_addr $amount $token_addr $is_forced $meta_bytes
        else
            cast send --legacy --private-key $skey --rpc-url $l1_rpc_url $bridge_addr $bridge_sig $destination_net $destination_addr $amount $token_addr $is_forced $meta_bytes
        fi
    fi
}

function claim() {
    readonly bridge_deposit_file=$(mktemp)
    readonly claimable_deposit_file=$(mktemp)
    echo "Getting full list of deposits" >&3
    curl -s "$bridge_api_url/bridges/$destination_addr?limit=100&offset=0" | jq '.' | tee $bridge_deposit_file
    
    echo "Looking for claimable deposits" >&3
    jq '[.deposits[] | select(.ready_for_claim == true and .claim_tx_hash == "" and .dest_net == '$destination_net')]' $bridge_deposit_file | tee $claimable_deposit_file
    readonly claimable_count=$(jq '. | length' $claimable_deposit_file)
    echo "Found $claimable_count claimable deposits" >&3

    if [[ $claimable_count == 0 ]]; then
        echo "We have no claimable deposits at this time" >&3
        exit
    fi
    # if [[ $rpc_network_id != $destination_net ]]; then
    #     echo "The bridge on the current rpc has network id $rpc_network_id but you are claming a transaction on network $destination_net - are you sure you're using the right RPC??" >&3
    #     exit 1
    # fi
    echo "We have $claimable_count claimable deposits on network $destination_net. Let's get this party started." >&3
    readonly current_deposit=$(mktemp)
    readonly current_proof=$(mktemp)
    while read deposit_idx; do
        echo "Starting claim for tx index: "$deposit_idx >&3
        echo "Deposit info:" >&3
        jq --arg idx $deposit_idx '.[($idx | tonumber)]' $claimable_deposit_file | tee $current_deposit >&3

        curr_deposit_cnt=$(jq -r '.deposit_cnt' $current_deposit)
        curr_network_id=$(jq -r '.network_id' $current_deposit)
        curl -s "$bridge_api_url/merkle-proof?deposit_cnt=$curr_deposit_cnt&net_id=$curr_network_id" | jq '.' | tee $current_proof

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
            cast call --rpc-url $l2_rpc_url $bridge_addr $claim_sig "$in_merkle_proof" "$in_rollup_merkle_proof" $in_global_index $in_main_exit_root $in_rollup_exit_root $in_orig_net $in_orig_addr $in_dest_net $in_dest_addr $in_amount $in_metadata
        else
            cast send --legacy --rpc-url $l2_rpc_url --private-key $skey $bridge_addr $claim_sig "$in_merkle_proof" "$in_rollup_merkle_proof" $in_global_index $in_main_exit_root $in_rollup_exit_root $in_orig_net $in_orig_addr $in_dest_net $in_dest_addr $in_amount $in_metadata
        fi


    done < <(seq 0 $((claimable_count - 1)) )
}
