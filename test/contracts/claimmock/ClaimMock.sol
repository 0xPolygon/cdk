// SPDX-License-Identifier: AGPL-3.0

pragma solidity 0.8.18;

contract ClaimMock {
    event ClaimEvent(
        uint256 globalIndex,
        uint32 originNetwork,
        address originAddress,
        address destinationAddress,
        uint256 amount
    );

    function claimAsset(
        bytes32[32] calldata smtProofLocalExitRoot,
        bytes32[32] calldata smtProofRollupExitRoot,
        uint256 globalIndex,
        bytes32 mainnetExitRoot,
        bytes32 rollupExitRoot,
        uint32 originNetwork,
        address originTokenAddress,
        uint32 destinationNetwork,
        address destinationAddress,
        uint256 amount,
        bytes calldata metadata
    ) external {
        emit ClaimEvent(
            globalIndex,
            originNetwork,
            originTokenAddress,
            destinationAddress,
            amount
        );
    }

    function claimMessage(
        bytes32[32] calldata smtProofLocalExitRoot,
        bytes32[32] calldata smtProofRollupExitRoot,
        uint256 globalIndex,
        bytes32 mainnetExitRoot,
        bytes32 rollupExitRoot,
        uint32 originNetwork,
        address originAddress,
        uint32 destinationNetwork,
        address destinationAddress,
        uint256 amount,
        bytes calldata metadata
    ) external {
        emit ClaimEvent(
            globalIndex,
            originNetwork,
            originAddress,
            destinationAddress,
            amount
        );
    }
}