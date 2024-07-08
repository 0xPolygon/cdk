package localbridgesync

import (
	"context"
	"fmt"
	"math/big"
	"slices"
	"testing"

	"github.com/ethereum/go-ethereum/common"
	"github.com/stretchr/testify/require"
)

func TestProceessor(t *testing.T) {
	path := t.TempDir()
	p, err := newProcessor(path)
	require.NoError(t, err)
	actions := []processAction{
		// processed: ~
		&getLastProcessedBlockAction{
			p:                          p,
			description:                "on an empty processor",
			ctx:                        context.Background(),
			expectedLastProcessedBlock: 0,
			expectedErr:                nil,
		},
		&reorgAction{
			p:                 p,
			description:       "on an empty processor: firstReorgedBlock = 0",
			firstReorgedBlock: 0,
			expectedErr:       nil,
		},
		&reorgAction{
			p:                 p,
			description:       "on an empty processor: firstReorgedBlock = 1",
			firstReorgedBlock: 1,
			expectedErr:       nil,
		},
		&getClaimsAndBridgesAction{
			p:               p,
			description:     "on an empty processor",
			ctx:             context.Background(),
			fromBlock:       0,
			toBlock:         2,
			expectedClaims:  nil,
			expectedBridges: nil,
			expectedErr:     ErrBlockNotProcessed,
		},
		&storeBridgeEventsAction{
			p:           p,
			description: "block1",
			blockNum:    block1.Num,
			block:       block1.Events,
			expectedErr: nil,
		},
		// processed: block1
		&getLastProcessedBlockAction{
			p:                          p,
			description:                "after block1",
			ctx:                        context.Background(),
			expectedLastProcessedBlock: 1,
			expectedErr:                nil,
		},
		&getClaimsAndBridgesAction{
			p:               p,
			description:     "after block1: range 0, 2",
			ctx:             context.Background(),
			fromBlock:       0,
			toBlock:         2,
			expectedClaims:  nil,
			expectedBridges: nil,
			expectedErr:     ErrBlockNotProcessed,
		},
		&getClaimsAndBridgesAction{
			p:               p,
			description:     "after block1: range 1, 1",
			ctx:             context.Background(),
			fromBlock:       1,
			toBlock:         1,
			expectedClaims:  block1.Events.Claims,
			expectedBridges: block1.Events.Bridges,
			expectedErr:     nil,
		},
		&reorgAction{
			p:                 p,
			description:       "after block1",
			firstReorgedBlock: 1,
			expectedErr:       nil,
		},
		// processed: ~
		&getClaimsAndBridgesAction{
			p:               p,
			description:     "after block1 reorged",
			ctx:             context.Background(),
			fromBlock:       0,
			toBlock:         2,
			expectedClaims:  nil,
			expectedBridges: nil,
			expectedErr:     ErrBlockNotProcessed,
		},
		&storeBridgeEventsAction{
			p:           p,
			description: "block1 (after it's reorged)",
			blockNum:    block1.Num,
			block:       block1.Events,
			expectedErr: nil,
		},
		// processed: block3
		&storeBridgeEventsAction{
			p:           p,
			description: "block3",
			blockNum:    block3.Num,
			block:       block3.Events,
			expectedErr: nil,
		},
		// processed: block1, block3
		&getLastProcessedBlockAction{
			p:                          p,
			description:                "after block3",
			ctx:                        context.Background(),
			expectedLastProcessedBlock: 3,
			expectedErr:                nil,
		},
		&getClaimsAndBridgesAction{
			p:               p,
			description:     "after block3: range 2, 2",
			ctx:             context.Background(),
			fromBlock:       2,
			toBlock:         2,
			expectedClaims:  []Claim{},
			expectedBridges: []Bridge{},
			expectedErr:     nil,
		},
		&getClaimsAndBridgesAction{
			p:               p,
			description:     "after block3: range 1, 3",
			ctx:             context.Background(),
			fromBlock:       1,
			toBlock:         3,
			expectedClaims:  append(block1.Events.Claims, block3.Events.Claims...),
			expectedBridges: append(block1.Events.Bridges, block3.Events.Bridges...),
			expectedErr:     nil,
		},
		&reorgAction{
			p:                 p,
			description:       "after block3, with value 3",
			firstReorgedBlock: 3,
			expectedErr:       nil,
		},
		// processed: block1
		&getLastProcessedBlockAction{
			p:                          p,
			description:                "after block3 reorged",
			ctx:                        context.Background(),
			expectedLastProcessedBlock: 2,
			expectedErr:                nil,
		},
		&reorgAction{
			p:                 p,
			description:       "after block3, with value 2",
			firstReorgedBlock: 2,
			expectedErr:       nil,
		},
		&getLastProcessedBlockAction{
			p:                          p,
			description:                "after block2 reorged",
			ctx:                        context.Background(),
			expectedLastProcessedBlock: 1,
			expectedErr:                nil,
		},
		&storeBridgeEventsAction{
			p:           p,
			description: "block3 after reorg",
			blockNum:    block3.Num,
			block:       block3.Events,
			expectedErr: nil,
		},
		// processed: block1, block3
		&storeBridgeEventsAction{
			p:           p,
			description: "block4",
			blockNum:    block4.Num,
			block:       block4.Events,
			expectedErr: nil,
		},
		// processed: block1, block3, block4
		&storeBridgeEventsAction{
			p:           p,
			description: "block5",
			blockNum:    block5.Num,
			block:       block5.Events,
			expectedErr: nil,
		},
		// processed: block1, block3, block4, block5
		&getLastProcessedBlockAction{
			p:                          p,
			description:                "after block5",
			ctx:                        context.Background(),
			expectedLastProcessedBlock: 5,
			expectedErr:                nil,
		},
		&getClaimsAndBridgesAction{
			p:               p,
			description:     "after block5: range 1, 3",
			ctx:             context.Background(),
			fromBlock:       1,
			toBlock:         3,
			expectedClaims:  append(block1.Events.Claims, block3.Events.Claims...),
			expectedBridges: append(block1.Events.Bridges, block3.Events.Bridges...),
			expectedErr:     nil,
		},
		&getClaimsAndBridgesAction{
			p:               p,
			description:     "after block5: range 4, 5",
			ctx:             context.Background(),
			fromBlock:       4,
			toBlock:         5,
			expectedClaims:  append(block4.Events.Claims, block5.Events.Claims...),
			expectedBridges: append(block4.Events.Bridges, block5.Events.Bridges...),
			expectedErr:     nil,
		},
		&getClaimsAndBridgesAction{
			p:           p,
			description: "after block5: range 0, 5",
			ctx:         context.Background(),
			fromBlock:   0,
			toBlock:     5,
			expectedClaims: slices.Concat(
				block1.Events.Claims,
				block3.Events.Claims,
				block4.Events.Claims,
				block5.Events.Claims,
			),
			expectedBridges: slices.Concat(
				block1.Events.Bridges,
				block3.Events.Bridges,
				block4.Events.Bridges,
				block5.Events.Bridges,
			),
			expectedErr: nil,
		},
	}

	for _, a := range actions {
		t.Run(fmt.Sprintf("%s: %s", a.method(), a.desc()), a.execute)
	}
}

// BOILERPLATE

// blocks

var (
	block1 = block{
		blockHeader: blockHeader{
			Num:  1,
			Hash: common.HexToHash("01"),
		},
		Events: bridgeEvents{
			Bridges: []Bridge{
				{
					LeafType:           1,
					OriginNetwork:      1,
					OriginAddress:      common.HexToAddress("01"),
					DestinationNetwork: 1,
					DestinationAddress: common.HexToAddress("01"),
					Amount:             big.NewInt(1),
					Metadata:           common.Hex2Bytes("01"),
					DepositCount:       1,
				},
			},
			Claims: []Claim{
				{
					GlobalIndex:        big.NewInt(1),
					OriginNetwork:      1,
					OriginAddress:      common.HexToAddress("01"),
					DestinationAddress: common.HexToAddress("01"),
					Amount:             big.NewInt(1),
				},
			},
		},
	}
	block3 = block{
		blockHeader: blockHeader{
			Num:  3,
			Hash: common.HexToHash("02"),
		},
		Events: bridgeEvents{
			Bridges: []Bridge{
				{
					LeafType:           2,
					OriginNetwork:      2,
					OriginAddress:      common.HexToAddress("02"),
					DestinationNetwork: 2,
					DestinationAddress: common.HexToAddress("02"),
					Amount:             big.NewInt(2),
					Metadata:           common.Hex2Bytes("02"),
					DepositCount:       2,
				},
				{
					LeafType:           3,
					OriginNetwork:      3,
					OriginAddress:      common.HexToAddress("03"),
					DestinationNetwork: 3,
					DestinationAddress: common.HexToAddress("03"),
					Amount:             nil,
					Metadata:           common.Hex2Bytes("03"),
					DepositCount:       3,
				},
			},
			Claims: []Claim{},
		},
	}
	block4 = block{
		blockHeader: blockHeader{
			Num:  4,
			Hash: common.HexToHash("03"),
		},
		Events: bridgeEvents{
			Bridges: []Bridge{},
			Claims:  []Claim{},
		},
	}
	block5 = block{
		blockHeader: blockHeader{
			Num:  5,
			Hash: common.HexToHash("04"),
		},
		Events: bridgeEvents{
			Bridges: []Bridge{},
			Claims: []Claim{
				{
					GlobalIndex:        big.NewInt(4),
					OriginNetwork:      4,
					OriginAddress:      common.HexToAddress("04"),
					DestinationAddress: common.HexToAddress("04"),
					Amount:             big.NewInt(4),
				},
				{
					GlobalIndex:        big.NewInt(5),
					OriginNetwork:      5,
					OriginAddress:      common.HexToAddress("05"),
					DestinationAddress: common.HexToAddress("05"),
					Amount:             big.NewInt(5),
				},
			},
		},
	}
)

// actions

type processAction interface {
	method() string
	desc() string
	execute(t *testing.T)
}

// GetClaimsAndBridges

type getClaimsAndBridgesAction struct {
	p               *processor
	description     string
	ctx             context.Context
	fromBlock       uint64
	toBlock         uint64
	expectedClaims  []Claim
	expectedBridges []Bridge
	expectedErr     error
}

func (a *getClaimsAndBridgesAction) method() string {
	return "GetClaimsAndBridges"
}

func (a *getClaimsAndBridgesAction) desc() string {
	return a.description
}

func (a *getClaimsAndBridgesAction) execute(t *testing.T) {
	actualClaims, actualBridges, actualErr := a.p.GetClaimsAndBridges(a.ctx, a.fromBlock, a.toBlock)
	require.Equal(t, a.expectedClaims, actualClaims)
	require.Equal(t, a.expectedBridges, actualBridges)
	require.Equal(t, a.expectedErr, actualErr)
}

// getLastProcessedBlock

type getLastProcessedBlockAction struct {
	p                          *processor
	description                string
	ctx                        context.Context
	expectedLastProcessedBlock uint64
	expectedErr                error
}

func (a *getLastProcessedBlockAction) method() string {
	return "getLastProcessedBlock"
}

func (a *getLastProcessedBlockAction) desc() string {
	return a.description
}

func (a *getLastProcessedBlockAction) execute(t *testing.T) {
	actualLastProcessedBlock, actualErr := a.p.getLastProcessedBlock(a.ctx)
	require.Equal(t, a.expectedLastProcessedBlock, actualLastProcessedBlock)
	require.Equal(t, a.expectedErr, actualErr)
}

// reorg

type reorgAction struct {
	p                 *processor
	description       string
	firstReorgedBlock uint64
	expectedErr       error
}

func (a *reorgAction) method() string {
	return "reorg"
}

func (a *reorgAction) desc() string {
	return a.description
}

func (a *reorgAction) execute(t *testing.T) {
	actualErr := a.p.reorg(a.firstReorgedBlock)
	require.Equal(t, a.expectedErr, actualErr)
}

// storeBridgeEvents

type storeBridgeEventsAction struct {
	p           *processor
	description string
	blockNum    uint64
	block       bridgeEvents
	expectedErr error
}

func (a *storeBridgeEventsAction) method() string {
	return "storeBridgeEvents"
}

func (a *storeBridgeEventsAction) desc() string {
	return a.description
}

func (a *storeBridgeEventsAction) execute(t *testing.T) {
	actualErr := a.p.storeBridgeEvents(a.blockNum, a.block)
	require.Equal(t, a.expectedErr, actualErr)
}
