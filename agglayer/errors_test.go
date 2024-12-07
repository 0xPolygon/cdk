package agglayer

import (
	"testing"

	"github.com/stretchr/testify/require"
)

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
