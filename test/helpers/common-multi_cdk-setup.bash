#!/usr/bin/env bash

_common_multi_setup() {
    load '../../helpers/common-setup'
    _common_setup
     # generated with cast wallet new
    readonly target_address=0xbecE3a31343c6019CDE0D5a4dF2AF8Df17ebcB0f
    readonly target_private_key=0x51caa196504216b1730280feb63ddd8c5ae194d13e57e58d559f1f1dc3eda7c9

    kurtosis service exec $enclave contracts-001 "cat /opt/zkevm/combined-001.json"  | tail -n +2 | jq '.' > combined-001.json
    kurtosis service exec $enclave contracts-002 "cat /opt/zkevm/combined-002.json"  | tail -n +2 | jq '.' > combined-002.json
    kurtosis service exec $enclave contracts-002 "cat /opt/zkevm-contracts/deployment/v2/create_rollup_parameters.json" | tail -n +2 | jq -r '.gasTokenAddress' > gas-token-address.json

    readonly private_key="0x12d7de8621a77640c9241b2595ba78ce443d05e94090365ab3bb5e19df82c625"
    readonly eth_address=$(cast wallet address --private-key $private_key)
    readonly l1_rpc_url=http://$(kurtosis port print $enclave el-1-geth-lighthouse rpc)
    readonly l2_pp1_url=$(kurtosis port print $enclave cdk-erigon-rpc-001 rpc)
    readonly l2_pp2_url=$(kurtosis port print $enclave cdk-erigon-rpc-002 rpc)
    readonly bridge_address=$(cat combined-001.json | jq -r .polygonZkEVMBridgeAddress)
    readonly pol_address=$(cat combined-001.json | jq -r .polTokenAddress)
    readonly gas_token_address=$(<gas-token-address.json)
    readonly l2_pp1b_url=$(kurtosis port print $enclave zkevm-bridge-service-001 rpc)
    readonly l2_pp2b_url=$(kurtosis port print $enclave zkevm-bridge-service-002 rpc)
    
    #readonly l1_rpc_network_id=$(cast call --rpc-url $l1_rpc_url $bridge_addr 'networkID() (uint32)')
    #readonly l2_pp1b_network_id=$(cast call --rpc-url $l2_pp1_url $bridge_addr 'networkID() (uint32)')
    #readonly l2_pp2b_network_id=$(cast call --rpc-url $l2_pp2_url $bridge_addr 'networkID() (uint32)')
    readonly l1_rpc_network_id=0
    readonly l2_pp1b_network_id=1
    readonly l2_pp2b_network_id=2
    echo "=== Bridge address=$bridge_address ===" >&3
    echo "=== POL address=$pol_address ===" >&3
    echo "=== Gas token address=$gas_token_address ===" >&3
    echo "=== L1 network id=$l1_rpc_network_id ===" >&3
    echo "=== L2 PP1 network id=$l2_pp1b_network_id ===" >&3
    echo "=== L2 PP2 network id=$l2_pp2b_network_id ===" >&3
    echo "=== L1 RPC URL=$l1_rpc_url ===" >&3
    echo "=== L2 PP1 URL=$l2_pp1_url ===" >&3
    echo "=== L2 PP2 URL=$l2_pp2_url ===" >&3
    echo "=== L2 PP1B URL=$l2_pp1b_url ===" >&3
    echo "=== L2 PP2B URL=$l2_pp2b_url ===" >&3
    
}

add_cdk_network2_to_agglayer(){
    echo "=== Checking if  network 2 is in agglayer ===" >&3
    local _prev=$(kurtosis service exec $enclave agglayer "grep \"2 = \" /etc/zkevm/agglayer-config.toml || true" | tail -n +2)
    if [ ! -z "$_prev" ]; then
        echo "Network 2 already added to agglayer" >&3
        return
    fi
    echo "=== Adding network 2 to agglayer === ($_prev)" >&3
    exit 1
    kurtosis service exec $enclave agglayer "sed -i 's/\[proof\-signers\]/2 = \"http:\/\/cdk-erigon-rpc-002:8123\"\n\[proof-signers\]/i' /etc/zkevm/agglayer-config.toml"
    kurtosis service stop $enclave agglayer
    kurtosis service start $enclave agglayer
}

fund_claim_tx_manager(){
    echo "=== Funding bridge auto-claim  ===" >&3
    cast send --legacy --value 100ether --rpc-url $l2_pp1_url --private-key $private_key 0x5f5dB0D4D58310F53713eF4Df80ba6717868A9f8
    cast send --legacy --value 100ether --rpc-url $l2_pp2_url --private-key $private_key 0x93F63c24735f45Cd0266E87353071B64dd86bc05
}


mint_pol_token(){
     echo "=== Mining POL  ===" >&3
    cast send \
     --rpc-url $l1_rpc_url \
     --private-key $private_key \
     $pol_address \
     'mint(address,uint256)' \
     $eth_address 10000000000000000000000
    # Allow bridge to spend it
    cast send \
     --rpc-url $l1_rpc_url \
     --private-key $private_key \
     $pol_address \
     'approve(address,uint256)' \
     $bridge_address 10000000000000000000000
}
