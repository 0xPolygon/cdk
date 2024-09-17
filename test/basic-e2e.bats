setup() {
    load 'helpers/common-setup'
    load 'helpers/common'
    _common_setup

    readonly enclave=${KURTOSIS_ENCLAVE:-cdk-v1}
    readonly node=${KURTOSIS_NODE:-cdk-erigon-node-001}
    readonly contracts_container=${KURTOSIS_CONTRACTS:-contracts-001}
    readonly sender_private_key=${SENDER_PRIVATE_KEY:-"12d7de8621a77640c9241b2595ba78ce443d05e94090365ab3bb5e19df82c625"}
    readonly receiver=${RECEIVER:-"0x85dA99c8a7C2C95964c8EfD687E95E632Fc533D6"}
    readonly contracts_service_wrapper=${KURTOSIS_CONTRACTS_WRAPPER:-"kurtosis service exec $enclave $contracts_container"}

    readonly l1_rpc_url=${L1_ETH_RPC_URL:-"$(kurtosis port print cdk-v1 el-1-geth-lighthouse rpc)"}
    readonly l2_rpc_url=${L2_ETH_RPC_URL:-$(kurtosis port print "$enclave" "$node" http-rpc)}

    readonly mint_fn_sig="function mint(address,uint256)"
    readonly balance_of_fn_sig="function balanceOf(address) (uint256)"
}

@test "Send EOA transaction" {
    local sender_addr=$(cast wallet address --private-key "$private_key")
    local initial_nonce=$(cast nonce "$sender_addr" --rpc-url "$rpc_url") || return 1
    local value="10ether"

    # case 1: Transaction successful sender has sufficient balance
    run sendTx "$l2_rpc_url" "$sender_private_key" "$receiver" "$value"
    assert_success
    assert_output --regexp "Transaction successful \(transaction hash: 0x[a-fA-F0-9]{64}\)"

    # case 2: Transaction rejected as sender attempts to transfer more than it has in its wallet.
    # Transaction will fail pre-validation check on the node and will be dropped subsequently from the pool
    # without recording it on the chain and hence nonce will not change
    local sender_balance=$(cast balance "$sender_addr" --ether --rpc-url "$rpc_url") || return 1
    local excessive_value=$(echo "$sender_balance + 1" | bc)"ether"
    run sendTx "$private_key" "$receiver" "$excessive_value"
    assert_failure

    # Check whether the sender's nonce was updated correctly
    local final_nonce=$(cast nonce "$sender_addr" --rpc-url "$rpc_url") || return 1
    assert_equal "$final_nonce" "$(echo "$initial_nonce + 1" | bc)"
}

@test "Deploy ERC20Mock contract" {
    local contract_artifact="./contracts/erc20mock/ERC20Mock.json"

    # Deploy ERC20Mock
    run deployContract "$l2_rpc_url" "$sender_private_key" "$contract_artifact"
    assert_success
    contract_addr=$(echo "$output" | tail -n 1)

    # Mint ERC20 tokens
    local amount="5"

    run sendTx "$l2_rpc_url" "$sender_private_key" "$contract_addr" "$mint_fn_sig" "$receiver" "$amount"
    assert_success
    assert_output --regexp "Transaction successful \(transaction hash: 0x[a-fA-F0-9]{64}\)"

    # Assert that balance is correct
    run queryContract "$l2_rpc_url" "$contract_addr" "$balance_of_fn_sig" "$receiver"
    assert_success
    receiverBalance=$(echo "$output" | tail -n 1)

    # Convert balance and amount to a standard format for comparison (e.g., remove any leading/trailing whitespace)
    receiverBalance=$(echo "$receiverBalance" | xargs)
    amount=$(echo "$amount" | xargs)

    # Check if the balance is equal to the amount
    assert_equal "$receiverBalance" "$amount"
}

@test "Custom native token transfer" {
    echo Enclave: "$enclave", container: "$contracts_container" >&3
    readonly rollup_params_file=/opt/zkevm/create_rollup_parameters.json

    # Retrieve the gas token address
    run bash -c "$contracts_service_wrapper 'cat $rollup_params_file' | tail -n +2 | jq -r '.gasTokenAddress'"
    assert_success
    assert_output --regexp "0x[a-fA-F0-9]{40}"
    local gas_token_addr=$output

    echo "Gas token addr $gas_token_addr, L1 RPC: $l1_rpc_url" >&3

    # Query for initial receiver balance
    run queryContract "$l1_rpc_url" "$gas_token_addr" "$balance_of_fn_sig" "$receiver"
    assert_success
    local initial_receiver_balance=$(echo "$output" | tail -n 1)
    echo "Initial receiver balance $initial_receiver_balance"

    # Mint gas token on L1
    local amount="1ether"
    run sendTx "$l1_rpc_url" "$sender_private_key" "$gas_token_addr" "$mint_fn_sig" "$receiver" "$amount"
    local wei_amount=$(cast --to-unit $amount wei)

    assert_success
    assert_output --regexp "Transaction successful \(transaction hash: 0x[a-fA-F0-9]{64}\)"

    # Assert that balance of gas token (on the L1) is correct
    run queryContract "$l1_rpc_url" "$gas_token_addr" "$balance_of_fn_sig" "$receiver"
    assert_success
    local receiver_balance=$(echo "$output" | tail -n 1)
    local expected_balance=$(echo "$initial_receiver_balance + $wei_amount" | bc)

    assert_equal "$receiver_balance" "$expected_balance"
}
