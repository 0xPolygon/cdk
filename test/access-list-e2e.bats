setup() {
    load 'helpers/common-setup'
    load 'helpers/common'
    _common_setup

    readonly enclave=${ENCLAVE:-cdk-v1}
    readonly sequencer=${KURTOSIS_NODE:-cdk-erigon-sequencer-001}
    readonly node=${KURTOSIS_NODE:-cdk-erigon-node-001}
    readonly rpc_url=${RPC_URL:-$(kurtosis port print "$enclave" "$node" http-rpc)}
}

teardown() {
    run kurtosis service exec $enclave $sequencer "acl mode --datadir /home/erigon/data/dynamic-kurtosis-sequencer/txpool/acls --mode disabled"
}

@test "Test Block List - Sending regular transaction when address not in block list" {
    local private_key=${RAW_PRIVATE_KEY:-"12d7de8621a77640c9241b2595ba78ce443d05e94090365ab3bb5e19df82c625"}
    local receiver="0x85dA99c8a7C2C95964c8EfD687E95E632Fc533D6"
    local value="10ether"

    run kurtosis service exec $enclave $sequencer "acl mode --datadir /home/erigon/data/dynamic-kurtosis-sequencer/txpool/acls --mode blocklist"
    run sendTx $private_key $receiver $value

    assert_success
    assert_output --regexp "Transaction successful \(transaction hash: 0x[a-fA-F0-9]{64}\)"
}

@test "Test Block List - Sending contracts deploy transaction when address not in block list" {
    local private_key="12d7de8621a77640c9241b2595ba78ce443d05e94090365ab3bb5e19df82c625"
    local contract_artifact="./contracts/erc20mock/ERC20Mock.json"

    run kurtosis service exec $enclave $sequencer "acl mode --datadir /home/erigon/data/dynamic-kurtosis-sequencer/txpool/acls --mode blocklist"
    run deployContract $private_key $contract_artifact

    assert_success
}

@test "Test Block List - Sending regular transaction when address is in block list" {
    local private_key="12d7de8621a77640c9241b2595ba78ce443d05e94090365ab3bb5e19df82c625"
    local sender=$(cast wallet address "$private_key")
    local receiver="0x85dA99c8a7C2C95964c8EfD687E95E632Fc533D6"
    local value="10ether"

    run kurtosis service exec $enclave $sequencer "acl mode --datadir /home/erigon/data/dynamic-kurtosis-sequencer/txpool/acls --mode blocklist"
    run kurtosis service exec $enclave $sequencer "acl add --datadir /home/erigon/data/dynamic-kurtosis-sequencer/txpool/acls --address $sender --type blocklist --policy sendTx"
    run sendTx $private_key $receiver $value

    assert_failure
    assert_output --partial "sender disallowed to send tx by ACL policy"
}

@test "Test Block List - Sending contracts deploy transaction when address is in block list" {
    local private_key="12d7de8621a77640c9241b2595ba78ce443d05e94090365ab3bb5e19df82c625"
    local sender=$(cast wallet address "$private_key")
    local contract_artifact="./contracts/erc20mock/ERC20Mock.json"

    run kurtosis service exec $enclave $sequencer "acl mode --datadir /home/erigon/data/dynamic-kurtosis-sequencer/txpool/acls --mode blocklist"
    run kurtosis service exec $enclave $sequencer "acl add --datadir /home/erigon/data/dynamic-kurtosis-sequencer/txpool/acls --address $sender --type blocklist --policy deploy"

    run deployContract $private_key $contract_artifact

    assert_failure
    assert_output --partial "sender disallowed to deploy contract by ACL policy"
}

@test "Test Allow List - Sending regular transaction when address not in allow list" {
    local private_key="12d7de8621a77640c9241b2595ba78ce443d05e94090365ab3bb5e19df82c625"
    local receiver="0x85dA99c8a7C2C95964c8EfD687E95E632Fc533D6"
    local value="10ether"

    run kurtosis service exec $enclave $sequencer "acl mode --datadir /home/erigon/data/dynamic-kurtosis-sequencer/txpool/acls --mode allowlist"
    run sendTx $private_key $receiver $value

    assert_failure
    assert_output --partial "sender disallowed to send tx by ACL policy"
}

@test "Test Allow List - Sending contracts deploy transaction when address not in allow list" {
    local private_key="12d7de8621a77640c9241b2595ba78ce443d05e94090365ab3bb5e19df82c625"
    local contract_artifact="./contracts/erc20mock/ERC20Mock.json"

    run kurtosis service exec $enclave $sequencer "acl mode --datadir /home/erigon/data/dynamic-kurtosis-sequencer/txpool/acls --mode allowlist"
    run deployContract $private_key $contract_artifact

    assert_failure
    assert_output --partial "sender disallowed to deploy contract by ACL policy"
}

@test "Test Allow List - Sending regular transaction when address is in allow list" {
    local private_key="12d7de8621a77640c9241b2595ba78ce443d05e94090365ab3bb5e19df82c625"
    local sender=$(cast wallet address "$private_key")
    local receiver="0x85dA99c8a7C2C95964c8EfD687E95E632Fc533D6"
    local value="10ether"

    run kurtosis service exec $enclave $sequencer "acl mode --datadir /home/erigon/data/dynamic-kurtosis-sequencer/txpool/acls --mode allowlist"
    run kurtosis service exec $enclave $sequencer "acl add --datadir /home/erigon/data/dynamic-kurtosis-sequencer/txpool/acls --address $sender --type allowlist --policy sendTx"
    run sendTx $private_key $receiver $value
    
    assert_success
    assert_output --regexp "Transaction successful \(transaction hash: 0x[a-fA-F0-9]{64}\)"
}

@test "Test Allow List - Sending contracts deploy transaction when address is in allow list" {
    local private_key="12d7de8621a77640c9241b2595ba78ce443d05e94090365ab3bb5e19df82c625"
    local sender=$(cast wallet address "$private_key")
    local contract_artifact="./contracts/erc20mock/ERC20Mock.json"

    run kurtosis service exec $enclave $sequencer "acl mode --datadir /home/erigon/data/dynamic-kurtosis-sequencer/txpool/acls --mode allowlist"
    run kurtosis service exec $enclave $sequencer "acl add --datadir /home/erigon/data/dynamic-kurtosis-sequencer/txpool/acls --address $sender --type allowlist --policy deploy"
    run deployContract $private_key $contract_artifact

    assert_success
}
