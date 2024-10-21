package aggsender

import (
	"context"
	"crypto/ecdsa"
	"errors"
	"fmt"
	"math/big"
	"testing"
	"time"

	"github.com/0xPolygon/cdk/agglayer"
	"github.com/0xPolygon/cdk/aggsender/db"
	aggsendertypes "github.com/0xPolygon/cdk/aggsender/types"
	"github.com/0xPolygon/cdk/bridgesync"
	"github.com/0xPolygon/cdk/config/types"
	"github.com/0xPolygon/cdk/etherman"
	"github.com/0xPolygon/cdk/l1infotreesync"
	"github.com/0xPolygon/cdk/log"
	treeTypes "github.com/0xPolygon/cdk/tree/types"
	"github.com/ethereum/go-ethereum/common"
	gethTypes "github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

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

func TestSignCertificate(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name          string
		certificate   *agglayer.Certificate
		sequencerKey  *ecdsa.PrivateKey
		expectedError bool
	}{
		{
			name: "Valid certificate",
			certificate: &agglayer.Certificate{
				NetworkID:         1,
				PrevLocalExitRoot: common.HexToHash("0x123"),
				NewLocalExitRoot:  common.HexToHash("0x456"),
				BridgeExits: []*agglayer.BridgeExit{
					{
						LeafType: agglayer.LeafTypeAsset,
						TokenInfo: &agglayer.TokenInfo{
							OriginNetwork:      1,
							OriginTokenAddress: common.HexToAddress("0x789"),
						},
						DestinationNetwork: 2,
						DestinationAddress: common.HexToAddress("0xabc"),
						Amount:             big.NewInt(100),
						Metadata:           []byte("metadata"),
					},
				},
				ImportedBridgeExits: []*agglayer.ImportedBridgeExit{},
				Height:              1,
			},
			sequencerKey: func() *ecdsa.PrivateKey {
				key, _ := crypto.GenerateKey()
				return key
			}(),
			expectedError: false,
		},
		{
			name: "Invalid certificate",
			certificate: &agglayer.Certificate{
				NetworkID:         1,
				PrevLocalExitRoot: common.HexToHash("0x123"),
				NewLocalExitRoot:  common.HexToHash("0x456"),
				BridgeExits:       []*agglayer.BridgeExit{},
				ImportedBridgeExits: []*agglayer.ImportedBridgeExit{
					{
						BridgeExit: &agglayer.BridgeExit{
							LeafType: agglayer.LeafTypeAsset,
							TokenInfo: &agglayer.TokenInfo{
								OriginNetwork:      1,
								OriginTokenAddress: common.HexToAddress("0x789"),
							},
							DestinationNetwork: 2,
							DestinationAddress: common.HexToAddress("0xabc"),
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
				Height: 1,
			},
			sequencerKey: func() *ecdsa.PrivateKey {
				key, _ := crypto.GenerateKey()
				return key
			}(),
			expectedError: false,
		},
	}

	for _, tt := range tests {
		tt := tt

		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			aggSender := &AggSender{
				sequencerKey: tt.sequencerKey,
			}
			signedCert, err := aggSender.signCertificate(tt.certificate)

			if tt.expectedError {
				require.Error(t, err)
				require.Nil(t, signedCert)
			} else {
				require.NoError(t, err)
				require.NotNil(t, signedCert)
				require.Equal(t, tt.certificate, signedCert.Certificate)
				require.NotEmpty(t, signedCert.Signature)
			}
		})
	}
}

//nolint:dupl
func TestGetImportedBridgeExits(t *testing.T) {
	t.Parallel()

	mockProof := generateTestProof(t)

	mockL1InfoTreeSyncer := NewL1InfoTreeSyncerMock(t)
	mockL1InfoTreeSyncer.On("GetInfoByGlobalExitRoot", mock.Anything).Return(&l1infotreesync.L1InfoTreeLeaf{
		L1InfoTreeIndex:   1,
		Timestamp:         123456789,
		PreviousBlockHash: common.HexToHash("0xabc"),
	}, nil)
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
					IsMessage:          false,
					OriginNetwork:      1,
					OriginAddress:      common.HexToAddress("0x1234"),
					DestinationNetwork: 2,
					DestinationAddress: common.HexToAddress("0x4567"),
					Amount:             big.NewInt(111),
					Metadata:           []byte("metadata1"),
					GlobalIndex:        big.NewInt(1),
					GlobalExitRoot:     common.HexToHash("0x7891"),
					RollupExitRoot:     common.HexToHash("0xaaab"),
					MainnetExitRoot:    common.HexToHash("0xbbba"),
					ProofLocalExitRoot: mockProof,
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
						RollupIndex: 0,
						LeafIndex:   1,
					},
					ClaimData: &agglayer.ClaimFromRollup{
						L1Leaf: agglayer.L1InfoTreeLeaf{
							L1InfoTreeIndex: 1,
							RollupExitRoot:  common.HexToHash("0xaaab"),
							MainnetExitRoot: common.HexToHash("0xbbba"),
							Inner: agglayer.L1InfoTreeLeafInner{
								GlobalExitRoot: common.HexToHash("0x7891"),
								Timestamp:      123456789,
								BlockHash:      common.HexToHash("0xabc"),
							},
						},
						ProofLeafLER: agglayer.MerkleProof{
							Root:  common.HexToHash("0xbbba"),
							Proof: mockProof,
						},
						ProofLERToRER: agglayer.MerkleProof{},
						ProofGERToL1Root: agglayer.MerkleProof{
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
					IsMessage:          false,
					OriginNetwork:      1,
					OriginAddress:      common.HexToAddress("0x123"),
					DestinationNetwork: 2,
					DestinationAddress: common.HexToAddress("0x456"),
					Amount:             big.NewInt(100),
					Metadata:           []byte("metadata"),
					GlobalIndex:        big.NewInt(1),
					GlobalExitRoot:     common.HexToHash("0x789"),
					RollupExitRoot:     common.HexToHash("0xaaa"),
					MainnetExitRoot:    common.HexToHash("0xbbb"),
					ProofLocalExitRoot: mockProof,
				},
				{
					IsMessage:          true,
					OriginNetwork:      3,
					OriginAddress:      common.HexToAddress("0x789"),
					DestinationNetwork: 4,
					DestinationAddress: common.HexToAddress("0xabc"),
					Amount:             big.NewInt(200),
					Metadata:           []byte("data"),
					GlobalIndex:        bridgesync.GenerateGlobalIndex(true, 0, 2),
					GlobalExitRoot:     common.HexToHash("0xdef"),
					RollupExitRoot:     common.HexToHash("0xbbb"),
					MainnetExitRoot:    common.HexToHash("0xccc"),
					ProofLocalExitRoot: mockProof,
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
						L1Leaf: agglayer.L1InfoTreeLeaf{
							L1InfoTreeIndex: 1,
							RollupExitRoot:  common.HexToHash("0xaaa"),
							MainnetExitRoot: common.HexToHash("0xbbb"),
							Inner: agglayer.L1InfoTreeLeafInner{
								GlobalExitRoot: common.HexToHash("0x789"),
								Timestamp:      123456789,
								BlockHash:      common.HexToHash("0xabc"),
							},
						},
						ProofLeafLER: agglayer.MerkleProof{
							Root:  common.HexToHash("0xbbb"),
							Proof: mockProof,
						},
						ProofLERToRER: agglayer.MerkleProof{},
						ProofGERToL1Root: agglayer.MerkleProof{
							Root:  common.HexToHash("0x789"),
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
						L1Leaf: agglayer.L1InfoTreeLeaf{
							L1InfoTreeIndex: 2,
							RollupExitRoot:  common.HexToHash("0xbbb"),
							MainnetExitRoot: common.HexToHash("0xccc"),
							Inner: agglayer.L1InfoTreeLeafInner{
								GlobalExitRoot: common.HexToHash("0x789"),
								Timestamp:      123456789,
								BlockHash:      common.HexToHash("0xabc"),
							},
						},
						ProofLeafMER: agglayer.MerkleProof{
							Root:  common.HexToHash("0xccc"),
							Proof: mockProof,
						},
						ProofGERToL1Root: agglayer.MerkleProof{
							Root:  common.HexToHash("0x789"),
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
	mockL2BridgeSyncer := NewL2BridgeSyncerMock(t)
	mockL1InfoTreeSyncer := NewL1InfoTreeSyncerMock(t)
	mockProof := generateTestProof(t)

	tests := []struct {
		name          string
		bridges       []bridgesync.Bridge
		claims        []bridgesync.Claim
		previousExit  common.Hash
		lastHeight    uint64
		mockFn        func()
		expectedCert  *agglayer.Certificate
		expectedError bool
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
					IsMessage:          false,
					OriginNetwork:      1,
					OriginAddress:      common.HexToAddress("0x1234"),
					DestinationNetwork: 2,
					DestinationAddress: common.HexToAddress("0x4567"),
					Amount:             big.NewInt(111),
					Metadata:           []byte("metadata1"),
					GlobalIndex:        big.NewInt(1),
					GlobalExitRoot:     common.HexToHash("0x7891"),
					RollupExitRoot:     common.HexToHash("0xaaab"),
					MainnetExitRoot:    common.HexToHash("0xbbba"),
					ProofLocalExitRoot: mockProof,
				},
			},
			previousExit: common.HexToHash("0x123"),
			lastHeight:   1,
			expectedCert: &agglayer.Certificate{
				NetworkID:         1,
				PrevLocalExitRoot: common.HexToHash("0x123"),
				NewLocalExitRoot:  common.HexToHash("0x789"),
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
							L1Leaf: agglayer.L1InfoTreeLeaf{
								L1InfoTreeIndex: 1,
								RollupExitRoot:  common.HexToHash("0xaaab"),
								MainnetExitRoot: common.HexToHash("0xbbba"),
								Inner: agglayer.L1InfoTreeLeafInner{
									GlobalExitRoot: common.HexToHash("0x7891"),
									Timestamp:      123456789,
									BlockHash:      common.HexToHash("0xabc"),
								},
							},
							ProofLeafLER: agglayer.MerkleProof{
								Root:  common.HexToHash("0xbbba"),
								Proof: mockProof,
							},
							ProofLERToRER: agglayer.MerkleProof{},
							ProofGERToL1Root: agglayer.MerkleProof{
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
				}, nil)
				mockL1InfoTreeSyncer.On("GetL1InfoTreeMerkleProofFromIndexToRoot", mock.Anything, mock.Anything, mock.Anything).Return(mockProof, nil)
			},
			expectedError: false,
		},
		{
			name:          "No bridges or claims",
			bridges:       []bridgesync.Bridge{},
			claims:        []bridgesync.Claim{},
			previousExit:  common.HexToHash("0x123"),
			lastHeight:    1,
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
			previousExit: common.HexToHash("0x123"),
			lastHeight:   1,
			mockFn: func() {
				mockL1InfoTreeSyncer.On("GetInfoByGlobalExitRoot", mock.Anything).Return(&l1infotreesync.L1InfoTreeLeaf{
					L1InfoTreeIndex:   1,
					Timestamp:         123456789,
					PreviousBlockHash: common.HexToHash("0xabc"),
				}, nil)
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
			}
			cert, err := aggSender.buildCertificate(context.Background(), tt.bridges, tt.claims, tt.previousExit, tt.lastHeight)

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

func TestShouldSendCertificate(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name                   string
		block                  uint64
		epochSize              uint64
		lastL1CertificateBlock uint64
		expectedResult         bool
	}{
		{
			name:           "Should send certificate",
			block:          8,
			epochSize:      10,
			expectedResult: true,
		},
		{
			name:           "Should send certificate - another case",
			block:          9,
			epochSize:      10,
			expectedResult: true,
		},
		{
			name:                   "Should not send certificate",
			block:                  25,
			epochSize:              10,
			lastL1CertificateBlock: 18,
			expectedResult:         false,
		},
		{
			name:           "Should not send certificate at zero block",
			block:          0,
			epochSize:      1,
			expectedResult: false,
		},
		{
			name:                   "Should send certificate with large epoch size",
			block:                  1998,
			epochSize:              1000,
			lastL1CertificateBlock: 998,
			expectedResult:         true,
		},
	}

	for _, tt := range tests {
		tt := tt

		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			aggSender := &AggSender{
				cfg: Config{
					EpochSize: tt.epochSize,
				},
				lastL1CertificateBlock: tt.lastL1CertificateBlock,
			}
			result := aggSender.shouldSendCertificate(tt.block)
			require.Equal(t, tt.expectedResult, result)
		})
	}
}

func TestCheckIfCertificatesAreSettled(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name                     string
		pendingCertificates      []*aggsendertypes.CertificateInfo
		certificateHeaders       map[common.Hash]*agglayer.CertificateHeader
		getFromDBError           error
		clientError              error
		updateDBError            error
		expectedErrorLogMessages []string
		expectedInfoMessages     []string
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
		},
	}

	for _, tt := range tests {
		tt := tt

		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			mockStorage := db.NewAggSenderStorageMock(t)
			mockAggLayerClient := agglayer.NewAgglayerClientMock(t)
			mockLogger := NewLoggerMock(t)

			mockStorage.On("GetCertificatesByStatus", mock.Anything, []agglayer.CertificateStatus{agglayer.Pending}).Return(tt.pendingCertificates, tt.getFromDBError)
			for certID, header := range tt.certificateHeaders {
				mockAggLayerClient.On("GetCertificateHeader", certID).Return(header, tt.clientError)
			}
			if tt.updateDBError != nil {
				mockStorage.On("UpdateCertificateStatus", mock.Anything, mock.Anything).Return(tt.updateDBError)
			} else if tt.clientError == nil && tt.getFromDBError == nil {
				mockStorage.On("UpdateCertificateStatus", mock.Anything, mock.Anything).Return(nil)
			}

			if tt.clientError != nil {
				for _, msg := range tt.expectedErrorLogMessages {
					mockLogger.On("Errorf", msg, mock.Anything, mock.Anything, mock.Anything).Return()
				}
			} else {
				for _, msg := range tt.expectedErrorLogMessages {
					mockLogger.On("Errorf", msg, mock.Anything).Return()
				}

				for _, msg := range tt.expectedInfoMessages {
					mockLogger.On("Infof", msg, mock.Anything, mock.Anything).Return()
				}
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

			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()

			go aggSender.checkIfCertificatesAreSettled(ctx)

			time.Sleep(2 * time.Second)
			cancel()

			mockLogger.AssertExpectations(t)
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
		name                    string
		sequencerKey            *ecdsa.PrivateKey
		l1BlockNumber           []interface{}
		getLastSentCertificate  []interface{}
		l2BlockFinality         []interface{}
		l2HeaderByNumber        []interface{}
		getCertificateHeader    []interface{}
		deleteCertificate       []interface{}
		getCertificateByHeight  []interface{}
		getBlockByLER           []interface{}
		getBridges              []interface{}
		getClaims               []interface{}
		getInfoByGlobalExitRoot []interface{}
		getExitRootByIndex      []interface{}
		originNetwork           []interface{}
		sendCertificate         []interface{}
		saveLastSentCertificate []interface{}
		expectedError           string
	}

	setupTest := func(cfg testCfg) (*AggSender, *db.AggSenderStorageMock, *EthClientMock, *L2BridgeSyncerMock,
		*EthClientMock, *agglayer.AgglayerClientMock, *L1InfoTreeSyncerMock) {
		var (
			aggsender = &AggSender{
				log:          log.WithFields("aggsender", 1),
				cfg:          Config{EpochSize: 10},
				sequencerKey: cfg.sequencerKey,
			}
			mockL1Client         *EthClientMock
			mockStorage          *db.AggSenderStorageMock
			mockL2Syncer         *L2BridgeSyncerMock
			mockL2Client         *EthClientMock
			mockAggLayerClient   *agglayer.AgglayerClientMock
			mockL1InfoTreeSyncer *L1InfoTreeSyncerMock
		)

		if cfg.l1BlockNumber != nil {
			mockL1Client = NewEthClientMock(t)
			mockL1Client.On("BlockNumber", mock.Anything).Return(cfg.l1BlockNumber...).Once()

			aggsender.l1Client = mockL1Client
		}

		if cfg.getLastSentCertificate != nil || cfg.deleteCertificate != nil ||
			cfg.getCertificateByHeight != nil || cfg.saveLastSentCertificate != nil {
			mockStorage = db.NewAggSenderStorageMock(t)
			mockStorage.On("GetLastSentCertificate", mock.Anything).Return(cfg.getLastSentCertificate...).Once()

			if cfg.deleteCertificate != nil {
				mockStorage.On("DeleteCertificate", mock.Anything, mock.Anything).Return(cfg.deleteCertificate...).Once()
			}

			if cfg.getCertificateByHeight != nil {
				mockStorage.On("GetCertificateByHeight", mock.Anything, mock.Anything).Return(cfg.getCertificateByHeight...).Once()
			}

			if cfg.saveLastSentCertificate != nil {
				mockStorage.On("SaveLastSentCertificate", mock.Anything, mock.Anything).Return(cfg.saveLastSentCertificate...).Once()
			}

			aggsender.storage = mockStorage
		}

		if cfg.l2BlockFinality != nil || cfg.getBlockByLER != nil || cfg.originNetwork != nil ||
			cfg.getBridges != nil || cfg.getClaims != nil || cfg.getInfoByGlobalExitRoot != nil {
			mockL2Syncer = NewL2BridgeSyncerMock(t)
			mockL2Syncer.On("BlockFinality").Return(cfg.l2BlockFinality...).Once()

			if cfg.getBlockByLER != nil {
				mockL2Syncer.On("GetBlockByLER", mock.Anything, mock.Anything).Return(cfg.getBlockByLER...).Once()
			}

			if cfg.getBridges != nil {
				mockL2Syncer.On("GetBridges", mock.Anything, mock.Anything, mock.Anything).Return(cfg.getBridges...).Once()
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

		if cfg.l2HeaderByNumber != nil {
			mockL2Client = NewEthClientMock(t)
			mockL2Client.On("HeaderByNumber", mock.Anything, mock.Anything).Return(cfg.l2HeaderByNumber...).Once()

			aggsender.l2Client = mockL2Client
		}

		if cfg.getCertificateHeader != nil || cfg.sendCertificate != nil {
			mockAggLayerClient = agglayer.NewAgglayerClientMock(t)
			mockAggLayerClient.On("GetCertificateHeader", mock.Anything).Return(cfg.getCertificateHeader...).Once()

			if cfg.sendCertificate != nil {
				mockAggLayerClient.On("SendCertificate", mock.Anything).Return(cfg.sendCertificate...).Once()
			}

			aggsender.aggLayerClient = mockAggLayerClient
		}

		if cfg.getInfoByGlobalExitRoot != nil {
			mockL1InfoTreeSyncer = NewL1InfoTreeSyncerMock(t)
			mockL1InfoTreeSyncer.On("GetInfoByGlobalExitRoot", mock.Anything).Return(cfg.getInfoByGlobalExitRoot...).Once()

			aggsender.l1infoTreeSyncer = mockL1InfoTreeSyncer
		}

		return aggsender, mockStorage, mockL1Client, mockL2Syncer, mockL2Client, mockAggLayerClient, mockL1InfoTreeSyncer
	}

	tests := []testCfg{
		{
			name:          "error getting L1 block",
			l1BlockNumber: []interface{}{uint64(0), errors.New("error getting block")},
			expectedError: "error getting block",
		},
		{
			name:          "should not send certificate",
			l1BlockNumber: []interface{}{uint64(1), nil},
		},
		{
			name:                   "error getting last sent certificate",
			l1BlockNumber:          []interface{}{uint64(8), nil},
			l2BlockFinality:        []interface{}{etherman.FinalizedBlock},
			l2HeaderByNumber:       []interface{}{&gethTypes.Header{Number: big.NewInt(8)}, nil},
			getLastSentCertificate: []interface{}{aggsendertypes.CertificateInfo{}, errors.New("error getting last sent certificate")},
			expectedError:          "error getting last sent certificate",
		},
		{
			name:            "error getting block finality",
			l1BlockNumber:   []interface{}{uint64(8), nil},
			l2BlockFinality: []interface{}{etherman.BlockNumberFinality("invalid")},
			expectedError:   "error getting block finality",
		},
		{
			name:             "error getting last l2 block",
			l1BlockNumber:    []interface{}{uint64(8), nil},
			l2BlockFinality:  []interface{}{etherman.FinalizedBlock},
			l2HeaderByNumber: []interface{}{nil, errors.New("error getting block")},
			expectedError:    "error getting block from l2",
		},
		{
			name:          "error getting last certificate header",
			l1BlockNumber: []interface{}{uint64(18), nil},
			getLastSentCertificate: []interface{}{aggsendertypes.CertificateInfo{
				Height:           1,
				CertificateID:    common.HexToHash("0x1"),
				NewLocalExitRoot: common.HexToHash("0x123"),
				FromBlock:        1,
				ToBlock:          10,
			}, nil},
			l2BlockFinality:      []interface{}{etherman.FinalizedBlock},
			l2HeaderByNumber:     []interface{}{&gethTypes.Header{Number: big.NewInt(20)}, nil},
			getCertificateHeader: []interface{}{nil, errors.New("error getting certificate header")},
			expectedError:        "error getting certificate",
		},
		{
			name:          "error deleting in error certificate",
			l1BlockNumber: []interface{}{uint64(18), nil},
			getLastSentCertificate: []interface{}{aggsendertypes.CertificateInfo{
				Height:           1,
				CertificateID:    common.HexToHash("0x1"),
				NewLocalExitRoot: common.HexToHash("0x123"),
				FromBlock:        1,
				ToBlock:          10,
			}, nil},
			l2BlockFinality:  []interface{}{etherman.FinalizedBlock},
			l2HeaderByNumber: []interface{}{&gethTypes.Header{Number: big.NewInt(20)}, nil},
			getCertificateHeader: []interface{}{&agglayer.CertificateHeader{
				Status: agglayer.InError,
			}, nil},
			deleteCertificate: []interface{}{errors.New("error deleting certificate")},
			expectedError:     "error deleting certificate",
		},
		{
			name:          "error getting certificate by height",
			l1BlockNumber: []interface{}{uint64(28), nil},
			getLastSentCertificate: []interface{}{aggsendertypes.CertificateInfo{
				Height:           11,
				CertificateID:    common.HexToHash("0x1"),
				NewLocalExitRoot: common.HexToHash("0x123"),
				FromBlock:        19,
				ToBlock:          28,
			}, nil},
			l2BlockFinality:  []interface{}{etherman.FinalizedBlock},
			l2HeaderByNumber: []interface{}{&gethTypes.Header{Number: big.NewInt(29)}, nil},
			getCertificateHeader: []interface{}{&agglayer.CertificateHeader{
				Status: agglayer.InError,
			}, nil},
			deleteCertificate:      []interface{}{nil},
			getCertificateByHeight: []interface{}{aggsendertypes.CertificateInfo{}, errors.New("error getting certificate by height")},
			expectedError:          "error getting certificate by height",
		},
		{
			name:          "error getting block by LER",
			l1BlockNumber: []interface{}{uint64(38), nil},
			getLastSentCertificate: []interface{}{aggsendertypes.CertificateInfo{
				Height:           21,
				CertificateID:    common.HexToHash("0x11"),
				NewLocalExitRoot: common.HexToHash("0x1223"),
				FromBlock:        29,
				ToBlock:          38,
			}, nil},
			l2BlockFinality:  []interface{}{etherman.PendingBlock},
			l2HeaderByNumber: []interface{}{&gethTypes.Header{Number: big.NewInt(38)}, nil},
			getCertificateHeader: []interface{}{&agglayer.CertificateHeader{
				Status: agglayer.InError,
			}, nil},
			deleteCertificate: []interface{}{nil},
			getCertificateByHeight: []interface{}{aggsendertypes.CertificateInfo{
				NewLocalExitRoot: common.HexToHash("0x1222"),
				Height:           20,
			}, nil},
			getBlockByLER: []interface{}{uint64(0), errors.New("error getting block by LER")},
			expectedError: "error getting block by LER",
		},
		{
			name:          "no new blocks to send certificate",
			l1BlockNumber: []interface{}{uint64(38), nil},
			getLastSentCertificate: []interface{}{aggsendertypes.CertificateInfo{
				Height:           41,
				CertificateID:    common.HexToHash("0x111"),
				NewLocalExitRoot: common.HexToHash("0x13223"),
				FromBlock:        31,
				ToBlock:          40,
			}, nil},
			l2BlockFinality:  []interface{}{etherman.FinalizedBlock},
			l2HeaderByNumber: []interface{}{&gethTypes.Header{Number: big.NewInt(41)}, nil},
			getCertificateHeader: []interface{}{&agglayer.CertificateHeader{
				Status: agglayer.Settled,
			}, nil},
			getBlockByLER: []interface{}{uint64(41), nil},
		},
		{
			name:          "get bridges error",
			l1BlockNumber: []interface{}{uint64(98), nil},
			getLastSentCertificate: []interface{}{aggsendertypes.CertificateInfo{
				Height:           50,
				CertificateID:    common.HexToHash("0x1111"),
				NewLocalExitRoot: common.HexToHash("0x132233"),
				FromBlock:        40,
				ToBlock:          49,
			}, nil},
			l2BlockFinality:  []interface{}{etherman.FinalizedBlock},
			l2HeaderByNumber: []interface{}{&gethTypes.Header{Number: big.NewInt(59)}, nil},
			getCertificateHeader: []interface{}{&agglayer.CertificateHeader{
				Status:        agglayer.Settled,
				CertificateID: common.HexToHash("0x1110"),
				Height:        49,
			}, nil},
			getBlockByLER: []interface{}{uint64(41), nil},
			getBridges:    []interface{}{nil, errors.New("error getting bridges")},
			expectedError: "error getting bridges",
		},
		{
			name:          "no bridges",
			l1BlockNumber: []interface{}{uint64(198), nil},
			getLastSentCertificate: []interface{}{aggsendertypes.CertificateInfo{
				Height:           60,
				CertificateID:    common.HexToHash("0x11111"),
				NewLocalExitRoot: common.HexToHash("0x1322233"),
				FromBlock:        50,
				ToBlock:          59,
			}, nil},
			l2BlockFinality:  []interface{}{etherman.PendingBlock},
			l2HeaderByNumber: []interface{}{&gethTypes.Header{Number: big.NewInt(69)}, nil},
			getCertificateHeader: []interface{}{&agglayer.CertificateHeader{
				Status:        agglayer.Settled,
				CertificateID: common.HexToHash("0x1110"),
				Height:        59,
			}, nil},
			getBlockByLER: []interface{}{uint64(51), nil},
			getBridges:    []interface{}{[]bridgesync.Bridge{}, nil},
		},
		{
			name:          "get claims error",
			l1BlockNumber: []interface{}{uint64(158), nil},
			getLastSentCertificate: []interface{}{aggsendertypes.CertificateInfo{
				Height:           70,
				CertificateID:    common.HexToHash("0x121111"),
				NewLocalExitRoot: common.HexToHash("0x13122233"),
				FromBlock:        60,
				ToBlock:          69,
			}, nil},
			l2BlockFinality:  []interface{}{etherman.FinalizedBlock},
			l2HeaderByNumber: []interface{}{&gethTypes.Header{Number: big.NewInt(79)}, nil},
			getCertificateHeader: []interface{}{&agglayer.CertificateHeader{
				Status:        agglayer.Settled,
				CertificateID: common.HexToHash("0x1110"),
				Height:        69,
			}, nil},
			getBlockByLER: []interface{}{uint64(61), nil},
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
			name:          "error building certificate",
			l1BlockNumber: []interface{}{uint64(148), nil},
			getLastSentCertificate: []interface{}{aggsendertypes.CertificateInfo{
				Height:           80,
				CertificateID:    common.HexToHash("0x1321111"),
				NewLocalExitRoot: common.HexToHash("0x131122233"),
				FromBlock:        70,
				ToBlock:          79,
			}, nil},
			l2BlockFinality:  []interface{}{etherman.PendingBlock},
			l2HeaderByNumber: []interface{}{&gethTypes.Header{Number: big.NewInt(89)}, nil},
			getCertificateHeader: []interface{}{&agglayer.CertificateHeader{
				Status:        agglayer.Settled,
				CertificateID: common.HexToHash("0x1110"),
				Height:        79,
			}, nil},
			getBlockByLER: []interface{}{uint64(71), nil},
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
			expectedError:           "error building certificate",
		},
		{
			name:          "send certificate error",
			l1BlockNumber: []interface{}{uint64(138), nil},
			getLastSentCertificate: []interface{}{aggsendertypes.CertificateInfo{
				Height:           90,
				CertificateID:    common.HexToHash("0x1121111"),
				NewLocalExitRoot: common.HexToHash("0x111122211"),
				FromBlock:        80,
				ToBlock:          89,
			}, nil},
			l2BlockFinality:  []interface{}{etherman.FinalizedBlock},
			l2HeaderByNumber: []interface{}{&gethTypes.Header{Number: big.NewInt(99)}, nil},
			getCertificateHeader: []interface{}{&agglayer.CertificateHeader{
				Status:        agglayer.Settled,
				CertificateID: common.HexToHash("0x1110"),
				Height:        89,
			}, nil},
			getBlockByLER: []interface{}{uint64(81), nil},
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
			name:          "store last sent certificate error",
			l1BlockNumber: []interface{}{uint64(128), nil},
			getLastSentCertificate: []interface{}{aggsendertypes.CertificateInfo{
				Height:           100,
				CertificateID:    common.HexToHash("0x11121111"),
				NewLocalExitRoot: common.HexToHash("0x1211122211"),
				FromBlock:        90,
				ToBlock:          99,
			}, nil},
			l2BlockFinality:  []interface{}{etherman.FinalizedBlock},
			l2HeaderByNumber: []interface{}{&gethTypes.Header{Number: big.NewInt(109)}, nil},
			getCertificateHeader: []interface{}{&agglayer.CertificateHeader{
				Status:        agglayer.Settled,
				CertificateID: common.HexToHash("0x11110"),
				Height:        99,
			}, nil},
			getBlockByLER: []interface{}{uint64(91), nil},
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
			name:          "successful sending of certificate",
			l1BlockNumber: []interface{}{uint64(178), nil},
			getLastSentCertificate: []interface{}{aggsendertypes.CertificateInfo{
				Height:           110,
				CertificateID:    common.HexToHash("0x12121111"),
				NewLocalExitRoot: common.HexToHash("0x1221122211"),
				FromBlock:        100,
				ToBlock:          109,
			}, nil},
			l2BlockFinality:  []interface{}{etherman.FinalizedBlock},
			l2HeaderByNumber: []interface{}{&gethTypes.Header{Number: big.NewInt(119)}, nil},
			getCertificateHeader: []interface{}{&agglayer.CertificateHeader{
				Status:        agglayer.Settled,
				CertificateID: common.HexToHash("0x11110"),
				Height:        109,
			}, nil},
			getBlockByLER: []interface{}{uint64(91), nil},
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

			aggsender, mockStorage, mockL1Client, mockL2Syncer, mockL2Client,
				mockAggLayerClient, mockL1InfoTreeSyncer := setupTest(tt)

			err := aggsender.sendCertificate(context.Background())

			if tt.expectedError != "" {
				require.ErrorContains(t, err, tt.expectedError)
			} else {
				require.NoError(t, err)
			}

			if mockStorage != nil {
				mockStorage.AssertExpectations(t)
			}

			if mockL1Client != nil {
				mockL1Client.AssertExpectations(t)
			}

			if mockL2Syncer != nil {
				mockL2Syncer.AssertExpectations(t)
			}

			if mockL2Client != nil {
				mockL2Client.AssertExpectations(t)
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
