#!/usr/bin/env bash

function deployContract() {
    local private_key="$1"
    local contract_artifact="$2"

    # Check if rpc_url is available
    if [[ -z "$rpc_url" ]]; then
        echo "Error: rpc_url environment variable is not set."
        return 1
    fi

    if [[ ! -f "$contract_artifact" ]]; then
        echo "Error: Contract artifact '$contract_artifact' does not exist."
        return 1
    fi

    # Get the sender address
    local sender_addr=$(cast wallet address "$private_key")
    if [[ $? -ne 0 ]]; then
        echo "Error: Failed to retrieve sender address."
        return 1
    fi

    echo "Attempting to deploy contract artifact '$contract_artifact' to $rpc_url (sender: $sender_addr)" >&3

    # Get bytecode from the contract artifact
    local bytecode=$(jq -r .bytecode "$contract_artifact")
    if [[ -z "$bytecode" || "$bytecode" == "null" ]]; then
        echo "Error: Failed to read bytecode from $contract_artifact"
        return 1
    fi

    # Send the transaction and capture the output
    local cast_output=$(cast send --rpc-url "$rpc_url" \
        --private-key "$private_key" \
        --legacy \
        --create "$bytecode" \
        2>&1)

    # Check if cast send was successful
    if [[ $? -ne 0 ]]; then
        echo "Error: Failed to send transaction."
        echo "$cast_output"
        return 1
    fi

    echo "Deploy contract output:" >&3
    echo "$cast_output" >&3

    # Extract the contract address from the output
    local deployed_contract_address=$(echo "$cast_output" | grep 'contractAddress' | sed 's/contractAddress\s\+//')
    echo "Deployed contract address: $deployed_contract_address" >&3

    if [[ -z "$deployed_contract_address" ]]; then
        echo "Error: Failed to extract deployed contract address"
        return 1
    fi

    if [[ ! "$deployed_contract_address" =~ ^0x[a-fA-F0-9]{40}$ ]]; then
        echo "Error: Invalid contract address $deployed_contract_address"
        return 1
    fi

    # Print contract address for return
    echo "$deployed_contract_address"

    return 0
}

function sendTx() {
    # Check if at least 3 arguments are provided
    if [[ $# -lt 3 ]]; then
        echo "Usage: sendTx <private_key> <receiver> <value_or_function_signature> [<param1> <param2> ...]"
        return 1
    fi

    local private_key="$1"           # Sender private key
    local account_addr="$2"          # Receiver address
    local value_or_function_sig="$3" # Value or function signature

    # Error handling: Ensure the receiver is a valid Ethereum address
    if [[ ! "$account_addr" =~ ^0x[a-fA-F0-9]{40}$ ]]; then
        echo "Error: Invalid receiver address '$account_addr'."
        return 1
    fi

    shift 3 # Shift the first 3 arguments (private_key, account_addr, value_or_function_sig)

    local sender_addr
    sender_addr=$(cast wallet address "$private_key")
    if [[ $? -ne 0 ]]; then
        echo "Error: Failed to extract the sender address for $private_key"
        return 1
    fi

    # Check initial ether balance of sender and receiver
    local sender_initial_balance=$(cast balance "$sender_addr" --ether --rpc-url "$rpc_url") || return 1
    local receiver_initial_balance=$(cast balance "$account_addr" --ether --rpc-url "$rpc_url") || return 1

    # Check if the first remaining argument is a numeric value (Ether to be transferred)
    if [[ "$value_or_function_sig" =~ ^[0-9]+(\.[0-9]+)?(ether)?$ ]]; then
        # Case: EOA transaction (Ether transfer)
        echo "Sending EOA transaction (RPC URL: $rpc_url, sender: $sender_addr) to: $account_addr " \
            "with value: $value_or_function_sig" >&3

        cast_output=$(cast send --rpc-url "$rpc_url" \
            --private-key "$private_key" \
            "$account_addr" --value "$value_or_function_sig" \
            --legacy \
            2>&1)
    else
        # Case: Smart contract transaction (contract interaction with function signature and parameters)
        local params=("$@") # Collect all remaining arguments as function parameters

        # Verify if the function signature starts with "function"
        if [[ ! "$value_or_function_sig" =~ ^function\ .+\(.+\)$ ]]; then
            echo "Error: Invalid function signature format '$value_or_function_sig'."
            return 1
        fi

        echo "Sending smart contract transaction (RPC URL: $rpc_url, sender: $sender_addr) to $account_addr" \
            "with function signature: '$value_or_function_sig' and params: ${params[*]}" >&3

        # Send the smart contract interaction using cast
        cast_output=$(cast send --rpc-url "$rpc_url" \
            --private-key "$private_key" \
            "$account_addr" "$value_or_function_sig" "${params[@]}" \
            --legacy \
            2>&1)
    fi

    # Check if the transaction was successful
    if [[ $? -ne 0 ]]; then
        echo "Error: Failed to send transaction. The cast send output:"
        echo "$cast_output"
        return 1
    fi

    # Extract the transaction hash from the output
    local tx_hash=$(echo "$cast_output" | grep 'transactionHash' | awk '{print $2}' | tail -n 1)
    echo "Tx hash: $tx_hash"

    if [[ -z "$tx_hash" ]]; then
        echo "Error: Failed to extract transaction hash."
        return 1
    fi

    if [[ "$value_or_function_sig" =~ ^[0-9]+(ether)?$ ]]; then
        checkTransactionSuccess "$sender_addr" "$receiver" "$value_or_function_sig" "$tx_hash" "$sender_initial_balance" "$receiver_initial_balance"
        if [[ $? -ne 0 ]]; then
            echo "Error: Balance not updated correctly."
            return 1
        fi
    fi

    echo "Transaction successful (transaction hash: $tx_hash)"

    return 0
}

function queryContract() {
    local addr="$1"          # Contract address
    local funcSignature="$2" # Function signature
    shift 2                  # Shift past the first two arguments
    local params=("$@")      # Collect remaining arguments as parameters array

    echo "Querying state of $addr account (RPC URL: $rpc_url) with function signature: '$funcSignature' and params: ${params[*]}" >&3

    # Check if rpc_url is available
    if [[ -z "$rpc_url" ]]; then
        echo "Error: rpc_url environment variable is not set."
        return 1
    fi

    # Check if the contract address is valid
    if [[ ! "$addr" =~ ^0x[a-fA-F0-9]{40}$ ]]; then
        echo "Error: Invalid contract address '$addr'."
        return 1
    fi

    # Call the contract using `cast call`
    local result
    result=$(cast call --rpc-url "$rpc_url" "$addr" "$funcSignature" "${params[@]}" 2>&1)

    # Check if the call was successful
    if [[ $? -ne 0 ]]; then
        echo "Error: Failed to query contract."
        echo "$result"
        return 1
    fi

    # Return the result (contract query response)
    echo "$result"

    return 0
}

function rpcQuery() {
    local method="$1" # The JSON-RPC method name
    shift
    local params=("$@") # Remaining arguments are the parameters for the RPC method

    # Check if rpc_url is available
    if [[ -z "$rpc_url" ]]; then
        echo "Error: rpc_url environment variable is not set."
        return 1
    fi

    # Use cast to perform a generic RPC call
    local response
    response=$(cast rpc "$method" "${params[@]}" --rpc-url "$rpc_url" 2>&1)

    # Check if the cast rpc command was successful
    if [[ $? -ne 0 ]]; then
        echo "Error: Failed to perform RPC query."
        echo "$response"
        return 1
    fi

    echo "$response"
    return 0
}

function checkTransactionSuccess() {
    local sender_addr="$1"
    local receiver="$2"
    local value_or_function_sig="$3"
    local tx_hash="$4"
    local sender_initial_balance="$5"
    local receiver_initial_balance="$6"

    local sender_final_balance=$(cast balance "$sender_addr" --ether --rpc-url "$rpc_url") || return 1
    local gas_used=$(cast tx "$tx_hash" --rpc-url "$rpc_url" | grep '^gas ' | awk '{print $2}')
    local gas_price=$(cast tx "$tx_hash" --rpc-url "$rpc_url" | grep '^gasPrice' | awk '{print $2}')
    local gas_fee=$(echo "$gas_used * $gas_price" | bc)
    local gas_fee_in_ether=$(cast to-unit "$gas_fee" ether)

    local sender_balance_change=$(echo "$sender_initial_balance - $sender_final_balance" | bc)
    echo "Sender balance changed by: '$sender_balance_change' wei"
    echo "Gas fee paid: '$gas_fee_in_ether' ether"

    local receiver_final_balance=$(cast balance "$receiver" --ether --rpc-url "$rpc_url") || return 1
    local receiver_balance_change=$(echo "$receiver_final_balance - $receiver_initial_balance" | bc)
    echo "Receiver balance changed by: '$receiver_balance_change' wei"

    # Trim 'ether' suffix from value_or_function_sig to get the numeric part
    local value_in_ether=$(echo "$value_or_function_sig" | sed 's/ether$//')

    if ! echo "$receiver_balance_change == $value_in_ether" | bc -l; then
        echo "Error: receiver balance updated incorrectly. Expected: $value_in_ether, Actual: $receiver_balance_change"
        return 1
    fi

    # Calculate expected sender balance change
    local expected_sender_change=$(echo "$value_in_ether + $gas_fee_in_ether" | bc)
    if ! echo "$sender_balance_change == $expected_sender_change" | bc -l; then
        echo "Error: sender balance updated incorrectly. Expected: $expected_sender_change, Actual: $sender_balance_change"
        return 1
    fi
}
