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

func TestExploratoryClient(t *testing.T) {
	t.Skip("This test is for exploratory purposes only")
	sut := NewAggLayerClient("http://127.0.0.1:32853")
	config, err := sut.GetEpochConfiguration()
	require.NoError(t, err)
	require.NotNil(t, config)
	fmt.Printf("Config: %s", config.String())
}

func TestGetEpochConfigurationResponseWithError(t *testing.T) {
	sut := NewAggLayerClient(testURL)
	response := rpc.Response{
		Error: &rpc.ErrorObject{},
	}
	jSONRPCCall = func(url, method string, params ...interface{}) (rpc.Response, error) {
		return response, nil
	}
	clockConfig, err := sut.GetEpochConfiguration()
	require.Nil(t, clockConfig)
	require.Error(t, err)
}

func TestGetEpochConfigurationResponseBadJson(t *testing.T) {
	sut := NewAggLayerClient(testURL)
	response := rpc.Response{
		Result: []byte(`{`),
	}
	jSONRPCCall = func(url, method string, params ...interface{}) (rpc.Response, error) {
		return response, nil
	}
	clockConfig, err := sut.GetEpochConfiguration()
	require.Nil(t, clockConfig)
	require.Error(t, err)
}

func TestGetEpochConfigurationErrorResponse(t *testing.T) {
	sut := NewAggLayerClient(testURL)

	jSONRPCCall = func(url, method string, params ...interface{}) (rpc.Response, error) {
		return rpc.Response{}, fmt.Errorf("unittest error")
	}
	clockConfig, err := sut.GetEpochConfiguration()
	require.Nil(t, clockConfig)
	require.Error(t, err)
}

func TestGetEpochConfigurationOkResponse(t *testing.T) {
	sut := NewAggLayerClient(testURL)
	response := rpc.Response{
		Result: []byte(`{"epoch_duration": 1, "genesis_block": 1}`),
	}
	jSONRPCCall = func(url, method string, params ...interface{}) (rpc.Response, error) {
		return response, nil
	}
	clockConfig, err := sut.GetEpochConfiguration()
	require.NotNil(t, clockConfig)
	require.NoError(t, err)
	require.Equal(t, ClockConfiguration{
		EpochDuration: 1,
		GenesisBlock:  1,
	}, *clockConfig)
}
