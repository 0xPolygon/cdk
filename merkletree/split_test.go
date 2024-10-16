package merkletree

import (
	"fmt"
	"math/big"
	"reflect"
	"testing"

	"github.com/0xPolygon/cdk/hex"
	"github.com/0xPolygon/cdk/test/testutils"
	"github.com/stretchr/testify/require"
)

func TestScalar2Fea(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected []uint64
	}{
		{
			name:     "Zero value",
			input:    "0",
			expected: []uint64{0, 0, 0, 0, 0, 0, 0, 0},
		},
		{
			name:     "Single 32-bit value",
			input:    "FFFFFFFF",
			expected: []uint64{0xFFFFFFFF, 0, 0, 0, 0, 0, 0, 0},
		},
		{
			name:     "Mixed bits across chunks (128-bit)",
			input:    "1234567890ABCDEF1234567890ABCDEF",
			expected: []uint64{0x90ABCDEF, 0x12345678, 0x90ABCDEF, 0x12345678, 0, 0, 0, 0},
		},
		{
			name:     "All bits set in each 32-bit chunk (256-bit)",
			input:    "FFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFF",
			expected: []uint64{0xFFFFFFFF, 0xFFFFFFFF, 0xFFFFFFFF, 0xFFFFFFFF, 0xFFFFFFFF, 0xFFFFFFFF, 0xFFFFFFFF, 0xFFFFFFFF},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			inputVal, success := new(big.Int).SetString(tt.input, 16)
			if !success {
				t.Fatalf("Invalid input value: %s", tt.input)
			}

			result := scalar2fea(inputVal)

			if !reflect.DeepEqual(result, tt.expected) {
				t.Errorf("scalar2fea(%s) = %v, want %v", tt.input, result, tt.expected)
			}
		})
	}
}

func Test_h4ToScalar(t *testing.T) {
	tcs := []struct {
		input    []uint64
		expected string
	}{
		{
			input:    []uint64{0, 0, 0, 0},
			expected: "0",
		},
		{
			input:    []uint64{0, 1, 2, 3},
			expected: "18831305206160042292187933003464876175252262292329349513216",
		},
	}

	for i, tc := range tcs {
		tc := tc
		t.Run(fmt.Sprintf("case %d", i), func(t *testing.T) {
			actual := h4ToScalar(tc.input)
			expected, ok := new(big.Int).SetString(tc.expected, 10)
			require.True(t, ok)
			require.Equal(t, expected, actual)
		})
	}
}

func Test_scalarToh4(t *testing.T) {
	tcs := []struct {
		input    string
		expected []uint64
	}{
		{
			input:    "0",
			expected: []uint64{0, 0, 0, 0},
		},
		{
			input:    "18831305206160042292187933003464876175252262292329349513216",
			expected: []uint64{0, 1, 2, 3},
		},
	}

	for i, tc := range tcs {
		tc := tc
		t.Run(fmt.Sprintf("case %d", i), func(t *testing.T) {
			bi, ok := new(big.Int).SetString(tc.input, 10)
			require.True(t, ok)

			actual := scalarToh4(bi)
			require.Equal(t, tc.expected, actual)
		})
	}
}

func Test_h4ToString(t *testing.T) {
	tcs := []struct {
		input    []uint64
		expected string
	}{
		{
			input:    []uint64{0, 0, 0, 0},
			expected: "0x0000000000000000000000000000000000000000000000000000000000000000",
		},
		{
			input:    []uint64{0, 1, 2, 3},
			expected: "0x0000000000000003000000000000000200000000000000010000000000000000",
		},
	}

	for i, tc := range tcs {
		tc := tc
		t.Run(fmt.Sprintf("case %d", i), func(t *testing.T) {
			actual := H4ToString(tc.input)
			require.Equal(t, tc.expected, actual)
		})
	}
}

func Test_Conversions(t *testing.T) {
	tcs := []struct {
		input []uint64
	}{
		{
			input: []uint64{0, 0, 0, 0},
		},
		{
			input: []uint64{0, 1, 2, 3},
		},
	}

	for i, tc := range tcs {
		tc := tc
		t.Run(fmt.Sprintf("case %d", i), func(t *testing.T) {
			resScalar := h4ToScalar(tc.input)
			init := scalarToh4(resScalar)
			require.Equal(t, tc.input, init)
		})
	}
}

func Test_stringToh4(t *testing.T) {
	tcs := []struct {
		description    string
		input          string
		expected       []uint64
		expectedErr    bool
		expectedErrMsg string
	}{
		{
			description: "happy path",
			input:       "cafe",
			expected:    []uint64{51966, 0, 0, 0},
		},
		{
			description: "0x prefix is allowed",
			input:       "0xcafe",
			expected:    []uint64{51966, 0, 0, 0},
		},

		{
			description:    "non hex input causes error",
			input:          "yu74",
			expectedErr:    true,
			expectedErrMsg: "could not convert \"yu74\" into big int",
		},
		{
			description:    "empty input causes error",
			input:          "",
			expectedErr:    true,
			expectedErrMsg: "could not convert \"\" into big int",
		},
	}

	for _, tc := range tcs {
		tc := tc
		t.Run(tc.description, func(t *testing.T) {
			actual, err := StringToh4(tc.input)
			require.NoError(t, testutils.CheckError(err, tc.expectedErr, tc.expectedErrMsg))

			require.Equal(t, tc.expected, actual)
		})
	}
}

func Test_ScalarToFilledByteSlice(t *testing.T) {
	tcs := []struct {
		input    string
		expected string
	}{
		{
			input:    "0",
			expected: "0x0000000000000000000000000000000000000000000000000000000000000000",
		},
		{
			input:    "256",
			expected: "0x0000000000000000000000000000000000000000000000000000000000000100",
		},
		{
			input:    "235938498573495379548793890390932048239042839490238",
			expected: "0x0000000000000000000000a16f882ee8972432c0a71c5e309ad5f7215690aebe",
		},
		{
			input:    "4309593458485959083095843905390485089430985490434080439904305093450934509490",
			expected: "0x098724b9a1bc97eee674cf5b6b56b8fafd83ac49c3da1f2c87c822548bbfdfb2",
		},
		{
			input:    "98999023430240239049320492430858334093493024832984092384902398409234090932489",
			expected: "0xdadf762a31e865f150a1456d7db7963c91361b771c8381a3fb879cf5bf91b909",
		},
	}

	for i, tc := range tcs {
		tc := tc
		t.Run(fmt.Sprintf("test case %d", i), func(t *testing.T) {
			input, ok := big.NewInt(0).SetString(tc.input, 10)
			require.True(t, ok)

			actualSlice := ScalarToFilledByteSlice(input)

			actual := hex.EncodeToHex(actualSlice)

			require.Equal(t, tc.expected, actual)
		})
	}
}

func Test_h4ToFilledByteSlice(t *testing.T) {
	tcs := []struct {
		input    []uint64
		expected string
	}{
		{
			input:    []uint64{0, 0, 0, 0},
			expected: "0x0000000000000000000000000000000000000000000000000000000000000000",
		},
		{
			input:    []uint64{0, 1, 2, 3},
			expected: "0x0000000000000003000000000000000200000000000000010000000000000000",
		},
		{
			input:    []uint64{55345354959, 991992992929, 2, 3},
			expected: "0x00000000000000030000000000000002000000e6f763d4a10000000ce2d718cf",
		},
		{
			input:    []uint64{8398349845894398543, 3485942349435495945, 734034022234249459, 5490434584389534589},
			expected: "0x4c31f12a390ec37d0a2fd00ddc52d8f330608e18f597e609748ceeb03ffe024f",
		},
	}

	for i, tc := range tcs {
		tc := tc
		t.Run(fmt.Sprintf("test case %d", i), func(t *testing.T) {
			actualSlice := h4ToFilledByteSlice(tc.input)

			actual := hex.EncodeToHex(actualSlice)

			require.Equal(t, tc.expected, actual)
		})
	}
}
