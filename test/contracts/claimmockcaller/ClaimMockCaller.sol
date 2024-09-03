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
        uint256 newAmount = amount;
        if(reverted) {
            newAmount = 0;
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
            mstore(add(x,2276),newAmount)
            let success := call(gas(), addr, 0, x, 0xaac, 0x20, 0)
        }
    }

    function claimAsset2(
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
        bool[2] calldata reverted
    ) external {
        address addr = address(claimMock);
        uint256 newAmount1 = amount;
        if(reverted[0]) {
            newAmount1 = 0;
        }
        uint256 globalIndex2 = globalIndex + 1;
        uint256 newAmount2 = amount+1;
        if(reverted[1]) {
            newAmount2 = 0;
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
            mstore(add(x,2276),newAmount1)
            let success := call(gas(), addr, 0, x, 0xaac, 0x20, 0)
        }
        assembly {
            let x := mload(0x40)   //Find empty storage location using "free memory pointer"
            mstore(x,argSig)
            mstore(add(x,164),value1)
            mstore(add(x,1156),value2)
            mstore(add(x,2052),globalIndex2)
            mstore(add(x,2084),mainnetExitRoot)
            mstore(add(x,2116),rollupExitRoot)
            mstore(add(x,2148),originNetwork)
            mstore(add(x,2180),originTokenAddress)
            mstore(add(x,2212),destinationNetwork)
            mstore(add(x,2244),destinationAddress)
            mstore(add(x,2276),newAmount2)
            let success := call(gas(), addr, 0, x, 0xaac, 0x20, 0)
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
        uint256 newAmount = amount;
        if(reverted) {
            newAmount = 0;
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
            mstore(add(x,2276),newAmount)
            let success := call(gas(), addr, 0, x, 0xaac, 0x20, 0)
        }
    }

    function claimMessage2(
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
        bool[2] calldata reverted
    ) external {
        address addr = address(claimMock);
        uint256 newAmount1 = amount;
        if(reverted[0]) {
            newAmount1 = 0;
        }
        uint256 globalIndex2 = globalIndex + 1;
        uint256 newAmount2 = amount+1;
        if(reverted[1]) {
            newAmount2 = 0;
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
            mstore(add(x,2276),newAmount1)
            let success := call(gas(), addr, 0, x, 0xaac, 0x20, 0)
        }
        assembly {
            let x := mload(0x40)   //Find empty storage location using "free memory pointer"
            mstore(x,argSig)
            mstore(add(x,164),value1)
            mstore(add(x,1156),value2)
            mstore(add(x,2052),globalIndex2)
            mstore(add(x,2084),mainnetExitRoot)
            mstore(add(x,2116),rollupExitRoot)
            mstore(add(x,2148),originNetwork)
            mstore(add(x,2180),originAddress)
            mstore(add(x,2212),destinationNetwork)
            mstore(add(x,2244),destinationAddress)
            mstore(add(x,2276),newAmount2)
            let success := call(gas(), addr, 0, x, 0xaac, 0x20, 0)
        }
    }
}