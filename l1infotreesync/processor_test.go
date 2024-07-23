package l1infotreesync

import (
	"context"
	"fmt"
	"slices"
	"testing"

	"github.com/ethereum/go-ethereum/common"
	"github.com/stretchr/testify/require"
)

func TestProceessor(t *testing.T) {
	path := t.TempDir()
	ctx := context.Background()
	p, err := newProcessor(ctx, path, 32)
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
			p:              p,
			description:    "on an empty processor",
			ctx:            context.Background(),
			fromBlock:      0,
			toBlock:        2,
			expectedEvents: nil,
			expectedErr:    ErrBlockNotProcessed,
		},
		&storeL1InfoTreeUpdatesAction{
			p:           p,
			description: "block1",
			b:           block1,
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
			p:              p,
			description:    "after block1: range 0, 2",
			ctx:            context.Background(),
			fromBlock:      0,
			toBlock:        2,
			expectedEvents: nil,
			expectedErr:    ErrBlockNotProcessed,
		},
		&getClaimsAndBridgesAction{
			p:              p,
			description:    "after block1: range 1, 1",
			ctx:            context.Background(),
			fromBlock:      1,
			toBlock:        1,
			expectedEvents: block1.Events,
			expectedErr:    nil,
		},
		&reorgAction{
			p:                 p,
			description:       "after block1",
			firstReorgedBlock: 1,
			expectedErr:       nil,
		},
		// processed: ~
		&getClaimsAndBridgesAction{
			p:              p,
			description:    "after block1 reorged",
			ctx:            context.Background(),
			fromBlock:      0,
			toBlock:        2,
			expectedEvents: nil,
			expectedErr:    ErrBlockNotProcessed,
		},
		&storeL1InfoTreeUpdatesAction{
			p:           p,
			description: "block1 (after it's reorged)",
			b:           block1,
			expectedErr: nil,
		},
		// processed: block3
		&storeL1InfoTreeUpdatesAction{
			p:           p,
			description: "block3",
			b:           block3,
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
			p:              p,
			description:    "after block3: range 2, 2",
			ctx:            context.Background(),
			fromBlock:      2,
			toBlock:        2,
			expectedEvents: []L1InfoTreeUpdate{},
			expectedErr:    nil,
		},
		&getClaimsAndBridgesAction{
			p:              p,
			description:    "after block3: range 1, 3",
			ctx:            context.Background(),
			fromBlock:      1,
			toBlock:        3,
			expectedEvents: append(block1.Events, block3.Events...),
			expectedErr:    nil,
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
		&storeL1InfoTreeUpdatesAction{
			p:           p,
			description: "block3 after reorg",
			b:           block3,
			expectedErr: nil,
		},
		// processed: block1, block3
		&storeL1InfoTreeUpdatesAction{
			p:           p,
			description: "block4",
			b:           block4,
			expectedErr: nil,
		},
		// processed: block1, block3, block4
		&storeL1InfoTreeUpdatesAction{
			p:           p,
			description: "block5",
			b:           block5,
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
			p:              p,
			description:    "after block5: range 1, 3",
			ctx:            context.Background(),
			fromBlock:      1,
			toBlock:        3,
			expectedEvents: append(block1.Events, block3.Events...),
			expectedErr:    nil,
		},
		&getClaimsAndBridgesAction{
			p:              p,
			description:    "after block5: range 4, 5",
			ctx:            context.Background(),
			fromBlock:      4,
			toBlock:        5,
			expectedEvents: append(block4.Events, block5.Events...),
			expectedErr:    nil,
		},
		&getClaimsAndBridgesAction{
			p:           p,
			description: "after block5: range 0, 5",
			ctx:         context.Background(),
			fromBlock:   0,
			toBlock:     5,
			expectedEvents: slices.Concat(
				block1.Events,
				block3.Events,
				block4.Events,
				block5.Events,
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
		Events: []L1InfoTreeUpdate{
			{RollupExitRoot: common.HexToHash("01")},
			{MainnetExitRoot: common.HexToHash("01")},
		},
	}
	block3 = block{
		blockHeader: blockHeader{
			Num:  3,
			Hash: common.HexToHash("02"),
		},
		Events: []L1InfoTreeUpdate{
			{RollupExitRoot: common.HexToHash("02"), MainnetExitRoot: common.HexToHash("02")},
		},
	}
	block4 = block{
		blockHeader: blockHeader{
			Num:  4,
			Hash: common.HexToHash("03"),
		},
		Events: []L1InfoTreeUpdate{},
	}
	block5 = block{
		blockHeader: blockHeader{
			Num:  5,
			Hash: common.HexToHash("04"),
		},
		Events: []L1InfoTreeUpdate{
			{RollupExitRoot: common.HexToHash("04")},
			{MainnetExitRoot: common.HexToHash("05")},
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
	p              *processor
	description    string
	ctx            context.Context
	fromBlock      uint64
	toBlock        uint64
	expectedEvents []L1InfoTreeUpdate
	expectedErr    error
}

func (a *getClaimsAndBridgesAction) method() string {
	return "GetClaimsAndBridges"
}

func (a *getClaimsAndBridgesAction) desc() string {
	return a.description
}

func (a *getClaimsAndBridgesAction) execute(t *testing.T) {
	// TODO: add relevant getters
	// actualEvents, actualErr := a.p.GetClaimsAndBridges(a.ctx, a.fromBlock, a.toBlock)
	// require.Equal(t, a.expectedEvents, actualEvents)
	// require.Equal(t, a.expectedErr, actualErr)
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

// storeL1InfoTreeUpdates

type storeL1InfoTreeUpdatesAction struct {
	p           *processor
	description string
	b           block
	expectedErr error
}

func (a *storeL1InfoTreeUpdatesAction) method() string {
	return "storeL1InfoTreeUpdates"
}

func (a *storeL1InfoTreeUpdatesAction) desc() string {
	return a.description
}

func (a *storeL1InfoTreeUpdatesAction) execute(t *testing.T) {
	actualErr := a.p.processBlock(a.b)
	require.Equal(t, a.expectedErr, actualErr)
}
