# based on: https://github.com/0xPolygon/kurtosis-cdk/blob/jhilliard/multi-pp-testing/multi-pp-test.sh.md

setup() {
    load '../../helpers/common-multi_cdk-setup'
    _common_multi_setup
    load '../../helpers/common'
    load '../../helpers/lxly-bridge-test'

    add_cdk_network2_to_agglayer
    fund_claim_tx_manager
    mint_pol_token

    ether_value=${ETHER_VALUE:-"0.0200000054"}
    amount=$(cast to-wei $ether_value ether)
    native_token_addr="0x0000000000000000000000000000000000000000" 
    readonly sender_private_key=${SENDER_PRIVATE_KEY:-"12d7de8621a77640c9241b2595ba78ce443d05e94090365ab3bb5e19df82c625"}
    readonly sender_addr="$(cast wallet address --private-key $sender_private_key)"
    # Params for lxly-bridge functions
    is_forced=${IS_FORCED:-"true"}
    bridge_addr=$bridge_address
    meta_bytes=${META_BYTES:-"0x1234"}
    destination_addr=$target_address
    timeout="600"
    claim_frequency="30"

    gas_price=$(cast gas-price --rpc-url "$l2_rpc_url")
}

@test "Test L2 to L2 bridge" {
    echo "=== Running LxLy bridge eth L1 to L2(PP1) amount:$amount" >&3
    destination_net=$l2_pp1b_network_id
    bridge_asset "$native_token_addr" "$l1_rpc_url"
    bridge_tx_hash_pp1=$bridge_tx_hash
    
    echo "=== Running LxLy bridge eth L1 to L2(PP2) amount:$amount" >&3
    destination_net=$l2_pp2b_network_id
    bridge_asset "$native_token_addr" "$l1_rpc_url"
    bridge_tx_hash_pp2=$bridge_tx_hash
    
    echo "=== Running LxLy claim L1 to L2(PP1) for $bridge_tx_hash_pp1" >&3
    run claim_tx_hash "$timeout" "$bridge_tx_hash_pp1" "$destination_addr" "$l2_pp1_url"  "$l2_pp1b_url"
    assert_success

    echo "=== Running LxLy claim L1 to L2(PP2) for $bridge_tx_hash_pp2" >&3
    run claim_tx_hash "$timeout" "$bridge_tx_hash_pp2" "$destination_addr" "$l2_pp2_url"  "$l2_pp2b_url"
    assert_success
    
    
   
    # reduce eth amount
    amount=1234567
    echo "=== Running LxLy bridge L2(PP2) to L2(PP1) amount:$amount" >&3
    destination_net=$l2_pp1b_network_id
    meta_bytes="0xbeef"
    bridge_asset "$native_token_addr" "$l2_pp2_url"
    
    echo "=== Running LxLy claim L2(PP2) to L2(PP1)  for: $bridge_tx_hash" >&3
    run claim_tx_hash "$timeout" "$bridge_tx_hash" "$destination_addr" "$l2_pp1_url"  "$l2_pp2b_url"
    assert_success
    
    # Now a need to do a bridge on L2(PP1) to trigger a certificate: 
    ether_value=${ETHER_VALUE:-"0.0100000054"}
    amount=$(cast to-wei $ether_value ether)
    echo "=== Running LxLy bridge eth L2(PP1) to L1 (trigger a certificate on PP1) amount:$amount" >&3
    destination_net=$l1_rpc_network_id
    meta_bytes="0xabcd"
    bridge_asset "$native_token_addr" "$l2_pp1_url"
    
    echo "=== Running LxLy claim L2(PP1) to L1 for $bridge_tx_hash" >&3
    run claim_tx_hash "$timeout" "$bridge_tx_hash" "$destination_addr" "$l1_rpc_url"  "$l2_pp1b_url"
    assert_success
}
