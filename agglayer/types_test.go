package agglayer

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"math/big"
	"reflect"
	"testing"

	cdkcommon "github.com/0xPolygon/cdk/common"
	"github.com/0xPolygon/cdk/log"
	"github.com/0xPolygon/cdk/tree/types"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/stretchr/testify/require"
)

const (
	expectedSignedCertificateEmptyMetadataJSON = `{"network_id":1,"height":1,"prev_local_exit_root":"0x0000000000000000000000000000000000000000000000000000000000000000","new_local_exit_root":"0x0000000000000000000000000000000000000000000000000000000000000000","bridge_exits":[{"leaf_type":"Transfer","token_info":null,"dest_network":0,"dest_address":"0x0000000000000000000000000000000000000000","amount":"1","metadata":null}],"imported_bridge_exits":[{"bridge_exit":{"leaf_type":"Transfer","token_info":null,"dest_network":0,"dest_address":"0x0000000000000000000000000000000000000000","amount":"1","metadata":null},"claim_data":null,"global_index":{"mainnet_flag":false,"rollup_index":1,"leaf_index":1}}],"metadata":"0x0000000000000000000000000000000000000000000000000000000000000000","signature":{"r":"0x0000000000000000000000000000000000000000000000000000000000000000","s":"0x0000000000000000000000000000000000000000000000000000000000000000","odd_y_parity":false}}`
	expectedSignedCertificateMetadataJSON      = `{"network_id":1,"height":1,"prev_local_exit_root":"0x0000000000000000000000000000000000000000000000000000000000000000","new_local_exit_root":"0x0000000000000000000000000000000000000000000000000000000000000000","bridge_exits":[{"leaf_type":"Transfer","token_info":null,"dest_network":0,"dest_address":"0x0000000000000000000000000000000000000000","amount":"1","metadata":[1,2,3]}],"imported_bridge_exits":[{"bridge_exit":{"leaf_type":"Transfer","token_info":null,"dest_network":0,"dest_address":"0x0000000000000000000000000000000000000000","amount":"1","metadata":null},"claim_data":null,"global_index":{"mainnet_flag":false,"rollup_index":1,"leaf_index":1}}],"metadata":"0x0000000000000000000000000000000000000000000000000000000000000000","signature":{"r":"0x0000000000000000000000000000000000000000000000000000000000000000","s":"0x0000000000000000000000000000000000000000000000000000000000000000","odd_y_parity":false}}`
)

func TestBridgeExitHash(t *testing.T) {
	MetadaHash := common.HexToHash("0x1234")
	bridge := BridgeExit{
		TokenInfo:        &TokenInfo{},
		IsMetadataHashed: true,
		Metadata:         MetadaHash[:],
	}
	require.Equal(t, "0x7d344cd1a895c66f0819be6a392d2a5d649c0cd5c8345706e11c757324da2943",
		bridge.Hash().String(), "use the hashed metadata, instead of calculating hash")

	bridge.IsMetadataHashed = false
	require.Equal(t, "0xa3ef92d7ca132432b864e424039077556b8757d2da4e01d6040c6ccbb39bef60",
		bridge.Hash().String(), "metadata is not hashed, calculate hash")

	bridge.IsMetadataHashed = false
	bridge.Metadata = []byte{}
	require.Equal(t, "0xad4224e96b39d42026b4795e5be83f43e0df757cdb13e781cd49e1a5363b193c",
		bridge.Hash().String(), "metadata is not hashed and it's empty, calculate hash")

	bridge.IsMetadataHashed = true
	bridge.Metadata = []byte{}
	require.Equal(t, "0x184125b2e3d1ded2ad3f82a383d9b09bd5bac4ccea4d41092f49523399598aca",
		bridge.Hash().String(), "metadata is a hashed and it's empty,use it")
}

func TestGenericError_Error(t *testing.T) {
	t.Parallel()

	err := GenericError{"test", "value"}
	require.Equal(t, "[Agglayer Error] test: value", err.Error())
}

func TestCertificateHeaderID(t *testing.T) {
	t.Parallel()

	certificate := CertificateHeader{
		Height:        1,
		CertificateID: common.HexToHash("0x123"),
	}
	require.Equal(t, "1/0x0000000000000000000000000000000000000000000000000000000000000123", certificate.ID())

	var certNil *CertificateHeader
	require.Equal(t, "nil", certNil.ID())
}

func TestCertificateHeaderString(t *testing.T) {
	t.Parallel()

	certificate := CertificateHeader{
		Height:        1,
		CertificateID: common.HexToHash("0x123"),
	}
	require.Equal(t, "Height: 1, CertificateID: 0x0000000000000000000000000000000000000000000000000000000000000123, PreviousLocalExitRoot: nil, NewLocalExitRoot: 0x0000000000000000000000000000000000000000000000000000000000000000. Status: Pending. Errors: []",
		certificate.String())

	var certNil *CertificateHeader
	require.Equal(t, "nil", certNil.String())
}

func TestMarshalJSON(t *testing.T) {
	t.Parallel()

	t.Run("MarshalJSON with empty proofs", func(t *testing.T) {
		t.Parallel()

		cert := SignedCertificate{
			Certificate: &Certificate{
				NetworkID:         1,
				Height:            1,
				PrevLocalExitRoot: common.Hash{},
				NewLocalExitRoot:  common.Hash{},
				BridgeExits: []*BridgeExit{
					{
						LeafType:           LeafTypeAsset,
						DestinationAddress: common.Address{},
						Amount:             big.NewInt(1),
					},
				},
				ImportedBridgeExits: []*ImportedBridgeExit{
					{
						BridgeExit: &BridgeExit{
							LeafType:           LeafTypeAsset,
							DestinationAddress: common.Address{},
							Amount:             big.NewInt(1),
							Metadata:           []byte{},
						},
						ClaimData: nil,
						GlobalIndex: &GlobalIndex{
							MainnetFlag: false,
							RollupIndex: 1,
							LeafIndex:   1,
						},
					},
				},
			},

			Signature: &Signature{
				R:         common.Hash{},
				S:         common.Hash{},
				OddParity: false,
			},
		}
		data, err := json.Marshal(cert)
		require.NoError(t, err)
		log.Info(string(data))
		require.Equal(t, expectedSignedCertificateEmptyMetadataJSON, string(data))

		cert.BridgeExits[0].Metadata = []byte{1, 2, 3}
		data, err = json.Marshal(cert)
		require.NoError(t, err)
		log.Info(string(data))
		require.Equal(t, expectedSignedCertificateMetadataJSON, string(data))
	})

	t.Run("MarshalJSON with proofs", func(t *testing.T) {
		t.Parallel()

		cert := SignedCertificate{
			Certificate: &Certificate{
				NetworkID:         11,
				Height:            111,
				PrevLocalExitRoot: common.HexToHash("0x111"),
				NewLocalExitRoot:  common.HexToHash("0x222"),
				BridgeExits: []*BridgeExit{
					{
						LeafType:           LeafTypeAsset,
						TokenInfo:          &TokenInfo{OriginNetwork: 1, OriginTokenAddress: common.HexToAddress("0x123")},
						DestinationNetwork: 2,
						DestinationAddress: common.HexToAddress("0x456"),
						Amount:             big.NewInt(1000),
						Metadata:           []byte{}, // we leave it empty on purpose to see when marshaled it will be null
					},
				},
				ImportedBridgeExits: []*ImportedBridgeExit{
					{
						BridgeExit: &BridgeExit{
							LeafType:           LeafTypeMessage,
							TokenInfo:          &TokenInfo{OriginNetwork: 1, OriginTokenAddress: common.HexToAddress("0x789")},
							DestinationNetwork: 2,
							DestinationAddress: common.HexToAddress("0xabc"),
							Amount:             big.NewInt(2000),
							Metadata:           []byte{0x03, 0x04},
						},
						GlobalIndex: &GlobalIndex{
							MainnetFlag: true,
							RollupIndex: 0,
							LeafIndex:   1,
						},
						ClaimData: &ClaimFromMainnnet{
							ProofLeafMER: &MerkleProof{
								Root:  common.HexToHash("0x333"),
								Proof: createDummyProof(t),
							},
							ProofGERToL1Root: &MerkleProof{
								Root:  common.HexToHash("0x444"),
								Proof: createDummyProof(t),
							},
							L1Leaf: &L1InfoTreeLeaf{
								L1InfoTreeIndex: 1,
								RollupExitRoot:  common.HexToHash("0x555"),
								MainnetExitRoot: common.HexToHash("0x123456"),
								Inner: &L1InfoTreeLeafInner{
									GlobalExitRoot: common.HexToHash("0x777"),
									BlockHash:      common.HexToHash("0x888"),
									Timestamp:      12345678,
								},
							},
						},
					},
					{
						BridgeExit: &BridgeExit{
							LeafType:           LeafTypeAsset,
							TokenInfo:          &TokenInfo{OriginNetwork: 1, OriginTokenAddress: common.HexToAddress("0x789")},
							DestinationNetwork: 2,
							DestinationAddress: common.HexToAddress("0xabcdef"),
							Amount:             big.NewInt(2201),
							Metadata:           []byte{0x05, 0x08},
						},
						GlobalIndex: &GlobalIndex{
							MainnetFlag: false,
							RollupIndex: 1,
							LeafIndex:   2,
						},
						ClaimData: &ClaimFromRollup{
							ProofLeafLER: &MerkleProof{
								Root:  common.HexToHash("0x333"),
								Proof: createDummyProof(t),
							},
							ProofLERToRER: &MerkleProof{
								Root:  common.HexToHash("0x444"),
								Proof: createDummyProof(t),
							},
							ProofGERToL1Root: &MerkleProof{
								Root:  common.HexToHash("0x555"),
								Proof: createDummyProof(t),
							},
							L1Leaf: &L1InfoTreeLeaf{
								L1InfoTreeIndex: 2,
								RollupExitRoot:  common.HexToHash("0x532"),
								MainnetExitRoot: common.HexToHash("0x654321"),
								Inner: &L1InfoTreeLeafInner{
									GlobalExitRoot: common.HexToHash("0x777"),
									BlockHash:      common.HexToHash("0x888"),
									Timestamp:      12345678,
								},
							},
						},
					},
				},
				Metadata: common.HexToHash("0xdef"),
			},
			Signature: &Signature{
				R:         common.HexToHash("0x111"),
				S:         common.HexToHash("0x222"),
				OddParity: true,
			},
		}

		expectedJSON := `{"network_id":11,"height":111,"prev_local_exit_root":"0x0000000000000000000000000000000000000000000000000000000000000111","new_local_exit_root":"0x0000000000000000000000000000000000000000000000000000000000000222","bridge_exits":[{"leaf_type":"Transfer","token_info":{"origin_network":1,"origin_token_address":"0x0000000000000000000000000000000000000123"},"dest_network":2,"dest_address":"0x0000000000000000000000000000000000000456","amount":"1000","metadata":null}],"imported_bridge_exits":[{"bridge_exit":{"leaf_type":"Message","token_info":{"origin_network":1,"origin_token_address":"0x0000000000000000000000000000000000000789"},"dest_network":2,"dest_address":"0x0000000000000000000000000000000000000abc","amount":"2000","metadata":[3,4]},"claim_data":{"Mainnet":{"l1_leaf":{"l1_info_tree_index":1,"rer":"0x0000000000000000000000000000000000000000000000000000000000000555","mer":"0x0000000000000000000000000000000000000000000000000000000000123456","inner":{"global_exit_root":"0x0000000000000000000000000000000000000000000000000000000000000777","block_hash":"0x0000000000000000000000000000000000000000000000000000000000000888","timestamp":12345678}},"proof_ger_l1root":{"root":"0x0000000000000000000000000000000000000000000000000000000000000444","proof":{"siblings":["0x0000000000000000000000000000000000000000000000000000000000000000","0x0000000000000000000000000000000000000000000000000000000000000001","0x0000000000000000000000000000000000000000000000000000000000000002","0x0000000000000000000000000000000000000000000000000000000000000003","0x0000000000000000000000000000000000000000000000000000000000000004","0x0000000000000000000000000000000000000000000000000000000000000005","0x0000000000000000000000000000000000000000000000000000000000000006","0x0000000000000000000000000000000000000000000000000000000000000007","0x0000000000000000000000000000000000000000000000000000000000000008","0x0000000000000000000000000000000000000000000000000000000000000009","0x000000000000000000000000000000000000000000000000000000000000000a","0x000000000000000000000000000000000000000000000000000000000000000b","0x000000000000000000000000000000000000000000000000000000000000000c","0x000000000000000000000000000000000000000000000000000000000000000d","0x000000000000000000000000000000000000000000000000000000000000000e","0x000000000000000000000000000000000000000000000000000000000000000f","0x0000000000000000000000000000000000000000000000000000000000000010","0x0000000000000000000000000000000000000000000000000000000000000011","0x0000000000000000000000000000000000000000000000000000000000000012","0x0000000000000000000000000000000000000000000000000000000000000013","0x0000000000000000000000000000000000000000000000000000000000000014","0x0000000000000000000000000000000000000000000000000000000000000015","0x0000000000000000000000000000000000000000000000000000000000000016","0x0000000000000000000000000000000000000000000000000000000000000017","0x0000000000000000000000000000000000000000000000000000000000000018","0x0000000000000000000000000000000000000000000000000000000000000019","0x000000000000000000000000000000000000000000000000000000000000001a","0x000000000000000000000000000000000000000000000000000000000000001b","0x000000000000000000000000000000000000000000000000000000000000001c","0x000000000000000000000000000000000000000000000000000000000000001d","0x000000000000000000000000000000000000000000000000000000000000001e","0x000000000000000000000000000000000000000000000000000000000000001f"]}},"proof_leaf_mer":{"root":"0x0000000000000000000000000000000000000000000000000000000000000333","proof":{"siblings":["0x0000000000000000000000000000000000000000000000000000000000000000","0x0000000000000000000000000000000000000000000000000000000000000001","0x0000000000000000000000000000000000000000000000000000000000000002","0x0000000000000000000000000000000000000000000000000000000000000003","0x0000000000000000000000000000000000000000000000000000000000000004","0x0000000000000000000000000000000000000000000000000000000000000005","0x0000000000000000000000000000000000000000000000000000000000000006","0x0000000000000000000000000000000000000000000000000000000000000007","0x0000000000000000000000000000000000000000000000000000000000000008","0x0000000000000000000000000000000000000000000000000000000000000009","0x000000000000000000000000000000000000000000000000000000000000000a","0x000000000000000000000000000000000000000000000000000000000000000b","0x000000000000000000000000000000000000000000000000000000000000000c","0x000000000000000000000000000000000000000000000000000000000000000d","0x000000000000000000000000000000000000000000000000000000000000000e","0x000000000000000000000000000000000000000000000000000000000000000f","0x0000000000000000000000000000000000000000000000000000000000000010","0x0000000000000000000000000000000000000000000000000000000000000011","0x0000000000000000000000000000000000000000000000000000000000000012","0x0000000000000000000000000000000000000000000000000000000000000013","0x0000000000000000000000000000000000000000000000000000000000000014","0x0000000000000000000000000000000000000000000000000000000000000015","0x0000000000000000000000000000000000000000000000000000000000000016","0x0000000000000000000000000000000000000000000000000000000000000017","0x0000000000000000000000000000000000000000000000000000000000000018","0x0000000000000000000000000000000000000000000000000000000000000019","0x000000000000000000000000000000000000000000000000000000000000001a","0x000000000000000000000000000000000000000000000000000000000000001b","0x000000000000000000000000000000000000000000000000000000000000001c","0x000000000000000000000000000000000000000000000000000000000000001d","0x000000000000000000000000000000000000000000000000000000000000001e","0x000000000000000000000000000000000000000000000000000000000000001f"]}}}},"global_index":{"mainnet_flag":true,"rollup_index":0,"leaf_index":1}},{"bridge_exit":{"leaf_type":"Transfer","token_info":{"origin_network":1,"origin_token_address":"0x0000000000000000000000000000000000000789"},"dest_network":2,"dest_address":"0x0000000000000000000000000000000000abcdef","amount":"2201","metadata":[5,8]},"claim_data":{"Rollup":{"l1_leaf":{"l1_info_tree_index":2,"rer":"0x0000000000000000000000000000000000000000000000000000000000000532","mer":"0x0000000000000000000000000000000000000000000000000000000000654321","inner":{"global_exit_root":"0x0000000000000000000000000000000000000000000000000000000000000777","block_hash":"0x0000000000000000000000000000000000000000000000000000000000000888","timestamp":12345678}},"proof_ger_l1root":{"root":"0x0000000000000000000000000000000000000000000000000000000000000555","proof":{"siblings":["0x0000000000000000000000000000000000000000000000000000000000000000","0x0000000000000000000000000000000000000000000000000000000000000001","0x0000000000000000000000000000000000000000000000000000000000000002","0x0000000000000000000000000000000000000000000000000000000000000003","0x0000000000000000000000000000000000000000000000000000000000000004","0x0000000000000000000000000000000000000000000000000000000000000005","0x0000000000000000000000000000000000000000000000000000000000000006","0x0000000000000000000000000000000000000000000000000000000000000007","0x0000000000000000000000000000000000000000000000000000000000000008","0x0000000000000000000000000000000000000000000000000000000000000009","0x000000000000000000000000000000000000000000000000000000000000000a","0x000000000000000000000000000000000000000000000000000000000000000b","0x000000000000000000000000000000000000000000000000000000000000000c","0x000000000000000000000000000000000000000000000000000000000000000d","0x000000000000000000000000000000000000000000000000000000000000000e","0x000000000000000000000000000000000000000000000000000000000000000f","0x0000000000000000000000000000000000000000000000000000000000000010","0x0000000000000000000000000000000000000000000000000000000000000011","0x0000000000000000000000000000000000000000000000000000000000000012","0x0000000000000000000000000000000000000000000000000000000000000013","0x0000000000000000000000000000000000000000000000000000000000000014","0x0000000000000000000000000000000000000000000000000000000000000015","0x0000000000000000000000000000000000000000000000000000000000000016","0x0000000000000000000000000000000000000000000000000000000000000017","0x0000000000000000000000000000000000000000000000000000000000000018","0x0000000000000000000000000000000000000000000000000000000000000019","0x000000000000000000000000000000000000000000000000000000000000001a","0x000000000000000000000000000000000000000000000000000000000000001b","0x000000000000000000000000000000000000000000000000000000000000001c","0x000000000000000000000000000000000000000000000000000000000000001d","0x000000000000000000000000000000000000000000000000000000000000001e","0x000000000000000000000000000000000000000000000000000000000000001f"]}},"proof_leaf_ler":{"root":"0x0000000000000000000000000000000000000000000000000000000000000333","proof":{"siblings":["0x0000000000000000000000000000000000000000000000000000000000000000","0x0000000000000000000000000000000000000000000000000000000000000001","0x0000000000000000000000000000000000000000000000000000000000000002","0x0000000000000000000000000000000000000000000000000000000000000003","0x0000000000000000000000000000000000000000000000000000000000000004","0x0000000000000000000000000000000000000000000000000000000000000005","0x0000000000000000000000000000000000000000000000000000000000000006","0x0000000000000000000000000000000000000000000000000000000000000007","0x0000000000000000000000000000000000000000000000000000000000000008","0x0000000000000000000000000000000000000000000000000000000000000009","0x000000000000000000000000000000000000000000000000000000000000000a","0x000000000000000000000000000000000000000000000000000000000000000b","0x000000000000000000000000000000000000000000000000000000000000000c","0x000000000000000000000000000000000000000000000000000000000000000d","0x000000000000000000000000000000000000000000000000000000000000000e","0x000000000000000000000000000000000000000000000000000000000000000f","0x0000000000000000000000000000000000000000000000000000000000000010","0x0000000000000000000000000000000000000000000000000000000000000011","0x0000000000000000000000000000000000000000000000000000000000000012","0x0000000000000000000000000000000000000000000000000000000000000013","0x0000000000000000000000000000000000000000000000000000000000000014","0x0000000000000000000000000000000000000000000000000000000000000015","0x0000000000000000000000000000000000000000000000000000000000000016","0x0000000000000000000000000000000000000000000000000000000000000017","0x0000000000000000000000000000000000000000000000000000000000000018","0x0000000000000000000000000000000000000000000000000000000000000019","0x000000000000000000000000000000000000000000000000000000000000001a","0x000000000000000000000000000000000000000000000000000000000000001b","0x000000000000000000000000000000000000000000000000000000000000001c","0x000000000000000000000000000000000000000000000000000000000000001d","0x000000000000000000000000000000000000000000000000000000000000001e","0x000000000000000000000000000000000000000000000000000000000000001f"]}},"proof_ler_rer":{"root":"0x0000000000000000000000000000000000000000000000000000000000000444","proof":{"siblings":["0x0000000000000000000000000000000000000000000000000000000000000000","0x0000000000000000000000000000000000000000000000000000000000000001","0x0000000000000000000000000000000000000000000000000000000000000002","0x0000000000000000000000000000000000000000000000000000000000000003","0x0000000000000000000000000000000000000000000000000000000000000004","0x0000000000000000000000000000000000000000000000000000000000000005","0x0000000000000000000000000000000000000000000000000000000000000006","0x0000000000000000000000000000000000000000000000000000000000000007","0x0000000000000000000000000000000000000000000000000000000000000008","0x0000000000000000000000000000000000000000000000000000000000000009","0x000000000000000000000000000000000000000000000000000000000000000a","0x000000000000000000000000000000000000000000000000000000000000000b","0x000000000000000000000000000000000000000000000000000000000000000c","0x000000000000000000000000000000000000000000000000000000000000000d","0x000000000000000000000000000000000000000000000000000000000000000e","0x000000000000000000000000000000000000000000000000000000000000000f","0x0000000000000000000000000000000000000000000000000000000000000010","0x0000000000000000000000000000000000000000000000000000000000000011","0x0000000000000000000000000000000000000000000000000000000000000012","0x0000000000000000000000000000000000000000000000000000000000000013","0x0000000000000000000000000000000000000000000000000000000000000014","0x0000000000000000000000000000000000000000000000000000000000000015","0x0000000000000000000000000000000000000000000000000000000000000016","0x0000000000000000000000000000000000000000000000000000000000000017","0x0000000000000000000000000000000000000000000000000000000000000018","0x0000000000000000000000000000000000000000000000000000000000000019","0x000000000000000000000000000000000000000000000000000000000000001a","0x000000000000000000000000000000000000000000000000000000000000001b","0x000000000000000000000000000000000000000000000000000000000000001c","0x000000000000000000000000000000000000000000000000000000000000001d","0x000000000000000000000000000000000000000000000000000000000000001e","0x000000000000000000000000000000000000000000000000000000000000001f"]}}}},"global_index":{"mainnet_flag":false,"rollup_index":1,"leaf_index":2}}],"metadata":"0x0000000000000000000000000000000000000000000000000000000000000def","signature":{"r":"0x0000000000000000000000000000000000000000000000000000000000000111","s":"0x0000000000000000000000000000000000000000000000000000000000000222","odd_y_parity":true}}`

		data, err := json.Marshal(cert)
		require.NoError(t, err)
		require.Equal(t, expectedJSON, string(data))

		require.Equal(t, "0x46da3ca29098968ffc46ed2b894757671fa73cf7ebd4b82c89e90cc36a2737ae", cert.Hash().String())
		require.Equal(t, "0xc82b12e7383d2a8c1ec290d575bf6b2ac48363ca824cd51fe9f8cd312d55cd7a", cert.BridgeExits[0].Hash().String())
		require.Equal(t, "0x23fbdb2d272c9a4ae45135986363363d2e87dd6c2f2494a62b86851396f3fed4", cert.ImportedBridgeExits[0].Hash().String())
		require.Equal(t, "0x7eb947fcd0ed89ba0f41ec3f85e600d7114ec9349eb99a9478e3dd4e456297b1", cert.ImportedBridgeExits[1].Hash().String())
	})
}

func TestSignedCertificate_Copy(t *testing.T) {
	t.Parallel()

	t.Run("copy with non-nil fields", func(t *testing.T) {
		t.Parallel()

		original := &SignedCertificate{
			Certificate: &Certificate{
				NetworkID:         1,
				Height:            100,
				PrevLocalExitRoot: [32]byte{0x01},
				NewLocalExitRoot:  [32]byte{0x02},
				BridgeExits: []*BridgeExit{
					{
						LeafType:           LeafTypeAsset,
						TokenInfo:          &TokenInfo{OriginNetwork: 1, OriginTokenAddress: common.HexToAddress("0x123")},
						DestinationNetwork: 2,
						DestinationAddress: common.HexToAddress("0x456"),
						Amount:             big.NewInt(1000),
						Metadata:           []byte{0x01, 0x02},
					},
				},
				ImportedBridgeExits: []*ImportedBridgeExit{
					{
						BridgeExit: &BridgeExit{
							LeafType:           LeafTypeMessage,
							TokenInfo:          &TokenInfo{OriginNetwork: 1, OriginTokenAddress: common.HexToAddress("0x789")},
							DestinationNetwork: 3,
							DestinationAddress: common.HexToAddress("0xabc"),
							Amount:             big.NewInt(2000),
							Metadata:           []byte{0x03, 0x04},
						},
						ClaimData:   &ClaimFromMainnnet{},
						GlobalIndex: &GlobalIndex{MainnetFlag: true, RollupIndex: 1, LeafIndex: 2},
					},
				},
				Metadata: common.HexToHash("0xdef"),
			},
			Signature: &Signature{
				R:         common.HexToHash("0x111"),
				S:         common.HexToHash("0x222"),
				OddParity: true,
			},
		}

		certificateCopy := original.CopyWithDefaulting()

		require.NotNil(t, certificateCopy)
		require.NotSame(t, original, certificateCopy)
		require.NotSame(t, original.Certificate, certificateCopy.Certificate)
		require.Same(t, original.Signature, certificateCopy.Signature)
		require.Equal(t, original, certificateCopy)
	})

	t.Run("copy with nil BridgeExits, ImportedBridgeExits and Signature", func(t *testing.T) {
		t.Parallel()

		original := &SignedCertificate{
			Certificate: &Certificate{
				NetworkID:           1,
				Height:              100,
				PrevLocalExitRoot:   [32]byte{0x01},
				NewLocalExitRoot:    [32]byte{0x02},
				BridgeExits:         nil,
				ImportedBridgeExits: nil,
				Metadata:            common.HexToHash("0xdef"),
			},
			Signature: nil,
		}

		certificateCopy := original.CopyWithDefaulting()

		require.NotNil(t, certificateCopy)
		require.NotSame(t, original, certificateCopy)
		require.NotSame(t, original.Certificate, certificateCopy.Certificate)
		require.NotNil(t, certificateCopy.Signature)
		require.Equal(t, original.NetworkID, certificateCopy.NetworkID)
		require.Equal(t, original.Height, certificateCopy.Height)
		require.Equal(t, original.PrevLocalExitRoot, certificateCopy.PrevLocalExitRoot)
		require.Equal(t, original.NewLocalExitRoot, certificateCopy.NewLocalExitRoot)
		require.Equal(t, original.Metadata, certificateCopy.Metadata)
		require.NotNil(t, certificateCopy.BridgeExits)
		require.NotNil(t, certificateCopy.ImportedBridgeExits)
		require.Empty(t, certificateCopy.BridgeExits)
		require.Empty(t, certificateCopy.ImportedBridgeExits)
	})
}

func TestGlobalIndex_UnmarshalFromMap(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		data    map[string]interface{}
		want    *GlobalIndex
		wantErr bool
	}{
		{
			name: "valid data",
			data: map[string]interface{}{
				"rollup_index": uint32(0),
				"leaf_index":   uint32(2),
				"mainnet_flag": true,
			},
			want: &GlobalIndex{
				RollupIndex: 0,
				LeafIndex:   2,
				MainnetFlag: true,
			},
			wantErr: false,
		},
		{
			name: "missing rollup_index",
			data: map[string]interface{}{
				"leaf_index":   uint32(2),
				"mainnet_flag": true,
			},
			want:    &GlobalIndex{},
			wantErr: true,
		},
		{
			name: "invalid rollup_index type",
			data: map[string]interface{}{
				"rollup_index": "invalid",
				"leaf_index":   uint32(2),
				"mainnet_flag": true,
			},
			want:    &GlobalIndex{},
			wantErr: true,
		},
		{
			name: "missing leaf_index",
			data: map[string]interface{}{
				"rollup_index": uint32(1),
				"mainnet_flag": true,
			},
			want:    &GlobalIndex{},
			wantErr: true,
		},
		{
			name: "invalid leaf_index type",
			data: map[string]interface{}{
				"rollup_index": uint32(1),
				"leaf_index":   "invalid",
				"mainnet_flag": true,
			},
			want:    &GlobalIndex{},
			wantErr: true,
		},
		{
			name: "missing mainnet_flag",
			data: map[string]interface{}{
				"rollup_index": uint32(1),
				"leaf_index":   uint32(2),
			},
			want:    &GlobalIndex{},
			wantErr: true,
		},
		{
			name: "invalid mainnet_flag type",
			data: map[string]interface{}{
				"rollup_index": uint32(1),
				"leaf_index":   uint32(2),
				"mainnet_flag": "invalid",
			},
			want:    &GlobalIndex{},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			g := &GlobalIndex{}
			err := g.UnmarshalFromMap(tt.data)
			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				require.Equal(t, tt.want, g)
			}
		})
	}
}

func TestUnmarshalCertificateHeaderUnknownError(t *testing.T) {
	t.Parallel()
	rawCertificateHeader := `{
		"network_id": 14,
		"height": 0,
		"epoch_number": null,
		"certificate_index": null,
		"certificate_id": "0x3af88c9ca106822bd141fdc680dcb888f4e9d4997fad1645ba3d5d747059eb32",
		"new_local_exit_root": "0x625e889ced3c31277c6653229096374d396a2fd3564a8894aaad2ff935d2fc8c",
		"metadata": "0x0000000000000000000000000000000000000000000000000000000000002f3d",
		"status": {
			"InError": {
				"error": {
					"ProofVerificationFailed": {
						"Plonk": "the verifying key does not match the inner plonk bn254 proof's committed verifying key"
					}
				}
			}
		}
	}`

	var result *CertificateHeader
	err := json.Unmarshal([]byte(rawCertificateHeader), &result)
	require.NoError(t, err)
	require.NotNil(t, result)

	expectedErr := &GenericError{
		Key:   "ProofVerificationFailed",
		Value: "{\"Plonk\":\"the verifying key does not match the inner plonk bn254 proof's committed verifying key\"}",
	}

	require.Equal(t, expectedErr, result.Error)
}

func TestConvertNumeric(t *testing.T) {
	tests := []struct {
		name        string
		value       float64
		target      reflect.Type
		expected    interface{}
		expectedErr error
	}{
		// Integer conversions
		{"FloatToInt", 42.5, reflect.TypeOf(int(0)), int(42), nil},
		{"FloatToInt8", 127.5, reflect.TypeOf(int8(0)), int8(127), nil},
		{"FloatToInt16", 32767.5, reflect.TypeOf(int16(0)), int16(32767), nil},
		{"FloatToInt32", 2147483647.5, reflect.TypeOf(int32(0)), int32(2147483647), nil},
		{"FloatToInt64", -10000000000000000.9, reflect.TypeOf(int64(0)), int64(-10000000000000000), nil},

		// Unsigned integer conversions
		{"FloatToUint", 42.5, reflect.TypeOf(uint(0)), uint(42), nil},
		{"FloatToUint8", 255.5, reflect.TypeOf(uint8(0)), uint8(255), nil},
		{"FloatToUint16", 65535.5, reflect.TypeOf(uint16(0)), uint16(65535), nil},
		{"FloatToUint32", 4294967295.5, reflect.TypeOf(uint32(0)), uint32(4294967295), nil},
		{"FloatToUint64", 10000000000000000.9, reflect.TypeOf(uint64(0)), uint64(10000000000000000), nil},

		// Float conversions
		{"FloatToFloat32", 3.14, reflect.TypeOf(float32(0)), float32(3.14), nil},
		{"FloatToFloat64", 3.14, reflect.TypeOf(float64(0)), float64(3.14), nil},

		// Unsupported type
		{"UnsupportedType", 3.14, reflect.TypeOf("string"), nil, errors.New("unsupported target type string")},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := convertNumeric(tt.value, tt.target)
			if tt.expectedErr != nil {
				require.ErrorContains(t, err, tt.expectedErr.Error())
			} else {
				require.NoError(t, err)
			}
			require.Equal(t, tt.expected, result)
		})
	}
}

func TestCertificateHash(t *testing.T) {
	// Test inputs
	prevLocalExitRoot := [common.HashLength]byte{}
	newLocalExitRoot := [common.HashLength]byte{}
	copy(prevLocalExitRoot[:], bytes.Repeat([]byte{0x01}, common.HashLength))
	copy(newLocalExitRoot[:], bytes.Repeat([]byte{0x02}, common.HashLength))

	// Create dummy BridgeExits
	bridgeExits := []*BridgeExit{
		{
			LeafType:           LeafTypeAsset,
			TokenInfo:          createDummyTokenInfo(t),
			DestinationNetwork: 1,
			DestinationAddress: common.HexToAddress("0x0000000000000000000000000000000000000001"),
			Amount:             big.NewInt(100),
			Metadata:           []byte("metadata1"),
		},
		{
			LeafType:           LeafTypeMessage,
			TokenInfo:          createDummyTokenInfo(t),
			DestinationNetwork: 2,
			DestinationAddress: common.HexToAddress("0x0000000000000000000000000000000000000002"),
			Amount:             big.NewInt(200),
			Metadata:           []byte("metadata2"),
		},
	}

	// Create dummy ImportedBridgeExits
	importedBridgeExits := []*ImportedBridgeExit{
		{
			BridgeExit: &BridgeExit{
				LeafType:           LeafTypeAsset,
				TokenInfo:          createDummyTokenInfo(t),
				DestinationNetwork: 3,
				DestinationAddress: common.HexToAddress("0x0000000000000000000000000000000000000003"),
				Amount:             big.NewInt(300),
				Metadata:           []byte("metadata3"),
			},
			ClaimData:   createDummyClaim(t),
			GlobalIndex: createDummyGlobalIndex(t),
		},
		{
			BridgeExit: &BridgeExit{
				LeafType:           LeafTypeAsset,
				TokenInfo:          createDummyTokenInfo(t),
				DestinationNetwork: 4,
				DestinationAddress: common.HexToAddress("0x0000000000000000000000000000000000000004"),
				Amount:             big.NewInt(400),
				Metadata:           []byte("metadata4"),
			},
			ClaimData:   createDummyClaim(t),
			GlobalIndex: createDummyGlobalIndex(t),
		},
	}

	metadata := common.HexToHash("0x123456789abcdef123456789abcdef123456789abcdef123456789abcdef1234")

	// Create the certificate
	certificate := &Certificate{
		NetworkID:           1,
		Height:              100,
		PrevLocalExitRoot:   prevLocalExitRoot,
		NewLocalExitRoot:    newLocalExitRoot,
		BridgeExits:         bridgeExits,
		ImportedBridgeExits: importedBridgeExits,
		Metadata:            metadata,
	}

	// Manually calculate the expected hash
	bridgeExitsHashes := [][]byte{
		bridgeExits[0].Hash().Bytes(),
		bridgeExits[1].Hash().Bytes(),
	}
	importedBridgeExitsHashes := [][]byte{
		importedBridgeExits[0].Hash().Bytes(),
		importedBridgeExits[1].Hash().Bytes(),
	}

	bridgeExitsPart := crypto.Keccak256(bridgeExitsHashes...)
	importedBridgeExitsPart := crypto.Keccak256(importedBridgeExitsHashes...)

	expectedHash := crypto.Keccak256Hash(
		cdkcommon.Uint32ToBytes(1),
		cdkcommon.Uint64ToBytes(100),
		prevLocalExitRoot[:],
		newLocalExitRoot[:],
		bridgeExitsPart,
		importedBridgeExitsPart,
	)

	// Test the certificate hash
	calculatedHash := certificate.Hash()

	require.Equal(t, calculatedHash, expectedHash)
}

func TestCertificate_HashToSign(t *testing.T) {
	c := &Certificate{
		NewLocalExitRoot: common.HexToHash("0xabcd"),
		ImportedBridgeExits: []*ImportedBridgeExit{
			{
				GlobalIndex: &GlobalIndex{
					MainnetFlag: true,
					RollupIndex: 23,
					LeafIndex:   1,
				},
			},
			{
				GlobalIndex: &GlobalIndex{
					MainnetFlag: false,
					RollupIndex: 15,
					LeafIndex:   2,
				},
			},
		},
	}

	globalIndexHashes := make([][]byte, len(c.ImportedBridgeExits))
	for i, importedBridgeExit := range c.ImportedBridgeExits {
		globalIndexHashes[i] = importedBridgeExit.GlobalIndex.Hash().Bytes()
	}

	expectedHash := crypto.Keccak256Hash(
		c.NewLocalExitRoot[:],
		crypto.Keccak256Hash(globalIndexHashes...).Bytes(),
	)

	certHash := c.HashToSign()
	require.Equal(t, expectedHash, certHash)
}

func TestClaimFromMainnnet_MarshalJSON(t *testing.T) {
	// Test data
	merkleProof := &MerkleProof{
		Root: common.HexToHash("0x1"),
		Proof: [types.DefaultHeight]common.Hash{
			common.HexToHash("0x2"),
			common.HexToHash("0x3"),
		},
	}

	l1InfoTreeLeaf := &L1InfoTreeLeaf{
		L1InfoTreeIndex: 42,
		RollupExitRoot:  [common.HashLength]byte{0xaa, 0xbb, 0xcc},
		MainnetExitRoot: [common.HashLength]byte{0xdd, 0xee, 0xff},
		Inner: &L1InfoTreeLeafInner{
			GlobalExitRoot: common.HexToHash("0x1"),
			BlockHash:      common.HexToHash("0x2"),
			Timestamp:      1672531200, // Example timestamp
		},
	}

	claim := &ClaimFromMainnnet{
		ProofLeafMER:     merkleProof,
		ProofGERToL1Root: merkleProof,
		L1Leaf:           l1InfoTreeLeaf,
	}

	// Marshal the ClaimFromMainnnet struct to JSON
	expectedJSON, err := claim.MarshalJSON()
	require.NoError(t, err)

	var actualClaim ClaimFromMainnnet
	err = json.Unmarshal(expectedJSON, &actualClaim)
	require.NoError(t, err)
}

func TestBridgeExitString(t *testing.T) {
	tests := []struct {
		name           string
		bridgeExit     *BridgeExit
		expectedOutput string
	}{
		{
			name: "With TokenInfo",
			bridgeExit: &BridgeExit{
				LeafType:           LeafTypeAsset,
				TokenInfo:          createDummyTokenInfo(t),
				DestinationNetwork: 100,
				DestinationAddress: common.HexToAddress("0x2"),
				Amount:             big.NewInt(1000),
				Metadata:           []byte{0x01, 0x02, 0x03},
			},
			expectedOutput: "LeafType: Transfer, DestinationNetwork: 100, DestinationAddress: 0x0000000000000000000000000000000000000002, Amount: 1000, Metadata: 010203, TokenInfo: OriginNetwork: 1, OriginTokenAddress: 0x0000000000000000000000000000000000002345",
		},
		{
			name: "Without TokenInfo",
			bridgeExit: &BridgeExit{
				LeafType:           LeafTypeMessage,
				DestinationNetwork: 200,
				DestinationAddress: common.HexToAddress("0x1"),
				Amount:             big.NewInt(5000),
				Metadata:           []byte{0xff, 0xee, 0xdd},
			},
			expectedOutput: "LeafType: Message, DestinationNetwork: 200, DestinationAddress: 0x0000000000000000000000000000000000000001, Amount: 5000, Metadata: ffeedd, TokenInfo: nil",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			actualOutput := tt.bridgeExit.String()
			require.Equal(t, tt.expectedOutput, actualOutput)
		})
	}
}

func TestCertificateStatusUnmarshalJSON(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		expected    CertificateStatus
		expectError bool
	}{
		{
			name:        "Valid status - Pending",
			input:       `"Pending"`,
			expected:    Pending,
			expectError: false,
		},
		{
			name:        "Valid status - Proven",
			input:       `"Proven"`,
			expected:    Proven,
			expectError: false,
		},
		{
			name:        "Valid status - Candidate",
			input:       `"Candidate"`,
			expected:    Candidate,
			expectError: false,
		},
		{
			name:        "Valid status - InError",
			input:       `"InError"`,
			expected:    InError,
			expectError: false,
		},
		{
			name:        "Valid status - Settled",
			input:       `"Settled"`,
			expected:    Settled,
			expectError: false,
		},
		{
			name:        "Invalid status",
			input:       `"InvalidStatus"`,
			expected:    0, // Unchanged (default value of CertificateStatus)
			expectError: true,
		},
		{
			name:        "Contains 'InError' string",
			input:       `"SomeStringWithInError"`,
			expected:    InError,
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var status CertificateStatus
			err := json.Unmarshal([]byte(tt.input), &status)

			if tt.expectError {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				require.Equal(t, tt.expected, status)
			}
		})
	}
}

func TestMerkleProofString(t *testing.T) {
	tests := []struct {
		name     string
		proof    MerkleProof
		expected string
	}{
		{
			name: "Empty Root and Empty Proof",
			proof: MerkleProof{
				Root:  common.Hash{},
				Proof: [types.DefaultHeight]common.Hash{},
			},
			expected: fmt.Sprintf("Root: %s, Proof: %v", common.Hash{}.String(), [types.DefaultHeight]common.Hash{}),
		},
		{
			name: "Non-Empty Root and Empty Proof",
			proof: MerkleProof{
				Root:  common.HexToHash("0xabc123"),
				Proof: [types.DefaultHeight]common.Hash{},
			},
			expected: fmt.Sprintf("Root: %s, Proof: %v", common.HexToHash("0xabc123").String(), [types.DefaultHeight]common.Hash{}),
		},
		{
			name: "Non-Empty Root and Partially Populated Proof",
			proof: MerkleProof{
				Root: common.HexToHash("0xabc123"),
				Proof: [types.DefaultHeight]common.Hash{
					common.HexToHash("0xdef456"),
					common.HexToHash("0x123789"),
				},
			},
			expected: fmt.Sprintf("Root: %s, Proof: %v", common.HexToHash("0xabc123").String(), [types.DefaultHeight]common.Hash{
				common.HexToHash("0xdef456"),
				common.HexToHash("0x123789"),
			}),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.proof.String()
			require.Equal(t, tt.expected, result)
		})
	}
}

func TestGlobalIndexString(t *testing.T) {
	tests := []struct {
		name     string
		input    GlobalIndex
		expected string
	}{
		{
			name: "All fields zero",
			input: GlobalIndex{
				MainnetFlag: false,
				RollupIndex: 0,
				LeafIndex:   0,
			},
			expected: "MainnetFlag: false, RollupIndex: 0, LeafIndex: 0",
		},
		{
			name: "MainnetFlag true, non-zero indices",
			input: GlobalIndex{
				MainnetFlag: true,
				RollupIndex: 123,
				LeafIndex:   456,
			},
			expected: "MainnetFlag: true, RollupIndex: 123, LeafIndex: 456",
		},
		{
			name: "MainnetFlag false, large indices",
			input: GlobalIndex{
				MainnetFlag: false,
				RollupIndex: 4294967295, // Maximum value of uint32
				LeafIndex:   2147483647, // Large but within uint32 range
			},
			expected: "MainnetFlag: false, RollupIndex: 4294967295, LeafIndex: 2147483647",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.input.String()
			require.Equal(t, tt.expected, result)
		})
	}
}

func TestL1InfoTreeLeafString(t *testing.T) {
	tests := []struct {
		name     string
		input    L1InfoTreeLeaf
		expected string
	}{
		{
			name: "With valid Inner",
			input: L1InfoTreeLeaf{
				L1InfoTreeIndex: 1,
				RollupExitRoot:  common.HexToHash("0x01"),
				MainnetExitRoot: common.HexToHash("0x02"),
				Inner: &L1InfoTreeLeafInner{
					GlobalExitRoot: common.HexToHash("0x03"),
					BlockHash:      common.HexToHash("0x04"),
					Timestamp:      1234567890,
				},
			},
			expected: "L1InfoTreeIndex: 1, RollupExitRoot: 0x0000000000000000000000000000000000000000000000000000000000000001, " +
				"MainnetExitRoot: 0x0000000000000000000000000000000000000000000000000000000000000002, " +
				"Inner: GlobalExitRoot: 0x0000000000000000000000000000000000000000000000000000000000000003, " +
				"BlockHash: 0x0000000000000000000000000000000000000000000000000000000000000004, Timestamp: 1234567890",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.input.String()
			if result != tt.expected {
				t.Errorf("L1InfoTreeLeaf.String() = %q, want %q", result, tt.expected)
			}
		})
	}
}

func TestClaimType(t *testing.T) {
	cases := []struct {
		name         string
		claim        Claim
		expectedType string
	}{
		{
			name:         "Mainnet claim",
			claim:        &ClaimFromMainnnet{},
			expectedType: "Mainnet",
		},
		{
			name:         "Rollup claim",
			claim:        &ClaimFromRollup{},
			expectedType: "Rollup",
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			actualType := c.claim.Type()
			require.Equal(t, c.expectedType, actualType)
		})
	}
}
