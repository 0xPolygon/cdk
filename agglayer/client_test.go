package agglayer

import (
	"fmt"
	"testing"

	"github.com/0xPolygon/cdk-rpc/rpc"
	"github.com/ethereum/go-ethereum/common"
	"github.com/stretchr/testify/require"
)

const (
	testURL = "http://localhost:8080"
)

func TestExploratoryClient(t *testing.T) {
	t.Skip("This test is for exploratory purposes only")
	sut := NewAggLayerClient("http://127.0.0.1:32781")
	config, err := sut.GetEpochConfiguration()
	require.NoError(t, err)
	require.NotNil(t, config)
	fmt.Printf("Config: %s", config.String())

	lastCert, err := sut.GetLatestKnownCertificateHeader(1)
	require.NoError(t, err)
	require.NotNil(t, lastCert)
	fmt.Printf("LastCert: %s", lastCert.String())
}

func TestExploratoryGetCertificateHeader(t *testing.T) {
	t.Skip("This test is exploratory and should be skipped")
	aggLayerClient := NewAggLayerClient("http://localhost:32796")
	certificateID := common.HexToHash("0xf153e75e24591432ac5deafaeaafba3fec0fd851261c86051b9c0d540b38c369")
	certificateHeader, err := aggLayerClient.GetCertificateHeader(certificateID)
	require.NoError(t, err)
	fmt.Print(certificateHeader)
}
func TestExploratoryGetEpochConfiguration(t *testing.T) {
	t.Skip("This test is exploratory and should be skipped")
	aggLayerClient := NewAggLayerClient("http://localhost:32796")
	clockConfig, err := aggLayerClient.GetEpochConfiguration()
	require.NoError(t, err)
	fmt.Print(clockConfig)
}

func TestExploratoryGetLatestKnownCertificateHeader(t *testing.T) {
	t.Skip("This test is exploratory and should be skipped")
	aggLayerClient := NewAggLayerClient("http://localhost:32843")
	cert, err := aggLayerClient.GetLatestKnownCertificateHeader(1)
	require.NoError(t, err)
	fmt.Print(cert)
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

func TestGetLatestKnownCertificateHeaderOkResponse(t *testing.T) {
	sut := NewAggLayerClient(testURL)
	response := rpc.Response{
		Result: []byte(`{"network_id":1,"height":0,"epoch_number":223,"certificate_index":0,"certificate_id":"0xf9179d2fbe535814b5a14496e2eed474f49c6131227a9dfc5d2d8caf9e212054","new_local_exit_root":"0x7ae06f4a5d0b6da7dd4973fb6ef40d82c9f2680899b3baaf9e564413b59cc160","metadata":"0x00000000000000000000000000000000000000000000000000000000000001a7","status":"Settled"}`),
	}
	jSONRPCCall = func(url, method string, params ...interface{}) (rpc.Response, error) {
		return response, nil
	}
	cert, err := sut.GetLatestKnownCertificateHeader(1)
	require.NotNil(t, cert)
	require.NoError(t, err)
	require.Nil(t, cert.PreviousLocalExitRoot)
}

func TestGetLatestKnownCertificateHeaderErrorResponse(t *testing.T) {
	sut := NewAggLayerClient(testURL)
	jSONRPCCall = func(url, method string, params ...interface{}) (rpc.Response, error) {
		return rpc.Response{}, fmt.Errorf("unittest error")
	}

	cert, err := sut.GetLatestKnownCertificateHeader(1)

	require.Nil(t, cert)
	require.Error(t, err)
}

func TestGetLatestKnownCertificateHeaderResponseBadJson(t *testing.T) {
	sut := NewAggLayerClient(testURL)
	response := rpc.Response{
		Result: []byte(`{`),
	}
	jSONRPCCall = func(url, method string, params ...interface{}) (rpc.Response, error) {
		return response, nil
	}

	cert, err := sut.GetLatestKnownCertificateHeader(1)

	require.Nil(t, cert)
	require.Error(t, err)
}

func TestGetLatestKnownCertificateHeaderWithPrevLERResponse(t *testing.T) {
	sut := NewAggLayerClient(testURL)
	response := rpc.Response{
		Result: []byte(`{"network_id":1,"height":0,"epoch_number":223,"certificate_index":0,"certificate_id":"0xf9179d2fbe535814b5a14496e2eed474f49c6131227a9dfc5d2d8caf9e212054","prev_local_exit_root":"0x27ae5ba08d7291c96c8cbddcc148bf48a6d68c7974b94356f53754ef6171d757","new_local_exit_root":"0x7ae06f4a5d0b6da7dd4973fb6ef40d82c9f2680899b3baaf9e564413b59cc160","metadata":"0x00000000000000000000000000000000000000000000000000000000000001a7","status":"Settled"}`),
	}
	jSONRPCCall = func(_, _ string, _ ...interface{}) (rpc.Response, error) {
		return response, nil
	}
	cert, err := sut.GetLatestKnownCertificateHeader(1)

	require.NoError(t, err)
	require.NotNil(t, cert)

	require.Equal(t, "0x27ae5ba08d7291c96c8cbddcc148bf48a6d68c7974b94356f53754ef6171d757", cert.PreviousLocalExitRoot.String())
}
