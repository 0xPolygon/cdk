setup() {
    load 'helpers/common-setup'
    load 'helpers/common'
    _common_setup

    readonly enclave=${ENCLAVE:-cdk-v1}
    readonly node=${KURTOSIS_NODE:-cdk-erigon-node-001}
    readonly rpc_url=${RPC_URL:-$(kurtosis port print "$enclave" "$node" http-rpc)}
    readonly private_key=${SENDER_PRIVATE_KEY:-"12d7de8621a77640c9241b2595ba78ce443d05e94090365ab3bb5e19df82c625"}
    readonly receiver_private_key=${RECEIVER_PRIVATE_KEY:-"14b34448b61f2b1599a37aaf3eecbc4e6f4955b19396f99c9ac00374966e84b9"}
    readonly receiver=${RECEIVER:-"0x5C053Dd37650d8BEdf3f44232B1d1A3adc2b14B8"}
}

@test "Send EOA transaction" {
    local sender_addr=$(cast wallet address --private-key "$private_key")
    local initial_nonce=$(cast nonce "$sender_addr" --rpc-url "$rpc_url") || return 1
    local value="10ether"

    # case 1: Transaction successful sender has sufficient balance
    run sendTx "$private_key" "$receiver" "$value"
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
    wallet_A_output=$(cast wallet new)
    address_A=$(echo "$wallet_A_output" | grep "Address" | awk '{print $2}')
    address_A_private_key=$(echo "$wallet_A_output" | grep "Private key" | awk '{print $3}')
    address_B=$(cast wallet new | grep "Address" | awk '{print $2}')

    ## Case 1: An address deploy ERC20Mock contract => Transaction successful
    run deployContract "$private_key" "$contract_artifact"
    assert_success
    contract_addr=$(echo "$output" | tail -n 1)

    # Mint ERC20 tokens
    local mintFnSig="function mint(address receiver, uint256 amount)"
    local amount="5"

    run sendTx "$private_key" "$contract_addr" "$mintFnSig" "$address_A" "$amount"
    assert_success
    assert_output --regexp "Transaction successful \(transaction hash: 0x[a-fA-F0-9]{64}\)"

    ## Case 2: Insufficient gas scenario => Transactions fails
    # Get bytecode from the contract artifact
    local bytecode=$(jq -r .bytecode "$contract_artifact")
    if [[ -z "$bytecode" || "$bytecode" == "null" ]]; then
        echo "Error: Failed to read bytecode from $contract_artifact"
        return 1
    fi

    # Estimate gas,gas price and gas cost
    local gas_units=$(cast estimate --create "$bytecode")
    local gas_price=$(cast gas-price)
    local value=$(echo "$gas_units * $gas_price - 20000" | bc)
    local value_ether=$(cast to-unit "$value" ether)"ether" 

    # Transfer insufficient funds
    cast_output=$(cast send --rpc-url "$rpc_url" --private-key "$private_key" "$address_A" --value "$value_ether" --legacy 2>&1)
    if [[ $? -ne 0 ]]; then
        echo "Error: Failed to send transaction. Output:"
        echo "$cast_output"
        return 1
    fi
    
    # Fetch initial nonce for address_A
    local address_A_initial_nonce=$(cast nonce "$address_A" --rpc-url "$rpc_url") || return 1
    # Attempt to deploy contract with insufficient gas
    run deployContract "$address_A_private_key" "$contract_artifact"
    assert_failure

    ## Case 3: Transaction should fail as address_A tries to transfer more tokens than it has
    # Transfer funds for gas fees to address_A
    value_ether="4ether" 
    cast_output=$(cast send --rpc-url "$rpc_url" --private-key "$private_key" "$address_A" --value "$value_ether" --legacy 2>&1)
    if [[ $? -ne 0 ]]; then
        echo "Error: Failed to send transaction. Output:"
        echo "$cast_output"
        return 1
    fi

    # Fetch balance of address_A to simulate excessive transfer
    local balanceOfFnSig="function balanceOf(address) (uint256)"
    run queryContract "$contract_addr" "$balanceOfFnSig" "$address_A"
    assert_success
    local address_A_Balance=$(echo "$output" | tail -n 1)
    address_A_Balance=$(echo "$address_A_Balance" | xargs)

    # Set excessive amount for transfer
    local excessive_amount=$(echo "$address_A_Balance + 1" | bc)

    # Attempt transfer of excessive amount from address_A to address_B
    local tranferFnSig="function transfer(address receiver, uint256 amount)"
    run sendTx "$address_A_private_key" "$contract_addr" "$tranferFnSig" "$address_B" "$excessive_amount"
    assert_failure

    # Verify balance of address_A after failed transaction
    run queryContract "$contract_addr" "$balanceOfFnSig" "$address_A"
    assert_success
    address_A_BalanceAfterFailedTx=$(echo "$output" | tail -n 1)
    address_A_BalanceAfterFailedTx=$(echo "$address_A_BalanceAfterFailedTx" | xargs)

    # Ensure balance is unchanged
    assert_equal "$address_A_BalanceAfterFailedTx" "$address_A_Balance"

    # Verify balance of address_B is still zero
    run queryContract "$contract_addr" "$balanceOfFnSig" "$address_B"
    assert_success
    local address_B_Balance=$(echo "$output" | tail -n 1)
    address_B_Balance=$(echo "$address_B_Balance" | xargs)

    assert_equal "$address_B_Balance" "0"

    # Check if nonce increased by 2
    local address_A_final_nonce=$(cast nonce "$address_A" --rpc-url "$rpc_url") || return 1
    assert_equal "$address_A_final_nonce" "$(echo "$address_A_initial_nonce + 2" | bc)"
}
