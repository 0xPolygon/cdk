// SPDX-License-Identifier: AGPL-3.0

pragma solidity 0.8.18;

interface IBasePolygonZkEVMGlobalExitRoot {
    /**
     * @dev Thrown when the caller is not the allowed contracts
     */
    error OnlyAllowedContracts();

    function updateExitRoot(bytes32 newRollupExitRoot) external;

    function globalExitRootMap(
        bytes32 globalExitRootNum
    ) external returns (uint256);
}

interface IPolygonZkEVMGlobalExitRootV2 is IBasePolygonZkEVMGlobalExitRoot {
    function getLastGlobalExitRoot() external view returns (bytes32);

    function getRoot() external view returns (bytes32);
}

contract VerifyBatchesMock {
    uint256 internal constant _EXIT_TREE_DEPTH = 32;
    IPolygonZkEVMGlobalExitRootV2 public immutable globalExitRootManager;
    uint32 public rollupCount;
    mapping(uint32 rollupID => bytes32) public rollupIDToLastExitRoot;

    event VerifyBatches(
        uint32 indexed rollupID,
        uint64 numBatch,
        bytes32 stateRoot,
        bytes32 exitRoot,
        address indexed aggregator
    );

    event VerifyBatchesTrustedAggregator(
        uint32 indexed rollupID,
        uint64 numBatch,
        bytes32 stateRoot,
        bytes32 exitRoot,
        address indexed aggregator
    );

    constructor(
        IPolygonZkEVMGlobalExitRootV2 _globalExitRootManager
    ) {
        globalExitRootManager = _globalExitRootManager;
    }

    function verifyBatches(
        uint32 rollupID,
        uint64 finalNewBatch,
        bytes32 newLocalExitRoot,
        bytes32 newStateRoot,
        bool updateGER
    ) external {
        if (rollupID > rollupCount) {
            rollupCount = rollupID;
        }
        rollupIDToLastExitRoot[rollupID] = newLocalExitRoot;
        if (updateGER) {
            globalExitRootManager.updateExitRoot(getRollupExitRoot());
        }

        emit VerifyBatches(
            rollupID,
            finalNewBatch,
            newStateRoot,
            newLocalExitRoot,
            msg.sender
        );
    }

    function verifyBatchesTrustedAggregator(
        uint32 rollupID,
        uint64 finalNewBatch,
        bytes32 newLocalExitRoot,
        bytes32 newStateRoot,
        bool updateGER
    ) external {
        if (rollupID > rollupCount) {
            rollupCount = rollupID;
        }
        rollupIDToLastExitRoot[rollupID] = newLocalExitRoot;
        if (updateGER) {
            globalExitRootManager.updateExitRoot(getRollupExitRoot());
        }

        emit VerifyBatchesTrustedAggregator(
            rollupID,
            finalNewBatch,
            newStateRoot,
            newLocalExitRoot,
            msg.sender
        );
    }


    function getRollupExitRoot() public view returns (bytes32) {
        uint256 currentNodes = rollupCount;

        // If there are no nodes return 0
        if (currentNodes == 0) {
            return bytes32(0);
        }

        // This array will contain the nodes of the current iteration
        bytes32[] memory tmpTree = new bytes32[](currentNodes);

        // In the first iteration the nodes will be the leafs which are the local exit roots of each network
        for (uint256 i = 0; i < currentNodes; i++) {
            // The first rollup ID starts on 1
            tmpTree[i] = rollupIDToLastExitRoot[uint32(i + 1)];
        }

        // This variable will keep track of the zero hashes
        bytes32 currentZeroHashHeight = 0;

        // This variable will keep track of the reamining levels to compute
        uint256 remainingLevels = _EXIT_TREE_DEPTH;

        // Calculate the root of the sub-tree that contains all the localExitRoots
        while (currentNodes != 1) {
            uint256 nextIterationNodes = currentNodes / 2 + (currentNodes % 2);
            bytes32[] memory nextTmpTree = new bytes32[](nextIterationNodes);
            for (uint256 i = 0; i < nextIterationNodes; i++) {
                // if we are on the last iteration of the current level and the nodes are odd
                if (i == nextIterationNodes - 1 && (currentNodes % 2) == 1) {
                    nextTmpTree[i] = keccak256(
                        abi.encodePacked(tmpTree[i * 2], currentZeroHashHeight)
                    );
                } else {
                    nextTmpTree[i] = keccak256(
                        abi.encodePacked(tmpTree[i * 2], tmpTree[(i * 2) + 1])
                    );
                }
            }

            // Update tree variables
            tmpTree = nextTmpTree;
            currentNodes = nextIterationNodes;
            currentZeroHashHeight = keccak256(
                abi.encodePacked(currentZeroHashHeight, currentZeroHashHeight)
            );
            remainingLevels--;
        }

        bytes32 currentRoot = tmpTree[0];

        // Calculate remaining levels, since it's a sequencial merkle tree, the rest of the tree are zeroes
        for (uint256 i = 0; i < remainingLevels; i++) {
            currentRoot = keccak256(
                abi.encodePacked(currentRoot, currentZeroHashHeight)
            );
            currentZeroHashHeight = keccak256(
                abi.encodePacked(currentZeroHashHeight, currentZeroHashHeight)
            );
        }
        return currentRoot;
    }
}