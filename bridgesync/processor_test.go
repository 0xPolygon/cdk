package bridgesync

import (
	"context"
	"encoding/json"
	"fmt"
	"math/big"
	"os"
	"path"
	"slices"
	"testing"

	migrationsBridge "github.com/0xPolygon/cdk/bridgesync/migrations"
	"github.com/0xPolygon/cdk/log"
	"github.com/0xPolygon/cdk/sync"
	"github.com/0xPolygon/cdk/tree/testvectors"
	"github.com/ethereum/go-ethereum/common"
	"github.com/stretchr/testify/require"
)

func TestProceessor(t *testing.T) {
	path := path.Join(t.TempDir(), "file::memory:?cache=shared")
	log.Debugf("sqlite path: %s", path)
	err := migrationsBridge.RunMigrations(path)
	require.NoError(t, err)
	p, err := newProcessor(path, "foo")
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
		&getClaims{
			p:              p,
			description:    "on an empty processor",
			ctx:            context.Background(),
			fromBlock:      0,
			toBlock:        2,
			expectedClaims: nil,
			expectedErr:    ErrBlockNotProcessed,
		},
		&getBridges{
			p:               p,
			description:     "on an empty processor",
			ctx:             context.Background(),
			fromBlock:       0,
			toBlock:         2,
			expectedBridges: nil,
			expectedErr:     ErrBlockNotProcessed,
		},
		&processBlockAction{
			p:           p,
			description: "block1",
			block:       block1,
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
		&getClaims{
			p:              p,
			description:    "after block1: range 0, 2",
			ctx:            context.Background(),
			fromBlock:      0,
			toBlock:        2,
			expectedClaims: nil,
			expectedErr:    ErrBlockNotProcessed,
		},
		&getBridges{
			p:               p,
			description:     "after block1: range 0, 2",
			ctx:             context.Background(),
			fromBlock:       0,
			toBlock:         2,
			expectedBridges: nil,
			expectedErr:     ErrBlockNotProcessed,
		},
		&getClaims{
			p:              p,
			description:    "after block1: range 1, 1",
			ctx:            context.Background(),
			fromBlock:      1,
			toBlock:        1,
			expectedClaims: eventsToClaims(block1.Events),
			expectedErr:    nil,
		},
		&getBridges{
			p:               p,
			description:     "after block1: range 1, 1",
			ctx:             context.Background(),
			fromBlock:       1,
			toBlock:         1,
			expectedBridges: eventsToBridges(block1.Events),
			expectedErr:     nil,
		},
		&reorgAction{
			p:                 p,
			description:       "after block1",
			firstReorgedBlock: 1,
			expectedErr:       nil,
		},
		// processed: ~
		&getClaims{
			p:              p,
			description:    "after block1 reorged",
			ctx:            context.Background(),
			fromBlock:      0,
			toBlock:        2,
			expectedClaims: nil,
			expectedErr:    ErrBlockNotProcessed,
		},
		&getBridges{
			p:               p,
			description:     "after block1 reorged",
			ctx:             context.Background(),
			fromBlock:       0,
			toBlock:         2,
			expectedBridges: nil,
			expectedErr:     ErrBlockNotProcessed,
		},
		&processBlockAction{
			p:           p,
			description: "block1 (after it's reorged)",
			block:       block1,
			expectedErr: nil,
		},
		// processed: block3
		&processBlockAction{
			p:           p,
			description: "block3",
			block:       block3,
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
		&getClaims{
			p:              p,
			description:    "after block3: range 2, 2",
			ctx:            context.Background(),
			fromBlock:      2,
			toBlock:        2,
			expectedClaims: []Claim{},
			expectedErr:    nil,
		},
		&getClaims{
			p:           p,
			description: "after block3: range 1, 3",
			ctx:         context.Background(),
			fromBlock:   1,
			toBlock:     3,
			expectedClaims: append(
				eventsToClaims(block1.Events),
				eventsToClaims(block3.Events)...,
			),
			expectedErr: nil,
		},
		&getBridges{
			p:               p,
			description:     "after block3: range 2, 2",
			ctx:             context.Background(),
			fromBlock:       2,
			toBlock:         2,
			expectedBridges: []Bridge{},
			expectedErr:     nil,
		},
		&getBridges{
			p:           p,
			description: "after block3: range 1, 3",
			ctx:         context.Background(),
			fromBlock:   1,
			toBlock:     3,
			expectedBridges: append(
				eventsToBridges(block1.Events),
				eventsToBridges(block3.Events)...,
			),
			expectedErr: nil,
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
			expectedLastProcessedBlock: 1,
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
		&processBlockAction{
			p:           p,
			description: "block3 after reorg",
			block:       block3,
			expectedErr: nil,
		},
		// processed: block1, block3
		&processBlockAction{
			p:           p,
			description: "block4",
			block:       block4,
			expectedErr: nil,
		},
		// processed: block1, block3, block4
		&processBlockAction{
			p:           p,
			description: "block5",
			block:       block5,
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
		&getClaims{
			p:           p,
			description: "after block5: range 1, 3",
			ctx:         context.Background(),
			fromBlock:   1,
			toBlock:     3,
			expectedClaims: append(
				eventsToClaims(block1.Events),
				eventsToClaims(block3.Events)...,
			),
			expectedErr: nil,
		},
		&getClaims{
			p:           p,
			description: "after block5: range 4, 5",
			ctx:         context.Background(),
			fromBlock:   4,
			toBlock:     5,
			expectedClaims: append(
				eventsToClaims(block4.Events),
				eventsToClaims(block5.Events)...,
			),
			expectedErr: nil,
		},
		&getClaims{
			p:           p,
			description: "after block5: range 0, 5",
			ctx:         context.Background(),
			fromBlock:   0,
			toBlock:     5,
			expectedClaims: slices.Concat(
				eventsToClaims(block1.Events),
				eventsToClaims(block3.Events),
				eventsToClaims(block4.Events),
				eventsToClaims(block5.Events),
			),
			expectedErr: nil,
		},
	}

	for _, a := range actions {
		log.Debugf("%s: %s", a.method(), a.desc())
		a.execute(t)
	}
}

// BOILERPLATE

// blocks

var (
	block1 = sync.Block{
		Num: 1,
		Events: []interface{}{
			Event{Bridge: &Bridge{
				BlockNum:           1,
				BlockPos:           0,
				LeafType:           1,
				OriginNetwork:      1,
				OriginAddress:      common.HexToAddress("01"),
				DestinationNetwork: 1,
				DestinationAddress: common.HexToAddress("01"),
				Amount:             big.NewInt(1),
				Metadata:           common.Hex2Bytes("01"),
				DepositCount:       0,
			}},
			Event{Claim: &Claim{
				BlockNum:           1,
				BlockPos:           1,
				GlobalIndex:        big.NewInt(1),
				OriginNetwork:      1,
				OriginAddress:      common.HexToAddress("01"),
				DestinationAddress: common.HexToAddress("01"),
				Amount:             big.NewInt(1),
			}},
		},
	}
	block3 = sync.Block{
		Num: 3,
		Events: []interface{}{
			Event{Bridge: &Bridge{
				BlockNum:           3,
				BlockPos:           0,
				LeafType:           2,
				OriginNetwork:      2,
				OriginAddress:      common.HexToAddress("02"),
				DestinationNetwork: 2,
				DestinationAddress: common.HexToAddress("02"),
				Amount:             big.NewInt(2),
				Metadata:           common.Hex2Bytes("02"),
				DepositCount:       1,
			}},
			Event{Bridge: &Bridge{
				BlockNum:           3,
				BlockPos:           1,
				LeafType:           3,
				OriginNetwork:      3,
				OriginAddress:      common.HexToAddress("03"),
				DestinationNetwork: 3,
				DestinationAddress: common.HexToAddress("03"),
				Amount:             big.NewInt(0),
				Metadata:           common.Hex2Bytes("03"),
				DepositCount:       2,
			}},
		},
	}
	block4 = sync.Block{
		Num:    4,
		Events: []interface{}{},
	}
	block5 = sync.Block{
		Num: 5,
		Events: []interface{}{
			Event{Claim: &Claim{
				BlockNum:           4,
				BlockPos:           0,
				GlobalIndex:        big.NewInt(4),
				OriginNetwork:      4,
				OriginAddress:      common.HexToAddress("04"),
				DestinationAddress: common.HexToAddress("04"),
				Amount:             big.NewInt(4),
			}},
			Event{Claim: &Claim{
				BlockNum:           4,
				BlockPos:           1,
				GlobalIndex:        big.NewInt(5),
				OriginNetwork:      5,
				OriginAddress:      common.HexToAddress("05"),
				DestinationAddress: common.HexToAddress("05"),
				Amount:             big.NewInt(5),
			}},
		},
	}
)

// actions

type processAction interface {
	method() string
	desc() string
	execute(t *testing.T)
}

// GetClaims

type getClaims struct {
	p              *processor
	description    string
	ctx            context.Context
	fromBlock      uint64
	toBlock        uint64
	expectedClaims []Claim
	expectedErr    error
}

func (a *getClaims) method() string {
	return "GetClaims"
}

func (a *getClaims) desc() string {
	return a.description
}

func (a *getClaims) execute(t *testing.T) { //nolint:thelper
	actualEvents, actualErr := a.p.GetClaims(a.ctx, a.fromBlock, a.toBlock)
	require.Equal(t, a.expectedClaims, actualEvents)
	require.Equal(t, a.expectedErr, actualErr)
}

// GetBridges

type getBridges struct {
	p               *processor
	description     string
	ctx             context.Context
	fromBlock       uint64
	toBlock         uint64
	expectedBridges []Bridge
	expectedErr     error
}

func (a *getBridges) method() string {
	return "GetBridges"
}

func (a *getBridges) desc() string {
	return a.description
}

func (a *getBridges) execute(t *testing.T) { //nolint:thelper
	actualEvents, actualErr := a.p.GetBridges(a.ctx, a.fromBlock, a.toBlock)
	require.Equal(t, a.expectedBridges, actualEvents)
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
	t.Helper()

	actualLastProcessedBlock, actualErr := a.p.GetLastProcessedBlock(a.ctx)
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
	t.Helper()

	actualErr := a.p.Reorg(context.Background(), a.firstReorgedBlock)
	require.Equal(t, a.expectedErr, actualErr)
}

// storeBridgeEvents

type processBlockAction struct {
	p           *processor
	description string
	block       sync.Block
	expectedErr error
}

func (a *processBlockAction) method() string {
	return "storeBridgeEvents"
}

func (a *processBlockAction) desc() string {
	return a.description
}

func (a *processBlockAction) execute(t *testing.T) {
	t.Helper()

	actualErr := a.p.ProcessBlock(context.Background(), a.block)
	require.Equal(t, a.expectedErr, actualErr)
}

func eventsToBridges(events []interface{}) []Bridge {
	bridges := []Bridge{}
	for _, event := range events {
		e, ok := event.(Event)
		if !ok {
			panic("should be ok")
		}
		if e.Bridge != nil {
			bridges = append(bridges, *e.Bridge)
		}
	}
	return bridges
}

func eventsToClaims(events []interface{}) []Claim {
	claims := []Claim{}
	for _, event := range events {
		e, ok := event.(Event)
		if !ok {
			panic("should be ok")
		}
		if e.Claim != nil {
			claims = append(claims, *e.Claim)
		}
	}
	return claims
}

func TestHashBridge(t *testing.T) {
	data, err := os.ReadFile("../tree/testvectors/leaf-vectors.json")
	require.NoError(t, err)

	var leafVectors []testvectors.DepositVectorRaw
	err = json.Unmarshal(data, &leafVectors)
	require.NoError(t, err)

	for ti, testVector := range leafVectors {
		t.Run(fmt.Sprintf("Test vector %d", ti), func(t *testing.T) {
			amount, err := big.NewInt(0).SetString(testVector.Amount, 0)
			require.True(t, err)

			bridge := Bridge{
				OriginNetwork:      testVector.OriginNetwork,
				OriginAddress:      common.HexToAddress(testVector.TokenAddress),
				Amount:             amount,
				DestinationNetwork: testVector.DestinationNetwork,
				DestinationAddress: common.HexToAddress(testVector.DestinationAddress),
				DepositCount:       uint32(ti + 1),
				Metadata:           common.FromHex(testVector.Metadata),
			}
			require.Equal(t, common.HexToHash(testVector.ExpectedHash), bridge.Hash())
		})
	}
}
