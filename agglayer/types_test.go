package agglayer

import (
	"encoding/json"
	"math/big"
	"testing"

	"github.com/0xPolygon/cdk/log"
	"github.com/ethereum/go-ethereum/common"
	"github.com/stretchr/testify/require"
)

const (
	expectedSignedCertificateEmptyMetadataJSON = `{"network_id":1,"height":1,"prev_local_exit_root":[0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0],"new_local_exit_root":[0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0],"bridge_exits":[{"leaf_type":"Transfer","token_info":null,"dest_network":0,"dest_address":"0x0000000000000000000000000000000000000000","amount":"1","metadata":[]}],"imported_bridge_exits":[{"bridge_exit":{"leaf_type":"Transfer","token_info":null,"dest_network":0,"dest_address":"0x0000000000000000000000000000000000000000","amount":"1","metadata":[]},"claim_data":null,"global_index":{"mainnet_flag":false,"rollup_index":1,"leaf_index":1}}],"metadata":"0x0000000000000000000000000000000000000000000000000000000000000000","signature":{"r":"0x0000000000000000000000000000000000000000000000000000000000000000","s":"0x0000000000000000000000000000000000000000000000000000000000000000","odd_y_parity":false}}`
	expectedSignedCertificateyMetadataJSON     = `{"network_id":1,"height":1,"prev_local_exit_root":[0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0],"new_local_exit_root":[0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0],"bridge_exits":[{"leaf_type":"Transfer","token_info":null,"dest_network":0,"dest_address":"0x0000000000000000000000000000000000000000","amount":"1","metadata":[1,2,3]}],"imported_bridge_exits":[{"bridge_exit":{"leaf_type":"Transfer","token_info":null,"dest_network":0,"dest_address":"0x0000000000000000000000000000000000000000","amount":"1","metadata":[]},"claim_data":null,"global_index":{"mainnet_flag":false,"rollup_index":1,"leaf_index":1}}],"metadata":"0x0000000000000000000000000000000000000000000000000000000000000000","signature":{"r":"0x0000000000000000000000000000000000000000000000000000000000000000","s":"0x0000000000000000000000000000000000000000000000000000000000000000","odd_y_parity":false}}`
	expectedCertificateJSONEmptyFields         = `{"network_id":0,"height":0,"prev_local_exit_root":[0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0],"new_local_exit_root":[0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0],"bridge_exits":null,"imported_bridge_exits":null,"metadata":"0x0000000000000000000000000000000000000000000000000000000000000000","signature":{"r":"0x0000000000000000000000000000000000000000000000000000000000000000","s":"0x0000000000000000000000000000000000000000000000000000000000000000","odd_y_parity":false}}`
)

func TestMarshalJSONEmptyFields(t *testing.T) {
	cert := &SignedCertificate{
		Certificate: &Certificate{
			BridgeExits:         nil,
			ImportedBridgeExits: nil,
		},
		Signature: &Signature{},
	}
	data, err := json.Marshal(cert)
	require.NoError(t, err)
	log.Info(string(data))
	require.Equal(t, expectedCertificateJSONEmptyFields, string(data))
}

func TestMarshalJSON(t *testing.T) {
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
	require.Equal(t, expectedSignedCertificateyMetadataJSON, string(data))
}
