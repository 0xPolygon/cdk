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
        bytes calldata metadata,
        bool reverted
    ) external {
        address addr = address(claimMock);
        uint256 value = 0;
        if(reverted) {
            value = 1;
        }
        bytes4 argSig = bytes4(keccak256("claimAsset(bytes32[32],bytes32[32],uint256,bytes32,bytes32,uint32,address,uint32,address,uint256,bytes)"));
        bytes32 value1 = smtProofLocalExitRoot[5];
        bytes32 value2 = smtProofRollupExitRoot[4];
        assembly {
            let x := mload(0x40)   //Find empty storage location using "free memory pointer"
            mstore(x,argSig)
            mstore(add(x,164),value1)
            mstore(add(x,1156),value2)
            mstore(add(x,2052),globalIndex)
            mstore(add(x,2084),mainnetExitRoot)
            mstore(add(x,2116),rollupExitRoot)
            mstore(add(x,2148),originNetwork)
            mstore(add(x,2180),originTokenAddress)
            mstore(add(x,2212),destinationNetwork)
            mstore(add(x,2244),destinationAddress)
            mstore(add(x,2276),amount)
            let success := call(gas(), addr, value, x, 0xaac, 0x20, 0)
        }
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
        bytes calldata metadata,
        bool reverted
    ) external {
        address addr = address(claimMock);
        uint256 value = 0;
        if(reverted) {
            value = 1;
        }
        bytes4 argSig = bytes4(keccak256("claimMessage(bytes32[32],bytes32[32],uint256,bytes32,bytes32,uint32,address,uint32,address,uint256,bytes)"));
        bytes32 value1 = smtProofLocalExitRoot[5];
        bytes32 value2 = smtProofRollupExitRoot[4];
        assembly {
            let x := mload(0x40)   //Find empty storage location using "free memory pointer"
            mstore(x,argSig)
            mstore(add(x,164),value1)
            mstore(add(x,1156),value2)
            mstore(add(x,2052),globalIndex)
            mstore(add(x,2084),mainnetExitRoot)
            mstore(add(x,2116),rollupExitRoot)
            mstore(add(x,2148),originNetwork)
            mstore(add(x,2180),originAddress)
            mstore(add(x,2212),destinationNetwork)
            mstore(add(x,2244),destinationAddress)
            mstore(add(x,2276),amount)
            let success := call(gas(), addr, value, x, 0xaac, 0x20, 0)
        }
    }

    function claimBytes(
        bytes memory claim,
        bool reverted
    ) external {
        address addr = address(claimMock);
        uint256 value = 0;
        if(reverted) {
            value = 1;
        }
        assembly {
            let success := call(gas(), addr, value, add(claim, 32), 0xaac, 0x20, 0)
        }
    }

    function claim2Bytes(
        bytes memory claim1,
        bytes memory claim2,
        bool[2] memory reverted
    ) external {
        address addr = address(claimMock);
        uint256 value1 = 0;
        if(reverted[0]) {
            value1 = 1;
        }
        uint256 value2 = 0;
        if(reverted[1]) {
            value2 = 1;
        }
        assembly {
            let success1 := call(gas(), addr, value1, add(claim1, 32), 0xaac, 0x20, 0)
        }
        assembly {
            let success2 :=  call(gas(), addr, value2, add(claim2, 32), 0xaac, 0x20, 0)
        }
    }

}