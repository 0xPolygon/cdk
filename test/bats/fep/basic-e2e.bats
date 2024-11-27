setup() {
    load '../../helpers/bats/common-setup'
    load '../../helpers/bats/common'
    
    _common_setup

    readonly sender_private_key=${SENDER_PRIVATE_KEY:-"12d7de8621a77640c9241b2595ba78ce443d05e94090365ab3bb5e19df82c625"}
    readonly receiver=${RECEIVER:-"0x85dA99c8a7C2C95964c8EfD687E95E632Fc533D6"}
}

@test "Send EOA transaction" {
    local sender_addr=$(cast wallet address --private-key "$sender_private_key")
    local initial_nonce=$(cast nonce "$sender_addr" --rpc-url "$l2_rpc_url") || {
        echo "Failed to retrieve nonce for sender: $sender_addr using RPC URL: $l2_rpc_url"
        return 1
    }
    local value="10ether"

    # case 1: Transaction successful sender has sufficient balance
    run send_tx "$l2_rpc_url" "$sender_private_key" "$receiver" "$value"
    assert_success
    assert_output --regexp "Transaction successful \(transaction hash: 0x[a-fA-F0-9]{64}\)"

    # case 2: Transaction rejected as sender attempts to transfer more than it has in its wallet.
    # Transaction will fail pre-validation check on the node and will be dropped subsequently from the pool
    # without recording it on the chain and hence nonce will not change
    local sender_balance=$(cast balance "$sender_addr" --ether --rpc-url "$l2_rpc_url") || {
        echo "Failed to retrieve balance for sender: $sender_addr using RPC URL: $l2_rpc_url"
        return 1
    }
    local excessive_value=$(echo "$sender_balance + 1" | bc)"ether"
    run send_tx "$l2_rpc_url" "$sender_private_key" "$receiver" "$excessive_value"
    assert_failure

    # Check whether the sender's nonce was updated correctly
    local final_nonce=$(cast nonce "$sender_addr" --rpc-url "$l2_rpc_url") || {
        echo "Failed to retrieve nonce for sender: $sender_addr using RPC URL: $l2_rpc_url"
        return 1
    }
    assert_equal "$final_nonce" "$(echo "$initial_nonce + 1" | bc)"
}

@test "Test ERC20Mock contract" {
    local contract_artifact="./contracts/erc20mock/ERC20Mock.json"
    wallet_A_output=$(cast wallet new)
    address_A=$(echo "$wallet_A_output" | grep "Address" | awk '{print $2}')
    address_A_private_key=$(echo "$wallet_A_output" | grep "Private key" | awk '{print $3}')
    address_B=$(cast wallet new | grep "Address" | awk '{print $2}')

    # Deploy ERC20Mock
    run deploy_contract "$l2_rpc_url" "$sender_private_key" "$contract_artifact"
    assert_success
    contract_addr=$(echo "$output" | tail -n 1)

    # Mint ERC20 tokens
    local amount="5"

    run send_tx "$l2_rpc_url" "$sender_private_key" "$contract_addr" "$mint_fn_sig" "$address_A" "$amount"
    assert_success
    assert_output --regexp "Transaction successful \(transaction hash: 0x[a-fA-F0-9]{64}\)"

    ## Case 2: Insufficient gas scenario => Transactions fails
    # nonce would not increase since transaction fails at the node's pre-validation check
    # Get bytecode from the contract artifact
    local bytecode=$(jq -r .bytecode "$contract_artifact")
    if [[ -z "$bytecode" || "$bytecode" == "null" ]]; then
        echo "Error: Failed to read bytecode from $contract_artifact"
        return 1
    fi

    # Estimate gas, gas price and gas cost
    local gas_units=$(cast estimate --rpc-url "$l2_rpc_url" --create "$bytecode")
    gas_units=$(echo "scale=0; $gas_units / 2" | bc)
    local gas_price=$(cast gas-price --rpc-url "$l2_rpc_url")
    local value=$(echo "$gas_units * $gas_price" | bc)
    local value_ether=$(cast to-unit "$value" ether)"ether" 

    # Transfer only half amount of tokens needed for contract deployment fees
    cast_output=$(cast send --rpc-url "$l2_rpc_url" --private-key "$sender_private_key" "$address_A" --value "$value_ether" --legacy 2>&1)
    if [[ $? -ne 0 ]]; then
        echo "Error: Failed to send transaction. Output:"
        echo "$cast_output"
        return 1
    fi
    
    # Fetch initial nonce for address_A
    local address_A_initial_nonce=$(cast nonce "$address_A" --rpc-url "$l2_rpc_url") || return 1
    # Attempt to deploy contract with insufficient gas
    run deploy_contract "$l2_rpc_url" "$address_A_private_key" "$contract_artifact"
    assert_failure

    ## Case 3: Transaction should fail as address_A tries to transfer more tokens than it has
    # nonce would not increase 
    # Transfer funds for gas fees to address_A
    value_ether="4ether" 
    cast_output=$(cast send --rpc-url "$l2_rpc_url" --private-key "$sender_private_key" "$address_A" --value "$value_ether" --legacy 2>&1)
    if [[ $? -ne 0 ]]; then
        echo "Error: Failed to send transaction. Output:"
        echo "$cast_output"
        return 1
    fi

    # Fetch balance of address_A to simulate excessive transfer
    run query_contract "$l2_rpc_url" "$contract_addr" "$balance_of_fn_sig" "$address_A"
    assert_success
    local address_A_Balance=$(echo "$output" | tail -n 1)
    address_A_Balance=$(echo "$address_A_Balance" | xargs)

    # Set excessive amount for transfer
    local excessive_amount=$(echo "$address_A_Balance + 1" | bc)

    # Attempt transfer of excessive amount from address_A to address_B
    local tranferFnSig="transfer(address,uint256)"
    run send_tx "$l2_rpc_url" "$address_A_private_key" "$contract_addr" "$tranferFnSig" "$address_B" "$excessive_amount"
    assert_failure

    # Verify balance of address_A after failed transaction
    run query_contract "$l2_rpc_url" "$contract_addr" "$balance_of_fn_sig" "$address_A"
    assert_success
    address_A_BalanceAfterFailedTx=$(echo "$output" | tail -n 1)
    address_A_BalanceAfterFailedTx=$(echo "$address_A_BalanceAfterFailedTx" | xargs)

    # Ensure balance is unchanged
    assert_equal "$address_A_BalanceAfterFailedTx" "$address_A_Balance"

    # Verify balance of address_B is still zero
    run query_contract "$l2_rpc_url" "$contract_addr" "$balance_of_fn_sig" "$address_B"
    assert_success
    local address_B_Balance=$(echo "$output" | tail -n 1)
    address_B_Balance=$(echo "$address_B_Balance" | xargs)

    assert_equal "$address_B_Balance" "0"

    # Nonce should not increase
    local address_A_final_nonce=$(cast nonce "$address_A" --rpc-url "$l2_rpc_url") || {
        echo "Failed to retrieve nonce for sender: $address_A using RPC URL: $l2_rpc_url"
        return 1
    }
    assert_equal "$address_A_final_nonce" "$address_A_initial_nonce"
}


@test "Deploy and test UniswapV3 contract" {
    # Generate new key pair
    wallet_A_output=$(cast wallet new)
    address_A=$(echo "$wallet_A_output" | grep "Address" | awk '{print $2}')
    address_A_private_key=$(echo "$wallet_A_output" | grep "Private key" | awk '{print $3}')

    # Transfer funds for gas
    local value_ether="50ether"
    cast_output=$(cast send --rpc-url "$l2_rpc_url" --private-key "$sender_private_key" "$address_A" --value "$value_ether" --legacy 2>&1)
    if [[ $? -ne 0 ]]; then
        echo "Error: Failed to send transaction. Output:"
        echo "$cast_output"
        return 1
    fi

    run polycli loadtest uniswapv3 --legacy -v 600 --rpc-url $l2_rpc_url --private-key $address_A_private_key
    assert_success

    # Remove ANSI escape codes from the output
    output=$(echo "$output" | sed -r "s/\x1B\[[0-9;]*[mGKH]//g")

    # Check if all required Uniswap contracts were deployed
    assert_output --regexp "Contract deployed address=0x[a-fA-F0-9]{40} name=WETH9"
    assert_output --regexp "Contract deployed address=0x[a-fA-F0-9]{40} name=UniswapV3Factory"
    assert_output --regexp "Contract deployed address=0x[a-fA-F0-9]{40} name=UniswapInterfaceMulticall"
    assert_output --regexp "Contract deployed address=0x[a-fA-F0-9]{40} name=ProxyAdmin"
    assert_output --regexp "Contract deployed address=0x[a-fA-F0-9]{40} name=TickLens"
    assert_output --regexp "Contract deployed address=0x[a-fA-F0-9]{40} name=NFTDescriptor"
    assert_output --regexp "Contract deployed address=0x[a-fA-F0-9]{40} name=NonfungibleTokenPositionDescriptor"
    assert_output --regexp "Contract deployed address=0x[a-fA-F0-9]{40} name=TransparentUpgradeableProxy"
    assert_output --regexp "Contract deployed address=0x[a-fA-F0-9]{40} name=NonfungiblePositionManager"
    assert_output --regexp "Contract deployed address=0x[a-fA-F0-9]{40} name=V3Migrator"
    assert_output --regexp "Contract deployed address=0x[a-fA-F0-9]{40} name=UniswapV3Staker"
    assert_output --regexp "Contract deployed address=0x[a-fA-F0-9]{40} name=QuoterV2"
    assert_output --regexp "Contract deployed address=0x[a-fA-F0-9]{40} name=SwapRouter02"

    # Check if ERC20 tokens were minted
    assert_output --regexp "Minted tokens amount=[0-9]+ recipient=0x[a-fA-F0-9]{40} token=SwapperA"
    assert_output --regexp "Minted tokens amount=[0-9]+ recipient=0x[a-fA-F0-9]{40} token=SwapperB"

    # Check if liquidity pool was created and initialized
    assert_output --regexp "Pool created and initialized fees=[0-9]+"

    # Check if liquidity was provided to the pool
    assert_output --regexp "Liquidity provided to the pool liquidity=[0-9]+"

    # Check if transaction got executed successfully
    assert_output --regexp "Starting main load test loop currentNonce=[0-9]+"
    assert_output --regexp "Finished main load test loop lastNonce=[0-9]+ startNonce=[0-9]+"
    assert_output --regexp "Got final block number currentNonce=[0-9]+ final block number=[0-9]+"
    assert_output --regexp "Num errors numErrors=0"
    assert_output --regexp "Finished"
}

