package aggsender

import (
	"context"
	"crypto/ecdsa"
	"fmt"
	"math/big"
	"testing"

	"github.com/0xPolygon/cdk/agglayer"
	"github.com/0xPolygon/cdk/bridgesync"
	"github.com/0xPolygon/cdk/l1infotreesync"
	treeTypes "github.com/0xPolygon/cdk/tree/types"
	"github.com/ethereum/go-ethereum/common"
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
					GlobalIndex:        big.NewInt(2),
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
						MainnetFlag: false,
						RollupIndex: 0,
						LeafIndex:   2,
					},
					ClaimData: &agglayer.ClaimFromRollup{
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
						ProofLeafLER: agglayer.MerkleProof{
							Root:  common.HexToHash("0xccc"),
							Proof: mockProof,
						},
						ProofLERToRER: agglayer.MerkleProof{},
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

func generateTestProof(t *testing.T) treeTypes.Proof {
	t.Helper()

	proof := treeTypes.Proof{}

	for i := 0; i < int(treeTypes.DefaultHeight) && i < 10; i++ {
		proof[i] = common.HexToHash(fmt.Sprintf("0x%d", i))
	}

	return proof
}
