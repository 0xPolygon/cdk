setup() {
    load 'helpers/common-setup'
    _common_setup
    load 'helpers/common'
    load 'helpers/lxly-bridge-test'

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
    destination_net=${DESTINATION_NET:-"1"}
    destination_addr=${DESTINATION_ADDRESS:-"0x0bb7AA0b4FdC2D2862c088424260e99ed6299148"}
    ether_value=${ETHER_VALUE:-"0.0200000054"}
    amount=$(cast to-wei $ether_value ether)
    token_addr=${TOKEN_ADDRESS:-"0x0000000000000000000000000000000000000000"}
    readonly is_forced=${IS_FORCED:-"true"}
    readonly bridge_addr=$BRIDGE_ADDRESS
    readonly meta_bytes=${META_BYTES:-"0x"}

    readonly l1_rpc_url=${L1_ETH_RPC_URL:-"$(kurtosis port print $enclave el-1-geth-lighthouse rpc)"}
    readonly bridge_api_url=${BRIDGE_API_URL:-"$(kurtosis port print $enclave zkevm-bridge-service-001 rpc)"}

    readonly dry_run=${DRY_RUN:-"false"}
    readonly sender_addr="$(cast wallet address --private-key $sender_private_key)"
    readonly l1_rpc_network_id=$(cast call --rpc-url $l1_rpc_url $bridge_addr 'networkID() (uint32)')
    readonly l2_rpc_network_id=$(cast call --rpc-url $l2_rpc_url $bridge_addr 'networkID() (uint32)')
    gas_price=$(cast gas-price --rpc-url "$l2_rpc_url")
}

@test "Run deposit" {
    echo "Running LxLy deposit" >&3
    run deposit
    assert_success
    assert_output --partial 'transactionHash'
}

@test "Run claim" {
    echo "Running LxLy claim" >&3

    timeout="120"
    claim_frequency="10"
    run wait_for_claim "$timeout" "$claim_frequency"
    assert_success
}

@test "Custom native token transfer" {
    # Use GAS_TOKEN_ADDR if provided, otherwise retrieve from file
    if [[ -n "$GAS_TOKEN_ADDR" ]]; then
        echo "Using provided GAS_TOKEN_ADDR: $GAS_TOKEN_ADDR" >&3
        local gas_token_addr="$GAS_TOKEN_ADDR"
    else
        echo "GAS_TOKEN_ADDR not provided, retrieving from rollup parameters file." >&3
        readonly rollup_params_file=/opt/zkevm/create_rollup_parameters.json
        run bash -c "$contracts_service_wrapper 'cat $rollup_params_file' | tail -n +2 | jq -r '.gasTokenAddress'"
        assert_success
        assert_output --regexp "0x[a-fA-F0-9]{40}"
        local gas_token_addr=$output
    fi

    echo "Gas token addr $gas_token_addr, L1 RPC: $l1_rpc_url" >&3

    # Set receiver address and query for its initial native token balance on the L2
    receiver=${RECEIVER:-"0x85dA99c8a7C2C95964c8EfD687E95E632Fc533D6"}
    local initial_receiver_balance=$(cast balance --ether "$receiver" --rpc-url "$l2_rpc_url")
    echo "Initial receiver balance of native token on L2 $initial_receiver_balance" >&3

    # Query for initial sender balance
    run query_contract "$l1_rpc_url" "$gas_token_addr" "$balance_of_fn_sig" "$sender_addr"
    assert_success
    local gas_token_init_sender_balance=$(echo "$output" | tail -n 1 | awk '{print $1}')
    echo "Initial sender balance $gas_token_init_sender_balance" of gas token on L1 >&3

    # Mint gas token on L1
    local tokens_amount="0.1ether"
    local wei_amount=$(cast --to-unit $tokens_amount wei)
    local minter_key=${MINTER_KEY:-"42b6e34dc21598a807dc19d7784c71b2a7a01f6480dc6f58258f78e539f1a1fa"}
    run mint_erc20_tokens "$l1_rpc_url" "$gas_token_addr" "$minter_key" "$sender_addr" "$tokens_amount"
    assert_success

    # Assert that balance of gas token (on the L1) is correct
    run query_contract "$l1_rpc_url" "$gas_token_addr" "$balance_of_fn_sig" "$sender_addr"
    assert_success
    local gas_token_final_sender_balance=$(echo "$output" |
        tail -n 1 |
        awk '{print $1}')
    local expected_balance=$(echo "$gas_token_init_sender_balance + $wei_amount" |
        bc |
        awk '{print $1}')

    echo "Sender balance ($sender_addr) (gas token L1): $gas_token_final_sender_balance" >&3
    assert_equal "$gas_token_final_sender_balance" "$expected_balance"

    # Send approve transaction to the gas token on L1
    deposit_ether_value="0.1ether"
    run send_tx "$l1_rpc_url" "$sender_private_key" "$gas_token_addr" "$approve_fn_sig" "$bridge_addr" "$deposit_ether_value"
    assert_success
    assert_output --regexp "Transaction successful \(transaction hash: 0x[a-fA-F0-9]{64}\)"

    # Deposit
    token_addr=$gas_token_addr
    destination_addr=$receiver
    destination_net=$l2_rpc_network_id
    amount=$wei_amount
    run deposit
    assert_success

    # Claim deposits (settle them on the L2)
    timeout="120"
    claim_frequency="10"
    run wait_for_claim "$timeout" "$claim_frequency"
    assert_success

    # Validate that the native token of receiver on L2 has increased by the bridge tokens amount
    run verify_native_token_balance "$l2_rpc_url" "$receiver" "$initial_receiver_balance" "$tokens_amount"
    assert_success
}
