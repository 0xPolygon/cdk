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
        echo "Error: Contract artifact $contract_artifact does not exist."
        return 1
    fi

    # Get the sender address
    local senderAddr=$(cast wallet address "$private_key")
    if [[ $? -ne 0 ]]; then
        echo "Error: Failed to retrieve sender address."
        return 1
    fi
    
    echo "Attempting to deploy contract artifact $contract_artifact to $rpc_url (sender: $senderAddr)"

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

    echo "$cast_output"

    # Extract the contract address from the output
    local deployed_contract_address=$(echo "$cast_output" | grep 'contractAddress' | sed 's/contractAddress\s\+//')

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
    local receiver="$2"              # Receiver address
    local value_or_function_sig="$3" # Value or function signature

    # Error handling: Ensure the receiver is a valid Ethereum address
    if [[ ! "$receiver" =~ ^0x[a-fA-F0-9]{40}$ ]]; then
        echo "Error: Invalid receiver address '$receiver'."
        return 1
    fi

    shift 3 # Shift the first 3 arguments (private_key, receiver, value_or_function_sig)

    local senderAddr
    senderAddr=$(cast wallet address "$private_key")
    if [[ $? -ne 0 ]]; then
        echo "Error: Failed to extract the sender address for $private_key"
        return 1
    fi

    # Check if the first remaining argument is a numeric value (Ether to be transferred)
    if [[ "$value_or_function_sig" =~ ^[0-9]+(ether)?$ ]]; then
        # Case: EOA transaction (Ether transfer)
        echo "Sending EOA transaction (RPC URL: $rpc_url, sender: $senderAddr) to: $receiver with value: $value_or_function_sig"
        cast_output=$(cast send --rpc-url "$rpc_url" \
            --private-key "$private_key" \
            "$receiver" --value "$value_or_function_sig" \
            --legacy \
            2>&1)
    else
        # Case: Smart contract transaction (contract interaction with function signature and parameters)
        local params=("$@") # Collect all remaining arguments as function parameters

        echo "$value_or_function_sig"

        # Verify if the function signature starts with "function"
        if [[ ! "$value_or_function_sig" =~ ^function\ .+\(.+\)$ ]]; then
            echo "Error: Invalid function signature format '$value_or_function_sig'."
            return 1
        fi

        echo "Sending smart contract transaction (RPC URL: $rpc_url, sender: $senderAddr) to $receiver with function signature: '$value_or_function_sig' and params: ${params[*]}"

        # Send the smart contract interaction using cast
        cast_output=$(cast send --rpc-url "$rpc_url" \
            --private-key "$private_key" \
            "$receiver" "$value_or_function_sig" "${params[@]}" \
            --legacy \
            2>&1)
    fi

    # Print the cast output
    echo "cast send output:"
    echo "$cast_output"

    # Check if the transaction was successful
    if [[ $? -ne 0 ]]; then
        echo "Error: Failed to send transaction. Output:"
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

    echo "Transaction successful (transaction hash: $tx_hash)"

    return 0
}

function queryContract() {
    local addr="$1"          # Contract address
    local funcSignature="$2" # Function signature
    shift 2                  # Shift past the first two arguments
    local params=("$@")      # Collect remaining arguments as parameters array

    echo "Querying state of $addr account (RPC URL: $rpc_url) with function signature: '$funcSignature' and params: ${params[*]}"

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
