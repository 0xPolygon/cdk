setup() {
    load 'helpers/common-setup'
    load 'helpers/common'
    _common_setup

    readonly enclave=${ENCLAVE:-cdk-v1}
    readonly node=${KURTOSIS_NODE:-cdk-erigon-node-001}
    readonly rpc_url=${RPC_URL:-$(kurtosis port print "$enclave" "$node" http-rpc)}
    readonly private_key=${SENDER_PRIVATE_KEY:-"12d7de8621a77640c9241b2595ba78ce443d05e94090365ab3bb5e19df82c625"}
    readonly receiver=${RECEIVER:-"0x85dA99c8a7C2C95964c8EfD687E95E632Fc533D6"}
}

@test "Send EOA transaction" {
    local value="10ether"

    run sendTx "$private_key" "$receiver" "$value"
    assert_success
    assert_output --regexp "Transaction successful \(transaction hash: 0x[a-fA-F0-9]{64}\)"
}

@test "Deploy ERC20Mock contract" {
    local contract_artifact="./contracts/erc20mock/ERC20Mock.json"

    # Deploy ERC20Mock
    run deployContract "$private_key" "$contract_artifact"
    assert_success
    contract_addr=$(echo "$output" | tail -n 1)

    # Mint ERC20 tokens
    local mintFnSig="function mint(address receiver, uint256 amount)"
    local amount="5"

    run sendTx "$private_key" "$contract_addr" "$mintFnSig" "$receiver" "$amount"
    assert_success
    assert_output --regexp "Transaction successful \(transaction hash: 0x[a-fA-F0-9]{64}\)"

    # Assert that balance is correct
    local balanceOfFnSig="function balanceOf(address) (uint256)"
    run queryContract "$contract_addr" "$balanceOfFnSig" "$receiver"
    assert_success
    receiverBalance=$(echo "$output" | tail -n 1)

    # Convert balance and amount to a standard format for comparison (e.g., remove any leading/trailing whitespace)
    receiverBalance=$(echo "$receiverBalance" | xargs)
    amount=$(echo "$amount" | xargs)

    # Check if the balance is equal to the amount
    assert_equal "$receiverBalance" "$amount"
}
