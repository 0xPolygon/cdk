package bridgesync

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"math/big"
	"os"
	"path"
	"slices"
	"testing"

	migrationsBridge "github.com/0xPolygon/cdk/bridgesync/migrations"
	"github.com/0xPolygon/cdk/db"
	"github.com/0xPolygon/cdk/log"
	"github.com/0xPolygon/cdk/sync"
	"github.com/0xPolygon/cdk/tree/testvectors"
	"github.com/0xPolygon/cdk/tree/types"
	"github.com/ethereum/go-ethereum/common"
	"github.com/russross/meddler"
	"github.com/stretchr/testify/require"
)

func TestBigIntString(t *testing.T) {
	globalIndex := GenerateGlobalIndex(true, 0, 1093)
	fmt.Println(globalIndex.String())

	_, ok := new(big.Int).SetString(globalIndex.String(), 10)
	require.True(t, ok)

	dbPath := path.Join(t.TempDir(), "file::memory:?cache=shared")

	err := migrationsBridge.RunMigrations(dbPath)
	require.NoError(t, err)
	db, err := db.NewSQLiteDB(dbPath)
	require.NoError(t, err)

	ctx := context.Background()
	tx, err := db.BeginTx(ctx, nil)
	require.NoError(t, err)

	claim := &Claim{
		BlockNum:            1,
		BlockPos:            0,
		GlobalIndex:         GenerateGlobalIndex(true, 0, 1093),
		OriginNetwork:       11,
		Amount:              big.NewInt(11),
		OriginAddress:       common.HexToAddress("0x11"),
		DestinationAddress:  common.HexToAddress("0x11"),
		ProofLocalExitRoot:  types.Proof{},
		ProofRollupExitRoot: types.Proof{},
		MainnetExitRoot:     common.Hash{},
		RollupExitRoot:      common.Hash{},
		GlobalExitRoot:      common.Hash{},
		DestinationNetwork:  12,
	}

	_, err = tx.Exec(`INSERT INTO block (num) VALUES ($1)`, claim.BlockNum)
	require.NoError(t, err)
	require.NoError(t, meddler.Insert(tx, "claim", claim))

	require.NoError(t, tx.Commit())

	tx, err = db.BeginTx(ctx, nil)
	require.NoError(t, err)

	rows, err := tx.Query(`
		SELECT * FROM claim
		WHERE block_num >= $1 AND block_num <= $2;
	`, claim.BlockNum, claim.BlockNum)
	require.NoError(t, err)

	claimsFromDB := []*Claim{}
	require.NoError(t, meddler.ScanAll(rows, &claimsFromDB))
	require.Len(t, claimsFromDB, 1)
	require.Equal(t, claim, claimsFromDB[0])
}

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
			expectedErr:    fmt.Errorf(errBlockNotProcessedFormat, 2, 0),
		},
		&getBridges{
			p:               p,
			description:     "on an empty processor",
			ctx:             context.Background(),
			fromBlock:       0,
			toBlock:         2,
			expectedBridges: nil,
			expectedErr:     fmt.Errorf(errBlockNotProcessedFormat, 2, 0),
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
			expectedErr:    fmt.Errorf(errBlockNotProcessedFormat, 2, 1),
		},
		&getBridges{
			p:               p,
			description:     "after block1: range 0, 2",
			ctx:             context.Background(),
			fromBlock:       0,
			toBlock:         2,
			expectedBridges: nil,
			expectedErr:     fmt.Errorf(errBlockNotProcessedFormat, 2, 1),
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
			expectedErr:    fmt.Errorf(errBlockNotProcessedFormat, 2, 0),
		},
		&getBridges{
			p:               p,
			description:     "after block1 reorged",
			ctx:             context.Background(),
			fromBlock:       0,
			toBlock:         2,
			expectedBridges: nil,
			expectedErr:     fmt.Errorf(errBlockNotProcessedFormat, 2, 0),
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

func (a *getClaims) execute(t *testing.T) {
	t.Helper()
	actualEvents, actualErr := a.p.GetClaims(a.ctx, a.fromBlock, a.toBlock)
	require.Equal(t, a.expectedErr, actualErr)
	require.Equal(t, a.expectedClaims, actualEvents)
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

func (a *getBridges) execute(t *testing.T) {
	t.Helper()
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

func TestDecodeGlobalIndex(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name                string
		globalIndex         *big.Int
		expectedMainnetFlag bool
		expectedRollupIndex uint32
		expectedLocalIndex  uint32
		expectedErr         error
	}{
		{
			name:                "Mainnet flag true, rollup index 0",
			globalIndex:         GenerateGlobalIndex(true, 0, 2),
			expectedMainnetFlag: true,
			expectedRollupIndex: 0,
			expectedLocalIndex:  2,
			expectedErr:         nil,
		},
		{
			name:                "Mainnet flag true, indexes 0",
			globalIndex:         GenerateGlobalIndex(true, 0, 0),
			expectedMainnetFlag: true,
			expectedRollupIndex: 0,
			expectedLocalIndex:  0,
			expectedErr:         nil,
		},
		{
			name:                "Mainnet flag false, rollup index 0",
			globalIndex:         GenerateGlobalIndex(false, 0, 2),
			expectedMainnetFlag: false,
			expectedRollupIndex: 0,
			expectedLocalIndex:  2,
			expectedErr:         nil,
		},
		{
			name:                "Mainnet flag false, rollup index non-zero",
			globalIndex:         GenerateGlobalIndex(false, 11, 0),
			expectedMainnetFlag: false,
			expectedRollupIndex: 11,
			expectedLocalIndex:  0,
			expectedErr:         nil,
		},
		{
			name:                "Mainnet flag false, indexes 0",
			globalIndex:         GenerateGlobalIndex(false, 0, 0),
			expectedMainnetFlag: false,
			expectedRollupIndex: 0,
			expectedLocalIndex:  0,
			expectedErr:         nil,
		},
		{
			name:                "Mainnet flag false, indexes non zero",
			globalIndex:         GenerateGlobalIndex(false, 1231, 111234),
			expectedMainnetFlag: false,
			expectedRollupIndex: 1231,
			expectedLocalIndex:  111234,
			expectedErr:         nil,
		},
		{
			name:                "Invalid global index length",
			globalIndex:         big.NewInt(0).SetBytes([]byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}),
			expectedMainnetFlag: false,
			expectedRollupIndex: 0,
			expectedLocalIndex:  0,
			expectedErr:         errors.New("invalid global index length"),
		},
	}

	for _, tt := range tests {
		tt := tt

		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			mainnetFlag, rollupIndex, localExitRootIndex, err := DecodeGlobalIndex(tt.globalIndex)
			if tt.expectedErr != nil {
				require.EqualError(t, err, tt.expectedErr.Error())
			} else {
				require.NoError(t, err)
			}
			require.Equal(t, tt.expectedMainnetFlag, mainnetFlag)
			require.Equal(t, tt.expectedRollupIndex, rollupIndex)
			require.Equal(t, tt.expectedLocalIndex, localExitRootIndex)
		})
	}
}

func TestInsertAndGetClaim(t *testing.T) {
	path := path.Join(t.TempDir(), "file::memory:?cache=shared")
	log.Debugf("sqlite path: %s", path)
	err := migrationsBridge.RunMigrations(path)
	require.NoError(t, err)
	p, err := newProcessor(path, "foo")
	require.NoError(t, err)

	tx, err := p.db.BeginTx(context.Background(), nil)
	require.NoError(t, err)

	// insert test claim
	testClaim := &Claim{
		BlockNum:            1,
		BlockPos:            0,
		GlobalIndex:         GenerateGlobalIndex(true, 0, 1093),
		OriginNetwork:       11,
		OriginAddress:       common.HexToAddress("0x11"),
		DestinationAddress:  common.HexToAddress("0x11"),
		Amount:              big.NewInt(11),
		ProofLocalExitRoot:  types.Proof{},
		ProofRollupExitRoot: types.Proof{},
		MainnetExitRoot:     common.Hash{},
		RollupExitRoot:      common.Hash{},
		GlobalExitRoot:      common.Hash{},
		DestinationNetwork:  12,
		Metadata:            []byte("0x11"),
		IsMessage:           false,
	}

	_, err = tx.Exec(`INSERT INTO block (num) VALUES ($1)`, testClaim.BlockNum)
	require.NoError(t, err)
	require.NoError(t, meddler.Insert(tx, "claim", testClaim))

	require.NoError(t, tx.Commit())

	// get test claim
	claims, err := p.GetClaims(context.Background(), 1, 1)
	require.NoError(t, err)
	require.Len(t, claims, 1)
	require.Equal(t, testClaim, &claims[0])
}

type mockBridgeContract struct {
	lastUpdatedDepositCount uint32
	err                     error
}

func (m *mockBridgeContract) LastUpdatedDepositCount(ctx context.Context, blockNumber uint64) (uint32, error) {
	return m.lastUpdatedDepositCount, m.err
}

func TestGetBridgesPublished(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name                    string
		fromBlock               uint64
		toBlock                 uint64
		bridges                 []Bridge
		lastUpdatedDepositCount uint32
		expectedBridges         []Bridge
		expectedError           error
	}{
		{
			name:                    "no bridges",
			fromBlock:               1,
			toBlock:                 10,
			bridges:                 []Bridge{},
			lastUpdatedDepositCount: 0,
			expectedBridges:         []Bridge{},
			expectedError:           nil,
		},
		{
			name:      "bridges within deposit count",
			fromBlock: 1,
			toBlock:   10,
			bridges: []Bridge{
				{DepositCount: 1, BlockNum: 1, Amount: big.NewInt(1)},
				{DepositCount: 2, BlockNum: 2, Amount: big.NewInt(1)},
			},
			lastUpdatedDepositCount: 2,
			expectedBridges: []Bridge{
				{DepositCount: 1, BlockNum: 1, Amount: big.NewInt(1)},
				{DepositCount: 2, BlockNum: 2, Amount: big.NewInt(1)},
			},
			expectedError: nil,
		},
	}

	for _, tc := range testCases {
		tc := tc

		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			path := path.Join(t.TempDir(), "file::memory:?cache=shared")
			require.NoError(t, migrationsBridge.RunMigrations(path))
			p, err := newProcessor(path, "foo")
			require.NoError(t, err)

			tx, err := p.db.BeginTx(context.Background(), nil)
			require.NoError(t, err)

			for i := tc.fromBlock; i <= tc.toBlock; i++ {
				_, err = tx.Exec(`INSERT INTO block (num) VALUES ($1)`, i)
				require.NoError(t, err)
			}

			for _, bridge := range tc.bridges {
				require.NoError(t, meddler.Insert(tx, "bridge", &bridge))
			}

			require.NoError(t, tx.Commit())

			ctx := context.Background()
			bridges, err := p.GetBridgesPublished(ctx, tc.fromBlock, tc.toBlock)

			if tc.expectedError != nil {
				require.Equal(t, tc.expectedError, err)
			} else {
				require.NoError(t, err)
				require.Equal(t, tc.expectedBridges, bridges)
			}
		})
	}
}
