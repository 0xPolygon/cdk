package aggsender

import (
	"context"
	"crypto/ecdsa"
	"encoding/json"
	"errors"
	"fmt"
	"math/big"
	"os"
	"runtime"
	"testing"
	"time"

	"github.com/0xPolygon/cdk/agglayer"
	"github.com/0xPolygon/cdk/aggsender/db"
	"github.com/0xPolygon/cdk/aggsender/mocks"
	aggsendertypes "github.com/0xPolygon/cdk/aggsender/types"
	"github.com/0xPolygon/cdk/bridgesync"
	"github.com/0xPolygon/cdk/config/types"
	"github.com/0xPolygon/cdk/l1infotreesync"
	"github.com/0xPolygon/cdk/log"
	treeTypes "github.com/0xPolygon/cdk/tree/types"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

const (
	networkIDTest = uint32(1234)
)

var (
	errTest = errors.New("unitest  error")
	ler1    = common.HexToHash("0x123")
)

func TestConfigString(t *testing.T) {
	config := Config{
		StoragePath:                 "/path/to/storage",
		AggLayerURL:                 "http://agglayer.url",
		AggsenderPrivateKey:         types.KeystoreFileConfig{Path: "/path/to/key", Password: "password"},
		URLRPCL2:                    "http://l2.rpc.url",
		BlockFinality:               "latestBlock",
		EpochNotificationPercentage: 50,
		SaveCertificatesToFilesPath: "/path/to/certificates",
	}

	expected := "StoragePath: /path/to/storage\n" +
		"AggLayerURL: http://agglayer.url\n" +
		"AggsenderPrivateKeyPath: /path/to/key\n" +
		"URLRPCL2: http://l2.rpc.url\n" +
		"BlockFinality: latestBlock\n" +
		"EpochNotificationPercentage: 50\n" +
		"SaveCertificatesToFilesPath: /path/to/certificates\n"

	require.Equal(t, expected, config.String())
}

func TestConvertClaimToImportedBridgeExit(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name          string
		claim         bridgesync.Claim
		expectedError bool
		expectedExit  *agglayer.ImportedBridgeExit
	}{
		{
			name: "Asset claim",
			claim: bridgesync.Claim{
				IsMessage:          false,
				OriginNetwork:      1,
				OriginAddress:      common.HexToAddress("0x123"),
				DestinationNetwork: 2,
				DestinationAddress: common.HexToAddress("0x456"),
				Amount:             big.NewInt(100),
				Metadata:           []byte("metadata"),
				GlobalIndex:        big.NewInt(1),
			},
			expectedError: false,
			expectedExit: &agglayer.ImportedBridgeExit{
				BridgeExit: &agglayer.BridgeExit{
					LeafType: agglayer.LeafTypeAsset,
					TokenInfo: &agglayer.TokenInfo{
						OriginNetwork:      1,
						OriginTokenAddress: common.HexToAddress("0x123"),
					},
					DestinationNetwork: 2,
					DestinationAddress: common.HexToAddress("0x456"),
					Amount:             big.NewInt(100),
					Metadata:           []byte("metadata"),
				},
				GlobalIndex: &agglayer.GlobalIndex{
					MainnetFlag: false,
					RollupIndex: 0,
					LeafIndex:   1,
				},
			},
		},
		{
			name: "Message claim",
			claim: bridgesync.Claim{
				IsMessage:          true,
				OriginNetwork:      1,
				OriginAddress:      common.HexToAddress("0x123"),
				DestinationNetwork: 2,
				DestinationAddress: common.HexToAddress("0x456"),
				Amount:             big.NewInt(100),
				Metadata:           []byte("metadata"),
				GlobalIndex:        big.NewInt(2),
			},
			expectedError: false,
			expectedExit: &agglayer.ImportedBridgeExit{
				BridgeExit: &agglayer.BridgeExit{
					LeafType: agglayer.LeafTypeMessage,
					TokenInfo: &agglayer.TokenInfo{
						OriginNetwork:      1,
						OriginTokenAddress: common.HexToAddress("0x123"),
					},
					DestinationNetwork: 2,
					DestinationAddress: common.HexToAddress("0x456"),
					Amount:             big.NewInt(100),
					Metadata:           []byte("metadata"),
				},
				GlobalIndex: &agglayer.GlobalIndex{
					MainnetFlag: false,
					RollupIndex: 0,
					LeafIndex:   2,
				},
			},
		},
		{
			name: "Invalid global index",
			claim: bridgesync.Claim{
				IsMessage:          false,
				OriginNetwork:      1,
				OriginAddress:      common.HexToAddress("0x123"),
				DestinationNetwork: 2,
				DestinationAddress: common.HexToAddress("0x456"),
				Amount:             big.NewInt(100),
				Metadata:           []byte("metadata"),
				GlobalIndex:        new(big.Int).SetBytes([]byte{0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11}),
			},
			expectedError: true,
			expectedExit:  nil,
		},
	}

	for _, tt := range tests {
		tt := tt

		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			aggSender := &AggSender{}
			exit, err := aggSender.convertClaimToImportedBridgeExit(tt.claim)

			if tt.expectedError {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				require.Equal(t, tt.expectedExit, exit)
			}
		})
	}
}

func TestGetBridgeExits(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name          string
		bridges       []bridgesync.Bridge
		expectedExits []*agglayer.BridgeExit
	}{
		{
			name: "Single bridge",
			bridges: []bridgesync.Bridge{
				{
					LeafType:           agglayer.LeafTypeAsset.Uint8(),
					OriginNetwork:      1,
					OriginAddress:      common.HexToAddress("0x123"),
					DestinationNetwork: 2,
					DestinationAddress: common.HexToAddress("0x456"),
					Amount:             big.NewInt(100),
					Metadata:           []byte("metadata"),
				},
			},
			expectedExits: []*agglayer.BridgeExit{
				{
					LeafType: agglayer.LeafTypeAsset,
					TokenInfo: &agglayer.TokenInfo{
						OriginNetwork:      1,
						OriginTokenAddress: common.HexToAddress("0x123"),
					},
					DestinationNetwork: 2,
					DestinationAddress: common.HexToAddress("0x456"),
					Amount:             big.NewInt(100),
					Metadata:           []byte("metadata"),
				},
			},
		},
		{
			name: "Multiple bridges",
			bridges: []bridgesync.Bridge{
				{
					LeafType:           agglayer.LeafTypeAsset.Uint8(),
					OriginNetwork:      1,
					OriginAddress:      common.HexToAddress("0x123"),
					DestinationNetwork: 2,
					DestinationAddress: common.HexToAddress("0x456"),
					Amount:             big.NewInt(100),
					Metadata:           []byte("metadata"),
				},
				{
					LeafType:           agglayer.LeafTypeMessage.Uint8(),
					OriginNetwork:      3,
					OriginAddress:      common.HexToAddress("0x789"),
					DestinationNetwork: 4,
					DestinationAddress: common.HexToAddress("0xabc"),
					Amount:             big.NewInt(200),
					Metadata:           []byte("data"),
				},
			},
			expectedExits: []*agglayer.BridgeExit{
				{
					LeafType: agglayer.LeafTypeAsset,
					TokenInfo: &agglayer.TokenInfo{
						OriginNetwork:      1,
						OriginTokenAddress: common.HexToAddress("0x123"),
					},
					DestinationNetwork: 2,
					DestinationAddress: common.HexToAddress("0x456"),
					Amount:             big.NewInt(100),
					Metadata:           []byte("metadata"),
				},
				{
					LeafType: agglayer.LeafTypeMessage,
					TokenInfo: &agglayer.TokenInfo{
						OriginNetwork:      3,
						OriginTokenAddress: common.HexToAddress("0x789"),
					},
					DestinationNetwork: 4,
					DestinationAddress: common.HexToAddress("0xabc"),
					Amount:             big.NewInt(200),
					Metadata:           []byte("data"),
				},
			},
		},
		{
			name:          "No bridges",
			bridges:       []bridgesync.Bridge{},
			expectedExits: []*agglayer.BridgeExit{},
		},
	}

	for _, tt := range tests {
		tt := tt

		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			aggSender := &AggSender{}
			exits := aggSender.getBridgeExits(tt.bridges)

			require.Equal(t, tt.expectedExits, exits)
		})
	}
}

func TestAggSenderStart(t *testing.T) {
	aggLayerMock := agglayer.NewAgglayerClientMock(t)
	epochNotifierMock := mocks.NewEpochNotifier(t)
	bridgeL2SyncerMock := mocks.NewL2BridgeSyncer(t)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	aggSender, err := New(
		ctx,
		log.WithFields("test", "unittest"),
		Config{
			StoragePath:          "file:TestAggSenderStart?mode=memory&cache=shared",
			DelayBeetweenRetries: types.Duration{Duration: 1 * time.Microsecond},
		},
		aggLayerMock,
		nil,
		bridgeL2SyncerMock,
		epochNotifierMock)
	require.NoError(t, err)
	require.NotNil(t, aggSender)
	ch := make(chan aggsendertypes.EpochEvent)
	epochNotifierMock.EXPECT().Subscribe("aggsender").Return(ch)
	bridgeL2SyncerMock.EXPECT().OriginNetwork().Return(uint32(1))
	bridgeL2SyncerMock.EXPECT().GetLastProcessedBlock(mock.Anything).Return(uint64(0), nil)
	aggLayerMock.EXPECT().GetLatestKnownCertificateHeader(mock.Anything).Return(nil, nil)

	go aggSender.Start(ctx)
	ch <- aggsendertypes.EpochEvent{
		Epoch: 1,
	}
	time.Sleep(200 * time.Millisecond)
}

func TestAggSenderSendCertificates(t *testing.T) {
	AggLayerMock := agglayer.NewAgglayerClientMock(t)
	epochNotifierMock := mocks.NewEpochNotifier(t)
	bridgeL2SyncerMock := mocks.NewL2BridgeSyncer(t)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	aggSender, err := New(
		ctx,
		log.WithFields("test", "unittest"),
		Config{
			StoragePath: "file::memory:?cache=shared",
		},
		AggLayerMock,
		nil,
		bridgeL2SyncerMock,
		epochNotifierMock)
	require.NoError(t, err)
	require.NotNil(t, aggSender)
	ch := make(chan aggsendertypes.EpochEvent, 2)
	epochNotifierMock.EXPECT().Subscribe("aggsender").Return(ch)
	err = aggSender.storage.SaveLastSentCertificate(ctx, aggsendertypes.CertificateInfo{
		Height: 1,
		Status: agglayer.Pending,
	})
	AggLayerMock.EXPECT().GetCertificateHeader(mock.Anything).Return(&agglayer.CertificateHeader{
		Status: agglayer.Pending,
	}, nil)
	require.NoError(t, err)
	ch <- aggsendertypes.EpochEvent{
		Epoch: 1,
	}
	go aggSender.sendCertificates(ctx)
	time.Sleep(200 * time.Millisecond)
}

//nolint:dupl
func TestGetImportedBridgeExits(t *testing.T) {
	t.Parallel()

	mockProof := generateTestProof(t)

	mockL1InfoTreeSyncer := mocks.NewL1InfoTreeSyncer(t)
	mockL1InfoTreeSyncer.On("GetInfoByGlobalExitRoot", mock.Anything).Return(&l1infotreesync.L1InfoTreeLeaf{
		L1InfoTreeIndex:   1,
		Timestamp:         123456789,
		PreviousBlockHash: common.HexToHash("0xabc"),
		GlobalExitRoot:    common.HexToHash("0x7891"),
	}, nil)
	mockL1InfoTreeSyncer.On("GetL1InfoTreeRootByIndex", mock.Anything, mock.Anything).Return(
		treeTypes.Root{Hash: common.HexToHash("0x7891")}, nil)
	mockL1InfoTreeSyncer.On("GetL1InfoTreeMerkleProofFromIndexToRoot", mock.Anything,
		mock.Anything, mock.Anything).Return(mockProof, nil)

	tests := []struct {
		name          string
		claims        []bridgesync.Claim
		expectedError bool
		expectedExits []*agglayer.ImportedBridgeExit
	}{
		{
			name: "Single claim",
			claims: []bridgesync.Claim{
				{
					IsMessage:           false,
					OriginNetwork:       1,
					OriginAddress:       common.HexToAddress("0x1234"),
					DestinationNetwork:  2,
					DestinationAddress:  common.HexToAddress("0x4567"),
					Amount:              big.NewInt(111),
					Metadata:            []byte("metadata1"),
					GlobalIndex:         bridgesync.GenerateGlobalIndex(false, 1, 1),
					GlobalExitRoot:      common.HexToHash("0x7891"),
					RollupExitRoot:      common.HexToHash("0xaaab"),
					MainnetExitRoot:     common.HexToHash("0xbbba"),
					ProofLocalExitRoot:  mockProof,
					ProofRollupExitRoot: mockProof,
				},
			},
			expectedError: false,
			expectedExits: []*agglayer.ImportedBridgeExit{
				{
					BridgeExit: &agglayer.BridgeExit{
						LeafType: agglayer.LeafTypeAsset,
						TokenInfo: &agglayer.TokenInfo{
							OriginNetwork:      1,
							OriginTokenAddress: common.HexToAddress("0x1234"),
						},
						DestinationNetwork: 2,
						DestinationAddress: common.HexToAddress("0x4567"),
						Amount:             big.NewInt(111),
						Metadata:           []byte("metadata1"),
					},
					GlobalIndex: &agglayer.GlobalIndex{
						MainnetFlag: false,
						RollupIndex: 1,
						LeafIndex:   1,
					},
					ClaimData: &agglayer.ClaimFromRollup{
						L1Leaf: &agglayer.L1InfoTreeLeaf{
							L1InfoTreeIndex: 1,
							RollupExitRoot:  common.HexToHash("0xaaab"),
							MainnetExitRoot: common.HexToHash("0xbbba"),
							Inner: &agglayer.L1InfoTreeLeafInner{
								GlobalExitRoot: common.HexToHash("0x7891"),
								Timestamp:      123456789,
								BlockHash:      common.HexToHash("0xabc"),
							},
						},
						ProofLeafLER: &agglayer.MerkleProof{
							Root:  common.HexToHash("0xc52019815b51acf67a715cae6794a20083d63fd9af45783b7adf69123dae92c8"),
							Proof: mockProof,
						},
						ProofLERToRER: &agglayer.MerkleProof{
							Root:  common.HexToHash("0xaaab"),
							Proof: mockProof,
						},
						ProofGERToL1Root: &agglayer.MerkleProof{
							Root:  common.HexToHash("0x7891"),
							Proof: mockProof,
						},
					},
				},
			},
		},
		{
			name: "Multiple claims",
			claims: []bridgesync.Claim{
				{
					IsMessage:           false,
					OriginNetwork:       1,
					OriginAddress:       common.HexToAddress("0x123"),
					DestinationNetwork:  2,
					DestinationAddress:  common.HexToAddress("0x456"),
					Amount:              big.NewInt(100),
					Metadata:            []byte("metadata"),
					GlobalIndex:         big.NewInt(1),
					GlobalExitRoot:      common.HexToHash("0x7891"),
					RollupExitRoot:      common.HexToHash("0xaaa"),
					MainnetExitRoot:     common.HexToHash("0xbbb"),
					ProofLocalExitRoot:  mockProof,
					ProofRollupExitRoot: mockProof,
				},
				{
					IsMessage:           true,
					OriginNetwork:       3,
					OriginAddress:       common.HexToAddress("0x789"),
					DestinationNetwork:  4,
					DestinationAddress:  common.HexToAddress("0xabc"),
					Amount:              big.NewInt(200),
					Metadata:            []byte("data"),
					GlobalIndex:         bridgesync.GenerateGlobalIndex(true, 0, 2),
					GlobalExitRoot:      common.HexToHash("0x7891"),
					RollupExitRoot:      common.HexToHash("0xbbb"),
					MainnetExitRoot:     common.HexToHash("0xccc"),
					ProofLocalExitRoot:  mockProof,
					ProofRollupExitRoot: mockProof,
				},
			},
			expectedError: false,
			expectedExits: []*agglayer.ImportedBridgeExit{
				{
					BridgeExit: &agglayer.BridgeExit{
						LeafType: agglayer.LeafTypeAsset,
						TokenInfo: &agglayer.TokenInfo{
							OriginNetwork:      1,
							OriginTokenAddress: common.HexToAddress("0x123"),
						},
						DestinationNetwork: 2,
						DestinationAddress: common.HexToAddress("0x456"),
						Amount:             big.NewInt(100),
						Metadata:           []byte("metadata"),
					},
					GlobalIndex: &agglayer.GlobalIndex{
						MainnetFlag: false,
						RollupIndex: 0,
						LeafIndex:   1,
					},
					ClaimData: &agglayer.ClaimFromRollup{
						L1Leaf: &agglayer.L1InfoTreeLeaf{
							L1InfoTreeIndex: 1,
							RollupExitRoot:  common.HexToHash("0xaaa"),
							MainnetExitRoot: common.HexToHash("0xbbb"),
							Inner: &agglayer.L1InfoTreeLeafInner{
								GlobalExitRoot: common.HexToHash("0x7891"),
								Timestamp:      123456789,
								BlockHash:      common.HexToHash("0xabc"),
							},
						},
						ProofLeafLER: &agglayer.MerkleProof{
							Root:  common.HexToHash("0x105e0f1144e57f6fb63f1dfc5083b1f59be3512be7cf5e63523779ad14a4d987"),
							Proof: mockProof,
						},
						ProofLERToRER: &agglayer.MerkleProof{
							Root:  common.HexToHash("0xaaa"),
							Proof: mockProof,
						},
						ProofGERToL1Root: &agglayer.MerkleProof{
							Root:  common.HexToHash("0x7891"),
							Proof: mockProof,
						},
					},
				},
				{
					BridgeExit: &agglayer.BridgeExit{
						LeafType: agglayer.LeafTypeMessage,
						TokenInfo: &agglayer.TokenInfo{
							OriginNetwork:      3,
							OriginTokenAddress: common.HexToAddress("0x789"),
						},
						DestinationNetwork: 4,
						DestinationAddress: common.HexToAddress("0xabc"),
						Amount:             big.NewInt(200),
						Metadata:           []byte("data"),
					},
					GlobalIndex: &agglayer.GlobalIndex{
						MainnetFlag: true,
						RollupIndex: 0,
						LeafIndex:   2,
					},
					ClaimData: &agglayer.ClaimFromMainnnet{
						L1Leaf: &agglayer.L1InfoTreeLeaf{
							L1InfoTreeIndex: 1,
							RollupExitRoot:  common.HexToHash("0xbbb"),
							MainnetExitRoot: common.HexToHash("0xccc"),
							Inner: &agglayer.L1InfoTreeLeafInner{
								GlobalExitRoot: common.HexToHash("0x7891"),
								Timestamp:      123456789,
								BlockHash:      common.HexToHash("0xabc"),
							},
						},
						ProofLeafMER: &agglayer.MerkleProof{
							Root:  common.HexToHash("0xccc"),
							Proof: mockProof,
						},
						ProofGERToL1Root: &agglayer.MerkleProof{
							Root:  common.HexToHash("0x7891"),
							Proof: mockProof,
						},
					},
				},
			},
		},
		{
			name:          "No claims",
			claims:        []bridgesync.Claim{},
			expectedError: false,
			expectedExits: []*agglayer.ImportedBridgeExit{},
		},
	}

	for _, tt := range tests {
		tt := tt

		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			aggSender := &AggSender{
				l1infoTreeSyncer: mockL1InfoTreeSyncer,
				log:              log.WithFields("test", "unittest"),
			}
			exits, err := aggSender.getImportedBridgeExits(context.Background(), tt.claims)

			if tt.expectedError {
				require.Error(t, err)
				require.Nil(t, exits)
			} else {
				require.NoError(t, err)
				require.Equal(t, tt.expectedExits, exits)
			}
		})
	}
}

func TestBuildCertificate(t *testing.T) {
	mockL2BridgeSyncer := mocks.NewL2BridgeSyncer(t)
	mockL1InfoTreeSyncer := mocks.NewL1InfoTreeSyncer(t)
	mockProof := generateTestProof(t)

	tests := []struct {
		name                    string
		bridges                 []bridgesync.Bridge
		claims                  []bridgesync.Claim
		lastSentCertificateInfo aggsendertypes.CertificateInfo
		toBlock                 uint64
		mockFn                  func()
		expectedCert            *agglayer.Certificate
		expectedError           bool
	}{
		{
			name: "Valid certificate with bridges and claims",
			bridges: []bridgesync.Bridge{
				{
					LeafType:           agglayer.LeafTypeAsset.Uint8(),
					OriginNetwork:      1,
					OriginAddress:      common.HexToAddress("0x123"),
					DestinationNetwork: 2,
					DestinationAddress: common.HexToAddress("0x456"),
					Amount:             big.NewInt(100),
					Metadata:           []byte("metadata"),
					DepositCount:       1,
				},
			},
			claims: []bridgesync.Claim{
				{
					IsMessage:           false,
					OriginNetwork:       1,
					OriginAddress:       common.HexToAddress("0x1234"),
					DestinationNetwork:  2,
					DestinationAddress:  common.HexToAddress("0x4567"),
					Amount:              big.NewInt(111),
					Metadata:            []byte("metadata1"),
					GlobalIndex:         big.NewInt(1),
					GlobalExitRoot:      common.HexToHash("0x7891"),
					RollupExitRoot:      common.HexToHash("0xaaab"),
					MainnetExitRoot:     common.HexToHash("0xbbba"),
					ProofLocalExitRoot:  mockProof,
					ProofRollupExitRoot: mockProof,
				},
			},
			lastSentCertificateInfo: aggsendertypes.CertificateInfo{
				NewLocalExitRoot: common.HexToHash("0x123"),
				Height:           1,
				Status:           agglayer.Settled,
			},
			toBlock: 10,
			expectedCert: &agglayer.Certificate{
				NetworkID:         1,
				PrevLocalExitRoot: common.HexToHash("0x123"),
				NewLocalExitRoot:  common.HexToHash("0x789"),
				Metadata:          aggsendertypes.NewCertificateMetadata(0, 10, 0).ToHash(),
				BridgeExits: []*agglayer.BridgeExit{
					{
						LeafType: agglayer.LeafTypeAsset,
						TokenInfo: &agglayer.TokenInfo{
							OriginNetwork:      1,
							OriginTokenAddress: common.HexToAddress("0x123"),
						},
						DestinationNetwork: 2,
						DestinationAddress: common.HexToAddress("0x456"),
						Amount:             big.NewInt(100),
						Metadata:           []byte("metadata"),
					},
				},
				ImportedBridgeExits: []*agglayer.ImportedBridgeExit{
					{
						BridgeExit: &agglayer.BridgeExit{
							LeafType: agglayer.LeafTypeAsset,
							TokenInfo: &agglayer.TokenInfo{
								OriginNetwork:      1,
								OriginTokenAddress: common.HexToAddress("0x1234"),
							},
							DestinationNetwork: 2,
							DestinationAddress: common.HexToAddress("0x4567"),
							Amount:             big.NewInt(111),
							Metadata:           []byte("metadata1"),
						},
						GlobalIndex: &agglayer.GlobalIndex{
							MainnetFlag: false,
							RollupIndex: 0,
							LeafIndex:   1,
						},
						ClaimData: &agglayer.ClaimFromRollup{
							L1Leaf: &agglayer.L1InfoTreeLeaf{
								L1InfoTreeIndex: 1,
								RollupExitRoot:  common.HexToHash("0xaaab"),
								MainnetExitRoot: common.HexToHash("0xbbba"),
								Inner: &agglayer.L1InfoTreeLeafInner{
									GlobalExitRoot: common.HexToHash("0x7891"),
									Timestamp:      123456789,
									BlockHash:      common.HexToHash("0xabc"),
								},
							},
							ProofLeafLER: &agglayer.MerkleProof{
								Root:  common.HexToHash("0xc52019815b51acf67a715cae6794a20083d63fd9af45783b7adf69123dae92c8"),
								Proof: mockProof,
							},
							ProofLERToRER: &agglayer.MerkleProof{
								Root:  common.HexToHash("0xaaab"),
								Proof: mockProof,
							},
							ProofGERToL1Root: &agglayer.MerkleProof{
								Root:  common.HexToHash("0x7891"),
								Proof: mockProof,
							},
						},
					},
				},
				Height: 2,
			},
			mockFn: func() {
				mockL2BridgeSyncer.On("OriginNetwork").Return(uint32(1))
				mockL2BridgeSyncer.On("GetExitRootByIndex", mock.Anything, mock.Anything).Return(treeTypes.Root{Hash: common.HexToHash("0x789")}, nil)

				mockL1InfoTreeSyncer.On("GetInfoByGlobalExitRoot", mock.Anything).Return(&l1infotreesync.L1InfoTreeLeaf{
					L1InfoTreeIndex:   1,
					Timestamp:         123456789,
					PreviousBlockHash: common.HexToHash("0xabc"),
					GlobalExitRoot:    common.HexToHash("0x7891"),
				}, nil)
				mockL1InfoTreeSyncer.On("GetL1InfoTreeRootByIndex", mock.Anything, mock.Anything).Return(treeTypes.Root{Hash: common.HexToHash("0x7891")}, nil)
				mockL1InfoTreeSyncer.On("GetL1InfoTreeMerkleProofFromIndexToRoot", mock.Anything, mock.Anything, mock.Anything).Return(mockProof, nil)
			},
			expectedError: false,
		},
		{
			name:    "No bridges or claims",
			bridges: []bridgesync.Bridge{},
			claims:  []bridgesync.Claim{},
			lastSentCertificateInfo: aggsendertypes.CertificateInfo{
				NewLocalExitRoot: common.HexToHash("0x123"),
				Height:           1,
			},
			expectedCert:  nil,
			expectedError: true,
		},
		{
			name: "Error getting imported bridge exits",
			bridges: []bridgesync.Bridge{
				{
					LeafType:           agglayer.LeafTypeAsset.Uint8(),
					OriginNetwork:      1,
					OriginAddress:      common.HexToAddress("0x123"),
					DestinationNetwork: 2,
					DestinationAddress: common.HexToAddress("0x456"),
					Amount:             big.NewInt(100),
					Metadata:           []byte("metadata"),
					DepositCount:       1,
				},
			},
			claims: []bridgesync.Claim{
				{
					IsMessage:          false,
					OriginNetwork:      1,
					OriginAddress:      common.HexToAddress("0x1234"),
					DestinationNetwork: 2,
					DestinationAddress: common.HexToAddress("0x4567"),
					Amount:             big.NewInt(111),
					Metadata:           []byte("metadata1"),
					GlobalIndex:        new(big.Int).SetBytes([]byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}),
					GlobalExitRoot:     common.HexToHash("0x7891"),
					RollupExitRoot:     common.HexToHash("0xaaab"),
					MainnetExitRoot:    common.HexToHash("0xbbba"),
					ProofLocalExitRoot: mockProof,
				},
			},
			lastSentCertificateInfo: aggsendertypes.CertificateInfo{
				NewLocalExitRoot: common.HexToHash("0x123"),
				Height:           1,
			},
			mockFn: func() {
				mockL1InfoTreeSyncer.On("GetInfoByGlobalExitRoot", mock.Anything).Return(&l1infotreesync.L1InfoTreeLeaf{
					L1InfoTreeIndex:   1,
					Timestamp:         123456789,
					PreviousBlockHash: common.HexToHash("0xabc"),
					GlobalExitRoot:    common.HexToHash("0x7891"),
				}, nil)
				mockL1InfoTreeSyncer.On("GetL1InfoTreeRootByIndex", mock.Anything, mock.Anything).Return(
					treeTypes.Root{Hash: common.HexToHash("0x7891")}, nil)
			},
			expectedCert:  nil,
			expectedError: true,
		},
	}

	for _, tt := range tests {
		tt := tt

		t.Run(tt.name, func(t *testing.T) {
			mockL1InfoTreeSyncer.ExpectedCalls = nil
			mockL2BridgeSyncer.ExpectedCalls = nil

			if tt.mockFn != nil {
				tt.mockFn()
			}

			aggSender := &AggSender{
				l2Syncer:         mockL2BridgeSyncer,
				l1infoTreeSyncer: mockL1InfoTreeSyncer,
				log:              log.WithFields("test", "unittest"),
			}

			certParam := &aggsendertypes.CertificateBuildParams{
				ToBlock: tt.toBlock,
				Bridges: tt.bridges,
				Claims:  tt.claims,
			}
			cert, err := aggSender.buildCertificate(context.Background(), certParam, &tt.lastSentCertificateInfo)

			if tt.expectedError {
				require.Error(t, err)
				require.Nil(t, cert)
			} else {
				require.NoError(t, err)
				require.Equal(t, tt.expectedCert, cert)
			}
		})
	}
}

func generateTestProof(t *testing.T) treeTypes.Proof {
	t.Helper()

	proof := treeTypes.Proof{}

	for i := 0; i < int(treeTypes.DefaultHeight) && i < 10; i++ {
		proof[i] = common.HexToHash(fmt.Sprintf("0x%d", i))
	}

	return proof
}

func TestCheckIfCertificatesAreSettled(t *testing.T) {
	tests := []struct {
		name                     string
		pendingCertificates      []*aggsendertypes.CertificateInfo
		certificateHeaders       map[common.Hash]*agglayer.CertificateHeader
		getFromDBError           error
		clientError              error
		updateDBError            error
		expectedErrorLogMessages []string
		expectedInfoMessages     []string
		expectedError            bool
	}{
		{
			name: "All certificates settled - update successful",
			pendingCertificates: []*aggsendertypes.CertificateInfo{
				{CertificateID: common.HexToHash("0x1"), Height: 1},
				{CertificateID: common.HexToHash("0x2"), Height: 2},
			},
			certificateHeaders: map[common.Hash]*agglayer.CertificateHeader{
				common.HexToHash("0x1"): {Status: agglayer.Settled},
				common.HexToHash("0x2"): {Status: agglayer.Settled},
			},
			expectedInfoMessages: []string{
				"certificate %s changed status to %s",
			},
		},
		{
			name: "Some certificates in error - update successful",
			pendingCertificates: []*aggsendertypes.CertificateInfo{
				{CertificateID: common.HexToHash("0x1"), Height: 1},
				{CertificateID: common.HexToHash("0x2"), Height: 2},
			},
			certificateHeaders: map[common.Hash]*agglayer.CertificateHeader{
				common.HexToHash("0x1"): {Status: agglayer.InError},
				common.HexToHash("0x2"): {Status: agglayer.Settled},
			},
			expectedInfoMessages: []string{
				"certificate %s changed status to %s",
			},
		},
		{
			name:           "Error getting pending certificates",
			getFromDBError: fmt.Errorf("storage error"),
			expectedErrorLogMessages: []string{
				"error getting pending certificates: %w",
			},
			expectedError: true,
		},
		{
			name: "Error getting certificate header",
			pendingCertificates: []*aggsendertypes.CertificateInfo{
				{CertificateID: common.HexToHash("0x1"), Height: 1},
			},
			certificateHeaders: map[common.Hash]*agglayer.CertificateHeader{
				common.HexToHash("0x1"): {Status: agglayer.InError},
			},
			clientError: fmt.Errorf("client error"),
			expectedErrorLogMessages: []string{
				"error getting header of certificate %s with height: %d from agglayer: %w",
			},
			expectedError: true,
		},
		{
			name: "Error updating certificate status",
			pendingCertificates: []*aggsendertypes.CertificateInfo{
				{CertificateID: common.HexToHash("0x1"), Height: 1},
			},
			certificateHeaders: map[common.Hash]*agglayer.CertificateHeader{
				common.HexToHash("0x1"): {Status: agglayer.Settled},
			},
			updateDBError: fmt.Errorf("update error"),
			expectedErrorLogMessages: []string{
				"error updating certificate status in storage: %w",
			},
			expectedInfoMessages: []string{
				"certificate %s changed status to %s",
			},
			expectedError: true,
		},
	}

	for _, tt := range tests {
		tt := tt

		t.Run(tt.name, func(t *testing.T) {
			mockStorage := mocks.NewAggSenderStorage(t)
			mockAggLayerClient := agglayer.NewAgglayerClientMock(t)
			mockLogger := log.WithFields("test", "unittest")

			mockStorage.On("GetCertificatesByStatus", agglayer.NonSettledStatuses).Return(
				tt.pendingCertificates, tt.getFromDBError)
			for certID, header := range tt.certificateHeaders {
				mockAggLayerClient.On("GetCertificateHeader", certID).Return(header, tt.clientError)
			}
			if tt.updateDBError != nil {
				mockStorage.On("UpdateCertificate", mock.Anything, mock.Anything).Return(tt.updateDBError)
			} else if tt.clientError == nil && tt.getFromDBError == nil {
				mockStorage.On("UpdateCertificate", mock.Anything, mock.Anything).Return(nil)
			}

			aggSender := &AggSender{
				log:            mockLogger,
				storage:        mockStorage,
				aggLayerClient: mockAggLayerClient,
				cfg:            Config{},
			}

			ctx := context.TODO()
			thereArePendingCerts := aggSender.checkPendingCertificatesStatus(ctx)
			require.Equal(t, tt.expectedError, thereArePendingCerts)
			mockAggLayerClient.AssertExpectations(t)
			mockStorage.AssertExpectations(t)
		})
	}
}

func TestSendCertificate(t *testing.T) {
	t.Parallel()

	privateKey, err := crypto.GenerateKey()
	require.NoError(t, err)

	type testCfg struct {
		name                                    string
		sequencerKey                            *ecdsa.PrivateKey
		shouldSendCertificate                   []interface{}
		getLastSentCertificate                  []interface{}
		lastL2BlockProcessed                    []interface{}
		getBridges                              []interface{}
		getClaims                               []interface{}
		getInfoByGlobalExitRoot                 []interface{}
		getL1InfoTreeRootByIndex                []interface{}
		getL1InfoTreeMerkleProofFromIndexToRoot []interface{}
		getExitRootByIndex                      []interface{}
		originNetwork                           []interface{}
		sendCertificate                         []interface{}
		saveLastSentCertificate                 []interface{}
		expectedError                           string
	}

	setupTest := func(cfg testCfg) (*AggSender, *mocks.AggSenderStorage, *mocks.L2BridgeSyncer,
		*agglayer.AgglayerClientMock, *mocks.L1InfoTreeSyncer) {
		var (
			aggsender = &AggSender{
				log:          log.WithFields("aggsender", 1),
				cfg:          Config{MaxRetriesStoreCertificate: 1},
				sequencerKey: cfg.sequencerKey,
			}
			mockStorage          *mocks.AggSenderStorage
			mockL2Syncer         *mocks.L2BridgeSyncer
			mockAggLayerClient   *agglayer.AgglayerClientMock
			mockL1InfoTreeSyncer *mocks.L1InfoTreeSyncer
		)

		if cfg.shouldSendCertificate != nil || cfg.getLastSentCertificate != nil ||
			cfg.saveLastSentCertificate != nil {
			mockStorage = mocks.NewAggSenderStorage(t)
			mockStorage.On("GetCertificatesByStatus", agglayer.NonSettledStatuses).
				Return(cfg.shouldSendCertificate...)

			aggsender.storage = mockStorage

			if cfg.getLastSentCertificate != nil {
				mockStorage.On("GetLastSentCertificate").Return(cfg.getLastSentCertificate...).Once()
			}

			if cfg.saveLastSentCertificate != nil {
				mockStorage.On("SaveLastSentCertificate", mock.Anything, mock.Anything).Return(cfg.saveLastSentCertificate...)
			}
		}

		if cfg.lastL2BlockProcessed != nil || cfg.originNetwork != nil ||
			cfg.getBridges != nil || cfg.getClaims != nil || cfg.getInfoByGlobalExitRoot != nil {
			mockL2Syncer = mocks.NewL2BridgeSyncer(t)

			mockL2Syncer.On("GetLastProcessedBlock", mock.Anything).Return(cfg.lastL2BlockProcessed...).Once()

			if cfg.getBridges != nil {
				mockL2Syncer.On("GetBridgesPublished", mock.Anything, mock.Anything, mock.Anything).Return(cfg.getBridges...)
			}

			if cfg.getClaims != nil {
				mockL2Syncer.On("GetClaims", mock.Anything, mock.Anything, mock.Anything).Return(cfg.getClaims...).Once()
			}

			if cfg.getExitRootByIndex != nil {
				mockL2Syncer.On("GetExitRootByIndex", mock.Anything, mock.Anything).Return(cfg.getExitRootByIndex...).Once()
			}

			if cfg.originNetwork != nil {
				mockL2Syncer.On("OriginNetwork").Return(cfg.originNetwork...).Once()
			}

			aggsender.l2Syncer = mockL2Syncer
		}

		if cfg.sendCertificate != nil {
			mockAggLayerClient = agglayer.NewAgglayerClientMock(t)
			mockAggLayerClient.On("SendCertificate", mock.Anything).Return(cfg.sendCertificate...).Once()

			aggsender.aggLayerClient = mockAggLayerClient
		}

		if cfg.getInfoByGlobalExitRoot != nil ||
			cfg.getL1InfoTreeRootByIndex != nil || cfg.getL1InfoTreeMerkleProofFromIndexToRoot != nil {
			mockL1InfoTreeSyncer = mocks.NewL1InfoTreeSyncer(t)
			mockL1InfoTreeSyncer.On("GetInfoByGlobalExitRoot", mock.Anything).Return(cfg.getInfoByGlobalExitRoot...).Once()

			if cfg.getL1InfoTreeRootByIndex != nil {
				mockL1InfoTreeSyncer.On("GetL1InfoTreeRootByIndex", mock.Anything, mock.Anything).Return(cfg.getL1InfoTreeRootByIndex...).Once()
			}

			if cfg.getL1InfoTreeMerkleProofFromIndexToRoot != nil {
				mockL1InfoTreeSyncer.On("GetL1InfoTreeMerkleProofFromIndexToRoot", mock.Anything, mock.Anything, mock.Anything).
					Return(cfg.getL1InfoTreeMerkleProofFromIndexToRoot...).Once()
			}

			aggsender.l1infoTreeSyncer = mockL1InfoTreeSyncer
		}

		return aggsender, mockStorage, mockL2Syncer, mockAggLayerClient, mockL1InfoTreeSyncer
	}

	tests := []testCfg{
		{
			name:                  "error getting pending certificates",
			shouldSendCertificate: []interface{}{nil, errors.New("error getting pending")},
			expectedError:         "error getting pending",
		},
		{
			name: "should not send certificate",
			shouldSendCertificate: []interface{}{[]*aggsendertypes.CertificateInfo{
				{Status: agglayer.Pending},
			}, nil},
		},
		{
			name:                   "error getting last sent certificate",
			shouldSendCertificate:  []interface{}{[]*aggsendertypes.CertificateInfo{}, nil},
			lastL2BlockProcessed:   []interface{}{uint64(8), nil},
			getLastSentCertificate: []interface{}{&aggsendertypes.CertificateInfo{}, errors.New("error getting last sent certificate")},
			expectedError:          "error getting last sent certificate",
		},
		{
			name:                  "no new blocks to send certificate",
			shouldSendCertificate: []interface{}{[]*aggsendertypes.CertificateInfo{}, nil},
			lastL2BlockProcessed:  []interface{}{uint64(41), nil},
			getLastSentCertificate: []interface{}{&aggsendertypes.CertificateInfo{
				Height:           41,
				CertificateID:    common.HexToHash("0x111"),
				NewLocalExitRoot: common.HexToHash("0x13223"),
				FromBlock:        31,
				ToBlock:          41,
			}, nil},
		},
		{
			name:                  "get bridges error",
			shouldSendCertificate: []interface{}{[]*aggsendertypes.CertificateInfo{}, nil},
			lastL2BlockProcessed:  []interface{}{uint64(59), nil},
			getLastSentCertificate: []interface{}{&aggsendertypes.CertificateInfo{
				Height:           50,
				CertificateID:    common.HexToHash("0x1111"),
				NewLocalExitRoot: common.HexToHash("0x132233"),
				FromBlock:        40,
				ToBlock:          41,
			}, nil},
			getBridges:    []interface{}{nil, errors.New("error getting bridges")},
			expectedError: "error getting bridges",
		},
		{
			name:                  "no bridges",
			shouldSendCertificate: []interface{}{[]*aggsendertypes.CertificateInfo{}, nil},
			lastL2BlockProcessed:  []interface{}{uint64(69), nil},
			getLastSentCertificate: []interface{}{&aggsendertypes.CertificateInfo{
				Height:           60,
				CertificateID:    common.HexToHash("0x11111"),
				NewLocalExitRoot: common.HexToHash("0x1322233"),
				FromBlock:        50,
				ToBlock:          51,
			}, nil},
			getBridges: []interface{}{[]bridgesync.Bridge{}, nil},
		},
		{
			name:                  "get claims error",
			shouldSendCertificate: []interface{}{[]*aggsendertypes.CertificateInfo{}, nil},
			lastL2BlockProcessed:  []interface{}{uint64(79), nil},
			getLastSentCertificate: []interface{}{&aggsendertypes.CertificateInfo{
				Height:           70,
				CertificateID:    common.HexToHash("0x121111"),
				NewLocalExitRoot: common.HexToHash("0x13122233"),
				FromBlock:        60,
				ToBlock:          61,
			}, nil},
			getBridges: []interface{}{[]bridgesync.Bridge{
				{
					BlockNum:      61,
					BlockPos:      0,
					LeafType:      agglayer.LeafTypeAsset.Uint8(),
					OriginNetwork: 1,
				},
			}, nil},
			getClaims:     []interface{}{nil, errors.New("error getting claims")},
			expectedError: "error getting claims",
		},
		{
			name:                  "error getting info by global exit root",
			shouldSendCertificate: []interface{}{[]*aggsendertypes.CertificateInfo{}, nil},
			lastL2BlockProcessed:  []interface{}{uint64(89), nil},
			getLastSentCertificate: []interface{}{&aggsendertypes.CertificateInfo{
				Height:           80,
				CertificateID:    common.HexToHash("0x1321111"),
				NewLocalExitRoot: common.HexToHash("0x131122233"),
				FromBlock:        70,
				ToBlock:          71,
			}, nil},
			getBridges: []interface{}{[]bridgesync.Bridge{
				{
					BlockNum:      71,
					BlockPos:      0,
					LeafType:      agglayer.LeafTypeAsset.Uint8(),
					OriginNetwork: 1,
				},
			}, nil},
			getClaims: []interface{}{[]bridgesync.Claim{
				{
					IsMessage: false,
				},
			}, nil},
			getInfoByGlobalExitRoot: []interface{}{nil, errors.New("error getting info by global exit root")},
			expectedError:           "error getting info by global exit root",
		},
		{
			name:                  "error getting L1 Info tree root by index",
			shouldSendCertificate: []interface{}{[]*aggsendertypes.CertificateInfo{}, nil},
			lastL2BlockProcessed:  []interface{}{uint64(89), nil},
			getLastSentCertificate: []interface{}{&aggsendertypes.CertificateInfo{
				Height:           80,
				CertificateID:    common.HexToHash("0x1321111"),
				NewLocalExitRoot: common.HexToHash("0x131122233"),
				FromBlock:        70,
				ToBlock:          71,
			}, nil},
			getBridges: []interface{}{[]bridgesync.Bridge{
				{
					BlockNum:      71,
					BlockPos:      0,
					LeafType:      agglayer.LeafTypeAsset.Uint8(),
					OriginNetwork: 1,
				},
			}, nil},
			getClaims: []interface{}{[]bridgesync.Claim{
				{
					IsMessage: false,
				},
			}, nil},
			getInfoByGlobalExitRoot: []interface{}{&l1infotreesync.L1InfoTreeLeaf{
				L1InfoTreeIndex:   1,
				BlockNumber:       1,
				BlockPosition:     0,
				PreviousBlockHash: common.HexToHash("0x123"),
				Timestamp:         123456789,
				MainnetExitRoot:   common.HexToHash("0xccc"),
				RollupExitRoot:    common.HexToHash("0xddd"),
				GlobalExitRoot:    common.HexToHash("0xeee"),
			}, nil},
			getL1InfoTreeRootByIndex: []interface{}{treeTypes.Root{}, errors.New("error getting L1 Info tree root by index")},
			expectedError:            "error getting L1 Info tree root by index",
		},
		{
			name:                  "error getting L1 Info tree merkle proof from index to root",
			shouldSendCertificate: []interface{}{[]*aggsendertypes.CertificateInfo{}, nil},
			lastL2BlockProcessed:  []interface{}{uint64(89), nil},
			getLastSentCertificate: []interface{}{&aggsendertypes.CertificateInfo{
				Height:           80,
				CertificateID:    common.HexToHash("0x1321111"),
				NewLocalExitRoot: common.HexToHash("0x131122233"),
				FromBlock:        70,
				ToBlock:          71,
			}, nil},
			getBridges: []interface{}{[]bridgesync.Bridge{
				{
					BlockNum:      71,
					BlockPos:      0,
					LeafType:      agglayer.LeafTypeAsset.Uint8(),
					OriginNetwork: 1,
				},
			}, nil},
			getClaims: []interface{}{[]bridgesync.Claim{
				{
					IsMessage:   false,
					GlobalIndex: big.NewInt(1),
				},
			}, nil},
			getInfoByGlobalExitRoot: []interface{}{&l1infotreesync.L1InfoTreeLeaf{
				L1InfoTreeIndex:   1,
				BlockNumber:       1,
				BlockPosition:     0,
				PreviousBlockHash: common.HexToHash("0x123"),
				Timestamp:         123456789,
				MainnetExitRoot:   common.HexToHash("0xccc"),
				RollupExitRoot:    common.HexToHash("0xddd"),
				GlobalExitRoot:    common.HexToHash("0xeee"),
			}, nil},
			getL1InfoTreeRootByIndex:                []interface{}{treeTypes.Root{Hash: common.HexToHash("0xeee")}, nil},
			getL1InfoTreeMerkleProofFromIndexToRoot: []interface{}{treeTypes.Proof{}, errors.New("error getting L1 Info tree merkle proof")},
			expectedError:                           "error getting L1 Info tree merkle proof for leaf index",
		},
		{
			name:                  "send certificate error",
			shouldSendCertificate: []interface{}{[]*aggsendertypes.CertificateInfo{}, nil},
			lastL2BlockProcessed:  []interface{}{uint64(99), nil},
			getLastSentCertificate: []interface{}{&aggsendertypes.CertificateInfo{
				Height:                90,
				CertificateID:         common.HexToHash("0x1121111"),
				NewLocalExitRoot:      common.HexToHash("0x111122211"),
				PreviousLocalExitRoot: &ler1,
				FromBlock:             80,
				ToBlock:               81,
				Status:                agglayer.Settled,
			}, nil},
			getBridges: []interface{}{[]bridgesync.Bridge{
				{
					BlockNum:      81,
					BlockPos:      0,
					LeafType:      agglayer.LeafTypeAsset.Uint8(),
					OriginNetwork: 1,
					DepositCount:  1,
				},
			}, nil},
			getClaims:          []interface{}{[]bridgesync.Claim{}, nil},
			getExitRootByIndex: []interface{}{treeTypes.Root{}, nil},
			originNetwork:      []interface{}{uint32(1), nil},
			sendCertificate:    []interface{}{common.Hash{}, errors.New("error sending certificate")},
			sequencerKey:       privateKey,
			expectedError:      "error sending certificate",
		},
		{
			name:                  "store last sent certificate error",
			shouldSendCertificate: []interface{}{[]*aggsendertypes.CertificateInfo{}, nil},
			lastL2BlockProcessed:  []interface{}{uint64(109), nil},
			getLastSentCertificate: []interface{}{&aggsendertypes.CertificateInfo{
				Height:           100,
				CertificateID:    common.HexToHash("0x11121111"),
				NewLocalExitRoot: common.HexToHash("0x1211122211"),
				FromBlock:        90,
				ToBlock:          91,
				Status:           agglayer.Settled,
			}, nil},
			getBridges: []interface{}{[]bridgesync.Bridge{
				{
					BlockNum:      91,
					BlockPos:      0,
					LeafType:      agglayer.LeafTypeAsset.Uint8(),
					OriginNetwork: 1,
					DepositCount:  1,
				},
			}, nil},
			getClaims:               []interface{}{[]bridgesync.Claim{}, nil},
			getExitRootByIndex:      []interface{}{treeTypes.Root{}, nil},
			originNetwork:           []interface{}{uint32(1), nil},
			sendCertificate:         []interface{}{common.Hash{}, nil},
			saveLastSentCertificate: []interface{}{errors.New("error saving last sent certificate in db")},
			sequencerKey:            privateKey,
			expectedError:           "error saving last sent certificate in db",
		},
		{
			name:                  "successful sending of certificate",
			shouldSendCertificate: []interface{}{[]*aggsendertypes.CertificateInfo{}, nil},
			lastL2BlockProcessed:  []interface{}{uint64(119), nil},
			getLastSentCertificate: []interface{}{&aggsendertypes.CertificateInfo{
				Height:           110,
				CertificateID:    common.HexToHash("0x12121111"),
				NewLocalExitRoot: common.HexToHash("0x1221122211"),
				FromBlock:        100,
				ToBlock:          101,
				Status:           agglayer.Settled,
			}, nil},
			getBridges: []interface{}{[]bridgesync.Bridge{
				{
					BlockNum:      101,
					BlockPos:      0,
					LeafType:      agglayer.LeafTypeAsset.Uint8(),
					OriginNetwork: 1,
					DepositCount:  1,
				},
			}, nil},
			getClaims:               []interface{}{[]bridgesync.Claim{}, nil},
			getExitRootByIndex:      []interface{}{treeTypes.Root{}, nil},
			originNetwork:           []interface{}{uint32(1), nil},
			sendCertificate:         []interface{}{common.Hash{}, nil},
			saveLastSentCertificate: []interface{}{nil},
			sequencerKey:            privateKey,
		},
	}

	for _, tt := range tests {
		tt := tt

		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			aggsender, mockStorage, mockL2Syncer,
				mockAggLayerClient, mockL1InfoTreeSyncer := setupTest(tt)

			_, err := aggsender.sendCertificate(context.Background())

			if tt.expectedError != "" {
				require.ErrorContains(t, err, tt.expectedError)
			} else {
				require.NoError(t, err)
			}

			if mockStorage != nil {
				mockStorage.AssertExpectations(t)
			}

			if mockL2Syncer != nil {
				mockL2Syncer.AssertExpectations(t)
			}

			if mockAggLayerClient != nil {
				mockAggLayerClient.AssertExpectations(t)
			}

			if mockL1InfoTreeSyncer != nil {
				mockL1InfoTreeSyncer.AssertExpectations(t)
			}
		})
	}
}

func TestExtractSignatureData(t *testing.T) {
	t.Parallel()

	testR := common.HexToHash("0x1")
	testV := common.HexToHash("0x2")

	tests := []struct {
		name              string
		signature         []byte
		expectedR         common.Hash
		expectedS         common.Hash
		expectedOddParity bool
		expectedError     error
	}{
		{
			name:              "Valid signature - odd parity",
			signature:         append(append(testR.Bytes(), testV.Bytes()...), 1),
			expectedR:         testR,
			expectedS:         testV,
			expectedOddParity: true,
			expectedError:     nil,
		},
		{
			name:              "Valid signature - even parity",
			signature:         append(append(testR.Bytes(), testV.Bytes()...), 2),
			expectedR:         testR,
			expectedS:         testV,
			expectedOddParity: false,
			expectedError:     nil,
		},
		{
			name:          "Invalid signature size",
			signature:     make([]byte, 64), // Invalid size
			expectedError: errInvalidSignatureSize,
		},
	}

	for _, tt := range tests {
		tt := tt

		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			r, s, isOddParity, err := extractSignatureData(tt.signature)

			if tt.expectedError != nil {
				require.Error(t, err)
				require.Equal(t, tt.expectedError, err)
			} else {
				require.NoError(t, err)
				require.Equal(t, tt.expectedR, r)
				require.Equal(t, tt.expectedS, s)
				require.Equal(t, tt.expectedOddParity, isOddParity)
			}
		})
	}
}

func TestExploratoryGenerateCert(t *testing.T) {
	t.Skip("This test is only for exploratory purposes, to generate json format of the certificate")

	key, err := crypto.GenerateKey()
	require.NoError(t, err)

	signature, err := crypto.Sign(common.HexToHash("0x1").Bytes(), key)
	require.NoError(t, err)

	r, s, v, err := extractSignatureData(signature)
	require.NoError(t, err)

	certificate := &agglayer.SignedCertificate{
		Certificate: &agglayer.Certificate{
			NetworkID:         1,
			Height:            1,
			PrevLocalExitRoot: common.HexToHash("0x1"),
			NewLocalExitRoot:  common.HexToHash("0x2"),
			BridgeExits: []*agglayer.BridgeExit{
				{
					LeafType: agglayer.LeafTypeAsset,
					TokenInfo: &agglayer.TokenInfo{
						OriginNetwork:      1,
						OriginTokenAddress: common.HexToAddress("0x11"),
					},
					DestinationNetwork: 2,
					DestinationAddress: common.HexToAddress("0x22"),
					Amount:             big.NewInt(100),
					Metadata:           []byte("metadata"),
				},
			},
			ImportedBridgeExits: []*agglayer.ImportedBridgeExit{
				{
					GlobalIndex: &agglayer.GlobalIndex{
						MainnetFlag: false,
						RollupIndex: 1,
						LeafIndex:   11,
					},
					BridgeExit: &agglayer.BridgeExit{
						LeafType: agglayer.LeafTypeAsset,
						TokenInfo: &agglayer.TokenInfo{
							OriginNetwork:      1,
							OriginTokenAddress: common.HexToAddress("0x11"),
						},
						DestinationNetwork: 2,
						DestinationAddress: common.HexToAddress("0x22"),
						Amount:             big.NewInt(100),
						Metadata:           []byte("metadata"),
					},
					ClaimData: &agglayer.ClaimFromMainnnet{
						ProofLeafMER: &agglayer.MerkleProof{
							Root:  common.HexToHash("0x1"),
							Proof: [32]common.Hash{},
						},
						ProofGERToL1Root: &agglayer.MerkleProof{
							Root:  common.HexToHash("0x3"),
							Proof: [32]common.Hash{},
						},
						L1Leaf: &agglayer.L1InfoTreeLeaf{
							L1InfoTreeIndex: 1,
							RollupExitRoot:  common.HexToHash("0x4"),
							MainnetExitRoot: common.HexToHash("0x5"),
							Inner: &agglayer.L1InfoTreeLeafInner{
								GlobalExitRoot: common.HexToHash("0x6"),
								BlockHash:      common.HexToHash("0x7"),
								Timestamp:      1231,
							},
						},
					},
				},
			},
		},
		Signature: &agglayer.Signature{
			R:         r,
			S:         s,
			OddParity: v,
		},
	}

	file, err := os.Create("test.json")
	require.NoError(t, err)

	defer file.Close()

	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")
	require.NoError(t, encoder.Encode(certificate))
}

func TestGetNextHeightAndPreviousLER(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name                           string
		lastSentCertificateInfo        *aggsendertypes.CertificateInfo
		lastSettleCertificateInfoCall  bool
		lastSettleCertificateInfo      *aggsendertypes.CertificateInfo
		lastSettleCertificateInfoError error
		expectedHeight                 uint64
		expectedPreviousLER            common.Hash
		expectedError                  bool
	}{
		{
			name: "Normal case",
			lastSentCertificateInfo: &aggsendertypes.CertificateInfo{
				Height:           10,
				NewLocalExitRoot: common.HexToHash("0x123"),
				Status:           agglayer.Settled,
			},
			expectedHeight:      11,
			expectedPreviousLER: common.HexToHash("0x123"),
		},
		{
			name:                    "First certificate",
			lastSentCertificateInfo: nil,
			expectedHeight:          0,
			expectedPreviousLER:     zeroLER,
		},
		{
			name: "First certificate error, with prevLER",
			lastSentCertificateInfo: &aggsendertypes.CertificateInfo{
				Height:                0,
				NewLocalExitRoot:      common.HexToHash("0x123"),
				Status:                agglayer.InError,
				PreviousLocalExitRoot: &ler1,
			},
			expectedHeight:      0,
			expectedPreviousLER: ler1,
		},
		{
			name: "First certificate error, no prevLER",
			lastSentCertificateInfo: &aggsendertypes.CertificateInfo{
				Height:           0,
				NewLocalExitRoot: common.HexToHash("0x123"),
				Status:           agglayer.InError,
			},
			expectedHeight:      0,
			expectedPreviousLER: zeroLER,
		},
		{
			name: "n certificate error, prevLER",
			lastSentCertificateInfo: &aggsendertypes.CertificateInfo{
				Height:                10,
				NewLocalExitRoot:      common.HexToHash("0x123"),
				PreviousLocalExitRoot: &ler1,
				Status:                agglayer.InError,
			},
			expectedHeight:      10,
			expectedPreviousLER: ler1,
		},
		{
			name: "last cert not closed, error",
			lastSentCertificateInfo: &aggsendertypes.CertificateInfo{
				Height:                10,
				NewLocalExitRoot:      common.HexToHash("0x123"),
				PreviousLocalExitRoot: &ler1,
				Status:                agglayer.Pending,
			},
			expectedHeight:      10,
			expectedPreviousLER: ler1,
			expectedError:       true,
		},
		{
			name: "Previous certificate in error, no prevLER",
			lastSentCertificateInfo: &aggsendertypes.CertificateInfo{
				Height:           10,
				NewLocalExitRoot: common.HexToHash("0x123"),
				Status:           agglayer.InError,
			},
			lastSettleCertificateInfo: &aggsendertypes.CertificateInfo{
				Height:           9,
				NewLocalExitRoot: common.HexToHash("0x3456"),
				Status:           agglayer.Settled,
			},
			expectedHeight:      10,
			expectedPreviousLER: common.HexToHash("0x3456"),
		},
		{
			name: "Previous certificate in error, no prevLER. Error getting previous cert",
			lastSentCertificateInfo: &aggsendertypes.CertificateInfo{
				Height:           10,
				NewLocalExitRoot: common.HexToHash("0x123"),
				Status:           agglayer.InError,
			},
			lastSettleCertificateInfo:      nil,
			lastSettleCertificateInfoError: errors.New("error getting last settle certificate"),
			expectedError:                  true,
		},
		{
			name: "Previous certificate in error, no prevLER. prev cert not available on storage",
			lastSentCertificateInfo: &aggsendertypes.CertificateInfo{
				Height:           10,
				NewLocalExitRoot: common.HexToHash("0x123"),
				Status:           agglayer.InError,
			},
			lastSettleCertificateInfoCall:  true,
			lastSettleCertificateInfo:      nil,
			lastSettleCertificateInfoError: nil,
			expectedError:                  true,
		},
		{
			name: "Previous certificate in error, no prevLER. prev cert not available on storage",
			lastSentCertificateInfo: &aggsendertypes.CertificateInfo{
				Height:           10,
				NewLocalExitRoot: common.HexToHash("0x123"),
				Status:           agglayer.InError,
			},
			lastSettleCertificateInfo: &aggsendertypes.CertificateInfo{
				Height:           9,
				NewLocalExitRoot: common.HexToHash("0x3456"),
				Status:           agglayer.InError,
			},
			lastSettleCertificateInfoError: nil,
			expectedError:                  true,
		},
	}

	for _, tt := range tests {
		tt := tt

		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			storageMock := mocks.NewAggSenderStorage(t)
			aggSender := &AggSender{log: log.WithFields("aggsender-test", "getNextHeightAndPreviousLER"), storage: storageMock}
			if tt.lastSettleCertificateInfoCall || tt.lastSettleCertificateInfo != nil || tt.lastSettleCertificateInfoError != nil {
				storageMock.EXPECT().GetCertificateByHeight(mock.Anything).Return(tt.lastSettleCertificateInfo, tt.lastSettleCertificateInfoError).Once()
			}
			height, previousLER, err := aggSender.getNextHeightAndPreviousLER(tt.lastSentCertificateInfo)
			if tt.expectedError {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				require.Equal(t, tt.expectedHeight, height)
				require.Equal(t, tt.expectedPreviousLER, previousLER)
			}
		})
	}
}

func TestSendCertificate_NoClaims(t *testing.T) {
	privateKey, err := crypto.GenerateKey()
	require.NoError(t, err)

	ctx := context.Background()
	mockStorage := mocks.NewAggSenderStorage(t)
	mockL2Syncer := mocks.NewL2BridgeSyncer(t)
	mockAggLayerClient := agglayer.NewAgglayerClientMock(t)
	mockL1InfoTreeSyncer := mocks.NewL1InfoTreeSyncer(t)

	aggSender := &AggSender{
		log:              log.WithFields("aggsender-test", "no claims test"),
		storage:          mockStorage,
		l2Syncer:         mockL2Syncer,
		aggLayerClient:   mockAggLayerClient,
		l1infoTreeSyncer: mockL1InfoTreeSyncer,
		sequencerKey:     privateKey,
		cfg:              Config{},
	}

	mockStorage.On("GetCertificatesByStatus", agglayer.NonSettledStatuses).Return([]*aggsendertypes.CertificateInfo{}, nil).Once()
	mockStorage.On("GetLastSentCertificate").Return(&aggsendertypes.CertificateInfo{
		NewLocalExitRoot: common.HexToHash("0x123"),
		Height:           1,
		FromBlock:        0,
		ToBlock:          10,
		Status:           agglayer.Settled,
	}, nil).Once()
	mockStorage.On("SaveLastSentCertificate", mock.Anything, mock.Anything).Return(nil).Once()
	mockL2Syncer.On("GetLastProcessedBlock", mock.Anything).Return(uint64(50), nil)
	mockL2Syncer.On("GetBridgesPublished", mock.Anything, uint64(11), uint64(50)).Return([]bridgesync.Bridge{
		{
			BlockNum:           30,
			BlockPos:           0,
			LeafType:           agglayer.LeafTypeAsset.Uint8(),
			OriginNetwork:      1,
			OriginAddress:      common.HexToAddress("0x1"),
			DestinationNetwork: 2,
			DestinationAddress: common.HexToAddress("0x2"),
			Amount:             big.NewInt(100),
			Metadata:           []byte("metadata"),
			DepositCount:       1,
		},
	}, nil)
	mockL2Syncer.On("GetClaims", mock.Anything, uint64(11), uint64(50)).Return([]bridgesync.Claim{}, nil)
	mockL2Syncer.On("GetExitRootByIndex", mock.Anything, uint32(1)).Return(treeTypes.Root{}, nil).Once()
	mockL2Syncer.On("OriginNetwork").Return(uint32(1), nil).Once()
	mockAggLayerClient.On("SendCertificate", mock.Anything).Return(common.Hash{}, nil).Once()

	signedCertificate, err := aggSender.sendCertificate(ctx)
	require.NoError(t, err)
	require.NotNil(t, signedCertificate)
	require.NotNil(t, signedCertificate.Signature)
	require.NotNil(t, signedCertificate.Certificate)
	require.NotNil(t, signedCertificate.Certificate.ImportedBridgeExits)
	require.Len(t, signedCertificate.Certificate.BridgeExits, 1)

	mockStorage.AssertExpectations(t)
	mockL2Syncer.AssertExpectations(t)
	mockAggLayerClient.AssertExpectations(t)
	mockL1InfoTreeSyncer.AssertExpectations(t)
}

func TestExtractFromCertificateMetadataToBlock(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		metadata common.Hash
		expected aggsendertypes.CertificateMetadata
	}{
		{
			name:     "Valid metadata",
			metadata: aggsendertypes.NewCertificateMetadata(0, 1000, 123567890).ToHash(),
			expected: aggsendertypes.CertificateMetadata{
				Version:   1,
				FromBlock: 0,
				Offset:    1000,
				CreatedAt: 123567890,
			},
		},
		{
			name:     "Zero metadata",
			metadata: aggsendertypes.NewCertificateMetadata(0, 0, 0).ToHash(),
			expected: aggsendertypes.CertificateMetadata{
				Version:   1,
				FromBlock: 0,
				Offset:    0,
				CreatedAt: 0,
			},
		},
	}

	for _, tt := range tests {
		tt := tt

		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			result := *aggsendertypes.NewCertificateMetadataFromHash(tt.metadata)
			require.Equal(t, tt.expected, result)
		})
	}
}

func TestCheckLastCertificateFromAgglayer_ErrorAggLayer(t *testing.T) {
	testData := newAggsenderTestData(t, testDataFlagMockStorage)
	testData.l2syncerMock.EXPECT().OriginNetwork().Return(networkIDTest).Once()
	testData.agglayerClientMock.EXPECT().GetLatestKnownCertificateHeader(networkIDTest).Return(nil, fmt.Errorf("unittest error")).Once()

	err := testData.sut.checkLastCertificateFromAgglayer(testData.ctx)

	require.Error(t, err)
}

func TestCheckLastCertificateFromAgglayer_ErrorStorageGetLastSentCertificate(t *testing.T) {
	testData := newAggsenderTestData(t, testDataFlagMockStorage)
	testData.l2syncerMock.EXPECT().OriginNetwork().Return(networkIDTest).Once()
	testData.agglayerClientMock.EXPECT().GetLatestKnownCertificateHeader(networkIDTest).Return(nil, nil).Once()
	testData.storageMock.EXPECT().GetLastSentCertificate().Return(nil, fmt.Errorf("unittest error"))

	err := testData.sut.checkLastCertificateFromAgglayer(testData.ctx)

	require.Error(t, err)
}

// TestCheckLastCertificateFromAgglayer_Case1NoCerts
// CASE 1: No certificates in local storage and agglayer
// Aggsender and agglayer are empty so it's ok
func TestCheckLastCertificateFromAgglayer_Case1NoCerts(t *testing.T) {
	testData := newAggsenderTestData(t, testDataFlagNone)
	testData.l2syncerMock.EXPECT().OriginNetwork().Return(networkIDTest).Once()
	testData.agglayerClientMock.EXPECT().GetLatestKnownCertificateHeader(networkIDTest).Return(nil, nil).Once()

	err := testData.sut.checkLastCertificateFromAgglayer(testData.ctx)

	require.NoError(t, err)
}

// TestCheckLastCertificateFromAgglayer_Case2NoCertLocalCertRemote
// CASE 2: No certificates in local storage but agglayer has one
// The local DB is empty and we set the lastCert reported by AggLayer
func TestCheckLastCertificateFromAgglayer_Case2NoCertLocalCertRemote(t *testing.T) {
	testData := newAggsenderTestData(t, testDataFlagNone)
	testData.l2syncerMock.EXPECT().OriginNetwork().Return(networkIDTest).Once()
	testData.agglayerClientMock.EXPECT().GetLatestKnownCertificateHeader(networkIDTest).
		Return(certInfoToCertHeader(t, &testData.testCerts[0], networkIDTest), nil).Once()

	err := testData.sut.checkLastCertificateFromAgglayer(testData.ctx)

	require.NoError(t, err)
	localCert, err := testData.sut.storage.GetLastSentCertificate()
	require.NoError(t, err)
	require.Equal(t, testData.testCerts[0].CertificateID, localCert.CertificateID)
}

// TestCheckLastCertificateFromAgglayer_Case2NoCertLocalCertRemoteErrorStorage
// sub case of previous one that fails to update local storage
func TestCheckLastCertificateFromAgglayer_Case2NoCertLocalCertRemoteErrorStorage(t *testing.T) {
	testData := newAggsenderTestData(t, testDataFlagMockStorage)
	testData.l2syncerMock.EXPECT().OriginNetwork().Return(networkIDTest).Once()
	testData.agglayerClientMock.EXPECT().GetLatestKnownCertificateHeader(networkIDTest).
		Return(certInfoToCertHeader(t, &testData.testCerts[0], networkIDTest), nil).Once()
	testData.storageMock.EXPECT().GetLastSentCertificate().Return(nil, nil)
	testData.storageMock.EXPECT().SaveLastSentCertificate(mock.Anything, mock.Anything).Return(errTest).Once()
	err := testData.sut.checkLastCertificateFromAgglayer(testData.ctx)

	require.Error(t, err)
}

// CASE 2.1: certificate in storage but not in agglayer
// sub case of previous one that fails to update local storage
func TestCheckLastCertificateFromAgglayer_Case2_1NoCertRemoteButCertLocal(t *testing.T) {
	testData := newAggsenderTestData(t, testDataFlagMockStorage)
	testData.l2syncerMock.EXPECT().OriginNetwork().Return(networkIDTest).Once()
	testData.agglayerClientMock.EXPECT().GetLatestKnownCertificateHeader(networkIDTest).
		Return(nil, nil).Once()
	testData.storageMock.EXPECT().GetLastSentCertificate().Return(&testData.testCerts[0], nil)
	err := testData.sut.checkLastCertificateFromAgglayer(testData.ctx)

	require.Error(t, err)
}

// CASE 3.1: the certificate on the agglayer has less height than the one stored in the local storage
func TestCheckLastCertificateFromAgglayer_Case3_1LessHeight(t *testing.T) {
	testData := newAggsenderTestData(t, testDataFlagMockStorage)
	testData.l2syncerMock.EXPECT().OriginNetwork().Return(networkIDTest).Once()
	testData.agglayerClientMock.EXPECT().GetLatestKnownCertificateHeader(networkIDTest).
		Return(certInfoToCertHeader(t, &testData.testCerts[0], networkIDTest), nil).Once()
	testData.storageMock.EXPECT().GetLastSentCertificate().Return(&testData.testCerts[1], nil)

	err := testData.sut.checkLastCertificateFromAgglayer(testData.ctx)

	require.ErrorContains(t, err, "recovery: the last certificate in the agglayer has less height (1) than the one in the local storage (2)")
}

// CASE 3.2: AggSender and AggLayer not same height. AggLayer has a new certificate
func TestCheckLastCertificateFromAgglayer_Case3_2Mismatch(t *testing.T) {
	testData := newAggsenderTestData(t, testDataFlagMockStorage)
	testData.l2syncerMock.EXPECT().OriginNetwork().Return(networkIDTest).Once()
	testData.agglayerClientMock.EXPECT().GetLatestKnownCertificateHeader(networkIDTest).
		Return(certInfoToCertHeader(t, &testData.testCerts[1], networkIDTest), nil).Once()
	testData.storageMock.EXPECT().GetLastSentCertificate().Return(&testData.testCerts[0], nil)
	testData.storageMock.EXPECT().SaveLastSentCertificate(mock.Anything, mock.Anything).Return(nil).Once()

	err := testData.sut.checkLastCertificateFromAgglayer(testData.ctx)

	require.NoError(t, err)
}

// CASE 4: AggSender and AggLayer not same certificateID
func TestCheckLastCertificateFromAgglayer_Case4Mismatch(t *testing.T) {
	testData := newAggsenderTestData(t, testDataFlagMockStorage)
	testData.l2syncerMock.EXPECT().OriginNetwork().Return(networkIDTest).Once()
	testData.agglayerClientMock.EXPECT().GetLatestKnownCertificateHeader(networkIDTest).
		Return(certInfoToCertHeader(t, &testData.testCerts[0], networkIDTest), nil).Once()
	testData.storageMock.EXPECT().GetLastSentCertificate().Return(&testData.testCerts[1], nil)

	err := testData.sut.checkLastCertificateFromAgglayer(testData.ctx)

	require.Error(t, err)
}

// CASE 5: AggSender and AggLayer same certificateID and same status
func TestCheckLastCertificateFromAgglayer_Case5SameStatus(t *testing.T) {
	testData := newAggsenderTestData(t, testDataFlagMockStorage)
	testData.l2syncerMock.EXPECT().OriginNetwork().Return(networkIDTest).Once()
	testData.agglayerClientMock.EXPECT().GetLatestKnownCertificateHeader(networkIDTest).
		Return(certInfoToCertHeader(t, &testData.testCerts[0], networkIDTest), nil).Once()
	testData.storageMock.EXPECT().GetLastSentCertificate().Return(&testData.testCerts[0], nil)

	err := testData.sut.checkLastCertificateFromAgglayer(testData.ctx)

	require.NoError(t, err)
}

// CASE 5: AggSender and AggLayer same certificateID and differ on status
func TestCheckLastCertificateFromAgglayer_Case5UpdateStatus(t *testing.T) {
	testData := newAggsenderTestData(t, testDataFlagMockStorage)
	testData.l2syncerMock.EXPECT().OriginNetwork().Return(networkIDTest).Once()
	aggLayerCert := certInfoToCertHeader(t, &testData.testCerts[0], networkIDTest)
	aggLayerCert.Status = agglayer.Settled
	testData.agglayerClientMock.EXPECT().GetLatestKnownCertificateHeader(networkIDTest).
		Return(aggLayerCert, nil).Once()
	testData.storageMock.EXPECT().GetLastSentCertificate().Return(&testData.testCerts[0], nil)
	testData.storageMock.EXPECT().UpdateCertificate(mock.Anything, mock.Anything).Return(nil).Once()

	err := testData.sut.checkLastCertificateFromAgglayer(testData.ctx)

	require.NoError(t, err)
}

// CASE 4: AggSender and AggLayer same certificateID and differ on status but fails update
func TestCheckLastCertificateFromAgglayer_Case4ErrorUpdateStatus(t *testing.T) {
	testData := newAggsenderTestData(t, testDataFlagMockStorage)
	testData.l2syncerMock.EXPECT().OriginNetwork().Return(networkIDTest).Once()
	aggLayerCert := certInfoToCertHeader(t, &testData.testCerts[0], networkIDTest)
	aggLayerCert.Status = agglayer.Settled
	testData.agglayerClientMock.EXPECT().GetLatestKnownCertificateHeader(networkIDTest).
		Return(aggLayerCert, nil).Once()
	testData.storageMock.EXPECT().GetLastSentCertificate().Return(&testData.testCerts[0], nil)
	testData.storageMock.EXPECT().UpdateCertificate(mock.Anything, mock.Anything).Return(errTest).Once()

	err := testData.sut.checkLastCertificateFromAgglayer(testData.ctx)

	require.Error(t, err)
}

func TestLimitSize_FirstOneFit(t *testing.T) {
	testData := newAggsenderTestData(t, testDataFlagMockStorage)
	certParams := &aggsendertypes.CertificateBuildParams{
		FromBlock: uint64(1),
		ToBlock:   uint64(20),
		Bridges:   NewBridgesData(t, 1, []uint64{1}),
	}
	newCert, err := testData.sut.limitCertSize(certParams)
	require.NoError(t, err)
	require.Equal(t, certParams, newCert)
}

func TestLimitSize_FirstMinusOneFit(t *testing.T) {
	testData := newAggsenderTestData(t, testDataFlagMockStorage)
	testData.sut.cfg.MaxCertSize = (aggsendertypes.EstimatedSizeBridgeExit * 3) + 1
	certParams := &aggsendertypes.CertificateBuildParams{
		FromBlock: uint64(1),
		ToBlock:   uint64(20),
		Bridges:   NewBridgesData(t, 0, []uint64{19, 19, 19, 20}),
	}
	newCert, err := testData.sut.limitCertSize(certParams)
	require.NoError(t, err)
	require.Equal(t, uint64(19), newCert.ToBlock)
}

func TestLimitSize_NoWayToFitInMaxSize(t *testing.T) {
	testData := newAggsenderTestData(t, testDataFlagMockStorage)
	testData.sut.cfg.MaxCertSize = (aggsendertypes.EstimatedSizeBridgeExit * 2) + 1
	certParams := &aggsendertypes.CertificateBuildParams{
		FromBlock: uint64(1),
		ToBlock:   uint64(20),
		Bridges:   NewBridgesData(t, 0, []uint64{19, 19, 19, 20}),
	}
	newCert, err := testData.sut.limitCertSize(certParams)
	require.NoError(t, err)
	require.Equal(t, uint64(19), newCert.ToBlock)
}

func TestLimitSize_MinNumBlocks(t *testing.T) {
	testData := newAggsenderTestData(t, testDataFlagMockStorage)
	testData.sut.cfg.MaxCertSize = (aggsendertypes.EstimatedSizeBridgeExit * 2) + 1
	certParams := &aggsendertypes.CertificateBuildParams{
		FromBlock: uint64(1),
		ToBlock:   uint64(2),
		Bridges:   NewBridgesData(t, 0, []uint64{1, 1, 1, 2, 2, 2}),
	}
	newCert, err := testData.sut.limitCertSize(certParams)
	require.NoError(t, err)
	require.Equal(t, uint64(1), newCert.ToBlock)
}

func TestGetLastSentBlockAndRetryCount(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name                    string
		lastSentCertificateInfo *aggsendertypes.CertificateInfo
		expectedBlock           uint64
		expectedRetryCount      int
	}{
		{
			name:                    "No last sent certificate",
			lastSentCertificateInfo: nil,
			expectedBlock:           0,
			expectedRetryCount:      0,
		},
		{
			name: "Last sent certificate with no error",
			lastSentCertificateInfo: &aggsendertypes.CertificateInfo{
				ToBlock: 10,
				Status:  agglayer.Settled,
			},
			expectedBlock:      10,
			expectedRetryCount: 0,
		},
		{
			name: "Last sent certificate with error and non-zero FromBlock",
			lastSentCertificateInfo: &aggsendertypes.CertificateInfo{
				FromBlock:  5,
				ToBlock:    10,
				Status:     agglayer.InError,
				RetryCount: 1,
			},
			expectedBlock:      4,
			expectedRetryCount: 2,
		},
		{
			name: "Last sent certificate with error and zero FromBlock",
			lastSentCertificateInfo: &aggsendertypes.CertificateInfo{
				FromBlock:  0,
				ToBlock:    10,
				Status:     agglayer.InError,
				RetryCount: 1,
			},
			expectedBlock:      10,
			expectedRetryCount: 2,
		},
	}

	for _, tt := range tests {
		tt := tt

		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			block, retryCount := getLastSentBlockAndRetryCount(tt.lastSentCertificateInfo)

			require.Equal(t, tt.expectedBlock, block)
			require.Equal(t, tt.expectedRetryCount, retryCount)
		})
	}
}

type testDataFlags = int

const (
	testDataFlagNone        testDataFlags = 0
	testDataFlagMockStorage testDataFlags = 1
)

type aggsenderTestData struct {
	ctx                  context.Context
	agglayerClientMock   *agglayer.AgglayerClientMock
	l2syncerMock         *mocks.L2BridgeSyncer
	l1InfoTreeSyncerMock *mocks.L1InfoTreeSyncer
	storageMock          *mocks.AggSenderStorage
	sut                  *AggSender
	testCerts            []aggsendertypes.CertificateInfo
}

func NewBridgesData(t *testing.T, num int, blockNum []uint64) []bridgesync.Bridge {
	t.Helper()
	if num == 0 {
		num = len(blockNum)
	}
	res := make([]bridgesync.Bridge, 0)
	for i := 0; i < num; i++ {
		res = append(res, bridgesync.Bridge{
			BlockNum:      blockNum[i%len(blockNum)],
			BlockPos:      0,
			LeafType:      agglayer.LeafTypeAsset.Uint8(),
			OriginNetwork: 1,
		})
	}
	return res
}

func NewClaimData(t *testing.T, num int, blockNum []uint64) []bridgesync.Claim {
	t.Helper()
	if num == 0 {
		num = len(blockNum)
	}
	res := make([]bridgesync.Claim, 0)
	for i := 0; i < num; i++ {
		res = append(res, bridgesync.Claim{
			BlockNum: blockNum[i%len(blockNum)],
			BlockPos: 0,
		})
	}
	return res
}

func certInfoToCertHeader(t *testing.T, certInfo *aggsendertypes.CertificateInfo, networkID uint32) *agglayer.CertificateHeader {
	t.Helper()
	if certInfo == nil {
		return nil
	}
	return &agglayer.CertificateHeader{
		Height:           certInfo.Height,
		NetworkID:        networkID,
		CertificateID:    certInfo.CertificateID,
		NewLocalExitRoot: certInfo.NewLocalExitRoot,
		Status:           agglayer.Pending,
		Metadata: aggsendertypes.NewCertificateMetadata(
			certInfo.FromBlock,
			uint32(certInfo.FromBlock-certInfo.ToBlock),
			certInfo.CreatedAt,
		).ToHash(),
	}
}

func newAggsenderTestData(t *testing.T, creationFlags testDataFlags) *aggsenderTestData {
	t.Helper()
	l2syncerMock := mocks.NewL2BridgeSyncer(t)
	agglayerClientMock := agglayer.NewAgglayerClientMock(t)
	l1InfoTreeSyncerMock := mocks.NewL1InfoTreeSyncer(t)
	logger := log.WithFields("aggsender-test", "checkLastCertificateFromAgglayer")
	var storageMock *mocks.AggSenderStorage
	var storage db.AggSenderStorage
	var err error
	if creationFlags&testDataFlagMockStorage != 0 {
		storageMock = mocks.NewAggSenderStorage(t)
		storage = storageMock
	} else {
		pc, _, _, _ := runtime.Caller(1)
		part := runtime.FuncForPC(pc)
		dbPath := fmt.Sprintf("file:%d?mode=memory&cache=shared", part.Entry())
		storageConfig := db.AggSenderSQLStorageConfig{
			DBPath:                  dbPath,
			KeepCertificatesHistory: true,
		}
		storage, err = db.NewAggSenderSQLStorage(logger, storageConfig)
		require.NoError(t, err)
	}

	ctx := context.TODO()
	sut := &AggSender{
		log:              logger,
		l2Syncer:         l2syncerMock,
		aggLayerClient:   agglayerClientMock,
		storage:          storage,
		l1infoTreeSyncer: l1InfoTreeSyncerMock,
		cfg: Config{
			MaxCertSize: 1024 * 1024,
		},
	}
	testCerts := []aggsendertypes.CertificateInfo{
		{
			Height:           1,
			CertificateID:    common.HexToHash("0x1"),
			NewLocalExitRoot: common.HexToHash("0x2"),
			Status:           agglayer.Pending,
		},
		{
			Height:           2,
			CertificateID:    common.HexToHash("0x1a111"),
			NewLocalExitRoot: common.HexToHash("0x2a2"),
			Status:           agglayer.Pending,
		},
	}

	return &aggsenderTestData{
		ctx:                  ctx,
		agglayerClientMock:   agglayerClientMock,
		l2syncerMock:         l2syncerMock,
		l1InfoTreeSyncerMock: l1InfoTreeSyncerMock,
		storageMock:          storageMock,
		sut:                  sut,
		testCerts:            testCerts,
	}
}
