setup() {
    load '../../helpers/common-setup'
    _common_setup
    load '../../helpers/common'
    load '../../helpers/lxly-bridge-test'

    if [ -z "$BRIDGE_ADDRESS" ]; then
        local combined_json_file="/opt/zkevm/combined.json"
        echo "BRIDGE_ADDRESS env variable is not provided, resolving the bridge address from the Kurtosis CDK '$combined_json_file'" >&3

        # Fetching the combined JSON output and filtering to get polygonZkEVMBridgeAddress
        combined_json_output=$($contracts_service_wrapper "cat $combined_json_file" | tail -n +2)
        bridge_default_address=$(echo "$combined_json_output" | jq -r .polygonZkEVMBridgeAddress)
        BRIDGE_ADDRESS=$bridge_default_address
    fi
    echo "Bridge address=$BRIDGE_ADDRESS" >&3

    readonly sender_private_key=${SENDER_PRIVATE_KEY:-"12d7de8621a77640c9241b2595ba78ce443d05e94090365ab3bb5e19df82c625"}
    readonly sender_addr="$(cast wallet address --private-key $sender_private_key)"
    destination_net=${DESTINATION_NET:-"1"}
    destination_addr=${DESTINATION_ADDRESS:-"0x0bb7AA0b4FdC2D2862c088424260e99ed6299148"}
    ether_value=${ETHER_VALUE:-"0.0200000054"}
    amount=$(cast to-wei $ether_value ether)
    readonly native_token_addr=${NATIVE_TOKEN_ADDRESS:-"0x0000000000000000000000000000000000000000"}
    if [[ -n "$GAS_TOKEN_ADDR" ]]; then
        echo "Using provided GAS_TOKEN_ADDR: $GAS_TOKEN_ADDR" >&3
        gas_token_addr="$GAS_TOKEN_ADDR"
    else
        echo "GAS_TOKEN_ADDR not provided, retrieving from rollup parameters file." >&3
        readonly rollup_params_file=/opt/zkevm/create_rollup_parameters.json
        run bash -c "$contracts_service_wrapper 'cat $rollup_params_file' | tail -n +2 | jq -r '.gasTokenAddress'"
        assert_success
        assert_output --regexp "0x[a-fA-F0-9]{40}"
        gas_token_addr=$output
    fi
    readonly is_forced=${IS_FORCED:-"true"}
    readonly bridge_addr=$BRIDGE_ADDRESS
    readonly meta_bytes=${META_BYTES:-"0x1234"}

    readonly l1_rpc_url=${L1_ETH_RPC_URL:-"$(kurtosis port print $enclave el-1-geth-lighthouse rpc)"}
    readonly bridge_api_url=${BRIDGE_API_URL:-"$(kurtosis port print $enclave zkevm-bridge-service-001 rpc)"}

    readonly dry_run=${DRY_RUN:-"false"}
    readonly l1_rpc_network_id=$(cast call --rpc-url $l1_rpc_url $bridge_addr 'networkID() (uint32)')
    readonly l2_rpc_network_id=$(cast call --rpc-url $l2_rpc_url $bridge_addr 'networkID() (uint32)')
    gas_price=$(cast gas-price --rpc-url "$l2_rpc_url")
    readonly weth_token_addr=$(cast call --rpc-url $l2_rpc_url $bridge_addr 'WETHToken()' | cast parse-bytes32-address)
}

@test "Native gas token deposit to WETH" {
    destination_addr=$sender_addr
    local initial_receiver_balance=$(cast call --rpc-url "$l2_rpc_url" "$weth_token_addr" "$balance_of_fn_sig" "$destination_addr" | awk '{print $1}')
    echo "Initial receiver balance of native token on L2 $initial_receiver_balance" >&3

    echo "=== Running LxLy deposit on L1 to network: $l2_rpc_network_id native_token: $native_token_addr" >&3
    
    destination_net=$l2_rpc_network_id
    run bridge_asset "$native_token_addr" "$l1_rpc_url"
    assert_success

    echo "=== Running LxLy claim on L2" >&3
    timeout="120"
    claim_frequency="10"
    run wait_for_claim "$timeout" "$claim_frequency" "$l2_rpc_url"
    assert_success

    echo "=== bridgeAsset L2 WETH: $weth_token_addr to L1 ETH" >&3
    destination_addr=$sender_addr
    destination_net=0
    run bridge_asset "$weth_token_addr" "$l2_rpc_url"
    assert_success
}

@test "transfer message" {
    echo "====== bridgeMessage L1 -> L2" >&3
    destination_addr=$sender_addr
    destination_net=$l2_rpc_network_id
    run bridge_message "$native_token_addr" "$l1_rpc_url"
    assert_success

    echo "====== Claim in L2" >&3
    timeout="120"
    claim_frequency="10"
    run wait_for_claim "$timeout" "$claim_frequency" "$l2_rpc_url" "bridgeMessage"
    assert_success

    echo "====== bridgeMessage L2->L1" >&3
    destination_net=0
    run bridge_message "0x0000000000000000000000000000000000000000" "$l2_rpc_url"
    assert_success
}