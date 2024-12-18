#!/usr/bin/env bash
# Error code reference https://hackmd.io/WwahVBZERJKdfK3BbKxzQQ
function bridge_message() {
    local token_addr="$1"
    local rpc_url="$2"
    local bridge_sig='bridgeMessage(uint32,address,bool,bytes)'

    if [[ $token_addr == "0x0000000000000000000000000000000000000000" ]]; then
        echo "The ETH balance for sender "$sender_addr":" >&3
        cast balance -e --rpc-url $rpc_url $sender_addr >&3
    else
        echo "The "$token_addr" token balance for sender "$sender_addr":" >&3
        echo "cast call --rpc-url $rpc_url $token_addr \"$balance_of_fn_sig\" $sender_addr" >&3
        balance_wei=$(cast call --rpc-url "$rpc_url" "$token_addr" "$balance_of_fn_sig" "$sender_addr" | awk '{print $1}')
        echo "$(cast --from-wei "$balance_wei")" >&3
    fi

    echo "Attempting to deposit $amount [wei] using bridgeMessage to $destination_addr, token $token_addr (sender=$sender_addr, network id=$destination_net, rpc url=$rpc_url)" >&3

    if [[ $dry_run == "true" ]]; then
        cast calldata $bridge_sig $destination_net $destination_addr $amount $token_addr $is_forced $meta_bytes
    else
        if [[ $token_addr == "0x0000000000000000000000000000000000000000" ]]; then
            echo "cast send --legacy --private-key $sender_private_key --value $amount --rpc-url $rpc_url $bridge_addr $bridge_sig $destination_net $destination_addr $is_forced $meta_bytes"
            cast send --legacy --private-key $sender_private_key --value $amount --rpc-url $rpc_url $bridge_addr $bridge_sig $destination_net $destination_addr $is_forced $meta_bytes
        else
            echo "cast send --legacy --private-key $sender_private_key --rpc-url $rpc_url $bridge_addr $bridge_sig $destination_net $destination_addr $is_forced $meta_bytes"
            cast send --legacy --private-key $sender_private_key --rpc-url $rpc_url $bridge_addr $bridge_sig $destination_net $destination_addr $is_forced $meta_bytes
        fi
    fi
}

# returns:
#  - bridge_tx_hash
function bridge_asset() {
    local token_addr="$1"
    local rpc_url="$2"
    local bridge_sig='bridgeAsset(uint32,address,uint256,address,bool,bytes)'

    if [[ $token_addr == "0x0000000000000000000000000000000000000000" ]]; then
        echo "...The ETH balance for sender "$sender_addr":" $(cast balance -e --rpc-url $rpc_url $sender_addr) >&3
    else
        echo "The "$token_addr" token balance for sender "$sender_addr":" >&3
        echo "cast call --rpc-url $rpc_url $token_addr \"$balance_of_fn_sig\" $sender_addr"
        balance_wei=$(cast call --rpc-url "$rpc_url" "$token_addr" "$balance_of_fn_sig" "$sender_addr" | awk '{print $1}')
        echo "$(cast --from-wei "$balance_wei")" >&3
    fi

    echo "....Attempting to deposit $amount [wei] using bridgeAsset to $destination_addr, token $token_addr (sender=$sender_addr, network id=$destination_net, rpc url=$rpc_url)" >&3

    if [[ $dry_run == "true" ]]; then
        cast calldata $bridge_sig $destination_net $destination_addr $amount $token_addr $is_forced $meta_bytes
    else
        local tmp_response_file=$(mktemp)
        if [[ $token_addr == "0x0000000000000000000000000000000000000000" ]]; then
            echo "cast send --legacy --private-key $sender_private_key --value $amount --rpc-url $rpc_url $bridge_addr $bridge_sig $destination_net $destination_addr $amount $token_addr $is_forced $meta_bytes" 
            cast send --legacy --private-key $sender_private_key --value $amount --rpc-url $rpc_url $bridge_addr $bridge_sig $destination_net $destination_addr $amount $token_addr $is_forced $meta_bytes > $tmp_response_file
        else
            echo "cast send --legacy --private-key $sender_private_key --rpc-url $rpc_url $bridge_addr $bridge_sig $destination_net $destination_addr $amount $token_addr $is_forced $meta_bytes"
            
            cast send --legacy --private-key $sender_private_key                 --rpc-url $rpc_url $bridge_addr $bridge_sig $destination_net $destination_addr $amount $token_addr $is_forced $meta_bytes > $tmp_response_file
        fi
        export bridge_tx_hash=$(grep "^transactionHash" $tmp_response_file | cut -f 2- -d ' ' | sed 's/ //g')
        echo "bridge_tx_hash=$bridge_tx_hash" 
    fi
}

function claim() {
    local destination_rpc_url="$1"
    local bridge_type="$2"
    local claim_sig="claimAsset(bytes32[32],bytes32[32],uint256,bytes32,bytes32,uint32,address,uint32,address,uint256,bytes)"
    if [[ $bridge_type == "bridgeMessage" ]]; then
        claim_sig="claimMessage(bytes32[32],bytes32[32],uint256,bytes32,bytes32,uint32,address,uint32,address,uint256,bytes)"
    fi
    
    readonly bridge_deposit_file=$(mktemp)
    readonly claimable_deposit_file=$(mktemp)
    echo "Getting full list of deposits" >&3
    echo "    curl -s \"$bridge_api_url/bridges/$destination_addr?limit=100&offset=0\"" >&3
    curl -s "$bridge_api_url/bridges/$destination_addr?limit=100&offset=0" | jq '.' | tee $bridge_deposit_file

    echo "Looking for claimable deposits" >&3
    jq '[.deposits[] | select(.ready_for_claim == true and .claim_tx_hash == "" and .dest_net == '$destination_net')]' $bridge_deposit_file | tee $claimable_deposit_file
    readonly claimable_count=$(jq '. | length' $claimable_deposit_file)
    echo "Found $claimable_count claimable deposits" >&3

    if [[ $claimable_count == 0 ]]; then
        echo "We have no claimable deposits at this time" >&3
        exit 1
    fi

    echo "We have $claimable_count claimable deposits on network $destination_net. Let's get this party started." >&3
    readonly current_deposit=$(mktemp)
    readonly current_proof=$(mktemp)
    local gas_price_factor=1
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
        else
            local comp_gas_price=$(bc -l <<< "$gas_price * 1.5" | sed 's/\..*//')
            if [[ $? -ne 0 ]]; then
                echo "Failed to calculate gas price" >&3
                exit 1
            fi
            
            echo "cast send --legacy --gas-price $comp_gas_price --rpc-url $destination_rpc_url --private-key $sender_private_key $bridge_addr \"$claim_sig\" \"$in_merkle_proof\" \"$in_rollup_merkle_proof\" $in_global_index $in_main_exit_root $in_rollup_exit_root $in_orig_net $in_orig_addr $in_dest_net $in_dest_addr $in_amount $in_metadata" >&3
            cast send --legacy --gas-price $comp_gas_price --rpc-url $destination_rpc_url --private-key $sender_private_key $bridge_addr "$claim_sig" "$in_merkle_proof" "$in_rollup_merkle_proof" $in_global_index $in_main_exit_root $in_rollup_exit_root $in_orig_net $in_orig_addr $in_dest_net $in_dest_addr $in_amount $in_metadata
        fi

    done < <(seq 0 $((claimable_count - 1)))
}

# This function is used to claim a concrete tx hash
# global vars:
# - destination_addr
# export:
# - global_index

function claim_tx_hash() {
    local timeout="$1" 
    tx_hash="$2"
    local destination_addr="$3"
    local destination_rpc_url="$4"
    local bridge_provide_merkel_proof="$5"
    
    readonly bridge_deposit_file=$(mktemp)
    local ready_for_claim="false"
    local start_time=$(date +%s)
    local current_time=$(date +%s)
    local end_time=$((current_time + timeout))
    while true; do
        current_time=$(date +%s)
        elpased_time=$((current_time - start_time))
        if ((current_time > end_time)); then
            echo "[$(date '+%Y-%m-%d %H:%M:%S')] ❌ Exiting... Timeout reached waiting for tx_hash [$tx_hash] timeout: $timeout! (elapsed: $elpased_time)"
            echo "     $current_time > $end_time" >&3
            exit 1
        fi
        curl -s "$bridge_provide_merkel_proof/bridges/$destination_addr?limit=100&offset=0" | jq  "[.deposits[] | select(.tx_hash == \"$tx_hash\" )]" > $bridge_deposit_file
        deposit_count=$(jq '. | length' $bridge_deposit_file)
        if [[ $deposit_count == 0 ]]; then
            echo "...[$(date '+%Y-%m-%d %H:%M:%S')] ❌  the tx_hash [$tx_hash] not found (elapsed: $elpased_time / timeout:$timeout)" >&3   
            sleep "$claim_frequency"
            continue
        fi
        local ready_for_claim=$(jq '.[0].ready_for_claim' $bridge_deposit_file)
        if [ $ready_for_claim != "true" ]; then
            echo ".... [$(date '+%Y-%m-%d %H:%M:%S')] ⏳ the tx_hash $tx_hash is not ready for claim yet (elapsed: $elpased_time / timeout:$timeout)" >&3
            sleep "$claim_frequency"
            continue
        else
            break
        fi
    done
    # Deposit is ready for claim
    echo "....[$(date '+%Y-%m-%d %H:%M:%S')] 🎉 the tx_hash $tx_hash is ready for claim! (elapsed: $elpased_time)" >&3
    local curr_claim_tx_hash=$(jq '.[0].claim_tx_hash' $bridge_deposit_file)
    if [ $curr_claim_tx_hash != "\"\"" ]; then
        echo "....[$(date '+%Y-%m-%d %H:%M:%S')] 🎉  the tx_hash $tx_hash is already claimed" >&3
        exit 0
    fi
    local curr_deposit_cnt=$(jq '.[0].deposit_cnt' $bridge_deposit_file)
    local curr_network_id=$(jq  '.[0].network_id' $bridge_deposit_file)
    readonly current_deposit=$(mktemp)
    jq '.[(0|tonumber)]' $bridge_deposit_file | tee $current_deposit
    readonly current_proof=$(mktemp)
    echo ".... requesting merkel proof for $tx_hash deposit_cnt=$curr_deposit_cnt network_id: $curr_network_id" >&3
    request_merkel_proof "$curr_deposit_cnt" "$curr_network_id" "$bridge_provide_merkel_proof" "$current_proof"
    echo "FILE current_deposit=$current_deposit" 
    echo "FILE bridge_deposit_file=$bridge_deposit_file" 
    echo "FILE current_proof=$current_proof" 

    while true; do 
        echo ".... requesting claim for $tx_hash" >&3
        request_claim $current_deposit $current_proof $destination_rpc_url
        request_result=$?
        echo "....[$(date '+%Y-%m-%d %H:%M:%S')] 🎉  request_claim returns $request_result" >&3
        if [ $request_result -eq 0 ]; then
            echo "....[$(date '+%Y-%m-%d %H:%M:%S')] 🎉   claim successful" >&3
            break
        fi
        if [ $request_result -eq 2 ]; then
            # GlobalExitRootInvalid() let's retry
            echo "....[$(date '+%Y-%m-%d %H:%M:%S')] ❌  claim failed, let's retry" >&3
            current_time=$(date +%s)
            elpased_time=$((current_time - start_time))
            if ((current_time > end_time)); then
                echo "[$(date '+%Y-%m-%d %H:%M:%S')] ❌ Exiting... Timeout reached waiting for tx_hash [$tx_hash] timeout: $timeout! (elapsed: $elpased_time)"
                echo "     $current_time > $end_time" >&3
                exit 1
            fi
            sleep $claim_frequency
            continue
        fi
        if [ $request_result -ne 0 ]; then
            echo "....[$(date '+%Y-%m-%d %H:%M:%S')] ✅  claim successful tx_hash [$tx_hash]" >&3
            exit 1
        fi
    done
    echo "....[$(date '+%Y-%m-%d %H:%M:%S')]   claimed" >&3
    export global_index=$(jq '.global_index' $current_deposit)
    # clean up temp files
    rm $current_deposit
    rm $current_proof
    rm $bridge_deposit_file
    
}
function request_merkel_proof(){
    local curr_deposit_cnt="$1"
    local curr_network_id="$2"
    local bridge_provide_merkel_proof="$3"
    local result_proof_file="$4"
    curl -s "$bridge_provide_merkel_proof/merkle-proof?deposit_cnt=$curr_deposit_cnt&net_id=$curr_network_id" | jq '.' > $result_proof_file
    echo "request_merkel_proof: $result_proof_file"
}

# This function is used to claim a concrete tx hash
# global vars:
#  -dry_run
#  -gas_price
#  -sender_private_key
#  -bridge_addr
function request_claim(){
    local deposit_file="$1"
    local proof_file="$2"
    local destination_rpc_url="$3"
    
    local leaf_type=$(jq -r '.leaf_type' $deposit_file)
    local claim_sig="claimAsset(bytes32[32],bytes32[32],uint256,bytes32,bytes32,uint32,address,uint32,address,uint256,bytes)"
    
    if [[ $leaf_type != "0" ]]; then
       claim_sig="claimMessage(bytes32[32],bytes32[32],uint256,bytes32,bytes32,uint32,address,uint32,address,uint256,bytes)"
    fi

    local in_merkle_proof="$(jq -r -c '.proof.merkle_proof' $proof_file | tr -d '"')"
    local in_rollup_merkle_proof="$(jq -r -c '.proof.rollup_merkle_proof' $proof_file | tr -d '"')"
    local in_global_index=$(jq -r '.global_index' $deposit_file)
    local in_main_exit_root=$(jq -r '.proof.main_exit_root' $proof_file)
    local in_rollup_exit_root=$(jq -r '.proof.rollup_exit_root' $proof_file)
    local in_orig_net=$(jq -r '.orig_net' $deposit_file)
    local in_orig_addr=$(jq -r '.orig_addr' $deposit_file)
    local in_dest_net=$(jq -r '.dest_net' $deposit_file)
    local in_dest_addr=$(jq -r '.dest_addr' $deposit_file)
    local in_amount=$(jq -r '.amount' $deposit_file)
    local in_metadata=$(jq -r '.metadata' $deposit_file)
    if [[ $dry_run == "true" ]]; then
            echo "... Not real cleaim (dry_run mode)" >&3
            cast calldata $claim_sig "$in_merkle_proof" "$in_rollup_merkle_proof" $in_global_index $in_main_exit_root $in_rollup_exit_root $in_orig_net $in_orig_addr $in_dest_net $in_dest_addr $in_amount $in_metadata
        else
            local comp_gas_price=$(bc -l <<< "$gas_price * 1.5" | sed 's/\..*//')
            if [[ $? -ne 0 ]]; then
                echo "Failed to calculate gas price" >&3
                exit 1
            fi
            echo "... Claiming deposit: global_index: $in_global_index orig_net: $in_orig_net dest_net: $in_dest_net  amount:$in_amount" >&3
            echo "claim: mainnetExitRoot=$in_main_exit_root  rollupExitRoot=$in_rollup_exit_root"
            echo "cast send --legacy --gas-price $comp_gas_price --rpc-url $destination_rpc_url --private-key $sender_private_key $bridge_addr \"$claim_sig\" \"$in_merkle_proof\" \"$in_rollup_merkle_proof\" $in_global_index $in_main_exit_root $in_rollup_exit_root $in_orig_net $in_orig_addr $in_dest_net $in_dest_addr $in_amount $in_metadata" 
            local tmp_response=$(mktemp)
            cast send --legacy --gas-price $comp_gas_price \
                        --rpc-url $destination_rpc_url \
                        --private-key $sender_private_key \
                        $bridge_addr "$claim_sig" "$in_merkle_proof" "$in_rollup_merkle_proof" $in_global_index $in_main_exit_root $in_rollup_exit_root $in_orig_net $in_orig_addr $in_dest_net $in_dest_addr $in_amount $in_metadata 2> $tmp_response ||  check_claim_revert_code $tmp_response 
        fi
}

function check_claim_revert_code(){
    local file_curl_reponse="$1"
    # 0x646cf558 -> AlreadyClaimed()
    echo "check revert " 
    cat $file_curl_reponse
    cat $file_curl_reponse | grep "0x646cf558" > /dev/null
    if [ $? -eq 0 ]; then
        echo "....[$(date '+%Y-%m-%d %H:%M:%S')] 🎉  deposit is already claimed (revert code 0x646cf558)" >&3
        return 0
    fi
    cat $file_curl_reponse | grep "0x002f6fad" > /dev/null
    if [ $? -eq 0 ]; then
        echo "....[$(date '+%Y-%m-%d %H:%M:%S')] 🎉  GlobalExitRootInvalid()(revert code 0x002f6fad)" >&3
        return 2
    fi
    echo "....[$(date '+%Y-%m-%d %H:%M:%S')] ❌  claim failed" >&3
    cat $file_curl_reponse >&3
    return 1
}

function wait_for_claim() {
    local timeout="$1"         # timeout (in seconds)
    local claim_frequency="$2" # claim frequency (in seconds)
    local destination_rpc_url="$3" # destination rpc url
    local bridge_type="$4"        # bridgeAsset or bridgeMessage
    local start_time=$(date +%s)
    local end_time=$((start_time + timeout))

    while true; do
        local current_time=$(date +%s)
        if ((current_time > end_time)); then
            echo "[$(date '+%Y-%m-%d %H:%M:%S')] ❌ Exiting... Timeout reached!"
            exit 1
        fi

        run claim $destination_rpc_url $bridge_type
        if [ $status -eq 0 ]; then
            break
        fi

        sleep "$claim_frequency"
    done
}
