package rpc

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/0xPolygon/cdk-rpc/rpc"
	"github.com/ethereum/go-ethereum/common"
	"github.com/stretchr/testify/require"
)

func Test_getBatchFromRPC(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name                 string
		batch                uint64
		getBatchByNumberResp string
		getBlockByHasResp    string
		getBatchByNumberErr  error
		getBlockByHashErr    error
		expectBlocks         int
		expectData           []byte
		expectTimestamp      uint64
		expectErr            error
	}{
		{
			name:                 "successfully fetched",
			getBatchByNumberResp: `{"jsonrpc":"2.0","id":1,"result":{"blocks":["1", "2", "3"],"batchL2Data":"0x1234567"}}`,
			getBlockByHasResp:    `{"jsonrpc":"2.0","id":1,"result":{"timestamp":"0x123456"}}`,
			batch:                0,
			expectBlocks:         3,
			expectData:           common.FromHex("0x1234567"),
			expectTimestamp:      1193046,
			expectErr:            nil,
		},
		{
			name:                 "invalid json",
			getBatchByNumberResp: `{"jsonrpc":"2.0","id":1,"result":{"blocks":invalid,"batchL2Data":"test"}}`,
			batch:                0,
			expectBlocks:         3,
			expectData:           nil,
			expectErr:            errors.New("invalid character 'i' looking for beginning of value"),
		},
		{
			name:                 "wrong json",
			getBatchByNumberResp: `{"jsonrpc":"2.0","id":1,"result":{"blocks":"invalid","batchL2Data":"test"}}`,
			batch:                0,
			expectBlocks:         3,
			expectData:           nil,
			expectErr:            errors.New("error unmarshalling the batch from the response calling zkevm_getBatchByNumber: json: cannot unmarshal string into Go struct field zkEVMBatch.blocks of type []string"),
		},
		{
			name:                 "error in the response",
			getBatchByNumberResp: `{"jsonrpc":"2.0","id":1,"result":null,"error":{"code":-32602,"message":"Invalid params"}}`,
			batch:                0,
			expectBlocks:         0,
			expectData:           nil,
			expectErr:            errors.New("error in the response calling zkevm_getBatchByNumber: &{-32602 Invalid params <nil>}"),
		},
		{
			name:                "http failed",
			getBatchByNumberErr: errors.New("failed to fetch"),
			batch:               0,
			expectBlocks:        0,
			expectData:          nil,
			expectErr:           errors.New("invalid status code, expected: 200, found: 500"),
		},
	}

	for _, tt := range tests {
		tt := tt

		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				var req rpc.Request
				err := json.NewDecoder(r.Body).Decode(&req)
				require.NoError(t, err)

				switch req.Method {
				case "zkevm_getBatchByNumber":
					if tt.getBatchByNumberErr != nil {
						http.Error(w, tt.getBatchByNumberErr.Error(), http.StatusInternalServerError)
						return
					}

					_, _ = w.Write([]byte(tt.getBatchByNumberResp))
				case "eth_getBlockByHash":
					if tt.getBlockByHashErr != nil {
						http.Error(w, tt.getBlockByHashErr.Error(), http.StatusInternalServerError)
						return
					}
					_, _ = w.Write([]byte(tt.getBlockByHasResp))
				default:
					http.Error(w, "method not found", http.StatusNotFound)
				}
			}))
			defer srv.Close()

			rcpBatchClient := NewBatchEndpoints(srv.URL)
			rpcBatch, err := rcpBatchClient.GetBatch(tt.batch)
			if tt.expectErr != nil {
				require.Equal(t, tt.expectErr.Error(), err.Error())
			} else {
				require.NoError(t, err)
				require.Equal(t, tt.expectTimestamp, rpcBatch.LastL2BLockTimestamp())
				require.Equal(t, tt.expectData, rpcBatch.L2Data())
			}
		})
	}
}
