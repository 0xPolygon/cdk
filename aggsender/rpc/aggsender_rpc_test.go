package aggsenderrpc

import (
	"fmt"
	"testing"

	"github.com/0xPolygon/cdk/aggsender/mocks"
	"github.com/0xPolygon/cdk/aggsender/types"
	"github.com/stretchr/testify/require"
)

func TestAggsenderRPCStatus(t *testing.T) {
	testData := newAggsenderData(t)
	testData.mockAggsender.EXPECT().Info().Return(types.AggsenderInfo{})
	res, err := testData.sut.Status()
	require.NoError(t, err)
	require.NotNil(t, res)
}

func TestAggsenderRPCGetCertificateHeaderPerHeight(t *testing.T) {
	testData := newAggsenderData(t)
	height := uint64(1)
	cases := []struct {
		name          string
		height        *uint64
		certResult    *types.CertificateInfo
		certError     error
		expectedError string
		expectedNil   bool
	}{
		{
			name:       "latest, no error",
			certResult: &types.CertificateInfo{},
			certError:  nil,
		},
		{
			name:          "latest,no error, no cert",
			certResult:    nil,
			certError:     nil,
			expectedError: "not found",
			expectedNil:   true,
		},
		{
			name:          "latest,error",
			certResult:    &types.CertificateInfo{},
			certError:     fmt.Errorf("my_error"),
			expectedError: "my_error",
			expectedNil:   true,
		},
		{
			name:       "hight, no error",
			height:     &height,
			certResult: &types.CertificateInfo{},
			certError:  nil,
		},
	}

	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			if tt.height == nil {
				testData.mockStore.EXPECT().GetLastSentCertificate().Return(tt.certResult, tt.certError).Once()
			} else {
				testData.mockStore.EXPECT().GetCertificateByHeight(*tt.height).Return(tt.certResult, tt.certError).Once()
			}
			res, err := testData.sut.GetCertificateHeaderPerHeight(tt.height)
			if tt.expectedError != "" {
				require.Error(t, err)
				require.Contains(t, err.Error(), tt.expectedError)
			} else {
				require.NoError(t, err)
			}
			if tt.expectedNil {
				require.Nil(t, res)
			} else {
				require.NotNil(t, res)
			}
		})
	}
}

type aggsenderRPCTestData struct {
	sut           *AggsenderRPC
	mockStore     *mocks.AggsenderStorer
	mockAggsender *mocks.AggsenderInterface
}

func newAggsenderData(t *testing.T) *aggsenderRPCTestData {
	t.Helper()
	mockStore := mocks.NewAggsenderStorer(t)
	mockAggsender := mocks.NewAggsenderInterface(t)
	sut := NewAggsenderRPC(nil, mockStore, mockAggsender)
	return &aggsenderRPCTestData{sut, mockStore, mockAggsender}
}
