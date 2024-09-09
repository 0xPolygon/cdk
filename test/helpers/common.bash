#!/usr/bin/env bash
function deployContract() {
    local rpc_url="$1"
    local private_key="$2"
    local contract_artifact="$3"

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
        echo "Usage: sendTx <rpc_url> <private_key> <receiver> [<value_or_function_signature> <param1> <param2> ...]"
        return 1
    fi

    # Assign variables from function arguments
    local rpc_url="$1"
    local private_key="$2"
    local receiver="$3"
    shift 3 # Shift the first 3 arguments (rpc_url, private_key, receiver)

    local first_remaining_arg="$1"
    shift # Shift the first remaining argument (value or function signature)

    # Error handling: Ensure the receiver is a valid Ethereum address
    if [[ ! "$receiver" =~ ^0x[a-fA-F0-9]{40}$ ]]; then
        echo "Error: Invalid receiver address '$receiver'."
        return 1
    fi

    # Check if the first remaining argument is a numeric value (Ether to be transferred)
    if [[ "$first_remaining_arg" =~ ^[0-9]+$ ]]; then
        # Case: EOA transaction (Ether transfer)
        echo "Sending Ether transaction to EOA: $receiver with value: $first_remaining_arg Wei"
        cast_output=$(cast send --rpc-url "$rpc_url" \
            --private-key "$private_key" \
            "$receiver" --value "$first_remaining_arg" \
            2>&1)
    else
        # Case: Smart contract transaction (contract interaction with function signature and parameters)
        local functionSignature="$first_remaining_arg"
        local params=("$@") # Collect all remaining arguments as function parameters
        echo "Sending smart contract transaction to $receiver with function signature: $functionSignature and params: ${params[*]}"

        # Prepare the function signature with parameters for cast send
        cast_output=$(cast send --rpc-url "$rpc_url" \
            --private-key "$private_key" \
            "$receiver" "$functionSignature" "${params[@]}" \
            2>&1)
    fi

    # Check if the transaction was successful
    if [[ $? -ne 0 ]]; then
        echo "Error: Failed to send transaction. Output:"
        echo "$cast_output"
        return 1
    fi

    # Transaction was successful, extract and display the transaction hash
    tx_hash=$(echo "$cast_output" | grep -oP '(?<=Transaction hash: )0x[a-fA-F0-9]+')

    if [[ -z "$tx_hash" ]]; then
        echo "Error: Failed to extract transaction hash."
        return 1
    fi

    echo "Transaction successful! Transaction hash: $tx_hash"
    return 0
}
