#!/usr/bin/env bash

function deployContract() {
    local rpc_url="$1"
    local private_key="$2"
    local contract_artifact="$3"

    # Check if rpc_url is available
    if [[ -z "$rpc_url" ]]; then
        echo "Error: rpc_url parameter is not set."
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
        echo "$cast_output"
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
    # Check if at least 4 arguments are provided
    if [[ $# -lt 4 ]]; then
        echo "Usage: sendTx <private_key> <receiver> <value_or_function_signature> [<param1> <param2> ...]"
        return 1
    fi

    local rpc_url="$1"
    local private_key="$2"           # Sender private key
    local account_addr="$3"          # Receiver address
    local value_or_function_sig="$4" # Value or function signature

    # Error handling: Ensure the receiver is a valid Ethereum address
    if [[ ! "$receiver_addr" =~ ^0x[a-fA-F0-9]{40}$ ]]; then
        echo "Error: Invalid receiver address '$receiver_addr'."
        return 1
    fi

    shift 4 # Shift the first 4 arguments (rpc_url, private_key, account_addr, value_or_function_sig)
    local params=("$@") # Collect all remaining arguments as function parameters

    # Get sender address from private key
    local sender_addr
    sender_addr=$(cast wallet address "$private_key") || {
        echo "Error: Failed to extract the sender address."
        return 1
    }

    # Get initial ether balances of sender and receiver
    local sender_initial_balance receiver_initial_balance
    sender_initial_balance=$(cast balance "$sender_addr" --ether --rpc-url "$rpc_url") || return 1
    receiver_initial_balance=$(cast balance "$receiver_addr" --ether --rpc-url "$rpc_url") || return 1

    # Check if the value_or_function_sig is a numeric value (Ether to be transferred)
    if [[ "$value_or_function_sig" =~ ^[0-9]+(\.[0-9]+)?(ether)?$ ]]; then
        # Case: Ether transfer (EOA transaction)
        send_eoa_transaction "$private_key" "$receiver_addr" "$value_or_function_sig" "$sender_addr" "$sender_initial_balance" "$receiver_initial_balance"
    else
        # Case: Smart contract interaction (contract interaction with function signature and parameters)
        send_smart_contract_transaction "$private_key" "$receiver_addr" "$value_or_function_sig" "$sender_addr" "${params[@]}"
    fi
}

function send_eoa_transaction() {
    local private_key="$1"
    local receiver_addr="$2"
    local value="$3"
    local sender_addr="$4"
    local sender_initial_balance="$5"
    local receiver_initial_balance="$6"

    echo "Sending EOA transaction to: $receiver_addr with value: $value" >&3

    # Send transaction via cast
    local cast_output tx_hash
    cast_output=$(cast send --rpc-url "$rpc_url" --private-key "$private_key" "$receiver_addr" --value "$value" --legacy 2>&1)
    if [[ $? -ne 0 ]]; then
        echo "Error: Failed to send transaction. Output:"
        echo "$cast_output"
        return 1
    fi

    tx_hash=$(extract_tx_hash "$cast_output")
    [[ -z "$tx_hash" ]] && {
        echo "Error: Failed to extract transaction hash."
        return 1
    }

    checkBalances "$sender_addr" "$receiver_addr" "$value" "$tx_hash" "$sender_initial_balance" "$receiver_initial_balance"
    if [[ $? -ne 0 ]]; then
        echo "Error: Balance not updated correctly."
        return 1
    fi

    echo "Transaction successful (transaction hash: $tx_hash)"
}

function send_smart_contract_transaction() {
    local private_key="$1"
    local receiver_addr="$2"
    local function_sig="$3"
    local sender_addr="$4"
    shift 4
    local params=("$@")

    # Verify if the function signature starts with "function"
    if [[ ! "$function_sig" =~ ^function\ .+\(.+\)$ ]]; then
        echo "Error: Invalid function signature format '$function_sig'."
        return 1
    fi

    echo "Sending smart contract transaction to $receiver_addr with function signature: '$function_sig' and params: ${params[*]}" >&3

    # Send the smart contract interaction using cast
    local cast_output tx_hash
    cast_output=$(cast send --rpc-url "$rpc_url" --private-key "$private_key" "$receiver_addr" "$function_sig" "${params[@]}" --legacy 2>&1)
    if [[ $? -ne 0 ]]; then
        echo "Error: Failed to send transaction. Output:"
        echo "$cast_output"
        return 1
    fi

    tx_hash=$(extract_tx_hash "$cast_output")
    [[ -z "$tx_hash" ]] && {
        echo "Error: Failed to extract transaction hash."
        return 1
    }

    echo "Transaction successful (transaction hash: $tx_hash)"
}

function extract_tx_hash() {
    local cast_output="$1"
    echo "$cast_output" | grep 'transactionHash' | awk '{print $2}' | tail -n 1
}

function queryContract() {
    local rpc_url="$1"       # RPC URL
    local addr="$2"          # Contract address
    local funcSignature="$3" # Function signature
    shift 3                  # Shift past the first 3 arguments
    local params=("$@")      # Collect remaining arguments as parameters array

    echo "Querying state of $addr account (RPC URL: $rpc_url) with function signature: '$funcSignature' and params: ${params[*]}" >&3

    # Check if url is available
    if [[ -z "$rpc_url" ]]; then
        echo "Error: rpc_url parameter is not provided."
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

function checkBalances() {
    local sender="$1"
    local receiver="$2"
    local amount="$3"
    local tx_hash="$4"
    local sender_initial_balance="$5"
    local receiver_initial_balance="$6"

    # Ethereum address regex: 0x followed by 40 hexadecimal characters
    if [[ ! "$sender" =~ ^0x[a-fA-F0-9]{40}$ ]]; then
        echo "Error: Invalid sender address '$sender'."
        return 1
    fi

    if [[ ! "$receiver" =~ ^0x[a-fA-F0-9]{40}$ ]]; then
        echo "Error: Invalid receiver address '$receiver'."
        return 1
    fi

    # Transaction hash regex: 0x followed by 64 hexadecimal characters
    if [[ ! "$tx_hash" =~ ^0x[a-fA-F0-9]{64}$ ]]; then
        echo "Error: Invalid transaction hash: $tx_hash".
        return 1
    fi

    local sender_final_balance=$(cast balance "$sender" --ether --rpc-url "$rpc_url") || return 1
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

    # Trim 'ether' suffix from amount to get the numeric part
    local value_in_ether=$(echo "$amount" | sed 's/ether$//')

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
