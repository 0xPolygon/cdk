#!/usr/bin/env bash
function deployContract() {
    local private_key="$1"
    local contract_artifact="$2"

    if [[ ! -f "$contract_artifact" ]]; then
        echo "Error: Contract artifact $contract_artifact does not exist."
        return 1
    fi

    echo "Attempting to deploy contract artifact $contract_artifact to $rpc_url"

    local bytecode
    bytecode=$(jq -r .bytecode "$contract_artifact")

    if [[ -z "$bytecode" || "$bytecode" == "null" ]]; then
        echo "Error: Failed to read bytecode from $contract_artifact"
        return 1
    fi

    # Send the transaction and capture the output
    cast_output=$(cast send --rpc-url "$rpc_url" \
        --private-key "$private_key" \
        --create "$bytecode" 2>&1)

    echo "$cast_output"

    # Extract the contract address from the output
    deployed_contract_address=$(echo "$cast_output" | grep -oP 'Deployed to: \K(0x[a-fA-F0-9]+)')

    if [[ -z "$deployed_contract_address" ]]; then
        echo "Error: Failed to extract deployed contract address"
        return 1
    fi

    echo "Deployed contract address: $deployed_contract_address"

    # Return the contract address as a return value
    return_value="$deployed_contract_address"
}

function sendTx() {
    # Check if at least 3 arguments are provided
    if [[ $# -lt 3 ]]; then
        echo "Usage: sendTx <private_key> <receiver> <value_or_function_signature>] [<param1> <param2> ...]"
        return 1
    fi

    # Assign variables from function arguments
    local private_key="$1"
    local receiver="$2"
    shift 2 # Shift the first 2 arguments (private_key, receiver)

    local first_remaining_arg="$1"
    shift # Shift the first remaining argument (value or function signature)

    # Error handling: Ensure the receiver is a valid Ethereum address
    if [[ ! "$receiver" =~ ^0x[a-fA-F0-9]{40}$ ]]; then
        echo "Error: Invalid receiver address '$receiver'."
        return 1
    fi

    # Check if the first remaining argument is a numeric value (Ether to be transferred)
    if [[ "$first_remaining_arg" =~ ^[0-9]+(ether)?$ ]]; then
        # Case: EOA transaction (Ether transfer)
        echo "Sending EOA transaction (RPC URL: $rpc_url) to: $receiver with value: $first_remaining_arg"
        cast_output=$(cast send --rpc-url "$rpc_url" \
            --private-key "$private_key" \
            "$receiver" --value "$first_remaining_arg" \
            --legacy \
            2>&1)
    else
        # Case: Smart contract transaction (contract interaction with function signature and parameters)
        local functionSignature="$first_remaining_arg"
        local params=("$@") # Collect all remaining arguments as function parameters
        echo "Sending smart contract transaction (RPC URL: $rpc_url) to $receiver with function signature: $functionSignature and params: ${params[*]}"

        # Prepare the function signature with parameters for cast send
        cast_output=$(cast send --rpc-url "$rpc_url" \
            --private-key "$private_key" \
            "$receiver" "$functionSignature" "${params[@]}" \
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
    local tx_hash=$(echo "$cast_output" | grep 'transactionHash' | sed 's/transactionHash\s\+//')
    echo "Tx hash: $tx_hash"

    if [[ -z "$tx_hash" ]]; then
        echo "Error: Failed to extract transaction hash."
        return 1
    fi

    echo "Transaction successful (transaction hash: $tx_hash)"
    return 0
}
