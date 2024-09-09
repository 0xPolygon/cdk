setup() {
    load '/usr/local/lib/bats/bats-support/load'
    load '/usr/local/lib/bats/bats-assert/load'

    # get the containing directory of this file
    # use $BATS_TEST_FILENAME instead of ${BASH_SOURCE[0]} or $0,
    # as those will point to the bats executable's location or the preprocessed file respectively
    DIR="$(cd "$(dirname "$BATS_TEST_FILENAME")" >/dev/null 2>&1 && pwd)"
    # make executables in src/ visible to PATH
    PATH="$DIR/../src:$PATH"

    readonly enclave=${ENCLAVE:-cdk-v1}
    readonly node=${KURTOSIS_NODE:-cdk-erigon-node-001}
    readonly rpc_url=${RPC_URL:-$(kurtosis port print "$enclave" "$node" http-rpc)}
}

@test "Send EOA transaction" {
    load 'helpers/common'

    local receiver="0x85dA99c8a7C2C95964c8EfD687E95E632Fc533D6"
    local value="10ether"
    local private_key="0x12d7de8621a77640c9241b2595ba78ce443d05e94090365ab3bb5e19df82c625"

    run sendTx $private_key $receiver $value

    assert_success
    assert_output --partial 'Transaction successful'
}
