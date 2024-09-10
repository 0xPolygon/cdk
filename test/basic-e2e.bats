setup() {
    load 'helpers/common-setup'
    _common_setup

    readonly enclave=${ENCLAVE:-cdk-v1}
    readonly node=${KURTOSIS_NODE:-cdk-erigon-node-001}
    readonly rpc_url=${RPC_URL:-$(kurtosis port print "$enclave" "$node" http-rpc)}
}

@test "Send EOA transaction" {
    load 'helpers/common'

    local private_key="12d7de8621a77640c9241b2595ba78ce443d05e94090365ab3bb5e19df82c625"
    local receiver="0x85dA99c8a7C2C95964c8EfD687E95E632Fc533D6"
    local value="10ether"

    run sendTx $private_key $receiver $value

    assert_success
    assert_output --partial 'Transaction successful'
}

@test "Deploy ERC20Mock contract" {
    load 'helpers/common'

    local private_key="12d7de8621a77640c9241b2595ba78ce443d05e94090365ab3bb5e19df82c625"
    local contract_artifact="./test/contracts/erc20mock/ERC20Mock.json"

    # Deploy ERC20Mock
    contract_addr=$(deployContract "$private_key" "$contract_artifact")
    [[ $? -eq 0 ]] || fail "Contract deployment failed"
    echo "Contract addr: $contract_addr"

    # Mint ERC20 tokens
    local receiver="0x85dA99c8a7C2C95964c8EfD687E95E632Fc533D6"
    local mintFnSig="function mint(address to, uint256 amount)"
    local amount="5"

    run sendTx "$private_key" "$contract_addr" "$mintFnSig" "$receiver" "$amount"
    assert_success
    assert_output --partial 'Transaction successful'

    # Assert that balance is correct
    local balanceOfFnSig="function balanceOf(address) (uint256)"
    receiverBalance=$(queryContract "$contract_addr" "$balanceOfFnSig" "$receiver")
    assert_success

    # Convert balance and amount to a standard format for comparison (e.g., remove any leading/trailing whitespace)
    receiverBalance=$(echo "$receiverBalance" | xargs)
    amount=$(echo "$amount" | xargs)

    # Check if the balance is equal to the amount
    assert_equal "$receiverBalance" "$amount"
}
