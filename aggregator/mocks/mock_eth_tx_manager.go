// Code generated by mockery. DO NOT EDIT.

package mocks

import (
	context "context"
	big "math/big"

	common "github.com/ethereum/go-ethereum/common"

	ethtxmanager "github.com/0xPolygon/zkevm-ethtx-manager/ethtxmanager"

	kzg4844 "github.com/ethereum/go-ethereum/crypto/kzg4844"

	mock "github.com/stretchr/testify/mock"

	types "github.com/ethereum/go-ethereum/core/types"

	zkevm_ethtx_managertypes "github.com/0xPolygon/zkevm-ethtx-manager/types"
)

// EthTxManagerClientMock is an autogenerated mock type for the EthTxManagerClient type
type EthTxManagerClientMock struct {
	mock.Mock
}

type EthTxManagerClientMock_Expecter struct {
	mock *mock.Mock
}

func (_m *EthTxManagerClientMock) EXPECT() *EthTxManagerClientMock_Expecter {
	return &EthTxManagerClientMock_Expecter{mock: &_m.Mock}
}

// Add provides a mock function with given fields: ctx, to, value, data, gasOffset, sidecar
func (_m *EthTxManagerClientMock) Add(ctx context.Context, to *common.Address, value *big.Int, data []byte, gasOffset uint64, sidecar *types.BlobTxSidecar) (common.Hash, error) {
	ret := _m.Called(ctx, to, value, data, gasOffset, sidecar)

	if len(ret) == 0 {
		panic("no return value specified for Add")
	}

	var r0 common.Hash
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, *common.Address, *big.Int, []byte, uint64, *types.BlobTxSidecar) (common.Hash, error)); ok {
		return rf(ctx, to, value, data, gasOffset, sidecar)
	}
	if rf, ok := ret.Get(0).(func(context.Context, *common.Address, *big.Int, []byte, uint64, *types.BlobTxSidecar) common.Hash); ok {
		r0 = rf(ctx, to, value, data, gasOffset, sidecar)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(common.Hash)
		}
	}

	if rf, ok := ret.Get(1).(func(context.Context, *common.Address, *big.Int, []byte, uint64, *types.BlobTxSidecar) error); ok {
		r1 = rf(ctx, to, value, data, gasOffset, sidecar)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// EthTxManagerClientMock_Add_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Add'
type EthTxManagerClientMock_Add_Call struct {
	*mock.Call
}

// Add is a helper method to define mock.On call
//   - ctx context.Context
//   - to *common.Address
//   - value *big.Int
//   - data []byte
//   - gasOffset uint64
//   - sidecar *types.BlobTxSidecar
func (_e *EthTxManagerClientMock_Expecter) Add(ctx interface{}, to interface{}, value interface{}, data interface{}, gasOffset interface{}, sidecar interface{}) *EthTxManagerClientMock_Add_Call {
	return &EthTxManagerClientMock_Add_Call{Call: _e.mock.On("Add", ctx, to, value, data, gasOffset, sidecar)}
}

func (_c *EthTxManagerClientMock_Add_Call) Run(run func(ctx context.Context, to *common.Address, value *big.Int, data []byte, gasOffset uint64, sidecar *types.BlobTxSidecar)) *EthTxManagerClientMock_Add_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context), args[1].(*common.Address), args[2].(*big.Int), args[3].([]byte), args[4].(uint64), args[5].(*types.BlobTxSidecar))
	})
	return _c
}

func (_c *EthTxManagerClientMock_Add_Call) Return(_a0 common.Hash, _a1 error) *EthTxManagerClientMock_Add_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

func (_c *EthTxManagerClientMock_Add_Call) RunAndReturn(run func(context.Context, *common.Address, *big.Int, []byte, uint64, *types.BlobTxSidecar) (common.Hash, error)) *EthTxManagerClientMock_Add_Call {
	_c.Call.Return(run)
	return _c
}

// AddWithGas provides a mock function with given fields: ctx, to, value, data, gasOffset, sidecar, gas
func (_m *EthTxManagerClientMock) AddWithGas(ctx context.Context, to *common.Address, value *big.Int, data []byte, gasOffset uint64, sidecar *types.BlobTxSidecar, gas uint64) (common.Hash, error) {
	ret := _m.Called(ctx, to, value, data, gasOffset, sidecar, gas)

	if len(ret) == 0 {
		panic("no return value specified for AddWithGas")
	}

	var r0 common.Hash
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, *common.Address, *big.Int, []byte, uint64, *types.BlobTxSidecar, uint64) (common.Hash, error)); ok {
		return rf(ctx, to, value, data, gasOffset, sidecar, gas)
	}
	if rf, ok := ret.Get(0).(func(context.Context, *common.Address, *big.Int, []byte, uint64, *types.BlobTxSidecar, uint64) common.Hash); ok {
		r0 = rf(ctx, to, value, data, gasOffset, sidecar, gas)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(common.Hash)
		}
	}

	if rf, ok := ret.Get(1).(func(context.Context, *common.Address, *big.Int, []byte, uint64, *types.BlobTxSidecar, uint64) error); ok {
		r1 = rf(ctx, to, value, data, gasOffset, sidecar, gas)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// EthTxManagerClientMock_AddWithGas_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'AddWithGas'
type EthTxManagerClientMock_AddWithGas_Call struct {
	*mock.Call
}

// AddWithGas is a helper method to define mock.On call
//   - ctx context.Context
//   - to *common.Address
//   - value *big.Int
//   - data []byte
//   - gasOffset uint64
//   - sidecar *types.BlobTxSidecar
//   - gas uint64
func (_e *EthTxManagerClientMock_Expecter) AddWithGas(ctx interface{}, to interface{}, value interface{}, data interface{}, gasOffset interface{}, sidecar interface{}, gas interface{}) *EthTxManagerClientMock_AddWithGas_Call {
	return &EthTxManagerClientMock_AddWithGas_Call{Call: _e.mock.On("AddWithGas", ctx, to, value, data, gasOffset, sidecar, gas)}
}

func (_c *EthTxManagerClientMock_AddWithGas_Call) Run(run func(ctx context.Context, to *common.Address, value *big.Int, data []byte, gasOffset uint64, sidecar *types.BlobTxSidecar, gas uint64)) *EthTxManagerClientMock_AddWithGas_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context), args[1].(*common.Address), args[2].(*big.Int), args[3].([]byte), args[4].(uint64), args[5].(*types.BlobTxSidecar), args[6].(uint64))
	})
	return _c
}

func (_c *EthTxManagerClientMock_AddWithGas_Call) Return(_a0 common.Hash, _a1 error) *EthTxManagerClientMock_AddWithGas_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

func (_c *EthTxManagerClientMock_AddWithGas_Call) RunAndReturn(run func(context.Context, *common.Address, *big.Int, []byte, uint64, *types.BlobTxSidecar, uint64) (common.Hash, error)) *EthTxManagerClientMock_AddWithGas_Call {
	_c.Call.Return(run)
	return _c
}

// EncodeBlobData provides a mock function with given fields: data
func (_m *EthTxManagerClientMock) EncodeBlobData(data []byte) (kzg4844.Blob, error) {
	ret := _m.Called(data)

	if len(ret) == 0 {
		panic("no return value specified for EncodeBlobData")
	}

	var r0 kzg4844.Blob
	var r1 error
	if rf, ok := ret.Get(0).(func([]byte) (kzg4844.Blob, error)); ok {
		return rf(data)
	}
	if rf, ok := ret.Get(0).(func([]byte) kzg4844.Blob); ok {
		r0 = rf(data)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(kzg4844.Blob)
		}
	}

	if rf, ok := ret.Get(1).(func([]byte) error); ok {
		r1 = rf(data)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// EthTxManagerClientMock_EncodeBlobData_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'EncodeBlobData'
type EthTxManagerClientMock_EncodeBlobData_Call struct {
	*mock.Call
}

// EncodeBlobData is a helper method to define mock.On call
//   - data []byte
func (_e *EthTxManagerClientMock_Expecter) EncodeBlobData(data interface{}) *EthTxManagerClientMock_EncodeBlobData_Call {
	return &EthTxManagerClientMock_EncodeBlobData_Call{Call: _e.mock.On("EncodeBlobData", data)}
}

func (_c *EthTxManagerClientMock_EncodeBlobData_Call) Run(run func(data []byte)) *EthTxManagerClientMock_EncodeBlobData_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].([]byte))
	})
	return _c
}

func (_c *EthTxManagerClientMock_EncodeBlobData_Call) Return(_a0 kzg4844.Blob, _a1 error) *EthTxManagerClientMock_EncodeBlobData_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

func (_c *EthTxManagerClientMock_EncodeBlobData_Call) RunAndReturn(run func([]byte) (kzg4844.Blob, error)) *EthTxManagerClientMock_EncodeBlobData_Call {
	_c.Call.Return(run)
	return _c
}

// MakeBlobSidecar provides a mock function with given fields: blobs
func (_m *EthTxManagerClientMock) MakeBlobSidecar(blobs []kzg4844.Blob) *types.BlobTxSidecar {
	ret := _m.Called(blobs)

	if len(ret) == 0 {
		panic("no return value specified for MakeBlobSidecar")
	}

	var r0 *types.BlobTxSidecar
	if rf, ok := ret.Get(0).(func([]kzg4844.Blob) *types.BlobTxSidecar); ok {
		r0 = rf(blobs)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*types.BlobTxSidecar)
		}
	}

	return r0
}

// EthTxManagerClientMock_MakeBlobSidecar_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'MakeBlobSidecar'
type EthTxManagerClientMock_MakeBlobSidecar_Call struct {
	*mock.Call
}

// MakeBlobSidecar is a helper method to define mock.On call
//   - blobs []kzg4844.Blob
func (_e *EthTxManagerClientMock_Expecter) MakeBlobSidecar(blobs interface{}) *EthTxManagerClientMock_MakeBlobSidecar_Call {
	return &EthTxManagerClientMock_MakeBlobSidecar_Call{Call: _e.mock.On("MakeBlobSidecar", blobs)}
}

func (_c *EthTxManagerClientMock_MakeBlobSidecar_Call) Run(run func(blobs []kzg4844.Blob)) *EthTxManagerClientMock_MakeBlobSidecar_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].([]kzg4844.Blob))
	})
	return _c
}

func (_c *EthTxManagerClientMock_MakeBlobSidecar_Call) Return(_a0 *types.BlobTxSidecar) *EthTxManagerClientMock_MakeBlobSidecar_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *EthTxManagerClientMock_MakeBlobSidecar_Call) RunAndReturn(run func([]kzg4844.Blob) *types.BlobTxSidecar) *EthTxManagerClientMock_MakeBlobSidecar_Call {
	_c.Call.Return(run)
	return _c
}

// ProcessPendingMonitoredTxs provides a mock function with given fields: ctx, resultHandler
func (_m *EthTxManagerClientMock) ProcessPendingMonitoredTxs(ctx context.Context, resultHandler ethtxmanager.ResultHandler) {
	_m.Called(ctx, resultHandler)
}

// EthTxManagerClientMock_ProcessPendingMonitoredTxs_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'ProcessPendingMonitoredTxs'
type EthTxManagerClientMock_ProcessPendingMonitoredTxs_Call struct {
	*mock.Call
}

// ProcessPendingMonitoredTxs is a helper method to define mock.On call
//   - ctx context.Context
//   - resultHandler ethtxmanager.ResultHandler
func (_e *EthTxManagerClientMock_Expecter) ProcessPendingMonitoredTxs(ctx interface{}, resultHandler interface{}) *EthTxManagerClientMock_ProcessPendingMonitoredTxs_Call {
	return &EthTxManagerClientMock_ProcessPendingMonitoredTxs_Call{Call: _e.mock.On("ProcessPendingMonitoredTxs", ctx, resultHandler)}
}

func (_c *EthTxManagerClientMock_ProcessPendingMonitoredTxs_Call) Run(run func(ctx context.Context, resultHandler ethtxmanager.ResultHandler)) *EthTxManagerClientMock_ProcessPendingMonitoredTxs_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context), args[1].(ethtxmanager.ResultHandler))
	})
	return _c
}

func (_c *EthTxManagerClientMock_ProcessPendingMonitoredTxs_Call) Return() *EthTxManagerClientMock_ProcessPendingMonitoredTxs_Call {
	_c.Call.Return()
	return _c
}

func (_c *EthTxManagerClientMock_ProcessPendingMonitoredTxs_Call) RunAndReturn(run func(context.Context, ethtxmanager.ResultHandler)) *EthTxManagerClientMock_ProcessPendingMonitoredTxs_Call {
	_c.Run(run)
	return _c
}

// Remove provides a mock function with given fields: ctx, id
func (_m *EthTxManagerClientMock) Remove(ctx context.Context, id common.Hash) error {
	ret := _m.Called(ctx, id)

	if len(ret) == 0 {
		panic("no return value specified for Remove")
	}

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, common.Hash) error); ok {
		r0 = rf(ctx, id)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// EthTxManagerClientMock_Remove_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Remove'
type EthTxManagerClientMock_Remove_Call struct {
	*mock.Call
}

// Remove is a helper method to define mock.On call
//   - ctx context.Context
//   - id common.Hash
func (_e *EthTxManagerClientMock_Expecter) Remove(ctx interface{}, id interface{}) *EthTxManagerClientMock_Remove_Call {
	return &EthTxManagerClientMock_Remove_Call{Call: _e.mock.On("Remove", ctx, id)}
}

func (_c *EthTxManagerClientMock_Remove_Call) Run(run func(ctx context.Context, id common.Hash)) *EthTxManagerClientMock_Remove_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context), args[1].(common.Hash))
	})
	return _c
}

func (_c *EthTxManagerClientMock_Remove_Call) Return(_a0 error) *EthTxManagerClientMock_Remove_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *EthTxManagerClientMock_Remove_Call) RunAndReturn(run func(context.Context, common.Hash) error) *EthTxManagerClientMock_Remove_Call {
	_c.Call.Return(run)
	return _c
}

// RemoveAll provides a mock function with given fields: ctx
func (_m *EthTxManagerClientMock) RemoveAll(ctx context.Context) error {
	ret := _m.Called(ctx)

	if len(ret) == 0 {
		panic("no return value specified for RemoveAll")
	}

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context) error); ok {
		r0 = rf(ctx)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// EthTxManagerClientMock_RemoveAll_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'RemoveAll'
type EthTxManagerClientMock_RemoveAll_Call struct {
	*mock.Call
}

// RemoveAll is a helper method to define mock.On call
//   - ctx context.Context
func (_e *EthTxManagerClientMock_Expecter) RemoveAll(ctx interface{}) *EthTxManagerClientMock_RemoveAll_Call {
	return &EthTxManagerClientMock_RemoveAll_Call{Call: _e.mock.On("RemoveAll", ctx)}
}

func (_c *EthTxManagerClientMock_RemoveAll_Call) Run(run func(ctx context.Context)) *EthTxManagerClientMock_RemoveAll_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context))
	})
	return _c
}

func (_c *EthTxManagerClientMock_RemoveAll_Call) Return(_a0 error) *EthTxManagerClientMock_RemoveAll_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *EthTxManagerClientMock_RemoveAll_Call) RunAndReturn(run func(context.Context) error) *EthTxManagerClientMock_RemoveAll_Call {
	_c.Call.Return(run)
	return _c
}

// Result provides a mock function with given fields: ctx, id
func (_m *EthTxManagerClientMock) Result(ctx context.Context, id common.Hash) (zkevm_ethtx_managertypes.MonitoredTxResult, error) {
	ret := _m.Called(ctx, id)

	if len(ret) == 0 {
		panic("no return value specified for Result")
	}

	var r0 zkevm_ethtx_managertypes.MonitoredTxResult
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, common.Hash) (zkevm_ethtx_managertypes.MonitoredTxResult, error)); ok {
		return rf(ctx, id)
	}
	if rf, ok := ret.Get(0).(func(context.Context, common.Hash) zkevm_ethtx_managertypes.MonitoredTxResult); ok {
		r0 = rf(ctx, id)
	} else {
		r0 = ret.Get(0).(zkevm_ethtx_managertypes.MonitoredTxResult)
	}

	if rf, ok := ret.Get(1).(func(context.Context, common.Hash) error); ok {
		r1 = rf(ctx, id)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// EthTxManagerClientMock_Result_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Result'
type EthTxManagerClientMock_Result_Call struct {
	*mock.Call
}

// Result is a helper method to define mock.On call
//   - ctx context.Context
//   - id common.Hash
func (_e *EthTxManagerClientMock_Expecter) Result(ctx interface{}, id interface{}) *EthTxManagerClientMock_Result_Call {
	return &EthTxManagerClientMock_Result_Call{Call: _e.mock.On("Result", ctx, id)}
}

func (_c *EthTxManagerClientMock_Result_Call) Run(run func(ctx context.Context, id common.Hash)) *EthTxManagerClientMock_Result_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context), args[1].(common.Hash))
	})
	return _c
}

func (_c *EthTxManagerClientMock_Result_Call) Return(_a0 zkevm_ethtx_managertypes.MonitoredTxResult, _a1 error) *EthTxManagerClientMock_Result_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

func (_c *EthTxManagerClientMock_Result_Call) RunAndReturn(run func(context.Context, common.Hash) (zkevm_ethtx_managertypes.MonitoredTxResult, error)) *EthTxManagerClientMock_Result_Call {
	_c.Call.Return(run)
	return _c
}

// ResultsByStatus provides a mock function with given fields: ctx, statuses
func (_m *EthTxManagerClientMock) ResultsByStatus(ctx context.Context, statuses []zkevm_ethtx_managertypes.MonitoredTxStatus) ([]zkevm_ethtx_managertypes.MonitoredTxResult, error) {
	ret := _m.Called(ctx, statuses)

	if len(ret) == 0 {
		panic("no return value specified for ResultsByStatus")
	}

	var r0 []zkevm_ethtx_managertypes.MonitoredTxResult
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, []zkevm_ethtx_managertypes.MonitoredTxStatus) ([]zkevm_ethtx_managertypes.MonitoredTxResult, error)); ok {
		return rf(ctx, statuses)
	}
	if rf, ok := ret.Get(0).(func(context.Context, []zkevm_ethtx_managertypes.MonitoredTxStatus) []zkevm_ethtx_managertypes.MonitoredTxResult); ok {
		r0 = rf(ctx, statuses)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]zkevm_ethtx_managertypes.MonitoredTxResult)
		}
	}

	if rf, ok := ret.Get(1).(func(context.Context, []zkevm_ethtx_managertypes.MonitoredTxStatus) error); ok {
		r1 = rf(ctx, statuses)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// EthTxManagerClientMock_ResultsByStatus_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'ResultsByStatus'
type EthTxManagerClientMock_ResultsByStatus_Call struct {
	*mock.Call
}

// ResultsByStatus is a helper method to define mock.On call
//   - ctx context.Context
//   - statuses []zkevm_ethtx_managertypes.MonitoredTxStatus
func (_e *EthTxManagerClientMock_Expecter) ResultsByStatus(ctx interface{}, statuses interface{}) *EthTxManagerClientMock_ResultsByStatus_Call {
	return &EthTxManagerClientMock_ResultsByStatus_Call{Call: _e.mock.On("ResultsByStatus", ctx, statuses)}
}

func (_c *EthTxManagerClientMock_ResultsByStatus_Call) Run(run func(ctx context.Context, statuses []zkevm_ethtx_managertypes.MonitoredTxStatus)) *EthTxManagerClientMock_ResultsByStatus_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context), args[1].([]zkevm_ethtx_managertypes.MonitoredTxStatus))
	})
	return _c
}

func (_c *EthTxManagerClientMock_ResultsByStatus_Call) Return(_a0 []zkevm_ethtx_managertypes.MonitoredTxResult, _a1 error) *EthTxManagerClientMock_ResultsByStatus_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

func (_c *EthTxManagerClientMock_ResultsByStatus_Call) RunAndReturn(run func(context.Context, []zkevm_ethtx_managertypes.MonitoredTxStatus) ([]zkevm_ethtx_managertypes.MonitoredTxResult, error)) *EthTxManagerClientMock_ResultsByStatus_Call {
	_c.Call.Return(run)
	return _c
}

// Start provides a mock function with no fields
func (_m *EthTxManagerClientMock) Start() {
	_m.Called()
}

// EthTxManagerClientMock_Start_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Start'
type EthTxManagerClientMock_Start_Call struct {
	*mock.Call
}

// Start is a helper method to define mock.On call
func (_e *EthTxManagerClientMock_Expecter) Start() *EthTxManagerClientMock_Start_Call {
	return &EthTxManagerClientMock_Start_Call{Call: _e.mock.On("Start")}
}

func (_c *EthTxManagerClientMock_Start_Call) Run(run func()) *EthTxManagerClientMock_Start_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run()
	})
	return _c
}

func (_c *EthTxManagerClientMock_Start_Call) Return() *EthTxManagerClientMock_Start_Call {
	_c.Call.Return()
	return _c
}

func (_c *EthTxManagerClientMock_Start_Call) RunAndReturn(run func()) *EthTxManagerClientMock_Start_Call {
	_c.Run(run)
	return _c
}

// Stop provides a mock function with no fields
func (_m *EthTxManagerClientMock) Stop() {
	_m.Called()
}

// EthTxManagerClientMock_Stop_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Stop'
type EthTxManagerClientMock_Stop_Call struct {
	*mock.Call
}

// Stop is a helper method to define mock.On call
func (_e *EthTxManagerClientMock_Expecter) Stop() *EthTxManagerClientMock_Stop_Call {
	return &EthTxManagerClientMock_Stop_Call{Call: _e.mock.On("Stop")}
}

func (_c *EthTxManagerClientMock_Stop_Call) Run(run func()) *EthTxManagerClientMock_Stop_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run()
	})
	return _c
}

func (_c *EthTxManagerClientMock_Stop_Call) Return() *EthTxManagerClientMock_Stop_Call {
	_c.Call.Return()
	return _c
}

func (_c *EthTxManagerClientMock_Stop_Call) RunAndReturn(run func()) *EthTxManagerClientMock_Stop_Call {
	_c.Run(run)
	return _c
}

// NewEthTxManagerClientMock creates a new instance of EthTxManagerClientMock. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
// The first argument is typically a *testing.T value.
func NewEthTxManagerClientMock(t interface {
	mock.TestingT
	Cleanup(func())
}) *EthTxManagerClientMock {
	mock := &EthTxManagerClientMock{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
