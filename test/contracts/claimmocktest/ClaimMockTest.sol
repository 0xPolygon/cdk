// SPDX-License-Identifier: AGPL-3.0

pragma solidity 0.8.18;

interface IClaimMock {
    event ClaimEvent(uint256 globalIndex, uint32 originNetwork, address originAddress, address destinationAddress, uint256 amount);
    function claimAsset(bytes32[32] calldata smtProofLocalExitRoot,bytes32[32] calldata smtProofRollupExitRoot,uint256 globalIndex,bytes32 mainnetExitRoot,bytes32 rollupExitRoot,uint32 originNetwork,address originTokenAddress,uint32 destinationNetwork,address destinationAddress,uint256 amount,bytes calldata metadata) external;
    function claimMessage(bytes32[32] calldata smtProofLocalExitRoot,bytes32[32] calldata smtProofRollupExitRoot,uint256 globalIndex,bytes32 mainnetExitRoot,bytes32 rollupExitRoot,uint32 originNetwork,address originAddress,uint32 destinationNetwork,address destinationAddress,uint256 amount,bytes calldata metadata) external;
}

interface IClaimMockCaller {
    function claimAsset(bytes32[32] calldata smtProofLocalExitRoot, bytes32[32] calldata smtProofRollupExitRoot, uint256 globalIndex, bytes32 mainnetExitRoot, bytes32 rollupExitRoot, uint32 originNetwork, address originTokenAddress, uint32 destinationNetwork, address destinationAddress, uint256 amount, bytes calldata metadata, bool reverted) external;
    function claimMessage(bytes32[32] calldata smtProofLocalExitRoot, bytes32[32] calldata smtProofRollupExitRoot, uint256 globalIndex, bytes32 mainnetExitRoot, bytes32 rollupExitRoot, uint32 originNetwork, address originAddress, uint32 destinationNetwork, address destinationAddress, uint256 amount, bytes calldata metadata, bool reverted) external;
    function claimBytes(bytes memory claim, bool reverted) external;
    function claim2Bytes(bytes memory claim1, bytes memory claim2, bool[2] memory reverted) external;
}

contract ClaimMockTest {
    IClaimMockCaller public immutable claimMockCaller;
    IClaimMock public immutable claimMock;

    uint8 constant _DEPOSIT_CONTRACT_TREE_DEPTH = 32;

    constructor(
        IClaimMock _claimMock,
        IClaimMockCaller _claimMockCaller
    ) {
        claimMock = _claimMock;
        claimMockCaller = _claimMockCaller;
    }

    function claimTestInternal(bytes memory claim, bool reverted) external {
        claimMockCaller.claimBytes(claim, reverted);
    }

    function claim2TestInternal(bytes memory claim1, bytes memory claim2, bool[2] memory reverted) external {
        claimMockCaller.claim2Bytes(claim1, claim2, reverted);
    }

    function claim3TestInternal(bytes memory claim1, bytes memory claim2, bytes memory claim3, bool[3] memory reverted) external {
        address addr = address(claimMock);
        uint256 value1 = 0;
        if(reverted[1]) {
            value1 = 1;
        }
        claimMockCaller.claimBytes(claim1, reverted[0]);
        assembly {
            let success1 := call(gas(), addr, value1, add(claim2, 32), 0xaac, 0x20, 0)
        }
        claimMockCaller.claimBytes(claim3, reverted[2]);
    }

}