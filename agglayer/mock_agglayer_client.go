// Code generated by mockery. DO NOT EDIT.

package agglayer

import (
	context "context"

	common "github.com/ethereum/go-ethereum/common"

	mock "github.com/stretchr/testify/mock"
)

// AgglayerClientMock is an autogenerated mock type for the AgglayerClientInterface type
type AgglayerClientMock struct {
	mock.Mock
}

type AgglayerClientMock_Expecter struct {
	mock *mock.Mock
}

func (_m *AgglayerClientMock) EXPECT() *AgglayerClientMock_Expecter {
	return &AgglayerClientMock_Expecter{mock: &_m.Mock}
}

// GetCertificateHeader provides a mock function with given fields: certificateHash
func (_m *AgglayerClientMock) GetCertificateHeader(certificateHash common.Hash) (*CertificateHeader, error) {
	ret := _m.Called(certificateHash)

	if len(ret) == 0 {
		panic("no return value specified for GetCertificateHeader")
	}

	var r0 *CertificateHeader
	var r1 error
	if rf, ok := ret.Get(0).(func(common.Hash) (*CertificateHeader, error)); ok {
		return rf(certificateHash)
	}
	if rf, ok := ret.Get(0).(func(common.Hash) *CertificateHeader); ok {
		r0 = rf(certificateHash)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*CertificateHeader)
		}
	}

	if rf, ok := ret.Get(1).(func(common.Hash) error); ok {
		r1 = rf(certificateHash)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// AgglayerClientMock_GetCertificateHeader_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'GetCertificateHeader'
type AgglayerClientMock_GetCertificateHeader_Call struct {
	*mock.Call
}

// GetCertificateHeader is a helper method to define mock.On call
//   - certificateHash common.Hash
func (_e *AgglayerClientMock_Expecter) GetCertificateHeader(certificateHash interface{}) *AgglayerClientMock_GetCertificateHeader_Call {
	return &AgglayerClientMock_GetCertificateHeader_Call{Call: _e.mock.On("GetCertificateHeader", certificateHash)}
}

func (_c *AgglayerClientMock_GetCertificateHeader_Call) Run(run func(certificateHash common.Hash)) *AgglayerClientMock_GetCertificateHeader_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(common.Hash))
	})
	return _c
}

func (_c *AgglayerClientMock_GetCertificateHeader_Call) Return(_a0 *CertificateHeader, _a1 error) *AgglayerClientMock_GetCertificateHeader_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

func (_c *AgglayerClientMock_GetCertificateHeader_Call) RunAndReturn(run func(common.Hash) (*CertificateHeader, error)) *AgglayerClientMock_GetCertificateHeader_Call {
	_c.Call.Return(run)
	return _c
}

// GetEpochConfiguration provides a mock function with given fields:
func (_m *AgglayerClientMock) GetEpochConfiguration() (*ClockConfiguration, error) {
	ret := _m.Called()

	if len(ret) == 0 {
		panic("no return value specified for GetEpochConfiguration")
	}

	var r0 *ClockConfiguration
	var r1 error
	if rf, ok := ret.Get(0).(func() (*ClockConfiguration, error)); ok {
		return rf()
	}
	if rf, ok := ret.Get(0).(func() *ClockConfiguration); ok {
		r0 = rf()
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*ClockConfiguration)
		}
	}

	if rf, ok := ret.Get(1).(func() error); ok {
		r1 = rf()
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// AgglayerClientMock_GetEpochConfiguration_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'GetEpochConfiguration'
type AgglayerClientMock_GetEpochConfiguration_Call struct {
	*mock.Call
}

// GetEpochConfiguration is a helper method to define mock.On call
func (_e *AgglayerClientMock_Expecter) GetEpochConfiguration() *AgglayerClientMock_GetEpochConfiguration_Call {
	return &AgglayerClientMock_GetEpochConfiguration_Call{Call: _e.mock.On("GetEpochConfiguration")}
}

func (_c *AgglayerClientMock_GetEpochConfiguration_Call) Run(run func()) *AgglayerClientMock_GetEpochConfiguration_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run()
	})
	return _c
}

func (_c *AgglayerClientMock_GetEpochConfiguration_Call) Return(_a0 *ClockConfiguration, _a1 error) *AgglayerClientMock_GetEpochConfiguration_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

func (_c *AgglayerClientMock_GetEpochConfiguration_Call) RunAndReturn(run func() (*ClockConfiguration, error)) *AgglayerClientMock_GetEpochConfiguration_Call {
	_c.Call.Return(run)
	return _c
}

// GetLatestKnownCertificateHeader provides a mock function with given fields: networkId
func (_m *AgglayerClientMock) GetLatestKnownCertificateHeader(networkId uint32) (*CertificateHeader, error) {
	ret := _m.Called(networkId)

	if len(ret) == 0 {
		panic("no return value specified for GetLatestKnownCertificateHeader")
	}

	var r0 *CertificateHeader
	var r1 error
	if rf, ok := ret.Get(0).(func(uint32) (*CertificateHeader, error)); ok {
		return rf(networkId)
	}
	if rf, ok := ret.Get(0).(func(uint32) *CertificateHeader); ok {
		r0 = rf(networkId)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*CertificateHeader)
		}
	}

	if rf, ok := ret.Get(1).(func(uint32) error); ok {
		r1 = rf(networkId)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// AgglayerClientMock_GetLatestKnownCertificateHeader_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'GetLatestKnownCertificateHeader'
type AgglayerClientMock_GetLatestKnownCertificateHeader_Call struct {
	*mock.Call
}

// GetLatestKnownCertificateHeader is a helper method to define mock.On call
//   - networkId uint32
func (_e *AgglayerClientMock_Expecter) GetLatestKnownCertificateHeader(networkId interface{}) *AgglayerClientMock_GetLatestKnownCertificateHeader_Call {
	return &AgglayerClientMock_GetLatestKnownCertificateHeader_Call{Call: _e.mock.On("GetLatestKnownCertificateHeader", networkId)}
}

func (_c *AgglayerClientMock_GetLatestKnownCertificateHeader_Call) Run(run func(networkId uint32)) *AgglayerClientMock_GetLatestKnownCertificateHeader_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(uint32))
	})
	return _c
}

func (_c *AgglayerClientMock_GetLatestKnownCertificateHeader_Call) Return(_a0 *CertificateHeader, _a1 error) *AgglayerClientMock_GetLatestKnownCertificateHeader_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

func (_c *AgglayerClientMock_GetLatestKnownCertificateHeader_Call) RunAndReturn(run func(uint32) (*CertificateHeader, error)) *AgglayerClientMock_GetLatestKnownCertificateHeader_Call {
	_c.Call.Return(run)
	return _c
}

// SendCertificate provides a mock function with given fields: certificate
func (_m *AgglayerClientMock) SendCertificate(certificate *SignedCertificate) (common.Hash, error) {
	ret := _m.Called(certificate)

	if len(ret) == 0 {
		panic("no return value specified for SendCertificate")
	}

	var r0 common.Hash
	var r1 error
	if rf, ok := ret.Get(0).(func(*SignedCertificate) (common.Hash, error)); ok {
		return rf(certificate)
	}
	if rf, ok := ret.Get(0).(func(*SignedCertificate) common.Hash); ok {
		r0 = rf(certificate)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(common.Hash)
		}
	}

	if rf, ok := ret.Get(1).(func(*SignedCertificate) error); ok {
		r1 = rf(certificate)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// AgglayerClientMock_SendCertificate_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'SendCertificate'
type AgglayerClientMock_SendCertificate_Call struct {
	*mock.Call
}

// SendCertificate is a helper method to define mock.On call
//   - certificate *SignedCertificate
func (_e *AgglayerClientMock_Expecter) SendCertificate(certificate interface{}) *AgglayerClientMock_SendCertificate_Call {
	return &AgglayerClientMock_SendCertificate_Call{Call: _e.mock.On("SendCertificate", certificate)}
}

func (_c *AgglayerClientMock_SendCertificate_Call) Run(run func(certificate *SignedCertificate)) *AgglayerClientMock_SendCertificate_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(*SignedCertificate))
	})
	return _c
}

func (_c *AgglayerClientMock_SendCertificate_Call) Return(_a0 common.Hash, _a1 error) *AgglayerClientMock_SendCertificate_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

func (_c *AgglayerClientMock_SendCertificate_Call) RunAndReturn(run func(*SignedCertificate) (common.Hash, error)) *AgglayerClientMock_SendCertificate_Call {
	_c.Call.Return(run)
	return _c
}

// SendTx provides a mock function with given fields: signedTx
func (_m *AgglayerClientMock) SendTx(signedTx SignedTx) (common.Hash, error) {
	ret := _m.Called(signedTx)

	if len(ret) == 0 {
		panic("no return value specified for SendTx")
	}

	var r0 common.Hash
	var r1 error
	if rf, ok := ret.Get(0).(func(SignedTx) (common.Hash, error)); ok {
		return rf(signedTx)
	}
	if rf, ok := ret.Get(0).(func(SignedTx) common.Hash); ok {
		r0 = rf(signedTx)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(common.Hash)
		}
	}

	if rf, ok := ret.Get(1).(func(SignedTx) error); ok {
		r1 = rf(signedTx)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// AgglayerClientMock_SendTx_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'SendTx'
type AgglayerClientMock_SendTx_Call struct {
	*mock.Call
}

// SendTx is a helper method to define mock.On call
//   - signedTx SignedTx
func (_e *AgglayerClientMock_Expecter) SendTx(signedTx interface{}) *AgglayerClientMock_SendTx_Call {
	return &AgglayerClientMock_SendTx_Call{Call: _e.mock.On("SendTx", signedTx)}
}

func (_c *AgglayerClientMock_SendTx_Call) Run(run func(signedTx SignedTx)) *AgglayerClientMock_SendTx_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(SignedTx))
	})
	return _c
}

func (_c *AgglayerClientMock_SendTx_Call) Return(_a0 common.Hash, _a1 error) *AgglayerClientMock_SendTx_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

func (_c *AgglayerClientMock_SendTx_Call) RunAndReturn(run func(SignedTx) (common.Hash, error)) *AgglayerClientMock_SendTx_Call {
	_c.Call.Return(run)
	return _c
}

// WaitTxToBeMined provides a mock function with given fields: hash, ctx
func (_m *AgglayerClientMock) WaitTxToBeMined(hash common.Hash, ctx context.Context) error {
	ret := _m.Called(hash, ctx)

	if len(ret) == 0 {
		panic("no return value specified for WaitTxToBeMined")
	}

	var r0 error
	if rf, ok := ret.Get(0).(func(common.Hash, context.Context) error); ok {
		r0 = rf(hash, ctx)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// AgglayerClientMock_WaitTxToBeMined_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'WaitTxToBeMined'
type AgglayerClientMock_WaitTxToBeMined_Call struct {
	*mock.Call
}

// WaitTxToBeMined is a helper method to define mock.On call
//   - hash common.Hash
//   - ctx context.Context
func (_e *AgglayerClientMock_Expecter) WaitTxToBeMined(hash interface{}, ctx interface{}) *AgglayerClientMock_WaitTxToBeMined_Call {
	return &AgglayerClientMock_WaitTxToBeMined_Call{Call: _e.mock.On("WaitTxToBeMined", hash, ctx)}
}

func (_c *AgglayerClientMock_WaitTxToBeMined_Call) Run(run func(hash common.Hash, ctx context.Context)) *AgglayerClientMock_WaitTxToBeMined_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(common.Hash), args[1].(context.Context))
	})
	return _c
}

func (_c *AgglayerClientMock_WaitTxToBeMined_Call) Return(_a0 error) *AgglayerClientMock_WaitTxToBeMined_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *AgglayerClientMock_WaitTxToBeMined_Call) RunAndReturn(run func(common.Hash, context.Context) error) *AgglayerClientMock_WaitTxToBeMined_Call {
	_c.Call.Return(run)
	return _c
}

// NewAgglayerClientMock creates a new instance of AgglayerClientMock. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
// The first argument is typically a *testing.T value.
func NewAgglayerClientMock(t interface {
	mock.TestingT
	Cleanup(func())
}) *AgglayerClientMock {
	mock := &AgglayerClientMock{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
