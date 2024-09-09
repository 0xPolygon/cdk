setup() {
    bats_load_library 'bats-support'
    bats_load_library 'bats-assert'

    # get the containing directory of this file
    # use $BATS_TEST_FILENAME instead of ${BASH_SOURCE[0]} or $0,
    # as those will point to the bats executable's location or the preprocessed file respectively
    DIR="$( cd "$( dirname "$BATS_TEST_FILENAME" )" >/dev/null 2>&1 && pwd )"
    # make executables in src/ visible to PATH
    PATH="$DIR/../src:$PATH"

    $DIR/scripts/kurtosis_prepare_params_yml.sh ../kurtosis-cdk cdk-validium

    # Check if the genesis file is already downloaded
    if [ ! -f "./tmp/cdk/genesis/genesis.json" ]; then
        kurtosis files download cdk-v1 genesis ./tmp/cdk/genesis
    fi
    # Download the genesis file
    readonly bridge_default_address=$(jq -r ".genesis[] | select(.contractName == \"PolygonZkEVMBridge proxy\") | .address" ./tmp/cdk/genesis/genesis.json)

    readonly skey=${RAW_PRIVATE_KEY:-"12d7de8621a77640c9241b2595ba78ce443d05e94090365ab3bb5e19df82c625"}
    readonly destination_net=${DESTINATION_NET:-"0"}
    readonly destination_addr=${DESTINATION_ADDRESS:-"0x0bb7AA0b4FdC2D2862c088424260e99ed6299148"}
    readonly ether_value=${ETHER_VALUE:-"0.0200000054"}
    readonly token_addr=${TOKEN_ADDRESS:-"0x0000000000000000000000000000000000000000"}
    readonly is_forced=${IS_FORCED:-"true"}
    readonly bridge_addr=${BRIDGE_ADDRESS:-$bridge_default_address}
    readonly meta_bytes=${META_BYTES:-"0x"}

    readonly rpc_url=${ETH_RPC_URL:-"$(kurtosis port print cdk-v1 cdk-erigon-node-001 http-rpc)"}
    readonly bridge_api_url=${BRIDGE_API_URL:-"$(kurtosis port print cdk-v1 zkevm-bridge-service-001 rpc)"}

    readonly dry_run=${DRY_RUN:-"false"}
    readonly claim_sig="claimAsset(bytes32[32],bytes32[32],uint256,bytes32,bytes32,uint32,address,uint32,address,uint256,bytes)"
    readonly bridge_sig='bridgeAsset(uint32,address,uint256,address,bool,bytes)'

    readonly amount=$(cast to-wei $ether_value ether)
    readonly current_addr="$(cast wallet address --private-key $skey)"
    readonly rpc_network_id=$(cast call --rpc-url $rpc_url $bridge_addr 'networkID()(uint32)')
}

@test "Run deposit" {
    load 'helpers/lxly-bridge-test'
    echo "Running LxLy deposit" >&3
    run deposit
    assert_success
    assert_output --partial 'transactionHash'
}

@test "Run claim" {
    load 'helpers/lxly-bridge-test'
    echo "Running LxLy claim"
    run claim
    assert_success
}
