package common

import (
	"fmt"
	"math/big"
	"testing"

	"github.com/ethereum/go-ethereum/common"
	"github.com/stretchr/testify/require"
)

func TestAsLittleEndianSlice(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		input    *big.Int
		expected []byte
	}{
		{
			name:     "Zero value",
			input:    big.NewInt(0),
			expected: make([]byte, 32),
		},
		{
			name:     "Positive value",
			input:    big.NewInt(123456789),
			expected: append([]byte{21, 205, 91, 7}, make([]byte, 28)...),
		},
		{
			name:     "Negative value",
			input:    big.NewInt(-123456789),
			expected: append([]byte{21, 205, 91, 7}, make([]byte, 28)...),
		},
		{
			name: "Large positive value",
			input: new(big.Int).SetBytes([]byte{0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff,
				0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff,
				0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff}),
			expected: []byte{0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff,
				0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff,
				0xff, 0xff, 0xff, 0xff, 0xff, 0xff},
		},
	}

	for _, tt := range tests {
		tt := tt

		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			result := BigIntToLittleEndianBytes(tt.input)
			require.Len(t, result, common.HashLength)

			for i := range result {
				require.Equal(t, tt.expected[i], result[i],
					fmt.Sprintf("expected byte at index %d to be %x, got %x", i, tt.expected[i], result[i]))
			}
		})
	}
}
