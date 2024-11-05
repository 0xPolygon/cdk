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
)

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
