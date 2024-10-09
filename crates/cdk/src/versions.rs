/// Command to print corresponding versions to the stdout.
pub(crate) fn versions() {
    // Multi-line string to print the versions.
    println!(
        r#"
# FORK 12 IMAGES
# yq .args .github/tests/forks/fork12.yml
zkevm_contracts_image: leovct/zkevm-contracts:v8.0.0-rc.4-fork.12
zkevm_prover_image: hermeznetwork/zkevm-prover:v8.0.0-RC12-fork.12
cdk_erigon_node_image: hermeznetwork/cdk-erigon:0948e33

# FORK 11 IMAGES
# yq .args .github/tests/forks/fork11.yml
# zkevm_contracts_image: leovct/zkevm-contracts:v7.0.0-rc.2-fork.11
# zkevm_prover_image: hermeznetwork/zkevm-prover:v7.0.2-fork.11
# cdk_erigon_node_image: hermeznetwork/cdk-erigon:acceptance-2.0.0-beta26-0f01107

zkevm_node_image: hermeznetwork/zkevm-node:v0.7.3-RC1
cdk_validium_node_image: 0xpolygon/cdk-validium-node:0.7.0-cdk

cdk_node_image: ghcr.io/0xpolygon/cdk:0.3.0-beta3
zkevm_da_image: 0xpolygon/cdk-data-availability:0.0.10
# agglayer_image: nulyjkdhthz/agglayer-rs:pr-318
agglayer_image: ghcr.io/agglayer/agglayer-rs:pr-96
zkevm_bridge_service_image: hermeznetwork/zkevm-bridge-service:v0.6.0-RC1
zkevm_bridge_ui_image: leovct/zkevm-bridge-ui:multi-network
zkevm_bridge_proxy_image: haproxy:3.0-bookworm
zkevm_sequence_sender_image: hermeznetwork/zkevm-sequence-sender:v0.2.0-RC12
zkevm_pool_manager_image: hermeznetwork/zkevm-pool-manager:v0.1.1
"#
    )
}
