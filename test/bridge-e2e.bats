setup() {
    load 'helpers/common-setup'
    _common_setup
    load 'helpers/common'
    load 'helpers/lxly-bridge-test'

    readonly data_availability_mode=${DATA_AVAILABILITY_MODE:-"cdk-validium"}
    $PROJECT_ROOT/test/scripts/kurtosis_prepare_params_yml.sh ../kurtosis-cdk $data_availability_mode
    [ $? -ne 0 ] && echo "Error preparing params.yml" && exit 1

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
    readonly amount=$(cast to-wei $ether_value ether)
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
}

@test "Run deposit" {
    echo "Running LxLy deposit" >&3
    run deposit
    assert_success
    assert_output --partial 'transactionHash'
}

@test "Run claim" {
    echo "Running LxLy claim"

    # The script timeout (in seconds).
    timeout="120"
    start_time=$(date +%s)
    end_time=$((start_time + timeout))

    while true; do
        current_time=$(date +%s)
        if ((current_time > end_time)); then
            echo "[$(date '+%Y-%m-%d %H:%M:%S')] ❌ Exiting... Timeout reached!"
            exit 1
        fi

        run claim
        if [ $status -eq 0 ]; then
            break
        fi
        sleep 10
    done

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

    receiver=${RECEIVER:-"0x85dA99c8a7C2C95964c8EfD687E95E632Fc533D6"}

    # Query for initial receiver balance
    run queryContract "$l1_rpc_url" "$gas_token_addr" "$balance_of_fn_sig" "$receiver"
    assert_success
    local initial_receiver_balance=$(echo "$output" | tail -n 1)
    echo "Initial receiver balance $initial_receiver_balance"

    # Mint gas token on L1
    local tokens_amount="1ether"
    run sendTx "$l1_rpc_url" "$sender_private_key" "$gas_token_addr" "$mint_fn_sig" "$receiver" "$tokens_amount"
    local wei_amount=$(cast --to-unit $tokens_amount wei)

    assert_success
    assert_output --regexp "Transaction successful \(transaction hash: 0x[a-fA-F0-9]{64}\)"

    # Assert that balance of gas token (on the L1) is correct
    run queryContract "$l1_rpc_url" "$gas_token_addr" "$balance_of_fn_sig" "$receiver"
    assert_success
    local receiver_balance=$(echo "$output" |
        tail -n 1 |
        awk '{print $1}')
    local expected_balance=$(echo "$initial_receiver_balance + $wei_amount" |
        bc |
        awk '{print $1}')

    echo "Receiver balance: $receiver_balance" >&3
    assert_equal "$receiver_balance" "$expected_balance"

    # Send approve transaction
    deposit_ether_value="0.1ether"
    run sendTx "$l1_rpc_url" "$sender_private_key" "$gas_token_addr" "$approve_fn_sig" "$bridge_addr" "$deposit_ether_value"
    assert_success
    assert_output --regexp "Transaction successful \(transaction hash: 0x[a-fA-F0-9]{64}\)"

    # Deposit
    token_addr=$gas_token_addr
    destination_addr=$receiver
    destination_net=$l2_rpc_network_id
    run deposit
    assert_success

    # Run claim?
    # Check whether native token balance on L2 has increased and send a dummy transaction
}
