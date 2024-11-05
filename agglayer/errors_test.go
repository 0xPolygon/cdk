package agglayer

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestErrorVectors(t *testing.T) {
	t.Parallel()

	type testCase struct {
		TestName              string `json:"test_name"`
		ExpectedError         string `json:"expected_error"`
		CertificateHeaderJSON string `json:"certificate_header"`
	}

	files, err := filepath.Glob("testdata/*/*.json")
	require.NoError(t, err)

	for _, file := range files {
		file := file

		t.Run(file, func(t *testing.T) {
			t.Parallel()

			data, err := os.ReadFile(file)
			require.NoError(t, err)

			var testCases []*testCase

			require.NoError(t, json.Unmarshal(data, &testCases))

			for _, tc := range testCases {
				certificateHeader := &CertificateHeader{}
				err = json.Unmarshal([]byte(tc.CertificateHeaderJSON), certificateHeader)

				if tc.ExpectedError == "" {
					require.NoError(t, err, "Test: %s not expected any unmarshal error, but got: %v", tc.TestName, err)
					require.NotNil(t, certificateHeader.Error, "Test: %s unpacked error is nil", tc.TestName)
					fmt.Println(certificateHeader.Error.String())
				} else {
					require.ErrorContains(t, err, tc.ExpectedError, "Test: %s expected error: %s. Got: %v", tc.TestName, tc.ExpectedError, err)
				}
			}
		})
	}
}

func TestConvertMapValue_String(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		data      map[string]interface{}
		key       string
		want      string
		errString string
	}{
		{
			name: "Key exists and type matches",
			data: map[string]interface{}{
				"key1": "value1",
			},
			key:  "key1",
			want: "value1",
		},
		{
			name: "Key exists but type does not match",
			data: map[string]interface{}{
				"key1": 1,
			},
			key:       "key1",
			want:      "",
			errString: "is not of type",
		},
		{
			name: "Key does not exist",
			data: map[string]interface{}{
				"key1": "value1",
			},
			key:       "key2",
			want:      "",
			errString: "key key2 not found in map",
		},
	}

	for _, tt := range tests {
		tt := tt

		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got, err := convertMapValue[string](tt.data, tt.key)
			if tt.errString != "" {
				require.ErrorContains(t, err, tt.errString)
			} else {
				require.Equal(t, tt.want, got)
			}
		})
	}
}

//nolint:dupl
func TestConvertMapValue_Uint32(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		data      map[string]interface{}
		key       string
		want      uint32
		errString string
	}{
		{
			name: "Key exists and type matches",
			data: map[string]interface{}{
				"key1": uint32(123),
			},
			key:  "key1",
			want: uint32(123),
		},
		{
			name: "Key exists but type does not match",
			data: map[string]interface{}{
				"key1": "value1",
			},
			key:       "key1",
			want:      0,
			errString: "is not of type",
		},
		{
			name: "Key does not exist",
			data: map[string]interface{}{
				"key1": uint32(123),
			},
			key:       "key2",
			want:      0,
			errString: "key key2 not found in map",
		},
	}

	for _, tt := range tests {
		tt := tt

		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got, err := convertMapValue[uint32](tt.data, tt.key)
			if tt.errString != "" {
				require.ErrorContains(t, err, tt.errString)
			} else {
				require.Equal(t, tt.want, got)
			}
		})
	}
}

//nolint:dupl
func TestConvertMapValue_Uint64(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		data      map[string]interface{}
		key       string
		want      uint64
		errString string
	}{
		{
			name: "Key exists and type matches",
			data: map[string]interface{}{
				"key1": uint64(3411),
			},
			key:  "key1",
			want: uint64(3411),
		},
		{
			name: "Key exists but type does not match",
			data: map[string]interface{}{
				"key1": "not a number",
			},
			key:       "key1",
			want:      0,
			errString: "is not of type",
		},
		{
			name: "Key does not exist",
			data: map[string]interface{}{
				"key1": uint64(123555),
			},
			key:       "key22",
			want:      0,
			errString: "key key22 not found in map",
		},
	}

	for _, tt := range tests {
		tt := tt

		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got, err := convertMapValue[uint64](tt.data, tt.key)
			if tt.errString != "" {
				require.ErrorContains(t, err, tt.errString)
			} else {
				require.Equal(t, tt.want, got)
			}
		})
	}
}

func TestConvertMapValue_Bool(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		data      map[string]interface{}
		key       string
		want      bool
		errString string
	}{
		{
			name: "Key exists and type matches",
			data: map[string]interface{}{
				"key1": true,
			},
			key:  "key1",
			want: true,
		},
		{
			name: "Key exists but type does not match",
			data: map[string]interface{}{
				"key1": "value1",
			},
			key:       "key1",
			want:      false,
			errString: "is not of type",
		},
		{
			name: "Key does not exist",
			data: map[string]interface{}{
				"key1": true,
			},
			key:       "key2",
			want:      false,
			errString: "key key2 not found in map",
		},
	}

	for _, tt := range tests {
		tt := tt

		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got, err := convertMapValue[bool](tt.data, tt.key)
			if tt.errString != "" {
				require.ErrorContains(t, err, tt.errString)
			} else {
				require.Equal(t, tt.want, got)
			}
		})
	}
}
