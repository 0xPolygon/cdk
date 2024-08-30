// SPDX-License-Identifier: AGPL-3.0

pragma solidity 0.8.18;


interface IClaimMock {
    function claimAsset(bytes32[32] calldata smtProofLocalExitRoot,bytes32[32] calldata smtProofRollupExitRoot,uint256 globalIndex,bytes32 mainnetExitRoot,bytes32 rollupExitRoot,uint32 originNetwork,address originTokenAddress,uint32 destinationNetwork,address destinationAddress,uint256 amount,bytes calldata metadata) external;
    function claimMessage(bytes32[32] calldata smtProofLocalExitRoot,bytes32[32] calldata smtProofRollupExitRoot,uint256 globalIndex,bytes32 mainnetExitRoot,bytes32 rollupExitRoot,uint32 originNetwork,address originAddress,uint32 destinationNetwork,address destinationAddress,uint256 amount,bytes calldata metadata) external;
}

contract ClaimMockCaller {
    IClaimMock public immutable claimMock;
    uint8 constant _DEPOSIT_CONTRACT_TREE_DEPTH = 32;

    constructor(
        IClaimMock _claimMock
    ) {
        claimMock = _claimMock;
    }

    function claimAsset(
        bytes32[_DEPOSIT_CONTRACT_TREE_DEPTH] calldata smtProofLocalExitRoot,
        bytes32[_DEPOSIT_CONTRACT_TREE_DEPTH] calldata smtProofRollupExitRoot,
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
        claimMock.claimAsset(
            smtProofLocalExitRoot,
            smtProofRollupExitRoot,
            globalIndex,
            mainnetExitRoot,
            rollupExitRoot,
            originNetwork,
            originTokenAddress,
            destinationNetwork,
            destinationAddress,
            amount,
            metadata
        );
    }

    function claimMessage(
        bytes32[_DEPOSIT_CONTRACT_TREE_DEPTH] calldata smtProofLocalExitRoot,
        bytes32[_DEPOSIT_CONTRACT_TREE_DEPTH] calldata smtProofRollupExitRoot,
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
        claimMock.claimMessage(
            smtProofLocalExitRoot,
            smtProofRollupExitRoot,
            globalIndex,
            mainnetExitRoot,
            rollupExitRoot,
            originNetwork,
            originAddress,
            destinationNetwork,
            destinationAddress,
            amount,
            metadata
        );  
    }
}