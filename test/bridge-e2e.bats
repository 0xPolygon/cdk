setup() {
    load 'helpers/common-setup'
    _common_setup

    $PROJECT_ROOT/test/scripts/kurtosis_prepare_params_yml.sh ../kurtosis-cdk cdk-validium

    # Check if the genesis file is already downloaded
    if [ ! -f "./tmp/cdk/genesis/genesis.json" ]; then
        mkdir -p ./tmp/cdk
        kurtosis files download cdk-v1 genesis ./tmp/cdk/genesis
    fi
    # Download the genesis file
    readonly bridge_default_address=$(jq -r ".genesis[] | select(.contractName == \"PolygonZkEVMBridge proxy\") | .address" ./tmp/cdk/genesis/genesis.json)

    readonly skey=${RAW_PRIVATE_KEY:-"12d7de8621a77640c9241b2595ba78ce443d05e94090365ab3bb5e19df82c625"}
    readonly destination_net=${DESTINATION_NET:-"1"}
    readonly destination_addr=${DESTINATION_ADDRESS:-"0x0bb7AA0b4FdC2D2862c088424260e99ed6299148"}
    readonly ether_value=${ETHER_VALUE:-"0.0200000054"}
    readonly token_addr=${TOKEN_ADDRESS:-"0x0000000000000000000000000000000000000000"}
    readonly is_forced=${IS_FORCED:-"true"}
    readonly bridge_addr=${BRIDGE_ADDRESS:-$bridge_default_address}
    readonly meta_bytes=${META_BYTES:-"0x"}

    readonly l1_rpc_url=${ETH_RPC_URL:-"$(kurtosis port print cdk-v1 el-1-geth-lighthouse rpc)"}
    readonly l2_rpc_url=${ETH_RPC_URL:-"$(kurtosis port print cdk-v1 cdk-erigon-node-001 http-rpc)"}
    readonly bridge_api_url=${BRIDGE_API_URL:-"$(kurtosis port print cdk-v1 zkevm-bridge-service-001 rpc)"}

    readonly dry_run=${DRY_RUN:-"false"}
    readonly claim_sig="claimAsset(bytes32[32],bytes32[32],uint256,bytes32,bytes32,uint32,address,uint32,address,uint256,bytes)"
    readonly bridge_sig='bridgeAsset(uint32,address,uint256,address,bool,bytes)'

    readonly amount=$(cast to-wei $ether_value ether)
    readonly current_addr="$(cast wallet address --private-key $skey)"
    readonly l1_rpc_network_id=$(cast call --rpc-url $l1_rpc_url $bridge_addr 'networkID()(uint32)')
    readonly l2_rpc_network_id=$(cast call --rpc-url $l2_rpc_url $bridge_addr 'networkID()(uint32)')
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

    # The script timeout (in seconds).
    timeout="120"
    start_time=$(date +%s)
    end_time=$((start_time + timeout))

    while true; do
        current_time=$(date +%s)
        if ((current_time > end_time)); then
            echo "[$(date '+%Y-%m-%d %H:%M:%S')] ❌ Exiting... Timeout reached!"
            exit 1
        fi
        
        run claim
        if [ $status -eq 0 ]; then
            break
        fi
        sleep 10
    done
    
    assert_success
}
