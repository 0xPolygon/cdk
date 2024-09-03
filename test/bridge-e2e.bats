setup() {
    bats_load_library 'bats-support' 
    bats_load_library 'bats-assert'

    # get the containing directory of this file
    # use $BATS_TEST_FILENAME instead of ${BASH_SOURCE[0]} or $0,
    # as those will point to the bats executable's location or the preprocessed file respectively
    DIR="$( cd "$( dirname "$BATS_TEST_FILENAME" )" >/dev/null 2>&1 && pwd )"
    # make executables in src/ visible to PATH
    PATH="$DIR/../src:$PATH"

    readonly skey=${RAW_PRIVATE_KEY:-"bc6a95c870cce28fe9686fdd70a3a595769e019bb396cd725bc16ec553f07c83"}
    readonly destination_net=${DESTINATION_NET:-"1"}
    readonly destination_addr=${DESTINATION_ADDRESS:-"0x0bb7AA0b4FdC2D2862c088424260e99ed6299148"}
    readonly ether_value=${ETHER_VALUE:-"0.0200000054"}
    readonly token_addr=${TOKEN_ADDRESS:-"0x0000000000000000000000000000000000000000"}
    readonly is_forced=${IS_FORCED:-"true"}
    readonly bridge_addr=${BRIDGE_ADDRESS:-"0x528e26b25a34a4A5d0dbDa1d57D318153d2ED582"}
    readonly meta_bytes=${META_BYTES:-"0x"}
    readonly subcommand=${1:-"deposit"}

    readonly rpc_url=${ETH_RPC_URL:-"https://rpc.cardona.zkevm-rpc.com"}
    readonly bridge_api_url=${BRIDGE_API_URL:-"https://bridge-api-cdk-validium-cardona-03-zkevm.polygondev.tools"}

    readonly dry_run=${DRY_RUN:-"false"}
    readonly claim_sig="claimAsset(bytes32[32],bytes32[32],uint256,bytes32,bytes32,uint32,address,uint32,address,uint256,bytes)"
    readonly bridge_sig='bridgeAsset(uint32,address,uint256,address,bool,bytes)'

    readonly amount=$(cast to-wei $ether_value ether)
    readonly current_addr="$(cast wallet address --private-key $skey)"
    readonly rpc_network_id=$(cast call --rpc-url $rpc_url $bridge_addr 'networkID()(uint32)')


    2>&1 echo "Running LxLy " $subcommand

    2>&1 echo "Checking the current network id: "
    2>&1 echo $rpc_network_id

    2>&1 echo "The current private key has address: "
    2>&1 echo $current_addr

    if [[ $token_addr == "0x0000000000000000000000000000000000000000" ]]; then
        2>&1 echo "Checking the current ETH balance: "
        2>&1 cast balance -e --rpc-url $rpc_url $current_addr
    else
        2>&1 echo "Checking the current token balance for token at $token_addr: "
        2>&1 cast call --rpc-url $rpc_url $token_addr 'balanceOf(address)(uint256)' $current_addr
    fi
}

@test "Run deposit" {
    load 'helpers/lxly-bridge-test'
    run deposit
    assert_output --partial 'foo'
}

@test "Run claim" {
    load 'helpers/lxly-bridge-test'
    run claim
    assert_output --partial 'execution reverted'
}
