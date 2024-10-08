setup() {
    load 'helpers/common-setup'
    load 'helpers/common'
    _common_setup

    readonly enclave=${ENCLAVE:-cdk-v1}
    readonly sequencer=${KURTOSIS_NODE:-cdk-erigon-sequencer-001}
    readonly node=${KURTOSIS_NODE:-cdk-erigon-node-001}
    readonly rpc_url=${RPC_URL:-$(kurtosis port print "$enclave" "$node" http-rpc)}
    readonly key=${SENDER_key:-"12d7de8621a77640c9241b2595ba78ce443d05e94090365ab3bb5e19df82c625"}
    readonly receiver=${RECEIVER:-"0x85dA99c8a7C2C95964c8EfD687E95E632Fc533D6"}
    readonly data_dir=${ACL_DATA_DIR:-"/home/erigon/data/dynamic-kurtosis-sequencer/txpool/acls"}
    readonly kurtosis_sequencer_wrapper=${KURTOSIS_WRAPPER:-"kurtosis service exec $enclave $sequencer"}
}

teardown() {
    run set_acl_mode "disabled"
}

# Helper function to add address to acl dynamically
add_to_access_list() {
    local acl_type="$1"
    local policy="$2"
    local sender=$(cast wallet address "$key")

    run $kurtosis_sequencer_wrapper "acl add --datadir $data_dir --address $sender --type $acl_type --policy $policy"
}

# Helper function to set the acl mode command dynamically
set_acl_mode() {
    local mode="$1"

    run $kurtosis_sequencer_wrapper "acl mode --datadir $data_dir --mode $mode"
}

@test "Test Block List - Sending regular transaction when address not in block list" {
    local value="10ether"
    run set_acl_mode "blocklist"
    run sendTx $key $receiver $value

    assert_success
    assert_output --regexp "Transaction successful \(transaction hash: 0x[a-fA-F0-9]{64}\)"
}

@test "Test Block List - Sending contracts deploy transaction when address not in block list" {
    local contract_artifact="./contracts/erc20mock/ERC20Mock.json"
    run set_acl_mode "blocklist"
    run deployContract $key $contract_artifact

    assert_success

    contract_addr=$(echo "$output" | tail -n 1)
    assert_output --regexp "0x[a-fA-F0-9]{40}"
}

@test "Test Block List - Sending regular transaction when address is in block list" {
    local value="10ether"

    run set_acl_mode "blocklist"
    run add_to_access_list "blocklist" "sendTx"

    run sendTx $key $receiver $value

    assert_failure
    assert_output --partial "sender disallowed to send tx by ACL policy"
}

@test "Test Block List - Sending contracts deploy transaction when address is in block list" {
    local contract_artifact="./contracts/erc20mock/ERC20Mock.json"

    run set_acl_mode "blocklist"
    run add_to_access_list "blocklist" "deploy"
    run deployContract $key $contract_artifact

    assert_failure
    assert_output --partial "sender disallowed to deploy contract by ACL policy"
}

@test "Test Allow List - Sending regular transaction when address not in allow list" {
    local value="10ether"

    run set_acl_mode "allowlist"
    run sendTx $key $receiver $value

    assert_failure
    assert_output --partial "sender disallowed to send tx by ACL policy"
}

@test "Test Allow List - Sending contracts deploy transaction when address not in allow list" {
    local contract_artifact="./contracts/erc20mock/ERC20Mock.json"

    run set_acl_mode "allowlist"
    run deployContract $key $contract_artifact

    assert_failure
    assert_output --partial "sender disallowed to deploy contract by ACL policy"
}

@test "Test Allow List - Sending regular transaction when address is in allow list" {
    local value="10ether"

    run set_acl_mode "allowlist"    
    run add_to_access_list "allowlist" "sendTx"    
    run sendTx $key $receiver $value
    
    assert_success
    assert_output --regexp "Transaction successful \(transaction hash: 0x[a-fA-F0-9]{64}\)"
}

@test "Test Allow List - Sending contracts deploy transaction when address is in allow list" {
    local contract_artifact="./contracts/erc20mock/ERC20Mock.json"

    run set_acl_mode "allowlist"    
    run add_to_access_list "allowlist" "deploy" 
    run deployContract $key $contract_artifact

    assert_success

    contract_addr=$(echo "$output" | tail -n 1)
    assert_output --regexp "0x[a-fA-F0-9]{40}"
}
