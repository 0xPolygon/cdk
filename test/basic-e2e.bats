setup() {
    load 'helpers/common-setup'
    load 'helpers/common'
    _common_setup

    readonly sender_private_key=${SENDER_PRIVATE_KEY:-"12d7de8621a77640c9241b2595ba78ce443d05e94090365ab3bb5e19df82c625"}
    readonly receiver=${RECEIVER:-"0x85dA99c8a7C2C95964c8EfD687E95E632Fc533D6"}
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
