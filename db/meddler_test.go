package db

import (
	"testing"

	"github.com/ethereum/go-ethereum/common"
	"github.com/stretchr/testify/assert"
)

func TestHashMeddler_PreWrite(t *testing.T) {
	t.Parallel()

	hex := "0x1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef"
	hash := common.HexToHash(hex)

	tests := []struct {
		name      string
		fieldPtr  interface{}
		wantValue interface{}
		wantErr   bool
	}{
		{
			name:      "Valid common.Hash",
			fieldPtr:  hash,
			wantValue: hex,
			wantErr:   false,
		},
		{
			name:      "Valid *common.Hash",
			fieldPtr:  &hash,
			wantValue: hex,
			wantErr:   false,
		},
		{
			name:      "Nil *common.Hash",
			fieldPtr:  (*common.Hash)(nil),
			wantValue: []byte{},
			wantErr:   false,
		},
		{
			name:      "Invalid type",
			fieldPtr:  "invalid",
			wantValue: nil,
			wantErr:   true,
		},
	}

	for _, tt := range tests {
		tt := tt

		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			h := HashMeddler{}
			gotValue, err := h.PreWrite(tt.fieldPtr)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.wantValue, gotValue)
			}
		})
	}
}
