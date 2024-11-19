package aggsender

import (
	"context"
	"crypto/ecdsa"
	"encoding/json"
	"errors"
	"fmt"
	"math/big"
	"os"
	"testing"
	"time"

	"github.com/0xPolygon/cdk/agglayer"
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

func TestExploratoryGetCertificateHeader(t *testing.T) {
	t.Skip("This test is exploratory and should be skipped")
	aggLayerClient := agglayer.NewAggLayerClient("http://localhost:32796")
	certificateID := common.HexToHash("0xf153e75e24591432ac5deafaeaafba3fec0fd851261c86051b9c0d540b38c369")
	certificateHeader, err := aggLayerClient.GetCertificateHeader(certificateID)
	require.NoError(t, err)
	fmt.Print(certificateHeader)
}
func TestExploratoryGetEpochConfiguration(t *testing.T) {
	t.Skip("This test is exploratory and should be skipped")
	aggLayerClient := agglayer.NewAggLayerClient("http://localhost:32796")
	clockConfig, err := aggLayerClient.GetEpochConfiguration()
	require.NoError(t, err)
	fmt.Print(clockConfig)
}

func TestConfigString(t *testing.T) {
	config := Config{
		StoragePath:                 "/path/to/storage",
		AggLayerURL:                 "http://agglayer.url",
		BlockGetInterval:            types.Duration{Duration: 10 * time.Second},
		CheckSettledInterval:        types.Duration{Duration: 20 * time.Second},
		AggsenderPrivateKey:         types.KeystoreFileConfig{Path: "/path/to/key", Password: "password"},
		URLRPCL2:                    "http://l2.rpc.url",
		BlockFinality:               "latestBlock",
		EpochNotificationPercentage: 50,
		SaveCertificatesToFilesPath: "/path/to/certificates",
	}

	expected := "StoragePath: /path/to/storage\n" +
		"AggLayerURL: http://agglayer.url\n" +
		"BlockGetInterval: 10s\n" +
		"CheckSettledInterval: 20s\n" +
		"AggsenderPrivateKeyPath: /path/to/key\n" +
		"AggsenderPrivateKeyPassword: password\n" +
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
	ch := make(chan aggsendertypes.EpochEvent)
	epochNotifierMock.EXPECT().Subscribe("aggsender").Return(ch)
	bridgeL2SyncerMock.EXPECT().GetLastProcessedBlock(mock.Anything).Return(uint64(0), nil)

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
							Root:  common.HexToHash("0xbbba"),
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
							Root:  common.HexToHash("0xbbb"),
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
			},
			toBlock: 10,
			expectedCert: &agglayer.Certificate{
				NetworkID:         1,
				PrevLocalExitRoot: common.HexToHash("0x123"),
				NewLocalExitRoot:  common.HexToHash("0x789"),
				Metadata:          createCertificateMetadata(10),
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
								Root:  common.HexToHash("0xbbba"),
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
			cert, err := aggSender.buildCertificate(context.Background(), tt.bridges, tt.claims, tt.lastSentCertificateInfo, tt.toBlock)

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

			mockStorage.On("GetCertificatesByStatus", nonSettledStatuses).Return(
				tt.pendingCertificates, tt.getFromDBError)
			for certID, header := range tt.certificateHeaders {
				mockAggLayerClient.On("GetCertificateHeader", certID).Return(header, tt.clientError)
			}
			if tt.updateDBError != nil {
				mockStorage.On("UpdateCertificateStatus", mock.Anything, mock.Anything).Return(tt.updateDBError)
			} else if tt.clientError == nil && tt.getFromDBError == nil {
				mockStorage.On("UpdateCertificateStatus", mock.Anything, mock.Anything).Return(nil)
			}

			aggSender := &AggSender{
				log:            mockLogger,
				storage:        mockStorage,
				aggLayerClient: mockAggLayerClient,
				cfg: Config{
					BlockGetInterval:     types.Duration{Duration: time.Second},
					CheckSettledInterval: types.Duration{Duration: time.Second},
				},
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
				cfg:          Config{},
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
			mockStorage.On("GetCertificatesByStatus", nonSettledStatuses).
				Return(cfg.shouldSendCertificate...).Once()

			aggsender.storage = mockStorage

			if cfg.getLastSentCertificate != nil {
				mockStorage.On("GetLastSentCertificate").Return(cfg.getLastSentCertificate...).Once()
			}

			if cfg.saveLastSentCertificate != nil {
				mockStorage.On("SaveLastSentCertificate", mock.Anything, mock.Anything).Return(cfg.saveLastSentCertificate...).Once()
			}
		}

		if cfg.lastL2BlockProcessed != nil || cfg.originNetwork != nil ||
			cfg.getBridges != nil || cfg.getClaims != nil || cfg.getInfoByGlobalExitRoot != nil {
			mockL2Syncer = mocks.NewL2BridgeSyncer(t)

			mockL2Syncer.On("GetLastProcessedBlock", mock.Anything).Return(cfg.lastL2BlockProcessed...).Once()

			if cfg.getBridges != nil {
				mockL2Syncer.On("GetBridgesPublished", mock.Anything, mock.Anything, mock.Anything).Return(cfg.getBridges...).Once()
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
			getLastSentCertificate: []interface{}{aggsendertypes.CertificateInfo{}, errors.New("error getting last sent certificate")},
			expectedError:          "error getting last sent certificate",
		},
		{
			name:                  "no new blocks to send certificate",
			shouldSendCertificate: []interface{}{[]*aggsendertypes.CertificateInfo{}, nil},
			lastL2BlockProcessed:  []interface{}{uint64(41), nil},
			getLastSentCertificate: []interface{}{aggsendertypes.CertificateInfo{
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
			getLastSentCertificate: []interface{}{aggsendertypes.CertificateInfo{
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
			getLastSentCertificate: []interface{}{aggsendertypes.CertificateInfo{
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
			getLastSentCertificate: []interface{}{aggsendertypes.CertificateInfo{
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
			getLastSentCertificate: []interface{}{aggsendertypes.CertificateInfo{
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
			getLastSentCertificate: []interface{}{aggsendertypes.CertificateInfo{
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
			getLastSentCertificate: []interface{}{aggsendertypes.CertificateInfo{
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
			getLastSentCertificate: []interface{}{aggsendertypes.CertificateInfo{
				Height:           90,
				CertificateID:    common.HexToHash("0x1121111"),
				NewLocalExitRoot: common.HexToHash("0x111122211"),
				FromBlock:        80,
				ToBlock:          81,
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
			getLastSentCertificate: []interface{}{aggsendertypes.CertificateInfo{
				Height:           100,
				CertificateID:    common.HexToHash("0x11121111"),
				NewLocalExitRoot: common.HexToHash("0x1211122211"),
				FromBlock:        90,
				ToBlock:          91,
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
			getLastSentCertificate: []interface{}{aggsendertypes.CertificateInfo{
				Height:           110,
				CertificateID:    common.HexToHash("0x12121111"),
				NewLocalExitRoot: common.HexToHash("0x1221122211"),
				FromBlock:        100,
				ToBlock:          101,
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
		name                    string
		lastSentCertificateInfo aggsendertypes.CertificateInfo
		expectedHeight          uint64
		expectedPreviousLER     common.Hash
	}{
		{
			name: "Normal case",
			lastSentCertificateInfo: aggsendertypes.CertificateInfo{
				Height:           10,
				NewLocalExitRoot: common.HexToHash("0x123"),
				Status:           agglayer.Settled,
			},
			expectedHeight:      11,
			expectedPreviousLER: common.HexToHash("0x123"),
		},
		{
			name: "Previous certificate in error",
			lastSentCertificateInfo: aggsendertypes.CertificateInfo{
				Height:           10,
				NewLocalExitRoot: common.HexToHash("0x123"),
				Status:           agglayer.InError,
			},
			expectedHeight:      10,
			expectedPreviousLER: common.HexToHash("0x123"),
		},
		{
			name: "First certificate",
			lastSentCertificateInfo: aggsendertypes.CertificateInfo{
				Height:           0,
				NewLocalExitRoot: common.Hash{},
				Status:           agglayer.Settled,
			},
			expectedHeight:      0,
			expectedPreviousLER: zeroLER,
		},
	}

	for _, tt := range tests {
		tt := tt

		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			aggSender := &AggSender{log: log.WithFields("aggsender-test", "getNextHeightAndPreviousLER")}
			height, previousLER := aggSender.getNextHeightAndPreviousLER(&tt.lastSentCertificateInfo)

			require.Equal(t, tt.expectedHeight, height)
			require.Equal(t, tt.expectedPreviousLER, previousLER)
		})
	}
}

func TestSendCertificate_NoClaims(t *testing.T) {
	t.Parallel()

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
		cfg: Config{
			BlockGetInterval:     types.Duration{Duration: time.Second},
			CheckSettledInterval: types.Duration{Duration: time.Second},
		},
	}

	mockStorage.On("GetCertificatesByStatus", nonSettledStatuses).Return([]*aggsendertypes.CertificateInfo{}, nil).Once()
	mockStorage.On("GetLastSentCertificate").Return(aggsendertypes.CertificateInfo{
		NewLocalExitRoot: common.HexToHash("0x123"),
		Height:           1,
		FromBlock:        0,
		ToBlock:          10,
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
	}, nil).Once()
	mockL2Syncer.On("GetClaims", mock.Anything, uint64(11), uint64(50)).Return([]bridgesync.Claim{}, nil).Once()
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
