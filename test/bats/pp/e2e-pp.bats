setup() {
    load '../../helpers/common-setup'
    _common_setup

    if [ -z "$BRIDGE_ADDRESS" ]; then
        local combined_json_file="/opt/zkevm/combined.json"
        echo "BRIDGE_ADDRESS env variable is not provided, resolving the bridge address from the Kurtosis CDK '$combined_json_file'" >&3

        # Fetching the combined JSON output and filtering to get polygonZkEVMBridgeAddress
        combined_json_output=$($contracts_service_wrapper "cat $combined_json_file" | tail -n +2)
        bridge_default_address=$(echo "$combined_json_output" | jq -r .polygonZkEVMBridgeAddress)
        BRIDGE_ADDRESS=$bridge_default_address
    fi
    echo "Bridge address=$BRIDGE_ADDRESS" >&3
}

@test "Verify certificate settlement" {
    echo "Waiting 10 minutes to get some settle certificate...." >&3

    readonly bridge_addr=$BRIDGE_ADDRESS
    readonly l2_rpc_network_id=$(cast call --rpc-url $l2_rpc_url $bridge_addr 'networkID() (uint32)')

    run $PROJECT_ROOT/../scripts/agglayer_certificates_monitor.sh 1 600 $l2_rpc_network_id
    assert_success
}
