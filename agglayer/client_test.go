package agglayer

import (
	"fmt"
	"testing"

	"github.com/0xPolygon/cdk-rpc/rpc"
	"github.com/stretchr/testify/require"
)

const (
	testURL = "http://localhost:8080"
)

func TestGetClockConfigurationResponseWithError(t *testing.T) {
	sut := NewAggLayerClient(testURL)
	response := rpc.Response{
		Error: &rpc.ErrorObject{},
	}
	jSONRPCCall = func(url, method string, params ...interface{}) (rpc.Response, error) {
		return response, nil
	}
	clockConfig, err := sut.GetClockConfiguration()
	require.Nil(t, clockConfig)
	require.Error(t, err)
}

func TestGetClockConfigurationResponseBadJson(t *testing.T) {
	sut := NewAggLayerClient(testURL)
	response := rpc.Response{
		Result: []byte(`{`),
	}
	jSONRPCCall = func(url, method string, params ...interface{}) (rpc.Response, error) {
		return response, nil
	}
	clockConfig, err := sut.GetClockConfiguration()
	require.Nil(t, clockConfig)
	require.Error(t, err)
}

func TestGetClockConfigurationErrorResponse(t *testing.T) {
	sut := NewAggLayerClient(testURL)

	jSONRPCCall = func(url, method string, params ...interface{}) (rpc.Response, error) {
		return rpc.Response{}, fmt.Errorf("unittest error")
	}
	clockConfig, err := sut.GetClockConfiguration()
	require.Nil(t, clockConfig)
	require.Error(t, err)
}

func TestGetClockConfigurationOkResponse(t *testing.T) {
	sut := NewAggLayerClient(testURL)
	response := rpc.Response{
		Result: []byte(`{"epoch_duration": 1, "genesis_block": 1}`),
	}
	jSONRPCCall = func(url, method string, params ...interface{}) (rpc.Response, error) {
		return response, nil
	}
	clockConfig, err := sut.GetClockConfiguration()
	require.NotNil(t, clockConfig)
	require.NoError(t, err)
	require.Equal(t, ClockConfiguration{
		EpochDuration: 1,
		GenesisBlock:  1,
	}, *clockConfig)
}
